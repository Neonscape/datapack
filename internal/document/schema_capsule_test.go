package document

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResearchCapsuleProofAndFallback(t *testing.T) {
	base := `package items
#PrimaryKey: string
#Row: {value: int, text: string}
rows: {one: {value: 1, text: "中文"}, two: {value: 2, text: "two"}}
_tool: {table: {namespace: "test", name: "items"}, formulas: {}, foreignKeys: {}}
`
	for _, tc := range []struct {
		name, source string
		eligible     bool
	}{
		{"independent scalar rows", base, true},
		{"schema depends on rows", strings.Replace(base, "value: int", "value: *rows.one.value | int", 1), false},
		{"row depends on other row", strings.Replace(base, "value: 2", "value: rows.one.value", 1), false},
		{"duplicate identity", strings.Replace(base, "two:", "one:", 1), false},
		{"duplicate field", strings.Replace(base, "value: 2", "value: 2, value: 3", 1), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "items.cue")
			if err := os.WriteFile(path, []byte(tc.source), 0644); err != nil {
				t.Fatal(err)
			}
			table, err := LoadTable(path)
			if err != nil {
				t.Fatal(err)
			}
			capsule, ok := SchemaCapsule(table)
			if ok != tc.eligible {
				t.Fatalf("eligible=%v expected %v", ok, tc.eligible)
			}
			if ok && !strings.Contains(string(capsule), "rows: {}") {
				t.Fatal("full rows not removed")
			}
		})
	}
}
