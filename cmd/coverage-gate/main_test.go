package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunPassesAtMinimum(t *testing.T) {
	profile := writeProfile(t, `mode: set
example.com/app/internal/a/a.go:1.1,2.2 9 1
example.com/app/internal/a/a.go:3.1,4.2 1 0
`)
	var stdout, stderr bytes.Buffer

	code := run([]string{"-profile", profile, "-minimum", "90"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("run() = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "90.0%") || !strings.Contains(stdout.String(), "PASS") {
		t.Errorf("stdout = %q, want coverage and pass status", stdout.String())
	}
}

func TestRunFailsBelowMinimum(t *testing.T) {
	profile := writeProfile(t, `mode: set
example.com/app/internal/a/a.go:1.1,2.2 8 1
example.com/app/internal/a/a.go:3.1,4.2 2 0
`)
	var stdout, stderr bytes.Buffer

	code := run([]string{"-profile", profile, "-minimum", "90"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("run() = %d, want 1", code)
	}
	if !strings.Contains(stdout.String(), "80.0%") || !strings.Contains(stdout.String(), "FAIL") {
		t.Errorf("stdout = %q, want coverage and fail status", stdout.String())
	}
}

func TestRunWritesJSONReport(t *testing.T) {
	profile := writeProfile(t, `mode: set
example.com/app/internal/a/a.go:1.1,2.2 10 1
`)
	jsonPath := filepath.Join(t.TempDir(), "coverage.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"-profile", profile, "-minimum", "90", "-json", jsonPath}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("run() = %d, stderr = %s", code, stderr.String())
	}
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read JSON report: %v", err)
	}
	if !strings.Contains(string(data), `"passed": true`) {
		t.Errorf("JSON report = %s, want passed true", data)
	}
}

func TestRunRejectsMissingProfile(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run([]string{"-profile", "missing.out"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("run() = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "missing.out") {
		t.Errorf("stderr = %q, want profile path", stderr.String())
	}
}

func TestRunRejectsInvalidMinimum(t *testing.T) {
	profile := writeProfile(t, `mode: set
example.com/app/internal/a/a.go:1.1,2.2 1 1
`)
	var stdout, stderr bytes.Buffer

	code := run([]string{"-profile", profile, "-minimum", "101"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("run() = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "between 0 and 100") {
		t.Errorf("stderr = %q, want minimum error", stderr.String())
	}
}

func TestRunReportsJSONWriteFailure(t *testing.T) {
	profile := writeProfile(t, `mode: set
example.com/app/internal/a/a.go:1.1,2.2 1 1
`)
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer

	code := run(
		[]string{"-profile", profile, "-minimum", "90", "-json", filepath.Join(blocker, "coverage.json")},
		&stdout,
		&stderr,
	)

	if code != 2 {
		t.Fatalf("run() = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "report directory") {
		t.Errorf("stderr = %q, want report directory error", stderr.String())
	}
}

func writeProfile(t *testing.T, profile string) string {
	t.Helper()
	filePath := filepath.Join(t.TempDir(), "coverage.out")
	if err := os.WriteFile(filePath, []byte(profile), 0o644); err != nil {
		t.Fatal(err)
	}
	return filePath
}
