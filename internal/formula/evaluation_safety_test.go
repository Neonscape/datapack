package formula

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"datapack/internal/document"
)

func safetyTable(t *testing.T, namespace, schema, rows, formulas string) document.Table {
	t.Helper()
	path := filepath.Join(t.TempDir(), "item.cue")
	source := fmt.Sprintf("package item\n#PrimaryKey: string\n#Row: %s\nrows: %s\n_tool: {table: {namespace: %q, name: \"item\"}, formulas: %s}\n", schema, rows, namespace, formulas)
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	table, err := document.LoadTable(path)
	if err != nil {
		t.Fatal(err)
	}
	return table
}

func TestSafetyFormulaDependenciesAndLexicalScope(t *testing.T) {
	for _, tc := range []struct{ name, ns, schema, rows, formulas, want string }{
		{"same row", "rpg", `{a: *0 | int, b: *0 | int}`, `{one: {}}`, `{one: {a: "b + 1", b: "10"}}`, "11"},
		{"static brackets", "rpg", `{a: *0 | int, b: *0 | int}`, `{one: {}, two: {}}`, `{one: {a: "rpg['item']['two']['b'] + 1"}, two: {b: "10"}}`, "11"},
		{"qualified field matches another column", "rpg", `{a: *0 | int, b: *0 | int}`, `{one: {}, two: {}}`, `{one: {a: "rpg.item.two.b + 1", b: "a + 1"}, two: {b: "10"}}`, "11"},
		{"local binding shadows column", "rpg", `{a: *0 | int, b: *0 | int}`, `{one: {}}`, `{one: {a: "let b = 10; b + 1", b: "a + 1"}}`, "11"},
		{"function name matches column", "rpg", `{a: *0 | int, abs: *0 | int}`, `{one: {}}`, `{one: {a: "abs(-2)", abs: "a + 1"}}`, "2"},
		{"namespace root matches column", "rpg", `{a: *0 | int, b: *0 | int, rpg: *0 | int}`, `{one: {}, two: {}}`, `{one: {a: "rpg.item.two.b + 1", rpg: "a + 1"}, two: {b: "10"}}`, "11"},
		{"bare and qualified namespace collision", "rpg", `{a: *0 | int, b: *0 | int, rpg: *0 | int}`, `{one: {}, two: {}}`, `{one: {a: "rpg + rpg.item.two.b", rpg: "10"}, two: {b: "10"}}`, "20"},
		{"map key is not a dependency", "rpg", `{a: *0 | int, b: *0 | int}`, `{one: {}}`, `{one: {a: "let m = {b: 10}; m.b + 1", b: "a + 1"}}`, "11"},
		{"dotted namespace", "rpg.balance", `{a: *0 | int, b: *0 | int}`, `{one: {}, two: {}}`, `{one: {a: "rpg.balance.item.two.b + 1"}, two: {b: "10"}}`, "11"},
		{"normal input default", "rpg", `{a: int, b: *10 | int}`, `{one: {}}`, `{one: {a: "b + 1"}}`, "11"},
		{"local namespace root", "rpg", `{a: *0 | int, b: *0 | int}`, `{one: {}}`, `{one: {a: "let rpg = {item: {one: {b: 10}}}; rpg.item.one.b + 1", b: "a + 1"}}`, "11"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			table := safetyTable(t, tc.ns, tc.schema, tc.rows, tc.formulas)
			before, err := os.ReadFile(table.Path)
			if err != nil {
				t.Fatal(err)
			}
			result := Evaluate(table)
			if len(result.Diagnostics) != 0 || fmt.Sprint(result.Rows["one"]["a"]) != tc.want {
				t.Fatalf("result=%#v, want a=%s", result, tc.want)
			}
			after, err := os.ReadFile(table.Path)
			if err != nil || string(before) != string(after) {
				t.Fatal("evaluation changed source")
			}
		})
	}
}

