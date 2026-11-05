package primpact

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type goPackage struct {
	ImportPath string
	Dir        string
	Imports    []string
}

type Dependent struct {
	ImportPath string
	Depth      int
}

type DependencyEdge struct {
	From string
	To   string
}

type PackageImpact struct {
	ChangedPackages []string
	Dependents      []Dependent
	Edges           []DependencyEdge
	BlastRadius     int
	Limit           int
	Passed          bool
	Cycles          int
}

func AnalyzePackages(
	ctx context.Context,
	repo string,
	changed []string,
	maxDepth int,
	limit int,
) (PackageImpact, error) {
	packages, err := listPackages(ctx, repo)
	if err != nil {
		return PackageImpact{}, err
	}

	byImport := make(map[string]goPackage, len(packages))
	byDir := make(map[string]string, len(packages))
	for _, pkg := range packages {
		byImport[pkg.ImportPath] = pkg
		byDir[filepath.Clean(pkg.Dir)] = pkg.ImportPath
	}

	changedSet := make(map[string]struct{})
	for _, file := range changed {
		if filepath.Ext(file) != ".go" {
			continue
		}
		dir := filepath.Clean(filepath.Join(repo, filepath.Dir(filepath.FromSlash(file))))
		if importPath, ok := byDir[dir]; ok {
			changedSet[importPath] = struct{}{}
		}
	}
	changedPackages := sortedKeys(changedSet)

	reverse := make(map[string][]string)
	for _, pkg := range packages {
		for _, imported := range pkg.Imports {
			if _, ok := byImport[imported]; ok {
				reverse[imported] = append(reverse[imported], pkg.ImportPath)
			}
		}
	}

	if maxDepth < 1 {
		maxDepth = 3
	}
	visited := make(map[string]int, len(changedSet))
	for _, importPath := range changedPackages {
		visited[importPath] = 0
	}
	queue := append([]string(nil), changedPackages...)
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		depth := visited[current]
		if depth >= maxDepth {
			continue
		}
		for _, dependent := range reverse[current] {
			nextDepth := depth + 1
			if previous, ok := visited[dependent]; ok && previous <= nextDepth {
				continue
			}
			visited[dependent] = nextDepth
			queue = append(queue, dependent)
		}
	}

	dependents := make([]Dependent, 0, len(visited))
	for importPath, depth := range visited {
		if _, changed := changedSet[importPath]; changed {
			continue
		}
		dependents = append(dependents, Dependent{ImportPath: importPath, Depth: depth})
	}
	sort.Slice(dependents, func(i, j int) bool {
		if dependents[i].Depth != dependents[j].Depth {
			return dependents[i].Depth < dependents[j].Depth
		}
		return dependents[i].ImportPath < dependents[j].ImportPath
	})
	included := make(map[string]struct{}, len(visited))
	for importPath := range visited {
		included[importPath] = struct{}{}
	}
	var edges []DependencyEdge
	for current := range included {
		for _, dependent := range reverse[current] {
			if _, ok := included[dependent]; ok {
				edges = append(edges, DependencyEdge{From: dependent, To: current})
			}
		}
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}
		return edges[i].To < edges[j].To
	})

	return PackageImpact{
		ChangedPackages: changedPackages,
		Dependents:      dependents,
		Edges:           edges,
		BlastRadius:     len(dependents),
		Limit:           limit,
		Passed:          len(dependents) <= limit,
	}, nil
}

func RenderPackageImpact(impact PackageImpact) string {
	var lines []string
	lines = append(lines, "flowchart LR")
	ids := make(map[string]string, len(impact.ChangedPackages)+len(impact.Dependents))
	ordered := append([]string(nil), impact.ChangedPackages...)
	for _, dependent := range impact.Dependents {
		ordered = append(ordered, dependent.ImportPath)
	}
	for index, importPath := range ordered {
		id := fmt.Sprintf("p%d", index)
		ids[importPath] = id
		lines = append(lines, fmt.Sprintf(`  %s["%s"]`, id, importPath))
	}
	for _, edge := range impact.Edges {
		from, fromOK := ids[edge.From]
		to, toOK := ids[edge.To]
		if fromOK && toOK {
			lines = append(lines, fmt.Sprintf("  %s --> %s", from, to))
		}
	}
	if len(impact.ChangedPackages) > 0 {
		lines = append(lines, "  classDef changed fill:#e8f5e9,stroke:#2e7d32")
		var classIDs []string
		for _, importPath := range impact.ChangedPackages {
			classIDs = append(classIDs, ids[importPath])
		}
		lines = append(lines, "  class "+strings.Join(classIDs, ",")+" changed")
	}
	return strings.Join(lines, "\n")
}

func listPackages(ctx context.Context, repo string) ([]goPackage, error) {
	command := exec.CommandContext(ctx, "go", "list", "-json", "./...")
	command.Dir = repo
	output, err := command.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("go list packages: %w\n%s", err, bytes.TrimSpace(output))
	}

	var packages []goPackage
	decoder := json.NewDecoder(bytes.NewReader(output))
	for {
		var pkg goPackage
		if err := decoder.Decode(&pkg); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("decode go list package: %w", err)
		}
		packages = append(packages, pkg)
	}
	return packages, nil
}

func sortedKeys(values map[string]struct{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
