package document

import (
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/cue/parser"
	"datapack/internal/cueformat"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestLocalSchemaRenamePreservesPublicAndRootDefinitions(t *testing.T) {
	root := t.TempDir()
	write := func(relative, source string) string {
		path := filepath.Join(root, relative)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	write("cue.mod/module.cue", `module:"example.com/game@v0"
language:{version:"v0.9.0"}`)
	common := write("schema/common/common.cue", `package common
#ID: string
_schema: {definitions: {ID: {kind: "type", type: "string"}}}`)
	table := write("tables/item/item.cue", `package item
import "example.com/game/schema/common"
#ID: string
#PrimaryKey: string
#Row: {local: _local.#ID, public: common.#ID, root: #ID}
rows: {one: {local: "a", public: "b", root: "c"}}
_local: {#ID: string, #Other: #ID}
_schema: {localDefinitions: {ID: {kind: "type", type: "string"}, Other: {kind: "type", type: "string"}}}
_tool: {table: {namespace: "game", name: "item"}}`)
	instances := load.Instances([]string{"."}, &load.Config{Dir: filepath.Dir(table), Package: "item"})
	if len(instances) != 1 {
		t.Fatal("unexpected package count")
	}
	if err := cuecontext.New().BuildInstance(instances[0]).Err(); err != nil {
		t.Fatalf("reproducer must compile before rename: %v", err)
	}
	assertScopePackageCompiles(t, table)
	before, _ := os.ReadFile(common)
	if _, err := RenameLocalSchemaDefinition(table, "ID", "LocalID", []string{common, table}); err != nil {
		t.Fatal(err)
	}
	assertScopePackageCompiles(t, table)
	after, _ := os.ReadFile(common)
	if string(after) != string(before) {
		t.Fatal("unrelated public definition changed")
	}
	data, _ := os.ReadFile(table)
	for _, want := range []string{"#ID: string", "public: common.#ID", "root: #ID", "local: _local.#LocalID", "#Other: #LocalID"} {
		if !regexp.MustCompile(strings.ReplaceAll(regexp.QuoteMeta(want), " ", `\s+`)).Match(data) {
			t.Fatalf("missing %s: %s", want, data)
		}
	}
}

func TestPublicRenameResolvesImportAliasesAndKeepsOtherScopes(t *testing.T) {
	root := t.TempDir()
	write := func(relative, source string) string {
		path := filepath.Join(root, relative)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	write("cue.mod/module.cue", `module: "example.com/game@v0"
language:{version:"v0.9.0"}`)
	target := write("schema/common/common.cue", `package common
#ID: string
#Other: #ID
_schema: {definitions: {ID: {kind: "type", type: "string"}, Other: {kind: "type", type: "string"}}}`)
	other := write("schema/other/other.cue", `package other
#ID: int
#Other: #ID
_schema: {definitions: {ID: {kind: "type", type: "integer"}, Other: {kind: "type", type: "integer"}}}`)
	table := write("tables/item/item.cue", `package item
import (
    shared "example.com/game/schema/common"
    other "example.com/game/schema/other"
)
#PrimaryKey: string
#Row: {public: shared.#ID, otherValue: other.#ID, local: _local.#ID}
rows: {one: {public: "x", otherValue: 1, local: "y"}}
_local: {#ID: string}
_schema: {localDefinitions: {ID: {kind: "type", type: "string"}}}
_tool: {table: {namespace: "game", name: "item"}}`)
	assertScopePackageCompiles(t, table)
	before, _ := os.ReadFile(other)
	if _, err := RenamePublicSchemaDefinition(target, "ID", "GameID", []string{target, other, table}); err != nil {
		t.Fatal(err)
	}
	assertScopePackageCompiles(t, table)
	after, _ := os.ReadFile(other)
	if string(before) != string(after) {
		t.Fatal("unrelated public source changed")
	}
	data, _ := os.ReadFile(table)
	for _, want := range []string{"public: shared.#GameID", "otherValue: other.#ID", "local: _local.#ID", "#ID: string"} {
		if !regexp.MustCompile(strings.ReplaceAll(regexp.QuoteMeta(want), " ", `\s+`)).Match(data) {
			t.Fatalf("missing %s: %s", want, data)
		}
	}
}

func assertScopePackageCompiles(t *testing.T, path string) {
	t.Helper()
	instances := load.Instances([]string{"."}, &load.Config{Dir: filepath.Dir(path)})
	if len(instances) != 1 {
		t.Fatal("expected one CUE package")
	}
	if err := cuecontext.New().BuildInstance(instances[0]).Err(); err != nil {
		t.Fatal(err)
	}
}

func TestSchemaReferenceBindingsAndCapture(t *testing.T) {
	for _, local := range []bool{false, true} {
		t.Run(map[bool]string{false: "public", true: "local"}[local], func(t *testing.T) {
			for _, test := range []struct {
				name, source, want string
				reject             bool
			}{
				{"nested old shadows target", `#Old:string
#Wrapper:{#Old:int,value:#Old}
#Use:#Old`, `#Wrapper: {#Old: int, value: #Old}`, false},
				{"new name captures reference", `#Old:string
#Wrapper:{#New:int,value:#Old}`, "", true},
				{"parenthesized selector", `#Old:string
#Use:(_local).#Old`, `#Use: (_local).#New`, false},
				{"let alias selector", `#Old:string
let L=_local
#Use:L.#Old`, `#Use: L.#New`, false},
				{"index wrapper rejected", `#Old:string
#Use:[_local][0].#Old`, "", true},
				{"nested package selector rejected", `#Old:string
#Use:_local.#Wrapper.#Old`, "", true},
				{"composite selector rejected", `#Old:string
#Use:(_local & {}).#Old`, "", true},
			} {
				t.Run(test.name, func(t *testing.T) {
					source := test.source
					if !local {
						source = strings.ReplaceAll(source, "_local", "common")
					}
					if local {
						source = "_local:{" + source + "}"
					}
					target, err := parser.ParseFile("common.cue", "package common\n"+source, parser.ParseComments)
					if err != nil {
						t.Fatal(err)
					}
					files := map[string]*ast.File{"common.cue": target}
					err = renameBoundSchemaReferences(files, "common.cue", "Old", "New", local)
					if test.reject {
						if err == nil {
							t.Fatal("unsafe rename accepted")
						}
						return
					}
					if err != nil {
						t.Fatal(err)
					}
					out, err := cueformat.Node(target)
					if err != nil {
						t.Fatal(err)
					}
					want := test.want
					if !local {
						want = strings.ReplaceAll(want, "_local", "common")
					}
					compact := func(x string) string { return strings.Join(strings.Fields(x), "") }
					if !strings.Contains(compact(string(out)), compact(want)) {
						t.Fatalf("want %s in %s", want, out)
					}
				})
			}
		})
	}
}

func TestSchemaRenamePreservesUntouchedSourcesAndRejectsAmbiguousPackage(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "a", "common.cue")
	other := filepath.Join(root, "b", "common.cue")
	table := filepath.Join(root, "item.cue")
	writeModuleFile(t, target, `package common
#Old:string
_schema:{definitions:{Old:{kind:"type",type:"string"}}}`)
	writeModuleFile(t, other, `package common
#Old:int
_schema:{definitions:{Old:{kind:"type",type:"integer"}}}`)
	writeModuleFile(t, table, `package item
#PrimaryKey:string
#Row:{value:common.#Old}
rows:{one:{value:"x"}}
_tool:{table:{namespace:"game",name:"item"}}`)
	originals := map[string]string{}
	for _, p := range []string{target, other, table} {
		b, _ := os.ReadFile(p)
		originals[p] = string(b)
	}
	if _, err := RenamePublicSchemaDefinition(target, "Old", "New", []string{target, other, table}); err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("ambiguous rename: %v", err)
	}
	for p, want := range originals {
		got, _ := os.ReadFile(p)
		if string(got) != want {
			t.Fatal("rejected rename changed " + p)
		}
	}
	// No reference to the target: whitespace and unrelated invalid input remain
	// untouched; rename is not a project-wide source formatter or repair command.
	writeModuleFile(t, table, `package item
#PrimaryKey:string
#Row:{value:string}
rows:{one:{value:"x"}}
_tool:{table:{namespace:"game",name:"item"}}`)
	before, _ := os.ReadFile(table)
	if _, err := RenamePublicSchemaDefinition(target, "Old", "New", []string{target, table}); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(table)
	if string(before) != string(after) {
		t.Fatal("untouched source reformatted")
	}
}

func TestPublicRenameMatchesImportPathAndPackageQualifier(t *testing.T) {
	root := t.TempDir()
	writeModuleFile(t, filepath.Join(root, "cue.mod", "module.cue"), `module:"example.com/game@v0"
language:{version:"v0.9.0"}`)
	target := filepath.Join(root, "schema", "shared", "rules.cue")
	other := filepath.Join(root, "schema", "other", "rules.cue")
	table := filepath.Join(root, "tables", "item", "item.cue")
	writeModuleFile(t, target, `package common
#Old:string
_schema:{definitions:{Old:{kind:"type",type:"string"}}}`)
	writeModuleFile(t, other, `package common
#Old:int
_schema:{definitions:{Old:{kind:"type",type:"integer"}}}`)
	writeModuleFile(t, table, `package item
import (
 first "example.com/game/schema/shared@v0:common"
 second "example.com/game/schema/other:common"
)
#PrimaryKey:string
#Row:{text:first.#Old, count:second.#Old}
rows:{one:{text:"x",count:1}}
_tool:{table:{namespace:"game",name:"item"}}`)
	assertScopePackageCompiles(t, table)
	before, _ := os.ReadFile(other)
	if _, err := RenamePublicSchemaDefinition(target, "Old", "New", []string{target, other, table}); err != nil {
		t.Fatal(err)
	}
	assertScopePackageCompiles(t, table)
	after, _ := os.ReadFile(other)
	if string(after) != string(before) {
		t.Fatal("other package with same name changed")
	}
	doc, err := LoadTable(table)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Columns[0].SchemaExpression != "first.#New" || doc.Columns[1].SchemaExpression != "second.#Old" {
		t.Fatalf("columns: %#v", doc.Columns)
	}
}

func TestLocalRenameLeavesOtherTablesWithSameDefinitionByteExact(t *testing.T) {
	root := t.TempDir()
	paths := []string{filepath.Join(root, "one.cue"), filepath.Join(root, "two.cue")}
	for i, path := range paths {
		source := `package one
#PrimaryKey:string
#Row:{value:_local.#Old}
rows:{one:{value:"x"}}
_local:{#Old:string}
_schema:{localDefinitions:{Old:{kind:"type",type:"string"}}}
_tool:{table:{namespace:"game",name:"one"}}`
		if i == 1 {
			source = strings.ReplaceAll(source, "one", "two")
		}
		writeModuleFile(t, path, source)
	}
	before, _ := os.ReadFile(paths[1])
	if _, err := RenameLocalSchemaDefinition(paths[0], "Old", "New", paths); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(paths[1])
	if string(before) != string(after) {
		t.Fatal("another table local definition was changed")
	}
}

func TestPublicRenameUpdatesSamePackageSiblingBareReferences(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "common.cue")
	sibling := filepath.Join(root, "extra.cue")
	writeModuleFile(t, target, `package common
#Old:string
_schema:{definitions:{Old:{kind:"type",type:"string"}}}`)
	writeModuleFile(t, sibling, `package common
#Other:#Old
_schema:{definitions:{Other:{kind:"type",type:"string"}}}`)
	if _, err := RenamePublicSchemaDefinition(target, "Old", "New", []string{target, sibling}); err != nil {
		t.Fatal(err)
	}
	doc, err := LoadPublicSchema(sibling)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Definitions[0].SchemaExpression != "#New" {
		t.Fatalf("sibling: %#v", doc)
	}
	assertScopePackageCompiles(t, target)
}

