package schemaregistry

import (
	"testing"

	"datapack/internal/document"
	"datapack/internal/schemamodel"
)

func TestBuildRegistersAndIndexesSchemaUsages(t *testing.T) {
	public := document.PublicSchema{
		Path:    "schema/common.cue",
		Package: "common",
		Definitions: []document.SchemaDefinition{
			{Name: "GameID", Kind: KindType, Schema: schemamodel.Node{Kind: schemamodel.KindString}},
			{Name: "NonNegative", Kind: KindConstraint, AppliesTo: []string{"integer", "number"}, Schema: schemamodel.Node{Kind: schemamodel.KindUnknown, Requirements: []schemamodel.Requirement{{Expression: ">=0"}}}},
		},
	}
	table := document.Table{
		Path:       "tables/item.cue",
		Namespace:  "rpg",
		Name:       "item",
		PrimaryKey: schemamodel.Node{Kind: schemamodel.KindReference, Reference: "common.#GameID"},
		Columns: []document.Column{
			{Name: "price", Schema: schemamodel.Node{Kind: schemamodel.KindInteger, Requirements: []schemamodel.Requirement{{Expression: "common.#NonNegative", Reference: "common.#NonNegative"}}}},
			{Name: "position", Schema: schemamodel.Node{Kind: schemamodel.KindReference, Reference: "_local.#Position"}},
		},
		LocalSchema: []document.SchemaDefinition{
			{Name: "Position", Kind: KindType, Schema: schemamodel.Node{Kind: schemamodel.KindStruct}},
		},
	}

	registry := Build([]document.PublicSchema{public}, []document.Table{table})
	if len(registry.Diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", registry.Diagnostics)
	}
	if _, exists := registry.Definitions[ID("common.#GameID")]; !exists {
		t.Fatal("GameID was not registered")
	}
	if _, exists := registry.Definitions[ID("rpg.item._local.#Position")]; !exists {
		t.Fatal("local Position was not registered")
	}
	if got := registry.Usages[ID("common.#NonNegative")]; len(got) != 1 || got[0].Field != "price" {
		t.Fatalf("NonNegative usages = %#v", got)
	}
	if got := registry.Usages[ID("rpg.item._local.#Position")]; len(got) != 1 || got[0].Field != "position" {
		t.Fatalf("Position usages = %#v", got)
	}
}

func TestBuildRejectsInvalidReferenceKinds(t *testing.T) {
	public := document.PublicSchema{
		Path:    "schema/common.cue",
		Package: "common",
		Definitions: []document.SchemaDefinition{
			{Name: "GameID", Kind: KindType, Schema: schemamodel.Node{Kind: schemamodel.KindString}},
			{Name: "NonNegative", Kind: KindConstraint, AppliesTo: []string{"integer"}, Schema: schemamodel.Node{Kind: schemamodel.KindUnknown}},
		},
	}
	table := document.Table{
		Path:       "tables/item.cue",
		Namespace:  "rpg",
		Name:       "item",
		PrimaryKey: schemamodel.Node{Kind: schemamodel.KindString},
		Columns: []document.Column{
			{Name: "badType", Schema: schemamodel.Node{Kind: schemamodel.KindReference, Reference: "common.#NonNegative"}},
			{Name: "badRequirement", Schema: schemamodel.Node{Kind: schemamodel.KindInteger, Requirements: []schemamodel.Requirement{{Expression: "common.#GameID", Reference: "common.#GameID"}}}},
		},
	}

	registry := Build([]document.PublicSchema{public}, []document.Table{table})
	if len(registry.Diagnostics) != 2 {
		t.Fatalf("diagnostics = %#v", registry.Diagnostics)
	}
}

func TestBuildRejectsConstraintAppliedToIncompatibleType(t *testing.T) {
	public := document.PublicSchema{
		Path:    "schema/common.cue",
		Package: "common",
		Definitions: []document.SchemaDefinition{
			{Name: "NonNegative", Kind: KindConstraint, AppliesTo: []string{"integer", "number"}, Schema: schemamodel.Node{Kind: schemamodel.KindUnknown}},
		},
	}
	table := document.Table{
		Path:       "tables/item.cue",
		Namespace:  "rpg",
		Name:       "item",
		PrimaryKey: schemamodel.Node{Kind: schemamodel.KindString},
		Columns: []document.Column{
			{Name: "name", Schema: schemamodel.Node{Kind: schemamodel.KindString, Requirements: []schemamodel.Requirement{{Expression: "common.#NonNegative", Reference: "common.#NonNegative"}}}},
		},
	}

	registry := Build([]document.PublicSchema{public}, []document.Table{table})
	if len(registry.Diagnostics) != 1 || registry.Diagnostics[0].Field != "name" {
		t.Fatalf("diagnostics = %#v", registry.Diagnostics)
	}
}

func TestBuildRejectsSchemaCycles(t *testing.T) {
	public := document.PublicSchema{
		Path:    "schema/common.cue",
		Package: "common",
		Definitions: []document.SchemaDefinition{
			{Name: "A", Kind: KindConstraint, AppliesTo: []string{"integer"}, Schema: schemamodel.Node{Kind: schemamodel.KindReference, Reference: "common.#B"}},
			{Name: "B", Kind: KindConstraint, AppliesTo: []string{"integer"}, Schema: schemamodel.Node{Kind: schemamodel.KindReference, Reference: "common.#A"}},
		},
	}

	registry := Build([]document.PublicSchema{public}, nil)
	if len(registry.Diagnostics) == 0 {
		t.Fatal("expected cycle diagnostic")
	}
}

func TestIndexOrdersDefinitionsAndIncludesUsages(t *testing.T) {
	registry := Build([]document.PublicSchema{{
		Path:    "schema/common.cue",
		Package: "common",
		Definitions: []document.SchemaDefinition{
			{Name: "Zebra", Kind: KindType, Schema: schemamodel.Node{Kind: schemamodel.KindString}},
			{Name: "Alpha", Kind: KindType, Schema: schemamodel.Node{Kind: schemamodel.KindString}},
		},
	}}, []document.Table{{
		Path: "tables/item.cue", Namespace: "rpg", Name: "item",
		PrimaryKey: schemamodel.Node{Kind: schemamodel.KindReference, Reference: "common.#Zebra"},
	}})

	index := registry.Index()
	if len(index.Definitions) != 2 || index.Definitions[0].ID != "common.#Alpha" || index.Definitions[1].ID != "common.#Zebra" {
		t.Fatalf("definitions = %#v", index.Definitions)
	}
	if got := index.Definitions[1].Usages; len(got) != 1 || got[0].Field != "#PrimaryKey" {
		t.Fatalf("usages = %#v", got)
	}
}

func TestIndexUsesEmptyDiagnosticsSlice(t *testing.T) {
	index := Build(nil, nil).Index()
	if index.Definitions == nil || index.Diagnostics == nil {
		t.Fatalf("Index() = %#v", index)
	}
}
