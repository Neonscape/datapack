package formula

import (
	"fmt"
	"strings"
	"testing"

	"datapack/internal/document"
	"datapack/internal/schemamodel"
)

func TestReviewRuntimeNamesAreIsolatedFromCells(t *testing.T) {
	for _, tc := range []struct{ name, schema, rows, formulas, target, want string }{
		{"fixed function collision", `{a: int, abs: int}`, `{one: {abs: 7}}`, `{one: {a: "abs(-2)"}}`, "a", "2"},
		{"default function collision", `{a: int, abs: *7 | int}`, `{one: {}}`, `{one: {a: "abs(-2)"}}`, "a", "2"},
		{"computed function collision", `{z: int, abs: *0 | int}`, `{one: {}}`, `{one: {abs: "7", z: "abs(-2)"}}`, "z", "2"},
		{"fixed local collision", `{a: int, b: int}`, `{one: {b: 5}}`, `{one: {a: "let b=10; b+1"}}`, "a", "11"},
		{"default local collision", `{a: int, b: *5 | int}`, `{one: {}}`, `{one: {a: "let b=10; b+1"}}`, "a", "11"},
		{"computed local collision", `{z: int, b: *0 | int}`, `{one: {}}`, `{one: {b: "5", z: "let b=10; b+1"}}`, "z", "11"},
		{"binding reads shadowed input", `{a: int, b: int}`, `{one: {b: 5}}`, `{one: {a: "let b=b+1; b+1"}}`, "a", "7"},
		{"private callback binding", `{a: int, b: int}`, `{one: {b: 7}}`, `{one: {a: "let __datapack_cell=1; let __datapack_cell_=2; rpg.item.one.b + __datapack_cell + __datapack_cell_"}}`, "a", "10"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			table := safetyTable(t, "rpg", tc.schema, tc.rows, tc.formulas)
			result := Evaluate(table)
			if len(result.Diagnostics) != 0 || fmt.Sprint(result.Rows["one"][tc.target]) != tc.want {
				t.Fatalf("result=%#v, want %s=%s", result, tc.target, tc.want)
			}
		})
	}
}

func TestReviewPrivateCallbackColumnCannotShadowExecution(t *testing.T) {
	table := document.Table{Namespace: "rpg", Name: "item", Columns: []document.Column{{Name: "result", Schema: schemamodel.Node{Kind: schemamodel.KindInteger}}, {Name: "__datapack_cell", Schema: schemamodel.Node{Kind: schemamodel.KindInteger}}}, Rows: []document.Row{{ID: "one", Values: map[string]document.Cell{"__datapack_cell": {Expression: "7"}}}}, Formulas: []document.Formula{{RowID: "one", Column: "result", Expression: "__datapack_cell + rpg.item.one.__datapack_cell"}}}
	result := Evaluate(table)
	if len(result.Diagnostics) != 0 || fmt.Sprint(result.Rows["one"]["result"]) != "14" {
		t.Fatal(result)
	}
}

func TestReviewMissingInputsAreDifferentFromNull(t *testing.T) {
	for _, expression := range []string{`b == nil ? 9 : b+1`, `rpg.item.one.b == nil ? 9 : rpg.item.one.b+1`} {
		for _, tc := range []struct {
			name, schema, rows string
			missing            bool
		}{
			{"required absent", `{a: *0 | int, b: int}`, `{one: {}}`, true},
			{"nullable absent", `{a: *0 | int, b: null | int}`, `{one: {}}`, true},
			{"explicit null", `{a: int, b: null | int}`, `{one: {b: null}}`, false},
			{"default null", `{a: int, b: *null | int}`, `{one: {}}`, false},
		} {
			t.Run(tc.name+expression, func(t *testing.T) {
				table := safetyTable(t, "rpg", tc.schema, tc.rows, `{one: {a: `+fmt.Sprintf("%q", expression)+`}}`)
				result := Evaluate(table)
				if tc.missing {
					if _, ok := result.Rows["one"]["a"]; ok {
						t.Fatalf("missing input produced result: %#v", result)
					}
					if len(result.Diagnostics) == 0 || !strings.Contains(result.Diagnostics[0].Message, "missing") {
						t.Fatal(result.Diagnostics)
					}
				} else if len(result.Diagnostics) != 0 || fmt.Sprint(result.Rows["one"]["a"]) != "9" {
					t.Fatalf("actual null was rejected: %#v", result)
				}
			})
		}
	}
}

func TestReviewArrayElementFenceCoversWrappers(t *testing.T) {
	for _, expression := range []string{`items[:][0]`, `[items][0][0]`, `(items)[0]`, `{x:items}.x[0]`, `items["0"]`, `[1,2][0]`} {
		t.Run(expression, func(t *testing.T) {
			table := safetyTable(t, "rpg", `{a: *0 | int, items: [...int]}`, `{one: {items: [1,2]}}`, `{one: {a: `+fmt.Sprintf("%q", expression)+`}}`)
			result := Evaluate(table)
			if _, ok := result.Rows["one"]["a"]; ok || len(result.Diagnostics) == 0 {
				t.Fatalf("array element fence bypassed: %#v", result)
			}
		})
	}
}
