package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"datapack/internal/document"
	"datapack/internal/foreignkey"
)

func TestCompileForeignKeysUseFinalDefaultAndNullValues(t *testing.T) {
	for _, tc := range []struct {
		name, schema, row string
		want              any
		invalid           bool
	}{
		{"scalar default", `*"burn" | string`, `{}`, "burn", false},
		{"array default", `*["burn","burn"] | [...string]`, `{}`, []any{"burn", "burn"}, false},
		{"nullable array default", `*[null,"burn"] | [...(null|string)]`, `{}`, []any{nil, "burn"}, false},
		{"nullable array literal", `[...(null|string)]`, `{target:[null,"burn"]}`, []any{nil, "burn"}, false},
		{"actual null", `null | string`, `{target:null}`, nil, false},
		{"null default", `*"burn" | null | string`, `{target:null}`, "burn", false},
		{"missing target default", `*"missing" | string`, `{}`, nil, true},
		{"missing target array", `*["burn","missing"] | [...string]`, `{}`, nil, true},
		{"missing input", `string`, `{}`, nil, true},
		{"null forbidden", `string`, `{target:null}`, nil, true},
		{"empty not null", `string`, `{target:""}`, nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			var tables []document.Table
			for _, source := range []struct{ name, body string }{
				{"effect", `#Row:{name:string}
rows:{burn:{name:"Burn"}}`},
				{"item", fmt.Sprintf("#Row:{target:%s}\nrows:{one:%s}", tc.schema, tc.row)},
			} {
				path := filepath.Join(root, source.name, source.name+".cue")
				keys := "{}"
				if source.name == "item" {
					keys = `{target:{table:"rpg.effect"}}`
				}
				writeFile(t, path, fmt.Sprintf("package %s\n#PrimaryKey:string\n%s\n_tool:{table:{namespace:\"rpg\",name:%q},foreignKeys:%s}", source.name, source.body, source.name, keys))
				table, err := document.LoadTable(path)
				if err != nil {
					t.Fatal(err)
				}
				tables = append(tables, table)
			}
			before, err := os.ReadFile(tables[1].Path)
			if err != nil {
				t.Fatal(err)
			}
			result := CompileAll(nil, tables)
			if tc.invalid {
				if result.Package != nil || len(result.Diagnostics) == 0 {
					t.Fatalf("invalid compile: %#v", result)
				}
			} else {
				if result.Package == nil || len(result.Diagnostics) > 0 {
					t.Fatalf("valid compile: %#v", result)
				}
				if got := result.Package.Tables[1].Rows[0].Values["target"]; !reflect.DeepEqual(got, tc.want) {
					t.Fatalf("got %#v want %#v", got, tc.want)
				}
			}
			after, err := os.ReadFile(tables[1].Path)
			if err != nil || string(before) != string(after) {
				t.Fatal("compile modified source")
			}
		})
	}
}

func TestImportedForeignKeyDefaultIsIndexedAndCompiled(t *testing.T) {
	for _, id := range []string{"burn", "missing"} {
		t.Run(id, func(t *testing.T) {
			root := t.TempDir()
			writeFile(t, filepath.Join(root, "cue.mod", "module.cue"), "module:\"example.com/game\"\nlanguage:{version:\"v0.9.0\"}")
			schemaPath := filepath.Join(root, "schema", "common.cue")
			writeFile(t, schemaPath, fmt.Sprintf("package common\n#EffectID: *%q | string\n_schema:{definitions:{EffectID:{kind:\"type\",type:\"string\"}}}", id))
			schema, err := document.LoadPublicSchema(schemaPath)
			if err != nil {
				t.Fatal(err)
			}
			var tables []document.Table
			for _, source := range []struct{ name, body string }{
				{"effect", `#PrimaryKey:string
#Row:{name:string}
rows:{burn:{name:"Burn"}}
_tool:{table:{namespace:"rpg",name:"effect"}}`},
				{"item", `import "example.com/game/schema:common"
#PrimaryKey:string
#Row:{target:common.#EffectID}
rows:{one:{}}
_tool:{table:{namespace:"rpg",name:"item"},foreignKeys:{target:{table:"rpg.effect"}}}`},
			} {
				path := filepath.Join(root, "tables", source.name, source.name+".cue")
				writeFile(t, path, "package "+source.name+"\n"+source.body)
				table, err := document.LoadTable(path)
				if err != nil {
					t.Fatal(err)
				}
				tables = append(tables, table)
			}
			index := foreignkey.Build(tables)
			usages := index.Usages[foreignkey.Target{Table: "rpg.effect", RowID: id}]
			if len(usages) != 1 || usages[0].Source != foreignkey.SourceDefault {
				t.Fatalf("default provenance: %#v", index)
			}
			result := CompileAll([]document.PublicSchema{schema}, tables)
			if id == "burn" {
				if result.Package == nil || len(result.Diagnostics) > 0 || len(index.Diagnostics) > 0 {
					t.Fatalf("result=%#v index=%#v", result, index)
				}
			} else if result.Package != nil || len(result.Diagnostics) == 0 || len(index.Diagnostics) == 0 {
				t.Fatalf("invalid imported default accepted: %#v %#v", result, index)
			}
		})
	}
}

func TestCompileRejectsInvalidForeignKeyDeclarationsOnEmptyTables(t *testing.T) {
	for _, tc := range []struct{ name, primary, column, key, target string }{
		{"invalid target primary key", "int", "string", "target", "rpg.effect"},
		{"invalid scalar column", "string", "int", "target", "rpg.effect"},
		{"invalid array column", "string", "[...int]", "target", "rpg.effect"},
		{"missing column", "string", "string", "absent", "rpg.effect"},
		{"missing table", "string", "string", "target", "rpg.absent"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			var tables []document.Table
			for _, source := range []struct{ name, primary, schema, keys string }{
				{"effect", tc.primary, `{name:string}`, `{}`},
				{"item", "string", "{target:" + tc.column + "}", fmt.Sprintf("{%s:{table:%q}}", tc.key, tc.target)},
			} {
				path := filepath.Join(root, source.name, source.name+".cue")
				writeFile(t, path, fmt.Sprintf("package %s\n#PrimaryKey:%s\n#Row:%s\nrows:{}\n_tool:{table:{namespace:\"rpg\",name:%q},foreignKeys:%s}", source.name, source.primary, source.schema, source.name, source.keys))
				table, err := document.LoadTable(path)
				if err != nil {
					t.Fatal(err)
				}
				tables = append(tables, table)
			}
			if result := CompileAll(nil, tables); result.Package != nil || len(result.Diagnostics) == 0 {
				t.Fatalf("invalid empty table accepted: %#v", result)
			}
			if index := foreignkey.Build(tables); !index.Incomplete || len(index.Diagnostics) == 0 {
				t.Fatalf("invalid index accepted: %#v", index)
			}
		})
	}
}