func TestLocalRenameRefusesSplitLocalPackageBeforeWrites(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "item.cue")
	sibling := filepath.Join(root, "extra.cue")
	writeModuleFile(t, target, `package item
#PrimaryKey:string
#Row:{value:_local.#Old}
rows:{one:{value:"x"}}
_local:{#Old:string}
_schema:{localDefinitions:{Old:{kind:"type",type:"string"}}}
_tool:{table:{namespace:"game",name:"item"}}`)
	writeModuleFile(t, sibling, `package item
_local:{#Other:#Old}`)
	before, _ := os.ReadFile(target)
	other, _ := os.ReadFile(sibling)
	if _, err := RenameLocalSchemaDefinition(target, "Old", "New", []string{target, sibling}); err == nil {
		t.Fatal("split local definitions accepted")
	}
	after, _ := os.ReadFile(target)
	otherAfter, _ := os.ReadFile(sibling)
	if string(before) != string(after) || string(other) != string(otherAfter) {
		t.Fatal("rejected split scope changed source")
	}
}

func TestSchemaRenameRejectsSplitDeclarationInOneFile(t *testing.T) {
	for _, local := range []bool{false, true} {
		t.Run(map[bool]string{false: "public", true: "local"}[local], func(t *testing.T) {
			source := `package common
#Old:int
#Old:>=0
_schema:{definitions:{Old:{kind:"type",type:"integer"}}}`
			if local {
				source = `package item
#PrimaryKey:string
#Row:{value:_local.#Old}
rows:{one:{value:1}}
_local:{#Old:int,#Old:>=0}
_schema:{localDefinitions:{Old:{kind:"type",type:"integer"}}}
_tool:{table:{namespace:"game",name:"item"}}`
			}
			path := writeSource(t, "source.cue", source)
			assertScopePackageCompiles(t, path)
			var err error
			if local {
				_, err = RenameLocalSchemaDefinition(path, "Old", "New", nil)
			} else {
				_, err = RenamePublicSchemaDefinition(path, "Old", "New", nil)
			}
			if err == nil || !strings.Contains(err.Error(), "split") {
				t.Fatalf("split declaration accepted: %v", err)
			}
			after, _ := os.ReadFile(path)
			if string(after) != source {
				t.Fatal("split declaration changed source")
			}
		})
	}
}
