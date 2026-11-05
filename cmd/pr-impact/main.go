package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/NicoYazawa/LedgerGate/internal/tooling/primpact"
)

type executeFunc func(context.Context, primpact.Options) (primpact.Report, error)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, primpact.Run))
}

func run(args []string, stdout, stderr io.Writer, execute executeFunc) int {
	flags := flag.NewFlagSet("pr-impact", flag.ContinueOnError)
	flags.SetOutput(stderr)
	var options primpact.Options
	var enforce bool
	flags.StringVar(&options.Repo, "repo", ".", "repository root")
	flags.StringVar(&options.Base, "base", "origin/main", "base git ref")
	flags.StringVar(&options.Out, "out", "artifacts/pr-impact", "artifact output directory")
	flags.StringVar(&options.Stage, "stage", "", "stage or branch name")
	flags.StringVar(&options.CodeGraphRoot, "codegraph-root", "docs/codegraph", "committed CodeGraph artifact root")
	flags.StringVar(&options.CodeGraphScripts, "codegraph-scripts", "", "CodeGraph validator scripts directory")
	flags.IntVar(&options.BlastRadius, "blast-radius", 20, "maximum reverse-dependent package count")
	flags.IntVar(&options.MaxDepth, "max-depth", 3, "maximum reverse-dependency traversal depth")
	flags.BoolVar(&options.AllowSignatures, "allow-signatures", false, "waive exported declaration changes")
	flags.BoolVar(&enforce, "enforce", false, "exit non-zero when the gate fails")
	if err := flags.Parse(args); err != nil {
		return 2
	}

	report, err := execute(context.Background(), options)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 2
	}
	status := "PASS"
	if !report.Passed {
		status = "FAIL"
	}
	_, _ = fmt.Fprintf(stdout, "Impact gate: %s\n", status)
	for _, predicate := range report.Predicates {
		_, _ = fmt.Fprintf(stdout, "- %s: %s (%s)\n", predicate.Name, predicate.Status, predicate.Detail)
	}
	_, _ = fmt.Fprintf(stdout, "Summary: %s/summary.md\n", options.Out)
	if enforce && !report.Passed {
		return 1
	}
	return 0
}
