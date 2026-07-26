package foreignkey

import (
	"testing"

	"datapack/internal/document"
)

func TestBuildIndexesScalarAndArrayReferences(t *testing.T) {
	tables := []document.Table{
		{
			Path: "effect.cue", Namespace: "rpg", Name: "effect",
			Rows: []document.Row{{ID: "burn"}, {ID: "freeze"}},
		},
		{
			Path: "item.cue", Namespace: "rpg", Name: "item",
			ForeignKeys: []document.ForeignKey{
				{Column: "effectID", Table: "rpg.effect"},
				{Column: "effectIDs", Table: "rpg.effect"},
			},
			Rows: []document.Row{{
				ID: "sword",
				Values: map[string]document.Cell{
					"effectID":  {Expression: `"burn"`},
					"effectIDs": {Expression: `["freeze", "burn"]`},
				},
			}},
		},
	}

	index := Build(tables)
	if len(index.Diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", index.Diagnostics)
	}
	if got := index.Usages[Target{Table: "rpg.effect", RowID: "burn"}]; len(got) != 2 {
		t.Fatalf("burn usages = %#v", got)
	}
}

func TestBuildReportsMissingForeignKeyTarget(t *testing.T) {
	index := Build([]document.Table{{
		Path:        "item.cue",
		Namespace:   "rpg",
		Name:        "item",
		ForeignKeys: []document.ForeignKey{{Column: "effectID", Table: "rpg.effect"}},
		Rows: []document.Row{{
			ID:     "sword",
			Values: map[string]document.Cell{"effectID": {Expression: `"missing"`}},
		}},
	}})
	if len(index.Diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v", index.Diagnostics)
	}
}
