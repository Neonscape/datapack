package compiler

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"cuelang.org/go/cue"
	"datapack/internal/cuerows"
	"datapack/internal/cueschema"
	"datapack/internal/document"
	"datapack/internal/formula"
)

func TestCompileMissingStructFieldDefaultsRetainPrecision(t *testing.T) {
	for _, tc := range []struct{ name, expression, want string }{
		{"large integer", "9007199254740993", "9007199254740993"},
		{"large integral decimal", "9007199254740993.0", "9007199254740993.0"},
		{"large fractional decimal", "9007199254740993.125", "9007199254740993.125"},
		{"negative decimal", "-9007199254740995.0", "-9007199254740995.0"},
		{"long decimal", "0.12345678901234567890123456789", "0.12345678901234567890123456789"},
		{"large exponent", "1e400", "1E+400"},
		{"small exponent", "1e-400", "1E-400"},
		{"ordinary decimal format", "0.1250", "0.125"},
		{"ordinary integral format", "2.0", "2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "defaults.cue")
			writeFile(t, path, `package defaults
#PrimaryKey: string
#Row: {q: {"exact-value": *`+tc.expression+` | number, enabled: *true | bool}}
rows: {missing: {}, explicit: {q: {"exact-value": `+tc.expression+`, enabled: true}}}
_tool: {table: {namespace: "probe", name: "defaults"}}
`)
			table, err := document.LoadTable(path)
			if err != nil {
				t.Fatal(err)
			}
			result := CompileAll(nil, []document.Table{table})
			if len(result.Diagnostics) != 0 || result.Package == nil {
				t.Fatalf("compile: %#v", result)
			}
			for _, row := range result.Package.Tables[0].Rows {
				encoded, err := json.Marshal(row.Values["q"])
				if err != nil {
					t.Fatal(err)
				}
				want := `{"enabled":true,"exact-value":` + tc.want + `}`
				if string(encoded) != want {
					t.Errorf("%s = %s, want %s", row.ID, encoded, want)
				}
			}
		})
	}
}

// Observe the actual boundary so the regression is tied to a selected CUE
// default, not to a manufactured float64 or a CSV-specific workaround.
func TestMissingStructDefaultPrecisionBoundary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "defaults.cue")
	writeFile(t, path, `package defaults
#PrimaryKey: string
#Row: {q: {"exact-value": *9007199254740993.0 | number}}
rows: {missing: {}}
_tool: {table: {namespace: "probe", name: "defaults"}}
`)
	table, err := document.LoadTable(path)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := cueschema.Load(table, table.Package)
	if err != nil {
		t.Fatal(err)
	}
	schema := loaded.LookupPath(cue.MakePath(cue.Def("Row")))
	evaluated := formula.EvaluateWithSchema(table, schema)
	values := exactValidationValues(table.Rows[0], evaluated.Rows["missing"], nil, schema)
	candidate, issues := cuerows.Validate(schema, loaded.LookupPath(cue.MakePath(cue.Def("PrimaryKey"))), "missing", values)
	if len(issues) != 0 {
		t.Fatal(issues)
	}
	field := candidate.LookupPath(cue.MakePath(cue.Str("q"), cue.Str("exact-value")))
	selected, hasDefault := field.Default()
	raw, err := field.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("evaluated=%#v exact=%#v field.Kind=%v selected.Kind=%v hasDefault=%v CUE_JSON=%s", evaluated.Rows["missing"], values, field.Kind(), selected.Kind(), hasDefault, raw)
	if !hasDefault || selected.Kind() != cue.FloatKind || string(raw) != "9007199254740993.0" {
		t.Fatalf("unexpected CUE default: %s", raw)
	}
	decoded, err := decodeExportValue(candidate)
	if err != nil {
		t.Fatal(err)
	}
	got, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"q":{"exact-value":9007199254740993.0}}` {
		t.Fatalf("compiled JSON = %s", got)
	}
}
