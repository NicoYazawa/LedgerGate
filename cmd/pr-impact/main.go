package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/NicoYazawa/LedgerGate/internal/tooling/primpact"
)

func main() {
	var options primpact.Options
	var enforce bool
	flag.StringVar(&options.Repo, "repo", ".", "repository root")
	flag.StringVar(&options.Base, "base", "origin/main", "base git ref")
	flag.StringVar(&options.Out, "out", "artifacts/pr-impact", "artifact output directory")
	flag.StringVar(&options.Stage, "stage", "", "stage or branch name")
	flag.StringVar(&options.CodeGraphRoot, "codegraph-root", "docs/codegraph", "committed CodeGraph artifact root")
	flag.StringVar(&options.CodeGraphScripts, "codegraph-scripts", "", "CodeGraph validator scripts directory")
	flag.IntVar(&options.BlastRadius, "blast-radius", 20, "maximum reverse-dependent package count")
	flag.IntVar(&options.MaxDepth, "max-depth", 3, "maximum reverse-dependency traversal depth")
	flag.BoolVar(&options.AllowSignatures, "allow-signatures", false, "waive exported declaration changes")
	flag.BoolVar(&enforce, "enforce", false, "exit non-zero when the gate fails")
	flag.Parse()

	report, err := primpact.Run(context.Background(), options)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	status := "PASS"
	if !report.Passed {
		status = "FAIL"
	}
	fmt.Printf("Impact gate: %s\n", status)
	for _, predicate := range report.Predicates {
		fmt.Printf("- %s: %s (%s)\n", predicate.Name, predicate.Status, predicate.Detail)
	}
	fmt.Printf("Summary: %s\n", options.Out+"/summary.md")
	if enforce && !report.Passed {
		os.Exit(1)
	}
}
