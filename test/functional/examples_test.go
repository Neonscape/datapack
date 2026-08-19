package functional

import (
	"os"
	"path/filepath"
	"testing"

	"datapack/internal/navigation"
	"datapack/internal/workspace"
)

func TestExampleProjectsCompileAndExportGoldenCSV(t *testing.T) {
	cases := []struct {
		example string
		files   []string
	}{
		{example: "basic-rpg-auto", files: []string{"rpg.effect.csv", "rpg.item.csv"}},
		{example: "character-growth-manual", files: []string{"combat.growth.csv"}},
		{example: "schema-composition", files: []string{"inventory.weapon.csv"}},
		{example: "skyharbor-content", files: []string{"content.effect.csv", "content.item.csv", "content.character.csv", "balance.growth.csv"}},
	}
	for _, test := range cases {
		t.Run(test.example, func(t *testing.T) {
			service, root := openExample(t, test.example)
			view, ok := service.Current()
			if !ok {
				t.Fatal("Current() returned no workspace")
			}
			requireCleanCompile(t, view)

			result, err := service.Export()
			if err != nil {
				t.Fatalf("Export(): %v", err)
			}
			if result.Directory != filepath.Join(root, "exports") {
				t.Fatalf("export directory = %q", result.Directory)
			}
			if len(result.Files) != len(test.files) {
				t.Fatalf("exported files = %#v, want %v", result.Files, test.files)
			}
			for _, name := range test.files {
				got := readFile(t, filepath.Join(result.Directory, name))
				want := readFile(t, goldenPath(t, test.example, name))
				if got != want {
					t.Fatalf("%s\ngot:\n%s\nwant:\n%s", name, got, want)
				}
			}
		})
	}
}

func TestExampleExportDoesNotRewriteSourceFiles(t *testing.T) {
	service, root := openExample(t, "character-growth-manual")
	before := readFile(t, filepath.Join(root, "tables", "growth", "growth.cue"))
	if _, err := service.Export(); err != nil {
		t.Fatal(err)
	}
	after := readFile(t, filepath.Join(root, "tables", "growth", "growth.cue"))
	if before != after {
		t.Fatal("Export() rewrote the table source")
	}
	if _, err := os.Stat(filepath.Join(root, "exports", "combat.growth.csv")); err != nil {
		t.Fatalf("exported CSV missing: %v", err)
	}
}

func TestExampleGetTableShowsFormulaPreview(t *testing.T) {
	service, root := openExample(t, "character-growth-manual")
	table, err := service.GetTable(filepath.Join(root, "tables", "growth", "growth.cue"))
	if err != nil {
		t.Fatal(err)
	}
	row := map[string]string{}
	for _, candidate := range table.Rows {
		if candidate.ID != "level_11" {
			continue
		}
		for name, cell := range candidate.Values {
			row[name] = cell.Display
		}
	}
	if row["baseAttack"] != "25" || row["attack"] != "45" {
		t.Fatalf("level_11 preview = %#v, want baseAttack 25 attack 45", row)
	}
	if len(table.Formulas) != 3 {
		t.Fatalf("formulas = %#v", table.Formulas)
	}
}

