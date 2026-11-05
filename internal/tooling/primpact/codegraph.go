package primpact

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
)

type CodeGraphOptions struct {
	Repo    string
	Stage   string
	Root    string
	Scripts string
}

type CodeGraphResult struct {
	Passed  bool
	Stage   string
	Bundle  string
	HTML    string
	Message string
	Output  string
}

func ResolveStageSlug(branch string) string {
	branch = strings.TrimSpace(strings.TrimPrefix(branch, "refs/heads/"))
	for _, prefix := range []string{"feat/", "fix/", "refactor/", "docs/"} {
		branch = strings.TrimPrefix(branch, prefix)
	}
	var builder strings.Builder
	hyphen := false
	for _, character := range strings.ToLower(branch) {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			builder.WriteRune(character)
			hyphen = false
			continue
		}
		if builder.Len() > 0 && !hyphen {
			builder.WriteByte('-')
			hyphen = true
		}
	}
	return strings.Trim(builder.String(), "-")
}

func ValidateCodeGraph(ctx context.Context, opts CodeGraphOptions) (CodeGraphResult, error) {
	stage := strings.TrimSpace(opts.Stage)
	if stage == "" {
		return CodeGraphResult{}, fmt.Errorf("CodeGraph stage is required")
	}
	root := opts.Root
	if root == "" {
		root = "docs/codegraph"
	}
	bundle := filepath.Join(opts.Repo, filepath.FromSlash(root), stage)
	result := CodeGraphResult{Stage: stage, Bundle: bundle, HTML: filepath.Join(bundle, "graph.html")}

	required := []string{"graph.mmd", "graph.locs.json", "graph.html"}
	for _, name := range required {
		path := filepath.Join(bundle, name)
		if info, err := os.Stat(path); err != nil || info.IsDir() {
			result.Message = fmt.Sprintf("missing CodeGraph artifact %s", filepath.ToSlash(filepath.Join(root, stage, name)))
			return result, nil
		}
	}

	scripts := opts.Scripts
	if scripts == "" {
		scripts = findCodeGraphScripts(opts.Repo)
	}
	if scripts == "" {
		result.Message = "CodeGraph validator scripts not found"
		return result, nil
	}

	var output strings.Builder
	validate := exec.CommandContext(
		ctx,
		"node",
		filepath.Join(scripts, "validate.mjs"),
		"--mmd", filepath.Join(bundle, "graph.mmd"),
		"--locs", filepath.Join(bundle, "graph.locs.json"),
		"--repo", opts.Repo,
	)
	validateOutput, err := validate.CombinedOutput()
	output.WriteString("validate.mjs:\n")
	output.Write(validateOutput)
	if err != nil {
		result.Output = output.String()
		result.Message = "CodeGraph location validation failed"
		return result, nil
	}

	verify := exec.CommandContext(
		ctx,
		"node",
		filepath.Join(scripts, "verify-html.mjs"),
		"--html", filepath.Join(bundle, "graph.html"),
	)
	verifyOutput, err := verify.CombinedOutput()
	output.WriteString("\nverify-html.mjs:\n")
	output.Write(verifyOutput)
	result.Output = strings.TrimSpace(output.String())
	if err != nil {
		result.Message = "CodeGraph HTML verification failed"
		return result, nil
	}

	result.Passed = true
	result.Message = "CodeGraph bundle is valid"
	return result, nil
}

func findCodeGraphScripts(repo string) string {
	candidates := []string{
		os.Getenv("CODEGRAPH_SCRIPTS"),
		filepath.Join(repo, "node_modules", "@longbowzz", "dsh-codegraph", "skills", "codegraph", "scripts"),
	}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, ".agents", "skills", "codegraph", "scripts"))
	}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if _, err := os.Stat(filepath.Join(candidate, "validate.mjs")); err == nil {
			return candidate
		}
	}
	return ""
}
