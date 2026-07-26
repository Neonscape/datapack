package schemamodel

import (
	"testing"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/parser"
)

func TestParseRepresentsStructuredFieldTypes(t *testing.T) {
	file, err := parser.ParseFile("schema.cue", `value: {
    name: string & != ""
    rarity: "N" | "R"
    tags: [...string]
    position: {x: number, y: number}
    description: string | null
    maxStack: *1 | int & >=1
}`)
	if err != nil {
		t.Fatal(err)
	}
	root := file.Decls[0].(*ast.Field).Value.(*ast.StructLit)
	fields := root.Elts

	name, err := Parse(fields[0].(*ast.Field).Value)
	if err != nil || name.Kind != KindString || len(name.Requirements) != 1 || name.Requirements[0].Expression != `!=""` {
		t.Fatalf("name = %#v, error = %v", name, err)
	}
	rarity, err := Parse(fields[1].(*ast.Field).Value)
	if err != nil || rarity.Kind != KindEnum || len(rarity.EnumValues) != 2 {
		t.Fatalf("rarity = %#v, error = %v", rarity, err)
	}
	tags, err := Parse(fields[2].(*ast.Field).Value)
	if err != nil || tags.Kind != KindArray || tags.Element == nil || tags.Element.Kind != KindString {
		t.Fatalf("tags = %#v, error = %v", tags, err)
	}
	position, err := Parse(fields[3].(*ast.Field).Value)
	if err != nil || position.Kind != KindStruct || len(position.Fields) != 2 {
		t.Fatalf("position = %#v, error = %v", position, err)
	}
	description, err := Parse(fields[4].(*ast.Field).Value)
	if err != nil || description.Kind != KindString || !description.Nullable {
		t.Fatalf("description = %#v, error = %v", description, err)
	}
	maxStack, err := Parse(fields[5].(*ast.Field).Value)
	if err != nil || maxStack.Kind != KindInteger || maxStack.Default == nil || maxStack.Default.Expression != "1" {
		t.Fatalf("maxStack = %#v, error = %v", maxStack, err)
	}
}

func TestParseKeepsReferencesAndConstraints(t *testing.T) {
	file, err := parser.ParseFile("schema.cue", `value: common.#GameID & common.#NonEmpty`)
	if err != nil {
		t.Fatal(err)
	}
	node, err := Parse(file.Decls[0].(*ast.Field).Value)
	if err != nil {
		t.Fatal(err)
	}
	if node.Kind != KindReference || node.Reference != "common.#GameID" || len(node.Requirements) != 1 || node.Requirements[0].Reference != "common.#NonEmpty" {
		t.Fatalf("node = %#v", node)
	}
}

func TestParseKeepsConstraintReferenceAfterBuiltinType(t *testing.T) {
	file, err := parser.ParseFile("schema.cue", `value: string & common.#NonEmpty`)
	if err != nil {
		t.Fatal(err)
	}
	node, err := Parse(file.Decls[0].(*ast.Field).Value)
	if err != nil {
		t.Fatal(err)
	}
	if node.Kind != KindString || len(node.Requirements) != 1 || node.Requirements[0].Reference != "common.#NonEmpty" {
		t.Fatalf("node = %#v", node)
	}
}
