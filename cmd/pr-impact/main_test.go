package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/NicoYazawa/LedgerGate/internal/tooling/primpact"
)

func TestRunPrintsPassingReport(t *testing.T) {
	var stdout, stderr bytes.Buffer
	var captured primpact.Options
	execute := func(_ context.Context, options primpact.Options) (primpact.Report, error) {
		captured = options
		return primpact.Report{
			Passed: true,
			Predicates: []primpact.Predicate{
				{Name: "coverage", Status: "PASS", Detail: "95%", Passed: true},
			},
		}, nil
	}

	code := run(
		[]string{"-base", "origin/test", "-stage", "feat/test", "-blast-radius", "12"},
		&stdout,
		&stderr,
		execute,
	)

	if code != 0 {
		t.Fatalf("run() = %d, stderr = %s", code, stderr.String())
	}
	if captured.Base != "origin/test" || captured.Stage != "feat/test" || captured.BlastRadius != 12 {
		t.Errorf("captured options = %+v", captured)
	}
	if !strings.Contains(stdout.String(), "Impact gate: PASS") ||
		!strings.Contains(stdout.String(), "coverage: PASS") {
		t.Errorf("stdout = %q", stdout.String())
	}
}

func TestRunAllowsAdvisoryFailure(t *testing.T) {
	var stdout, stderr bytes.Buffer
	execute := func(context.Context, primpact.Options) (primpact.Report, error) {
		return primpact.Report{Passed: false}, nil
	}

	code := run(nil, &stdout, &stderr, execute)

	if code != 0 {
		t.Fatalf("run() = %d, want 0 without enforcement", code)
	}
	if !strings.Contains(stdout.String(), "Impact gate: FAIL") {
		t.Errorf("stdout = %q, want FAIL", stdout.String())
	}
}

func TestRunEnforcesFailingReport(t *testing.T) {
	var stdout, stderr bytes.Buffer
	execute := func(context.Context, primpact.Options) (primpact.Report, error) {
		return primpact.Report{Passed: false}, nil
	}

	code := run([]string{"-enforce"}, &stdout, &stderr, execute)

	if code != 1 {
		t.Fatalf("run() = %d, want 1", code)
	}
}

func TestRunReportsExecutionError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	execute := func(context.Context, primpact.Options) (primpact.Report, error) {
		return primpact.Report{}, errors.New("git failed")
	}

	code := run(nil, &stdout, &stderr, execute)

	if code != 2 {
		t.Fatalf("run() = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "git failed") {
		t.Errorf("stderr = %q, want execution error", stderr.String())
	}
}

func TestRunRejectsInvalidFlags(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run([]string{"-unknown"}, &stdout, &stderr, func(context.Context, primpact.Options) (primpact.Report, error) {
		t.Fatal("execute called for invalid flags")
		return primpact.Report{}, nil
	})

	if code != 2 {
		t.Fatalf("run() = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "flag provided but not defined") {
		t.Errorf("stderr = %q, want flag error", stderr.String())
	}
}
