package coverage

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"path"
	"sort"
	"strconv"
	"strings"
)

type PackageCoverage struct {
	Package    string  `json:"package"`
	Covered    int     `json:"covered"`
	Statements int     `json:"statements"`
	Percent    float64 `json:"percent"`
}

type Report struct {
	Covered    int               `json:"covered"`
	Statements int               `json:"statements"`
	Percent    float64           `json:"percent"`
	Minimum    float64           `json:"minimum"`
	Passed     bool              `json:"passed"`
	Packages   []PackageCoverage `json:"packages"`
}

func Check(reader io.Reader, minimum float64) (Report, error) {
	if minimum < 0 || minimum > 100 {
		return Report{}, fmt.Errorf("minimum must be between 0 and 100")
	}
	type totals struct {
		covered    int
		statements int
	}
	byPackage := make(map[string]*totals)
	scanner := bufio.NewScanner(reader)
	lines := 0
	for scanner.Scan() {
		lines++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "mode:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			return Report{}, fmt.Errorf("invalid coverage line %d: %q", lines, line)
		}
		fileName := fields[0]
		separator := strings.LastIndex(fileName, ":")
		if separator <= 0 {
			return Report{}, fmt.Errorf("invalid coverage file on line %d: %q", lines, fileName)
		}
		statements, err := strconv.Atoi(fields[1])
		if err != nil || statements < 0 {
			return Report{}, fmt.Errorf("invalid statement count on line %d: %q", lines, fields[1])
		}
		count, err := strconv.Atoi(fields[2])
		if err != nil || count < 0 {
			return Report{}, fmt.Errorf("invalid execution count on line %d: %q", lines, fields[2])
		}
		pkg := path.Dir(fileName[:separator])
		total := byPackage[pkg]
		if total == nil {
			total = &totals{}
			byPackage[pkg] = total
		}
		total.statements += statements
		if count > 0 {
			total.covered += statements
		}
	}
	if err := scanner.Err(); err != nil {
		return Report{}, fmt.Errorf("read coverage profile: %w", err)
	}
	if len(byPackage) == 0 {
		return Report{}, fmt.Errorf("coverage profile contains no statements")
	}

	report := Report{Minimum: minimum}
	for pkg, total := range byPackage {
		percent := percent(total.covered, total.statements)
		report.Packages = append(report.Packages, PackageCoverage{
			Package: pkg, Covered: total.covered, Statements: total.statements, Percent: percent,
		})
		report.Covered += total.covered
		report.Statements += total.statements
	}
	sort.Slice(report.Packages, func(i, j int) bool {
		return report.Packages[i].Package < report.Packages[j].Package
	})
	report.Percent = percent(report.Covered, report.Statements)
	report.Passed = report.Percent >= minimum
	return report, nil
}

func percent(covered, statements int) float64 {
	if statements == 0 {
		return 0
	}
	return math.Round((float64(covered) * 1000 / float64(statements))) / 10
}
