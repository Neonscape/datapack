package functional

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExampleProjectsCompileAndExportGoldenCSV(t *testing.T) {
	cases := []struct {
		example string
		files   []string
	}{
		{example: "basic-rpg-auto", files: []string{"rpg.effect.csv", "rpg.item.csv"}},
		{example: "character-growth-manual", files: []string{"combat.growth.csv"}},
		{example: "schema-composition", files: []string{"inventory.weapon.csv"}},
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
