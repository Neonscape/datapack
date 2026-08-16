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

func TestParseClassifiesKnownPredicates(t *testing.T) {
	file, err := parser.ParseFile("schema.cue", `value: string & != "" & =~"^[a-z]+$" & strings.MinRunes(1) & strings.MaxRunes(8)`)
	if err != nil {
		t.Fatal(err)
	}
	node, err := Parse(file.Decls[0].(*ast.Field).Value)
	if err != nil {
		t.Fatal(err)
	}
	requirements := node.Requirements
	if len(requirements) != 4 {
		t.Fatalf("requirements = %#v", requirements)
	}
	if requirements[0].Kind != RequirementNonEmpty {
		t.Fatalf("nonEmpty = %#v", requirements[0])
	}
	if requirements[1].Kind != RequirementRegex || requirements[1].Pattern != "^[a-z]+$" || requirements[1].Op != "=~" {
		t.Fatalf("regex = %#v", requirements[1])
	}
	if requirements[2].Kind != RequirementLength || requirements[2].Min == nil || *requirements[2].Min != 1 {
		t.Fatalf("min = %#v", requirements[2])
	}
	if requirements[3].Kind != RequirementLength || requirements[3].Max == nil || *requirements[3].Max != 8 {
		t.Fatalf("max = %#v", requirements[3])
	}
}

func TestParseClassifiesComparisons(t *testing.T) {
	file, err := parser.ParseFile("schema.cue", `value: int & >=0 & <9999`)
	if err != nil {
		t.Fatal(err)
	}
	node, err := Parse(file.Decls[0].(*ast.Field).Value)
	if err != nil {
		t.Fatal(err)
	}
	if len(node.Requirements) != 2 || node.Requirements[0].Kind != RequirementCompare || node.Requirements[0].Op != ">=" || node.Requirements[0].Value != "0" {
		t.Fatalf("requirements = %#v", node.Requirements)
	}
}

func TestPrintRoundTripsSupportedSchema(t *testing.T) {
	cases := []string{
		`string`,
		`string & != ""`,
		`int & >= 0 & < 9999`,
		`"N" | "R" | "SR"`,
		`string | null`,
		`[...string]`,
		`{x: number, y: number, label?: string}`,
		`common.#GameID`,
		`string & common.#NonEmpty`,
		`(int & >= 1) | *1`,
		`string & != "" & =~"^[a-z]+$" & strings.MinRunes(1) & strings.MaxRunes(8)`,
	}
	for _, expression := range cases {
		parsed, err := parser.ParseExpr("schema.cue", expression)
		if err != nil {
			t.Fatalf("parse %q: %v", expression, err)
		}
		node, err := Parse(parsed)
		if err != nil {
			t.Fatalf("Parse %q: %v", expression, err)
		}
		printed, err := Print(node)
		if err != nil {
			t.Fatalf("Print %q: %v", expression, err)
		}
		reparsed, err := parser.ParseExpr("schema.cue", printed)
		if err != nil {
			t.Fatalf("re-parse %q -> %q: %v", expression, printed, err)
		}
		reprint, err := Print(MustParse(t, reparsed))
		if err != nil {
			t.Fatalf("re-print %q: %v", printed, err)
		}
		if reprint != printed {
			t.Fatalf("round-trip %q -> %q -> %q", expression, printed, reprint)
		}
	}
}

func MustParse(t *testing.T, expression ast.Expr) Node {
	t.Helper()
	node, err := Parse(expression)
	if err != nil {
		t.Fatal(err)
	}
	return node
}
