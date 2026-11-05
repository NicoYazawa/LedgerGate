package primpact

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestResolveStageSlug(t *testing.T) {
	tests := map[string]string{
		"feat/02-shared-value-primitives": "02-shared-value-primitives",
		"refactor/codegraph_gate":         "codegraph-gate",
		"fix/01-buildable-gateway-health": "01-buildable-gateway-health",
	}
	for branch, want := range tests {
		if got := ResolveStageSlug(branch); got != want {
			t.Errorf("ResolveStageSlug(%q) = %q, want %q", branch, got, want)
		}
	}
}

func TestValidateCodeGraphRequiresBundle(t *testing.T) {
	repo := t.TempDir()

	result, err := ValidateCodeGraph(context.Background(), CodeGraphOptions{
		Repo:  repo,
		Stage: "02-codegraph-gate",
	})
	if err != nil {
		t.Fatalf("ValidateCodeGraph() error = %v", err)
	}
	if result.Passed {
		t.Error("Passed = true, want false for a missing bundle")
	}
	if !strings.Contains(result.Message, "graph.mmd") {
		t.Errorf("Message = %q, want missing graph.mmd", result.Message)
	}
}

func TestValidateCodeGraphAcceptsValidBundle(t *testing.T) {
	scripts := installedCodeGraphScripts(t)
	repo := writeCodeGraphFixture(t)
	buildCodeGraphHTML(t, scripts, repo)

	result, err := ValidateCodeGraph(context.Background(), CodeGraphOptions{
		Repo:    repo,
		Stage:   "02-codegraph-gate",
		Scripts: scripts,
	})
	if err != nil {
		t.Fatalf("ValidateCodeGraph() error = %v", err)
	}
	if !result.Passed {
		t.Fatalf("Passed = false: %s\n%s", result.Message, result.Output)
	}
}

func TestValidateCodeGraphRejectsLocationOutsideRepo(t *testing.T) {
	scripts := installedCodeGraphScripts(t)
	repo := writeCodeGraphFixture(t)
	locs := filepath.Join(repo, "docs", "codegraph", "02-codegraph-gate", "graph.locs.json")
	contents, err := os.ReadFile(locs)
	if err != nil {
		t.Fatal(err)
	}
	invalid := strings.Replace(string(contents), `"main.go"`, `"../outside.go"`, 1)
	if err := os.WriteFile(locs, []byte(invalid), 0o644); err != nil {
		t.Fatal(err)
	}
	buildCodeGraphHTML(t, scripts, repo)

	result, err := ValidateCodeGraph(context.Background(), CodeGraphOptions{
		Repo:    repo,
		Stage:   "02-codegraph-gate",
		Scripts: scripts,
	})
	if err != nil {
		t.Fatalf("ValidateCodeGraph() error = %v", err)
	}
	if result.Passed {
		t.Error("Passed = true, want false for a source path outside the repository")
	}
}

func writeCodeGraphFixture(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	bundle := filepath.Join(repo, "docs", "codegraph", "02-codegraph-gate")
	if err := os.MkdirAll(filepath.Join(repo, ".codegraph"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(bundle, 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"main.go": `package main

func main() {}
`,
		".codegraph/config.json": `{"editor":{"id":"web","label":"Web"},"outputDir":".codegraph/output"}`,
		"docs/codegraph/02-codegraph-gate/graph.mmd": `flowchart LR
  A["进入入口"] --> B["完成启动"]
`,
		"docs/codegraph/02-codegraph-gate/graph.locs.json": `{
  "flowchart": {
    "nodes": {
      "A": ["main.go", 1],
      "B": ["main.go", 3]
    },
    "edges": [
      {"from":"A","to":"B","occurrence":1,"loc":["main.go",3]}
    ]
  }
}`,
	}
	for path, contents := range files {
		fullPath := filepath.Join(repo, filepath.FromSlash(path))
		if err := os.WriteFile(fullPath, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return repo
}

func buildCodeGraphHTML(t *testing.T, scripts, repo string) {
	t.Helper()
	bundle := filepath.Join(repo, "docs", "codegraph", "02-codegraph-gate")
	command := exec.Command(
		"node",
		filepath.Join(scripts, "build-html.mjs"),
		"--mmd", filepath.Join(bundle, "graph.mmd"),
		"--locs", filepath.Join(bundle, "graph.locs.json"),
		"--repo", repo,
		"--out", filepath.Join(bundle, "graph.html"),
		"--editor-config", filepath.Join(repo, ".codegraph", "config.json"),
	)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build CodeGraph fixture: %v\n%s", err, output)
	}
}

func installedCodeGraphScripts(t *testing.T) string {
	t.Helper()
	if scripts := os.Getenv("CODEGRAPH_SCRIPTS"); scripts != "" {
		return scripts
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	scripts := filepath.Join(home, ".agents", "skills", "codegraph", "scripts")
	if runtime.GOOS == "windows" {
		scripts = filepath.FromSlash(strings.ReplaceAll(scripts, `\`, "/"))
	}
	if _, err := os.Stat(filepath.Join(scripts, "validate.mjs")); err != nil {
		t.Skip("CodeGraph skill scripts not installed; set CODEGRAPH_SCRIPTS")
	}
	return scripts
}
