package cuesource

import "testing"

func TestParseSchemaExtractsExplicitDefinitionKinds(t *testing.T) {
	schema := ParseSchema("common.cue", []byte(`package common

#GameID: string & != ""
#NonNegative: >=0

_schema: {
    definitions: {
        GameID: {kind: "type", type: "string"}
        NonNegative: {kind: "constraint", appliesTo: ["integer", "number"]}
    }
}`))

	if len(schema.Diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", schema.Diagnostics)
	}
	if schema.Package != "common" {
		t.Fatalf("package = %q, want common", schema.Package)
	}
	if len(schema.Definitions) != 2 {
		t.Fatalf("definitions = %#v", schema.Definitions)
	}
	if schema.Definitions[0].Name != "GameID" || schema.Definitions[0].Kind != "type" || schema.Definitions[0].Type != "string" {
		t.Fatalf("type definition = %#v", schema.Definitions[0])
	}
	if schema.Definitions[1].Name != "NonNegative" || schema.Definitions[1].Kind != "constraint" || len(schema.Definitions[1].AppliesTo) != 2 {
		t.Fatalf("constraint definition = %#v", schema.Definitions[1])
	}
}

func TestParseSchemaRejectsUnregisteredDefinitions(t *testing.T) {
	schema := ParseSchema("broken.cue", []byte(`package common
#GameID: string
_schema: {definitions: {}}
`))

	if len(schema.Diagnostics) != 1 {
		t.Fatalf("diagnostic count = %d, want 1: %#v", len(schema.Diagnostics), schema.Diagnostics)
	}
}