func TestSkyharborExampleIndexesEveryLearningConcept(t *testing.T) {
	service, root := openExample(t, "skyharbor-content")
	view, ok := service.Current()
	if !ok {
		t.Fatal("Current() returned no workspace")
	}
	requireCleanCompile(t, view)

	for _, name := range []string{"effect", "item", "character", "growth"} {
		if findNode(view.Tree.Root, name) == nil {
			t.Fatalf("table %s is missing from project tree", name)
		}
	}
	for _, definition := range []string{"common.#GameID", "common.#Rarity", "common.#NonNegative", "content.item._local.#Position", "content.character._local.#GrowthRate", "balance.growth._local.#Level"} {
		found := false
		for _, entry := range view.Schemas.Definitions {
			if string(entry.ID) == definition {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Schema definition %s is missing from registry", definition)
		}
	}

	itemPath := filepath.Join(root, "tables", "content", "item", "item.cue")
	item, err := service.GetTable(itemPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(item.ForeignKeys) != 2 || len(item.Formulas) != 2 || item.CSVExport == nil {
		t.Fatalf("item metadata = foreign keys %d, formulas %d, csv %#v", len(item.ForeignKeys), len(item.Formulas), item.CSVExport)
	}
	if len(item.LocalSchema) != 2 {
		t.Fatalf("item local schemas = %#v", item.LocalSchema)
	}
	if item.Columns[4].Schema.Kind != "array" || item.Columns[5].Schema.Kind != "reference" {
		t.Fatalf("item structured columns = %#v, %#v", item.Columns[4].Schema, item.Columns[5].Schema)
	}
	if target, err := service.Navigate(navigation.Request{Kind: navigation.KindForeignKey, Path: itemPath, RowID: "burn", Column: "effectID"}); err != nil || target.Path == "" || target.RowID != "burn" {
		t.Fatalf("foreign-key navigation = %#v, %v", target, err)
	}
	if target, err := service.Navigate(navigation.Request{Kind: navigation.KindSchemaReference, Path: itemPath, Reference: "_local.#Position"}); err != nil || target.Definition != "Position" || !target.Local {
		t.Fatalf("local Schema navigation = %#v, %v", target, err)
	}
	characterPath := filepath.Join(root, "tables", "content", "character", "character.cue")
	character, err := service.GetTable(characterPath)
	if err != nil {
		t.Fatal(err)
	}
	if character.Rows[0].Values["attack"].Display != "64" || !character.Rows[0].Values["attack"].Formula {
		t.Fatalf("character formula preview = %#v", character.Rows[0].Values["attack"])
	}

	if _, err := service.SetCellComment(workspace.CellCommentSet{Path: characterPath, RowID: "captain_lyra", Column: "attack", Comment: "computed from baseAttack"}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Undo(); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Redo(); err != nil {
		t.Fatal(err)
	}
	updated, err := service.GetTable(characterPath)
	if err != nil || updated.CellComments["captain_lyra"]["attack"] != "computed from baseAttack" {
		t.Fatalf("comment after redo = %#v, %v", updated.CellComments, err)
	}
}

func TestSkyharborExampleSupportsSourceBackedLearningEdits(t *testing.T) {
	service, root := openExample(t, "skyharbor-content")
	effectPath := filepath.Join(root, "tables", "content", "effect", "effect.cue")
	if _, err := service.SetCompileMode("manual"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateRow(workspace.RowCreate{Path: effectPath, RowID: "shock", Values: map[string]string{"name": `"Shock"`, "kind": `"damage"`, "power": "12", "chance": "0.5", "duration": "1", "description": `"A short arc."`, "tags": `["electric"]`}}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetCellText(workspace.CellUpdate{Path: effectPath, RowID: "shock", Column: "description", Expression: "A multiline\ndescription"}); err != nil {
		t.Fatal(err)
	}
	view, ok := service.Current()
	if !ok || !view.Compile.Pending {
		t.Fatalf("manual edit state = %#v", view.Compile)
	}
	if _, err := service.Compile(); err != nil {
		t.Fatal(err)
	}
	view, _ = service.Current()
	requireCleanCompile(t, view)

	effect, err := service.GetTable(effectPath)
	if err != nil {
		t.Fatal(err)
	}
	if effect.Rows[len(effect.Rows)-1].ID != "shock" || effect.Rows[len(effect.Rows)-1].Values["description"].Display != "A multiline\ndescription" {
		t.Fatalf("created row = %#v", effect.Rows[len(effect.Rows)-1])
	}
	if _, err := service.DeleteRow(workspace.RowDelete{Path: effectPath, RowID: "shock"}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Undo(); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Redo(); err != nil {
		t.Fatal(err)
	}
}
