package primpact

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSkipsCodeGraphForDocsOnlyChange(t *testing.T) {
	repo := initGitRepo(t, map[string]string{"README.md": "first\n"})
	commitFiles(t, repo, map[string]string{"README.md": "second\n"}, "docs: update readme")
	out := t.TempDir()

	report, err := Run(context.Background(), Options{
		Repo: repo,
		Base: "HEAD~1",
		Out:  out,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !report.Passed {
		t.Fatalf("Passed = false: %+v", report.Predicates)
	}
	if report.CodeFileCount != 0 {
		t.Errorf("CodeFileCount = %d, want 0", report.CodeFileCount)
	}
	assertFileExists(t, filepath.Join(out, "summary.md"))
	if _, err := os.Stat(filepath.Join(out, "impact.mmd")); !os.IsNotExist(err) {
		t.Error("impact.mmd exists for a docs-only change")
	}
}

func TestRunWritesCodeImpactArtifacts(t *testing.T) {
	scripts := installedCodeGraphScripts(t)
	repo := initCodeGraphRepo(t, scripts)
	commitFiles(t, repo, map[string]string{
		"internal/base/base.go": "package base\n\nfunc Value() int { return 2 }\n",
	}, "feat: change base value")
	out := t.TempDir()

	report, err := Run(context.Background(), Options{
		Repo:             repo,
		Base:             "HEAD~1",
		Out:              out,
		Stage:            "refactor/02-codegraph-gate",
		CodeGraphScripts: scripts,
		BlastRadius:      20,
		MaxDepth:         3,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !report.Passed {
		t.Fatalf("Passed = false: %+v", report.Predicates)
	}
	if report.PackageImpact.BlastRadius != 2 {
		t.Errorf("BlastRadius = %d, want 2", report.PackageImpact.BlastRadius)
	}
	for _, name := range []string{"summary.md", "impact.mmd", "check.json"} {
		assertFileExists(t, filepath.Join(out, name))
	}
	summary, err := os.ReadFile(filepath.Join(out, "summary.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"CodeGraph validation", "blast-radius", "docs/codegraph/02-codegraph-gate/graph.html"} {
		if !strings.Contains(string(summary), text) {
			t.Errorf("summary missing %q:\n%s", text, summary)
		}
	}
}

func TestRunFailsBreakingExportUnlessAllowed(t *testing.T) {
	scripts := installedCodeGraphScripts(t)
	repo := initCodeGraphRepo(t, scripts)
	commitFiles(t, repo, map[string]string{
		"main.go": "package main\n\nfunc Exported(value int) int { return value }\n\nfunc main() {}\n",
	}, "feat: change exported signature")
	out := t.TempDir()
	options := Options{
		Repo:             repo,
		Base:             "HEAD~1",
		Out:              out,
		Stage:            "02-codegraph-gate",
		CodeGraphScripts: scripts,
		BlastRadius:      20,
		MaxDepth:         3,
	}

	report, err := Run(context.Background(), options)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if report.Passed {
		t.Error("Passed = true, want false for a breaking exported declaration")
	}
	if len(report.BreakingExports) == 0 {
		t.Error("BreakingExports is empty")
	}

	options.AllowSignatures = true
	report, err = Run(context.Background(), options)
	if err != nil {
		t.Fatalf("Run() with signature waiver error = %v", err)
	}
	if !report.Passed {
		t.Errorf("Passed = false with signature waiver: %+v", report.Predicates)
	}
}

func initCodeGraphRepo(t *testing.T, scripts string) string {
	t.Helper()
	repo := initGitRepo(t, map[string]string{
		"go.mod": "module example.com/impact\n\ngo 1.27.0\n",
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
		"main.go": `package main

func Exported() int { return 1 }

func main() {}
`,
		".codegraph/config.json": `{"editor":{"id":"web","label":"Web"},"outputDir":".codegraph/output"}`,
		"docs/codegraph/02-codegraph-gate/graph.mmd": `flowchart LR
  A["进入入口"] --> B["完成启动"]
`,
		"docs/codegraph/02-codegraph-gate/graph.locs.json": `{
  "flowchart": {
    "nodes": {
      "A": ["main.go", 3],
      "B": ["main.go", 5]
    },
    "edges": [
      {"from":"A","to":"B","occurrence":1,"loc":["main.go",5]}
    ]
  }
}`,
	})
	buildCodeGraphHTML(t, scripts, repo)
	runGit(t, repo, "add", ".")
	runGit(t, repo, "commit", "-m", "test: add CodeGraph fixture")
	return repo
}

func initGitRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	repo := t.TempDir()
	runGit(t, repo, "init")
	runGit(t, repo, "config", "user.name", "LedgerGate Test")
	runGit(t, repo, "config", "user.email", "test@example.com")
	writeRepoFiles(t, repo, files)
	runGit(t, repo, "add", ".")
	runGit(t, repo, "commit", "-m", "test: initial")
	return repo
}

func commitFiles(t *testing.T, repo string, files map[string]string, message string) {
	t.Helper()
	writeRepoFiles(t, repo, files)
	runGit(t, repo, "add", ".")
	runGit(t, repo, "commit", "-m", message)
}

func writeRepoFiles(t *testing.T, repo string, files map[string]string) {
	t.Helper()
	for path, contents := range files {
		fullPath := filepath.Join(repo, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func runGit(t *testing.T, repo string, args ...string) string {
	t.Helper()
	commandArgs := append([]string{"-C", repo}, args...)
	output, err := exec.Command("git", commandArgs...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, output)
	}
	return strings.TrimSpace(string(output))
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("%s is missing: %v", path, err)
	}
}
