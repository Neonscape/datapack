package schemaregistry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"datapack/internal/document"
)

func aliasSource(t *testing.T, root, relative, source string) string {
	t.Helper()
	path := filepath.Join(root, relative)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
func aliasModule(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	aliasSource(t, root, "cue.mod/module.cue", `module:"example.com/game@v0"
language:{version:"v0.9.0"}`)
	return root
}
func loadAliasSchema(t *testing.T, path string) document.PublicSchema {
	t.Helper()
	d, err := document.LoadPublicSchema(path)
	if err != nil {
		t.Fatal(err)
	}
	return d
}
func loadAliasTable(t *testing.T, path string) document.Table {
	t.Helper()
	d, err := document.LoadTable(path)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestImportAliasesTrackTypesConstraintsArraysNullableAndUsages(t *testing.T) {
	root := aliasModule(t)
	schema := loadAliasSchema(t, aliasSource(t, root, "schema/shared/rules.cue", `package common
#Value:int
#Positive:>=0
_schema:{definitions:{Value:{kind:"type",type:"integer"},Positive:{kind:"constraint",appliesTo:["integer"]}}}`))
	table := loadAliasTable(t, aliasSource(t, root, "tables/item/item.cue", `package item
import shared "example.com/game/schema/shared@v0:common"
#PrimaryKey:string
#Row:{amount:shared.#Value, values:[...shared.#Value], optional:null|shared.#Value, constrained:int & shared.#Positive}
rows:{one:{amount:1,values:[1,2],optional:null,constrained:2}}
_tool:{table:{namespace:"game",name:"item"}}`))
	registry := Build([]document.PublicSchema{schema}, []document.Table{table})
	if len(registry.Diagnostics) != 0 {
		t.Fatalf("diagnostics: %#v", registry.Diagnostics)
	}
	if len(registry.Usages[ID("common.#Value")]) != 3 || len(registry.Usages[ID("common.#Positive")]) != 1 {
		t.Fatalf("usages: %#v", registry.Usages)
	}
	if table.Columns[0].Schema.Reference != "shared.#Value" {
		t.Fatal("lexical reference changed")
	}
	data, err := json.Marshal(table)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "referenceImport") || strings.Contains(string(data), "sourcePackage") || strings.Contains(string(data), "example.com/game") {
		t.Fatalf("private binding context leaked into DTO: %s", data)
	}
}

func TestImportAliasDoesNotBindDifferentPackagePathOrVersion(t *testing.T) {
	for _, uri := range []string{"example.com/game/schema/other:common", "example.com/game/schema/shared@v1:common"} {
		t.Run(uri, func(t *testing.T) {
			root := aliasModule(t)
			schema := loadAliasSchema(t, aliasSource(t, root, "schema/shared/rules.cue", `package common
#Value:int
_schema:{definitions:{Value:{kind:"type",type:"integer"}}}`))
			source := `package item
import shared "` + uri + `"
#PrimaryKey:string
#Row:{amount:shared.#Value}
rows:{one:{amount:1}}
_tool:{table:{namespace:"game",name:"item"}}`
			table := loadAliasTable(t, aliasSource(t, root, "tables/item/item.cue", source))
			registry := Build([]document.PublicSchema{schema}, []document.Table{table})
			if len(registry.Diagnostics) != 1 || registry.Diagnostics[0].Key != "schema.referenceDoesNotExist" || len(registry.Usages[ID("common.#Value")]) != 0 {
				t.Fatalf("wrong package matched: %#v", registry)
			}
		})
	}
}

func TestAliasCyclesAcrossPublicDefinitionsUseCanonicalIDs(t *testing.T) {
	root := aliasModule(t)
	first := loadAliasSchema(t, aliasSource(t, root, "schema/first/rules.cue", `package first
import b "example.com/game/schema/second"
#Value:b.#Value
_schema:{definitions:{Value:{kind:"type",type:"integer"}}}`))
	second := loadAliasSchema(t, aliasSource(t, root, "schema/second/rules.cue", `package second
import a "example.com/game/schema/first"
#Value:a.#Value
_schema:{definitions:{Value:{kind:"type",type:"integer"}}}`))
	registry := Build([]document.PublicSchema{first, second}, nil)
	if len(registry.Diagnostics) != 1 || registry.Diagnostics[0].Key != "schema.cycle" {
		t.Fatalf("cycle: %#v", registry.Diagnostics)
	}
}

func TestShadowedImportNameIsNotReboundToPublicDefinition(t *testing.T) {
	root := aliasModule(t)
	table := loadAliasTable(t, aliasSource(t, root, "tables/item/item.cue", `package item
import shared "example.com/game/schema/shared:common"
#PrimaryKey:string
#Row:{direct:shared.#Value, nested:{shared:{#Value:string}, value:shared.#Value}}
rows:{one:{direct:1,nested:{value:"x"}}}
_tool:{table:{namespace:"game",name:"item"}}`))
	direct, nested := table.Columns[0].Schema, table.Columns[1].Schema.Fields[1].Schema
	if direct.ReferenceImport() == "" || nested.ReferenceImport() != "" {
		t.Fatalf("lexical binding: %#v %#v", direct, nested)
	}
}
