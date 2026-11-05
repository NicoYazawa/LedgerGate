package primpact

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Options struct {
	Repo             string
	Base             string
	Out              string
	Stage            string
	CodeGraphRoot    string
	CodeGraphScripts string
	BlastRadius      int
	MaxDepth         int
	AllowSignatures  bool
}

type Predicate struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail"`
	Passed bool   `json:"passed"`
}

type Report struct {
	Base             string          `json:"base"`
	BaseSHA          string          `json:"base_sha"`
	HeadSHA          string          `json:"head_sha"`
	ChangedFiles     []string        `json:"changed_files"`
	CodeFiles        []string        `json:"code_files"`
	CodeFileCount    int             `json:"code_file_count"`
	Stage            string          `json:"stage"`
	PackageImpact    PackageImpact   `json:"package_impact"`
	BreakingExports  []string        `json:"breaking_exports"`
	CodeGraph        CodeGraphResult `json:"codegraph"`
	Predicates       []Predicate     `json:"predicates"`
	Passed           bool            `json:"passed"`
	ImpactMermaid    string          `json:"impact_mermaid"`
	CodeGraphHTML    string          `json:"codegraph_html"`
	CodeGraphOutput  string          `json:"codegraph_output"`
	CodeGraphMessage string          `json:"codegraph_message"`
}

type checkOutput struct {
	Base            string          `json:"base"`
	BaseSHA         string          `json:"base_sha"`
	HeadSHA         string          `json:"head_sha"`
	CodeFileCount   int             `json:"code_file_count"`
	Stage           string          `json:"stage"`
	PackageImpact   PackageImpact   `json:"package_impact"`
	BreakingExports []string        `json:"breaking_exports"`
	CodeGraph       CodeGraphResult `json:"codegraph"`
	Predicates      []Predicate     `json:"predicates"`
	Passed          bool            `json:"passed"`
}

