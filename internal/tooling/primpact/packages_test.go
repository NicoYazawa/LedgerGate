package primpact

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzePackagesFindsTransitiveDependents(t *testing.T) {
	repo := newPackageFixture(t)

	impact, err := AnalyzePackages(
		context.Background(),
		repo,
		[]string{"internal/base/base.go"},
		3,
		2,
	)
	if err != nil {
		t.Fatalf("AnalyzePackages() error = %v", err)
	}

	assertStrings(t, impact.ChangedPackages, []string{"example.com/impact/internal/base"})
	if impact.BlastRadius != 2 {
		t.Errorf("BlastRadius = %d, want 2", impact.BlastRadius)
	}
	if !impact.Passed {
		t.Error("Passed = false, want true at the configured limit")
	}
	assertDependent(t, impact.Dependents, "example.com/impact/internal/mid", 1)
	assertDependent(t, impact.Dependents, "example.com/impact/cmd/app", 2)
}

func TestAnalyzePackagesFailsAboveBlastRadius(t *testing.T) {
	repo := newPackageFixture(t)

	impact, err := AnalyzePackages(
		context.Background(),
		repo,
		[]string{"internal/base/base.go"},
		3,
		1,
	)
	if err != nil {
		t.Fatalf("AnalyzePackages() error = %v", err)
	}

	if impact.BlastRadius != 2 {
		t.Errorf("BlastRadius = %d, want 2", impact.BlastRadius)
	}
	if impact.Passed {
		t.Error("Passed = true, want false above the configured limit")
	}
}

func TestRenderPackageImpactShowsDependencyDirection(t *testing.T) {
	repo := newPackageFixture(t)
	impact, err := AnalyzePackages(
		context.Background(),
		repo,
		[]string{"internal/base/base.go"},
		3,
		20,
	)
	if err != nil {
		t.Fatalf("AnalyzePackages() error = %v", err)
	}

	diagram := RenderPackageImpact(impact)

	want := `flowchart LR
  p0["example.com/impact/internal/base"]
  p1["example.com/impact/internal/mid"]
  p2["example.com/impact/cmd/app"]
  p2 --> p1
  p1 --> p0
  classDef changed fill:#e8f5e9,stroke:#2e7d32
  class p0 changed`
	if diagram != want {
		t.Errorf("RenderPackageImpact() =\n%s\nwant:\n%s", diagram, want)
	}
}

func newPackageFixture(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	files := map[string]string{
		"go.mod": `module example.com/impact

go 1.27.0
`,
		"internal/base/base.go": `package base

func Value() int { return 1 }
`,
		"internal/mid/mid.go": `package mid

import "example.com/impact/internal/base"

func Value() int { return base.Value() }
`,
		"cmd/app/main.go": `package main

import (
	"fmt"

	"example.com/impact/internal/mid"
)

func main() { fmt.Println(mid.Value()) }
`,
	}
	for path, contents := range files {
		fullPath := filepath.Join(repo, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("MkdirAll(%s): %v", path, err)
		}
		if err := os.WriteFile(fullPath, []byte(contents), 0o644); err != nil {
			t.Fatalf("WriteFile(%s): %v", path, err)
		}
	}
	return repo
}

func assertDependent(t *testing.T, dependents []Dependent, importPath string, depth int) {
	t.Helper()
	for _, dependent := range dependents {
		if dependent.ImportPath == importPath {
			if dependent.Depth != depth {
				t.Errorf("%s depth = %d, want %d", importPath, dependent.Depth, depth)
			}
			return
		}
	}
	t.Errorf("dependent %s not found in %+v", importPath, dependents)
}

func assertStrings(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("item %d = %q, want %q", i, got[i], want[i])
		}
	}
}
