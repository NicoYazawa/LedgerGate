package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/NicoYazawa/LedgerGate/internal/tooling/coverage"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("coverage-gate", flag.ContinueOnError)
	flags.SetOutput(stderr)
	profile := flags.String("profile", "coverage.out", "Go coverage profile")
	minimum := flags.Float64("minimum", 90, "minimum total statement coverage")
	jsonPath := flags.String("json", "", "optional JSON report path")
	if err := flags.Parse(args); err != nil {
		return 2
	}

	file, err := os.Open(*profile)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "coverage-gate: %v\n", err)
		return 2
	}
	defer func() {
		_ = file.Close()
	}()

	report, err := coverage.Check(file, *minimum)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "coverage-gate: %v\n", err)
		return 2
	}
	status := "PASS"
	if !report.Passed {
		status = "FAIL"
	}
	_, _ = fmt.Fprintf(
		stdout,
		"Coverage: %.1f%% (%d/%d statements) %s minimum %.1f%%\n",
		report.Percent,
		report.Covered,
		report.Statements,
		status,
		report.Minimum,
	)
	for _, pkg := range report.Packages {
		_, _ = fmt.Fprintf(stdout, "  %-60s %6.1f%% (%d/%d)\n", pkg.Package, pkg.Percent, pkg.Covered, pkg.Statements)
	}
	if *jsonPath != "" {
		if err := writeJSON(*jsonPath, report); err != nil {
			_, _ = fmt.Fprintf(stderr, "coverage-gate: %v\n", err)
			return 2
		}
	}
	if !report.Passed {
		return 1
	}
	return 0
}

func writeJSON(path string, report coverage.Report) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create report directory: %w", err)
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("encode report: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	return nil
}