func Run(ctx context.Context, opts Options) (Report, error) {
	if opts.Repo == "" {
		opts.Repo = "."
	}
	repo, err := filepath.Abs(opts.Repo)
	if err != nil {
		return Report{}, fmt.Errorf("resolve repository: %w", err)
	}
	if opts.Base == "" {
		opts.Base = "origin/main"
	}
	if opts.Out == "" {
		opts.Out = "artifacts/pr-impact"
	}
	if opts.BlastRadius == 0 {
		opts.BlastRadius = 20
	}
	if opts.MaxDepth == 0 {
		opts.MaxDepth = 3
	}
	out := opts.Out
	if !filepath.IsAbs(out) {
		out = filepath.Join(repo, filepath.FromSlash(out))
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return Report{}, fmt.Errorf("create output directory: %w", err)
	}

	baseSHA, err := gitOutput(ctx, repo, "merge-base", opts.Base, "HEAD")
	if err != nil {
		return Report{}, err
	}
	headSHA, err := gitOutput(ctx, repo, "rev-parse", "HEAD")
	if err != nil {
		return Report{}, err
	}
	changedText, err := gitOutput(ctx, repo, "diff", "--name-only", baseSHA+"...HEAD")
	if err != nil {
		return Report{}, err
	}
	changed := nonEmptyLines(changedText)
	codeFiles := filterCodeFiles(changed)
	goFiles := filterFilesByExtension(codeFiles, ".go")
	stage := ResolveStageSlug(opts.Stage)
	if stage == "" {
		branch, branchErr := gitOutput(ctx, repo, "branch", "--show-current")
		if branchErr == nil {
			stage = ResolveStageSlug(branch)
		}
	}

	report := Report{
		Base:          opts.Base,
		BaseSHA:       baseSHA,
		HeadSHA:       headSHA,
		ChangedFiles:  changed,
		CodeFiles:     codeFiles,
		CodeFileCount: len(codeFiles),
		Stage:         stage,
		Predicates:    []Predicate{},
	}

	if len(codeFiles) == 0 {
		report.Passed = true
		if err := writeSummary(out, report, true); err != nil {
			return Report{}, err
		}
		if err := writeJSONFile(filepath.Join(out, "check.json"), newCheckOutput(report)); err != nil {
			return Report{}, err
		}
		return report, nil
	}

	packagePassed := true
	if len(goFiles) == 0 {
		report.PackageImpact = PackageImpact{Passed: true}
		report.Predicates = append(report.Predicates,
			Predicate{Name: "cycles", Status: "PASS", Detail: "no Go files changed", Passed: true},
			Predicate{Name: "blast-radius", Status: "PASS", Detail: "not applicable", Passed: true},
		)
	} else {
		impact, analyzeErr := AnalyzePackages(ctx, repo, goFiles, opts.MaxDepth, opts.BlastRadius)
		if analyzeErr != nil {
			packagePassed = false
			report.Predicates = append(report.Predicates,
				Predicate{Name: "cycles", Status: "FAIL", Detail: analyzeErr.Error(), Passed: false},
			)
			report.PackageImpact = PackageImpact{Limit: opts.BlastRadius}
		} else {
			report.PackageImpact = impact
			packagePassed = impact.Passed
			report.Predicates = append(report.Predicates,
				Predicate{Name: "cycles", Status: "PASS", Detail: "Go package graph is valid", Passed: true},
				Predicate{Name: "blast-radius", Status: passFail(impact.Passed), Detail: fmt.Sprintf("%d dependents, limit %d", impact.BlastRadius, impact.Limit), Passed: impact.Passed},
			)
			report.ImpactMermaid = RenderPackageImpact(impact)
		}
	}
	if report.ImpactMermaid == "" {
		report.ImpactMermaid = "flowchart LR\n  none[\"No Go package impact\"]"
	}

	if len(goFiles) > 0 {
		diff, diffErr := gitOutput(ctx, repo, "diff", "--unified=0", baseSHA+"...HEAD", "--", "*.go")
		if diffErr != nil {
			return Report{}, diffErr
		}
		report.BreakingExports = FindBreakingExports(diff)
	}
	signaturePassed := len(report.BreakingExports) == 0 || opts.AllowSignatures
	signatureDetail := "no breaking exported declarations"
	if len(report.BreakingExports) > 0 {
		signatureDetail = strings.Join(report.BreakingExports, "; ")
		if opts.AllowSignatures {
			signatureDetail += " (waived)"
		}
	}
	report.Predicates = append(report.Predicates,
		Predicate{Name: "signatures", Status: passFail(signaturePassed), Detail: signatureDetail, Passed: signaturePassed},
	)

	if stage == "" {
		report.CodeGraph = CodeGraphResult{Passed: false, Message: "CodeGraph stage is required"}
	} else {
		report.CodeGraph, err = ValidateCodeGraph(ctx, CodeGraphOptions{
			Repo:    repo,
			Stage:   stage,
			Root:    opts.CodeGraphRoot,
			Scripts: opts.CodeGraphScripts,
		})
		if err != nil {
			return Report{}, err
		}
	}
	report.CodeGraphHTML = filepath.ToSlash(filepath.Join(firstNonEmpty(opts.CodeGraphRoot, "docs/codegraph"), stage, "graph.html"))
	report.CodeGraphMessage = report.CodeGraph.Message
	report.CodeGraphOutput = report.CodeGraph.Output
	report.Predicates = append(report.Predicates,
		Predicate{Name: "codegraph-bundle", Status: passFail(report.CodeGraph.Passed), Detail: report.CodeGraph.Message, Passed: report.CodeGraph.Passed},
	)

	report.Passed = packagePassed && signaturePassed && report.CodeGraph.Passed
	if err := os.WriteFile(filepath.Join(out, "impact.mmd"), []byte(report.ImpactMermaid+"\n"), 0o644); err != nil {
		return Report{}, fmt.Errorf("write impact diagram: %w", err)
	}
	if err := writeSummary(out, report, false); err != nil {
		return Report{}, err
	}
	if err := writeJSONFile(filepath.Join(out, "check.json"), newCheckOutput(report)); err != nil {
		return Report{}, err
	}
	return report, nil
}

func gitOutput(ctx context.Context, repo string, args ...string) (string, error) {
	commandArgs := append([]string{"-C", repo}, args...)
	command := exec.CommandContext(ctx, "git", commandArgs...)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return strings.TrimSpace(string(output)), nil
}