func TestSafetyFailedFormulasNeverLeaveDefaultResults(t *testing.T) {
	for _, tc := range []struct{ name, rows, formulas string }{
		{"self cycle", `{one: {}}`, `{one: {a: "a + 1"}}`},
		{"two cells", `{one: {}}`, `{one: {a: "b + 1", b: "a + 1"}}`},
		{"cross row cycle", `{one: {}, two: {}}`, `{one: {a: "rpg.item.two.a + 1"}, two: {a: "rpg.item.one.a + 1"}}`},
		{"parse failure", `{one: {}}`, `{one: {a: "b == nil ? 9 : b + 1", b: "("}}`},
		{"unknown function", `{one: {}}`, `{one: {a: "b == nil ? 9 : b + 1", b: "unknown(1)"}}`},
		{"runtime failure", `{one: {}}`, `{one: {a: "b == nil ? 9 : b + 1", b: "[1][8]"}}`},
		{"infinite result", `{one: {}}`, `{one: {a: "b == nil ? 9 : b + 1", b: "1 / 0"}}`},
		{"fixed value conflict", `{one: {b: 3}}`, `{one: {a: "b + 1", b: "10"}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			table := safetyTable(t, "rpg", `{a: *0 | int, b: *0 | int}`, tc.rows, tc.formulas)
			result := Evaluate(table)
			if len(result.Diagnostics) == 0 {
				t.Fatalf("invalid formulas have no diagnostics: %#v", result)
			}
			for _, f := range table.Formulas {
				if value, ok := result.Rows[f.RowID][f.Column]; ok {
					t.Errorf("failed formula %s.%s retained %v", f.RowID, f.Column, value)
				}
			}
		})
	}
}

func TestSafetyRejectsUntrackableAndForeignReferences(t *testing.T) {
	for _, expression := range []string{`rpg.item[selector].b`, `rpg.item.two[selector]`, `$env[selector]`, `rpg.other.two.b`, `rpg.item.missing.b`, `rpg.item.two.missing`} {
		t.Run(expression, func(t *testing.T) {
			table := safetyTable(t, "rpg", `{a: *0 | int, b: *0 | int, selector: string}`, `{one: {selector: "two"}, two: {selector: "b"}}`, `{one: {a: `+fmt.Sprintf("%q", expression)+`}, two: {b: "10"}}`)
			result := Evaluate(table)
			if len(result.Diagnostics) == 0 {
				t.Fatal("unsafe reference accepted")
			}
			if value, ok := result.Rows["one"]["a"]; ok {
				t.Fatalf("unsafe formula returned %v", value)
			}
		})
	}
}

func TestSafetyFunctionChecksIgnoreStringContents(t *testing.T) {
	table := safetyTable(t, "rpg", `{text: string}`, `{one: {}}`, `{one: {text: "\"Hello(there)\""}}`)
	result := Evaluate(table)
	if len(result.Diagnostics) != 0 || result.Rows["one"]["text"] != "Hello(there)" {
		t.Fatalf("result=%#v", result)
	}
}

func TestSafetyReferencesDoNotIncludeCalleesRootsOrLocals(t *testing.T) {
	table := document.Table{Namespace: "rpg", Name: "item", Columns: []document.Column{{Name: "a"}, {Name: "b"}, {Name: "rpg"}, {Name: "abs"}}}
	f := document.Formula{RowID: "one", Column: "a", Expression: "let b=2; abs(b) + rpg.item.two.a"}
	refs := References(table, f)
	if len(refs) != 1 || refs[0] != (Reference{RowID: "two", Column: "a"}) {
		t.Fatalf("references=%#v", refs)
	}
}

func TestSafetyFailedFixedInputCannotFallBackToDefault(t *testing.T) {
	table := safetyTable(t, "rpg", `{a: int, b: *10 | int}`, `{one: {b: missingIdentifier}}`, `{one: {a: "b + 1"}}`)
	result := Evaluate(table)
	if len(result.Diagnostics) == 0 {
		t.Fatal("bad fixed input accepted")
	}
	if _, ok := result.Rows["one"]["a"]; ok {
		t.Fatalf("failed fixed input fell back: %#v", result)
	}
}

func TestSafetyIntegerFormulaNormalization(t *testing.T) {
	table := safetyTable(t, "rpg", `{a: *0 | int}`, `{one: {}}`, `{one: {a: "round(1.4)"}}`)
	result := Evaluate(table)
	if len(result.Diagnostics) != 0 {
		t.Fatal(result.Diagnostics)
	}
	if _, ok := result.Rows["one"]["a"].(int64); !ok || fmt.Sprint(result.Rows["one"]["a"]) != "1" {
		t.Fatal(result.Rows)
	}
}

func TestSafetyQualifiedForeignKeyReadIsRejected(t *testing.T) {
	table := safetyTable(t, "rpg", `{a: *0 | int, b: string}`, `{one: {b: "one"}}`, `{one: {a: "len(rpg.item.one.b)"}}`)
	table.ForeignKeys = []document.ForeignKey{{Column: "b", Table: "rpg.item"}}
	result := Evaluate(table)
	if len(result.Diagnostics) == 0 || !strings.Contains(result.Diagnostics[0].Message, "foreign key") {
		t.Fatal(result.Diagnostics)
	}
}

func TestSafetyArrayElementReferencesAreRejected(t *testing.T) {
	for _, expression := range []string{"items[0]", "items[b]", "rpg.item.one.items[0]", "let x=items; x[0]"} {
		t.Run(expression, func(t *testing.T) {
			table := safetyTable(t, "rpg", `{a: *0 | int, b: *0 | int, items: [...int]}`, `{one: {items: [1, 2]}}`, `{one: {a: `+fmt.Sprintf("%q", expression)+`}}`)
			result := Evaluate(table)
			if len(result.Diagnostics) == 0 {
				t.Fatal("documented array reference boundary bypassed")
			}
			if _, ok := result.Rows["one"]["a"]; ok {
				t.Fatal(result.Rows)
			}
		})
	}
}
