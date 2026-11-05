package coverage

import (
	"strings"
	"testing"
)

func TestCheckUsesStatementWeights(t *testing.T) {
	profile := `mode: set
example.com/app/internal/a/a.go:1.1,2.2 2 1
example.com/app/internal/a/a.go:5.1,6.2 1 0
example.com/app/cmd/app/main.go:1.1,2.2 3 0
`

	report, err := Check(strings.NewReader(profile), 90)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if report.Covered != 2 || report.Statements != 6 {
		t.Errorf("covered/statements = %d/%d, want 2/6", report.Covered, report.Statements)
	}
	if report.Percent != 33.3 {
		t.Errorf("Percent = %.1f, want 33.3", report.Percent)
	}
	if report.Passed {
		t.Error("Passed = true, want false below minimum")
	}
	if len(report.Packages) != 2 {
		t.Fatalf("Packages count = %d, want 2", len(report.Packages))
	}
	if report.Packages[0].Package != "example.com/app/cmd/app" {
		t.Errorf("first package = %q, want example.com/app/cmd/app", report.Packages[0].Package)
	}
	if report.Packages[1].Package != "example.com/app/internal/a" {
		t.Errorf("second package = %q, want example.com/app/internal/a", report.Packages[1].Package)
	}
}

func TestCheckPassesAtExactMinimum(t *testing.T) {
	profile := `mode: set
example.com/app/internal/a/a.go:1.1,2.2 9 1
example.com/app/internal/a/a.go:3.1,4.2 1 0
`

	report, err := Check(strings.NewReader(profile), 90)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if report.Percent != 90 {
		t.Errorf("Percent = %.1f, want 90.0", report.Percent)
	}
	if !report.Passed {
		t.Error("Passed = false, want true at exact minimum")
	}
}

func TestCheckRejectsMalformedProfile(t *testing.T) {
	profile := `mode: set
not-a-coverage-line
`

	_, err := Check(strings.NewReader(profile), 90)
	if err == nil {
		t.Fatal("Check() error = nil, want malformed profile error")
	}
}

func TestCheckRejectsEmptyProfile(t *testing.T) {
	_, err := Check(strings.NewReader("mode: set\n"), 90)
	if err == nil {
		t.Fatal("Check() error = nil, want empty profile error")
	}
}

func TestCheckRejectsInvalidMinimum(t *testing.T) {
	_, err := Check(strings.NewReader("mode: set\nexample.com/a.go:1.1,2.1 1 1\n"), -1)
	if err == nil {
		t.Fatal("Check() error = nil, want minimum error")
	}
}

func TestCheckRejectsInvalidProfileFields(t *testing.T) {
	tests := []string{
		"example.com/a.go:1.1,2.1 x 1",
		"example.com/a.go:1.1,2.1 1 x",
		"example.com/a.go 1 1",
	}
	for _, line := range tests {
		_, err := Check(strings.NewReader("mode: set\n"+line+"\n"), 90)
		if err == nil {
			t.Errorf("Check(%q) error = nil, want malformed profile error", line)
		}
	}
}
