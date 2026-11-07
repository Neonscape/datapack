package functional

import (
	"path/filepath"
	"testing"
)

func TestExportBoundaryDecimalPrecision(t *testing.T) {
	root, project := emptyProject(t)
	writeFile(t, filepath.Join(root, "tables", "value", "value.cue"), `package value
#PrimaryKey: string
#Row: {value: number}
rows: {one: {value: 0.12345678901234567890123456789}}
_tool: {table: {namespace: "boundary", name: "value"}}
_export: {csv: {enabled: true, columns: ["id", "value"]}}
`)
	service := openProject(t, project)
	result, err := service.Export()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := readFile(t, result.Files[0]), "id,value\none,0.12345678901234567890123456789\n"; got != want {
		t.Fatalf("CSV = %q, want %q", got, want)
	}
}

func TestExportBoundaryDecimalDefaultsAndFormula(t *testing.T) {
	root, project := emptyProject(t)
	writeFile(t, filepath.Join(root, "tables", "value", "value.cue"), `package value
#PrimaryKey: string
#Row: {
    fixed: number
    defaulted: *0.12345678901234567890123456789 | number
    nullableDefault: *0.12345678901234567890123456789 | number | null
    array: *[0.12345678901234567890123456789] | [...number]
    structure: *{value: 0.12345678901234567890123456789} | {value: number}
    computed: *99 | number
}
rows: {one: {fixed: 0.125, nullableDefault: null}}
_tool: {table: {namespace: "boundary", name: "value"}, formulas: {one: {computed: "fixed * 2"}}}
_export: {csv: {enabled: true, columns: ["id", "fixed", "defaulted", "nullableDefault", "array", "structure", "computed"], encoding: {array: "json", structure: "json"}}}
`)
	service := openProject(t, project)
	result, err := service.Export()
	if err != nil {
		t.Fatal(err)
	}
	want := "id,fixed,defaulted,nullableDefault,array,structure,computed\none,0.125,0.12345678901234567890123456789,,[0.12345678901234567890123456789],\"{\"\"value\"\":0.12345678901234567890123456789}\",0.25\n"
	if got := readFile(t, result.Files[0]); got != want {
		t.Fatalf("CSV = %q, want %q", got, want)
	}
}

func TestExportBoundaryIntegerDecimalValidation(t *testing.T) {
	for _, test := range []struct {
		name, expression, want string
		schema                 string
		reject                 bool
	}{
		{name: "invalid fraction", expression: "1.0000000000000000000001", reject: true},
		{name: "large integral decimal", expression: "9007199254740993.0", want: "9007199254740993"},
		{name: "negative integral decimal", expression: "-9007199254740993.0", want: "-9007199254740993"},
		{name: "integral exponent", expression: "9.007199254740993e15", want: "9007199254740993"},
		{name: "small integral decimal", expression: "2.0", want: "2"},
		{name: "large decimal source default", expression: "*9007199254740993.0 | number", want: "9007199254740993"},
		{name: "small decimal source default", expression: "*2.0 | number", want: "2"},
		{name: "fractional source default", expression: "*1.0000000000000000000001 | number", reject: true},
		{name: "large integer default", schema: "*9007199254740993 | int", want: "9007199254740993"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root, project := emptyProject(t)
			schema := test.schema
			if schema == "" {
				schema = "int"
			}
			row := "{}"
			if test.expression != "" {
				row = "{value: " + test.expression + "}"
			}
			writeFile(t, filepath.Join(root, "tables", "value", "value.cue"), `package value
#PrimaryKey: string
#Row: {value: `+schema+`}
rows: {one: `+row+`}
_tool: {table: {namespace: "boundary", name: "value"}}
_export: {csv: {enabled: true, columns: ["id", "value"]}}
`)
			service := openProject(t, project)
			result, err := service.Export()
			if test.reject {
				if err == nil {
					t.Fatalf("fractional integer exported: %s", readFile(t, result.Files[0]))
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got, want := readFile(t, result.Files[0]), "id,value\none,"+test.want+"\n"; got != want {
				t.Fatalf("CSV = %q, want %q", got, want)
			}
		})
	}
}

func TestExportBoundaryPreservesSelectedDefaultsDuringValidation(t *testing.T) {
	for _, test := range []struct{ name, schema, expression string }{
		{"scalar", "number & >=0.7", "*0.5 | 0.75"},
		{"structure", "{x: number & >=0.7}", "{x: *0.5 | 0.75}"},
		{"array", "[...number & >=0.7]", "[*0.5 | 0.75]"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root, project := emptyProject(t)
			writeFile(t, filepath.Join(root, "tables", "value", "value.cue"), `package value
#PrimaryKey: string
#Row: {value: `+test.schema+`}
rows: {one: {value: `+test.expression+`}}
_tool: {table: {namespace: "boundary", name: "value"}}
_export: {csv: {enabled: true, columns: ["id", "value"], encoding: {value: "json"}}}
`)
			service := openProject(t, project)
			result, err := service.Export()
			if err == nil {
				t.Fatalf("validation replaced selected default 0.5 with another branch: %s", readFile(t, result.Files[0]))
			}
		})
	}
}
