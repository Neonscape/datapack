package workspace

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"datapack/internal/compiler"
	"datapack/internal/document"
	"datapack/internal/formula"
)

func validationSource(t *testing.T, source string) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "cue.mod/module.cue"), `module:"example.com/validation@v0"
language:{version:"v0.9.0"}`)
	path := filepath.Join(root, "tables/item/item.cue")
	writeFile(t, path, source)
	return path
}
func validationDocument(t *testing.T, path string) document.Table {
	t.Helper()
	table, err := document.LoadTable(path)
	if err != nil {
		t.Fatal(err)
	}
	return table
}

func TestCompilerDecodeDefaultsDoesNotHydrateSourcePreview(t *testing.T) {
	path := validationSource(t, `package item
#PrimaryKey:string
#Row:{amount:*7|int,nullable:null|int,tags:*[]|[...string]}
rows:{one:{nullable:null}}
_tool:{table:{namespace:"game",name:"item"}}`)
	before, _ := os.ReadFile(path)
	table := validationDocument(t, path)
	result := compiler.CompileAll(nil, []document.Table{table})
	if result.Package == nil || len(result.Diagnostics) != 0 {
		t.Fatalf("compile: %#v", result.Diagnostics)
	}
	values := result.Package.Tables[0].Rows[0].Values
	if !reflect.DeepEqual(values["amount"], int64(7)) || !reflect.DeepEqual(values["tags"], []any{}) {
		t.Fatalf("default decode: %#v", values)
	}
	if value, ok := values["nullable"]; !ok || value != nil {
		t.Fatal("explicit null lost")
	}
	appendCueDiagnostics(&table, formula.Evaluate(table).Rows)
	if len(table.Diagnostics) != 0 {
		t.Fatal(table.Diagnostics)
	}
	if _, exists := table.Rows[0].Values["amount"]; exists {
		t.Fatal("preview hydrated a source cell with a default")
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Fatal("validation changed source")
	}
}

func TestCompilerAndPreviewKeepFieldAndRowDiagnosticObservations(t *testing.T) {
	path := validationSource(t, `package item
#PrimaryKey:string
#Row:{price:int & >=10}
rows:{one:{price:1}}
_tool:{table:{namespace:"game",name:"item"}}`)
	table := validationDocument(t, path)
	compiled := compiler.CompileAll(nil, []document.Table{table})
	appendCueDiagnostics(&table, formula.Evaluate(table).Rows)
	if compiled.Package != nil || len(compiled.Diagnostics) != 2 || len(table.Diagnostics) != 2 {
		t.Fatalf("diagnostics: %#v / %#v", compiled.Diagnostics, table.Diagnostics)
	}
	for i, diagnostic := range compiled.Diagnostics {
		if diagnostic.RowID != "one" || diagnostic.Column != table.Diagnostics[i].Column || diagnostic.Message != table.Diagnostics[i].Message {
			t.Fatalf("observation %d: %#v / %#v", i, diagnostic, table.Diagnostics[i])
		}
	}
	if compiled.Diagnostics[0].Column != "price" || compiled.Diagnostics[1].Column != "" {
		t.Fatal("field/row observation order changed")
	}
}

func TestPrimaryKeyDiagnosticPolicyStaysWithCaller(t *testing.T) {
	path := validationSource(t, `package item
#PrimaryKey:string & =~"^[a-z]+$"
#Row:{price:int}
rows:{UPPER:{price:10}}
_tool:{table:{namespace:"game",name:"item"}}`)
	table := validationDocument(t, path)
	compiled := compiler.CompileAll(nil, []document.Table{table})
	appendCueDiagnostics(&table, formula.Evaluate(table).Rows)
	if len(compiled.Diagnostics) != 1 || len(table.Diagnostics) != 1 || compiled.Diagnostics[0].Key != "compile.invalidID" || compiled.Diagnostics[0].RowID != "UPPER" || compiled.Diagnostics[0].Message != table.Diagnostics[0].Message {
		t.Fatalf("ID policy: %#v / %#v", compiled.Diagnostics, table.Diagnostics)
	}
}

func TestCompilerAndPreviewRetainDifferentPackageSelection(t *testing.T) {
	path := validationSource(t, `package item
#PrimaryKey:string
#Row:{price:int}
rows:{one:{price:10}}
_tool:{table:{namespace:"game",name:"item"}}`)
	writeFile(t, filepath.Join(filepath.Dir(path), "other.cue"), `package other
#Unrelated:string`)
	table := validationDocument(t, path)
	compiled := compiler.CompileAll(nil, []document.Table{table})
	appendCueDiagnostics(&table, formula.Evaluate(table).Rows)
	if compiled.Package == nil || len(compiled.Diagnostics) != 0 {
		t.Fatalf("compiler lost its explicit package selector: %#v", compiled.Diagnostics)
	}
	if len(table.Diagnostics) == 0 {
		t.Fatal("preview's unqualified package selection unexpectedly changed")
	}
}
