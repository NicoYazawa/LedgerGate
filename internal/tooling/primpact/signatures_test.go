package primpact

import "testing"

func TestFindBreakingExports(t *testing.T) {
	tests := []struct {
		name string
		diff string
		want []string
	}{
		{
			name: "removed function",
			diff: `@@ -1 +0,0 @@
-func Public(value int) string {`,
			want: []string{"func Public(value int) string"},
		},
		{
			name: "changed function signature",
			diff: `@@ -1 +1 @@
-func Public(value int) string {
+func Public(value string) string {`,
			want: []string{"func Public(value int) string"},
		},
		{
			name: "added function",
			diff: `@@ -0,0 +1 @@
+func Public(value int) string {`,
		},
		{
			name: "moved unchanged function",
			diff: `@@ -1 +1 @@
-func Public(value int) string {
+func    Public(value int) string {`,
		},
		{
			name: "removed method",
			diff: `@@ -1 +0,0 @@
-func (service *Service) Public(value int) error {`,
			want: []string{"func (service *Service) Public(value int) error"},
		},
		{
			name: "removed exported type",
			diff: `@@ -1 +0,0 @@
-type Public struct {`,
			want: []string{"type Public struct"},
		},
		{
			name: "body edit",
			diff: `@@ -4 +4 @@
-	return "old"
+	return "new"`,
		},
		{
			name: "one-line body edit",
			diff: `@@ -1 +1 @@
-func Public() int { return 1 }
+func Public() int { return 2 }`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := FindBreakingExports(test.diff)
			assertStrings(t, got, test.want)
		})
	}
}
