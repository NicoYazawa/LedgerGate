package primpact

import (
	"regexp"
	"sort"
	"strings"
)

var exportedDeclaration = regexp.MustCompile(
	`^(?:(?:func\s+(?:\([^)]*\)\s+)?)|(?:type\s+)|(?:var\s+)|(?:const\s+))[A-Z][A-Za-z0-9_]*\b`,
)

func FindBreakingExports(diff string) []string {
	removed := make(map[string]int)
	added := make(map[string]int)
	for _, line := range strings.Split(diff, "\n") {
		if len(line) < 2 || strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
			continue
		}
		declaration := strings.Join(strings.Fields(line[1:]), " ")
		if body := strings.Index(declaration, "{"); body >= 0 {
			declaration = strings.TrimSpace(declaration[:body])
		}
		if !exportedDeclaration.MatchString(declaration) {
			continue
		}
		switch line[0] {
		case '-':
			removed[declaration]++
		case '+':
			added[declaration]++
		}
	}

	var breaking []string
	for declaration, count := range removed {
		if count > added[declaration] {
			breaking = append(breaking, declaration)
		}
	}
	sort.Strings(breaking)
	return breaking
}
