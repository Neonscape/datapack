package formula

import (
	"os"
	"testing"

	"datapack/internal/document"
	"datapack/internal/schemamodel"
)

func TestEvaluateCalculatesSameTableDependencies(t *testing.T) {
	table := document.Table{
		Namespace: "combat",
		Name:      "growth",
		Columns:   []document.Column{{Name: "baseAttack"}, {Name: "attack"}},
		Rows: []document.Row{
			{ID: "level_01", Values: map[string]document.Cell{"baseAttack": {Expression: "10"}}},
			{ID: "level_10", Values: map[string]document.Cell{"baseAttack": {Expression: "20"}}},
		},
		Formulas: []document.Formula{
			{RowID: "level_01", Column: "attack", Expression: "baseAttack * 2"},
			{RowID: "level_10", Column: "attack", Expression: "combat.growth.level_01.attack + baseAttack"},
		},
	}

	result := Evaluate(table)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", result.Diagnostics)
	}
	if result.Rows["level_01"]["attack"] != 20 || result.Rows["level_10"]["attack"] != 40 {
		t.Fatalf("rows = %#v", result.Rows)
	}
}

func TestEvaluateReportsFormulaCycles(t *testing.T) {
	table := document.Table{
		Columns: []document.Column{{Name: "attack"}},
		Rows:    []document.Row{{ID: "level_01", Values: map[string]document.Cell{}}},
		Formulas: []document.Formula{
			{RowID: "level_01", Column: "attack", Expression: "attack + 1"},
		},
	}

	result := Evaluate(table)
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Message != "formula has a circular or missing dependency" {
		t.Fatalf("diagnostics = %#v", result.Diagnostics)
	}
}

func TestEvaluateRejectsForeignKeyReads(t *testing.T) {
	table := document.Table{
		Columns:     []document.Column{{Name: "itemID"}, {Name: "price"}},
		ForeignKeys: []document.ForeignKey{{Column: "itemID", Table: "rpg.item"}},
		Rows:        []document.Row{{ID: "sale", Values: map[string]document.Cell{"itemID": {Expression: `"sword"`}}}},
		Formulas:    []document.Formula{{RowID: "sale", Column: "price", Expression: "itemID == 'sword' ? 10 : 0"}},
	}

	result := Evaluate(table)
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Message != `formulas cannot read foreign key column "itemID"` {
		t.Fatalf("diagnostics = %#v", result.Diagnostics)
	}
}

func TestEvaluateUsesCueDefaultsAsFormulaInputs(t *testing.T) {
	root := t.TempDir()
	path := root + "/growth.cue"
	if err := os.WriteFile(path, []byte(`package growth
#PrimaryKey: string
#Row: {baseAttack: *10 | int, attack: int}
rows: {level_01: {}}
_tool: {table: {namespace: "combat", name: "growth"}, formulas: {level_01: {attack: "baseAttack * 2"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	table, err := document.LoadTable(path)
	if err != nil {
		t.Fatal(err)
	}
	result := Evaluate(table)
	if len(result.Diagnostics) != 0 || result.Rows["level_01"]["attack"] != 20 {
		t.Fatalf("result = %#v", result)
	}
}

func TestEvaluateRejectsNonScalarFormulaTargets(t *testing.T) {
	table := document.Table{
		Columns:  []document.Column{{Name: "tags", Schema: schemamodel.Node{Kind: schemamodel.KindArray}}},
		Rows:     []document.Row{{ID: "item", Values: map[string]document.Cell{}}},
		Formulas: []document.Formula{{RowID: "item", Column: "tags", Expression: "[]"}},
	}

	result := Evaluate(table)
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Message != "formulas only support scalar or enum columns" {
		t.Fatalf("diagnostics = %#v", result.Diagnostics)
	}
}