func nonEmptyLines(text string) []string {
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, filepath.ToSlash(line))
		}
	}
	return lines
}

func filterCodeFiles(files []string) []string {
	extensions := map[string]struct{}{
		".go": {}, ".js": {}, ".jsx": {}, ".ts": {}, ".tsx": {}, ".py": {}, ".rb": {}, ".java": {},
		".kt": {}, ".rs": {}, ".c": {}, ".cc": {}, ".cpp": {}, ".h": {}, ".hpp": {}, ".cs": {},
	}
	var code []string
	for _, file := range files {
		if _, ok := extensions[strings.ToLower(filepath.Ext(file))]; ok {
			code = append(code, file)
		}
	}
	return code
}

func filterFilesByExtension(files []string, extension string) []string {
	var filtered []string
	for _, file := range files {
		if strings.EqualFold(filepath.Ext(file), extension) {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

func passFail(passed bool) string {
	if passed {
		return "PASS"
	}
	return "FAIL"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func newCheckOutput(report Report) checkOutput {
	return checkOutput{
		Base:            report.Base,
		BaseSHA:         report.BaseSHA,
		HeadSHA:         report.HeadSHA,
		CodeFileCount:   report.CodeFileCount,
		Stage:           report.Stage,
		PackageImpact:   report.PackageImpact,
		BreakingExports: report.BreakingExports,
		CodeGraph:       report.CodeGraph,
		Predicates:      report.Predicates,
		Passed:          report.Passed,
	}
}

func writeJSONFile(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode JSON artifact: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write JSON artifact: %w", err)
	}
	return nil
}

func writeSummary(out string, report Report, skipped bool) error {
	status := passFail(report.Passed)
	lines := []string{
		"<!-- codegraph-impact -->",
		"",
		"## Impact gate: " + status,
		"",
		"| Field | Value |",
		"|---|---|",
		fmt.Sprintf("| Base | `%s` (%s) |", report.Base, shortSHA(report.BaseSHA)),
		fmt.Sprintf("| Head | `%s` |", shortSHA(report.HeadSHA)),
		fmt.Sprintf("| Stage | `%s` |", report.Stage),
		fmt.Sprintf("| Changed files | %d |", len(report.ChangedFiles)),
		fmt.Sprintf("| Code files changed | %d |", report.CodeFileCount),
	}
	if skipped {
		lines = append(lines,
			"",
			"No code files changed. CodeGraph and Go package impact checks were skipped.",
		)
	} else {
		lines = append(lines,
			"",
			"### Package impact",
			"",
			"```mermaid",
			report.ImpactMermaid,
			"```",
			"",
			"### Predicates",
			"",
			"| Predicate | Result | Detail |",
			"|---|---|---|",
		)
		for _, predicate := range report.Predicates {
			lines = append(lines, fmt.Sprintf("| %s | %s | %s |", predicate.Name, predicate.Status, escapeTable(predicate.Detail)))
		}
		lines = append(lines,
			"",
			"### CodeGraph validation",
			"",
			fmt.Sprintf("- Bundle: `%s`", report.CodeGraphHTML),
			fmt.Sprintf("- Result: %s", report.CodeGraphMessage),
		)
		if report.CodeGraphOutput != "" {
			lines = append(lines, "", "```text", report.CodeGraphOutput, "```")
		}
		if len(report.BreakingExports) > 0 {
			lines = append(lines, "", "### Breaking exports", "")
			for _, declaration := range report.BreakingExports {
				lines = append(lines, "- `"+declaration+"`")
			}
		}
	}
	lines = append(lines,
		"",
		"### Notes",
		"",
		"- Go impact is computed from `go list` reverse dependencies.",
		"- CodeGraph validates Mermaid structure, source locations, and generated HTML; semantic labels still require review.",
		"- A green gate does not replace tests, risk tiers, or independent review.",
	)
	if err := os.WriteFile(filepath.Join(out, "summary.md"), []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		return fmt.Errorf("write summary: %w", err)
	}
	return nil
}

func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

func escapeTable(value string) string {
	value = strings.ReplaceAll(value, "|", "\\|")
	value = strings.ReplaceAll(value, "\n", " ")
	return value
}
