package functional

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"unicode"

	"datapack/internal/document"
	"datapack/internal/workspace"
)

var yunlanCounts = map[string]int{"localization.text": 480, "content.item": 72, "content.effect": 36, "content.skill": 48, "content.hero": 12, "balance.growth": 60, "world.enemy": 24, "world.loot_group": 12, "world.loot_entry": 36, "world.stage": 24, "quest.quest": 48, "craft.recipe": 36, "craft.ingredient": 72, "economy.shop": 12, "economy.product": 24, "system.parameter": 6}

func yunlanPath(root, logical string) string {
	parts := strings.Split(logical, ".")
	return filepath.Join(root, "tables", parts[0], parts[1], parts[1]+".cue")
}
func sourceHashes(t *testing.T, root string) map[string][32]byte {
	t.Helper()
	result := map[string][32]byte{}
	err := filepath.WalkDir(root, func(path string, e fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if e.IsDir() {
			if e.Name() == "exports" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		result[rel] = sha256.Sum256(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}
func yunlanCSV(t *testing.T, root, logical string) map[string]map[string]string {
	t.Helper()
	records, err := csv.NewReader(strings.NewReader(readFile(t, filepath.Join(root, "exports", logical+".csv")))).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	result := map[string]map[string]string{}
	for _, record := range records[1:] {
		row := map[string]string{}
		for i, header := range records[0] {
			row[header] = record[i]
		}
		result[record[0]] = row
	}
	return result
}
func TestYunlanProjectCompilesAndExportsGoldenCSV(t *testing.T) {
	s, root := openExample(t, "yunlan-rpg")
	view, _ := s.Current()
	requireCleanCompile(t, view)
	before := sourceHashes(t, root)
	count, comments := 0, 0
	for logical, want := range yunlanCounts {
		table, err := document.LoadTable(yunlanPath(root, logical))
		if err != nil {
			t.Fatal(err)
		}
		if len(table.Rows) != want {
			t.Fatalf("%s rows=%d want%d", logical, len(table.Rows), want)
		}
		count += len(table.Rows)
		for _, columns := range table.CellComments {
			comments += len(columns)
		}
		for _, column := range table.Columns {
			if !strings.ContainsFunc(column.Name, func(r rune) bool { return unicode.Is(unicode.Han, r) }) {
				t.Fatalf("non-Chinese business field %s.%s", logical, column.Name)
			}
			if column.DisplayName != "" {
				t.Fatalf("redundant display alias %s.%s", logical, column.Name)
			}
		}
	}
	if count != 1002 || comments < 20 {
		t.Fatalf("rows=%d comments=%d", count, comments)
	}
	result, err := s.Export()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Files) != 15 {
		t.Fatalf("exported %d files, want15", len(result.Files))
	}
	for _, path := range result.Files {
		name := filepath.Base(path)
		if name == "system.parameter.csv" {
			t.Fatal("excluded parameters exported")
		}
		if got, want := readFile(t, path), readFile(t, goldenPath(t, "yunlan-rpg", name)); got != want {
			t.Fatalf("golden mismatch: %s", name)
		}
	}
	if !reflect.DeepEqual(before, sourceHashes(t, root)) {
		t.Fatal("compile/export rewrote example source")
	}
	product := yunlanCSV(t, root, "economy.product")
	if product["product_01_1"]["实售价"] != "32" {
		t.Fatal(product["product_01_1"])
	}
	if _, exported := product["product_01_1"]["策划备注"]; exported {
		t.Fatal("planning notes leaked to CSV")
	}
	growth := yunlanCSV(t, root, "balance.growth")
	if growth["level_01"]["参考战力"] != "115" || growth["level_02"]["累计经验"] != "103" {
		t.Fatal(growth["level_01"], growth["level_02"])
	}
	cumulative := 0
	for level := 1; level < 60; level++ {
		cumulative += 80 + 20*level + 3*level*level
	}
	if growth["level_60"]["累计经验"] != strconv.Itoa(cumulative) {
		t.Fatal("cross-row cumulative curve changed")
	}
	hero := yunlanCSV(t, root, "content.hero")["hero_01"]
	var skills []string
	if err := json.Unmarshal([]byte(hero["技能ID"]), &skills); err != nil || !reflect.DeepEqual(skills, []string{"skill_01", "skill_02", "skill_03", "skill_04"}) {
		t.Fatalf("skill JSON=%s, %v", hero["技能ID"], err)
	}
	var point map[string]float64
	if err := json.Unmarshal([]byte(hero["出生坐标"]), &point); err != nil || point["横"] != 0 || point["纵"] != 0 {
		t.Fatal("Chinese struct JSON changed", err)
	}
	text := yunlanCSV(t, root, "localization.text")
	if text[hero["名称文本ID"]]["简体中文"] != "陆青禾" {
		t.Fatal("hero-to-localization link broken")
	}
	skill := yunlanCSV(t, root, "content.skill")[skills[0]]
	if skill["主效果ID"] != "effect_01" || yunlanCSV(t, root, "content.effect")[skill["主效果ID"]]["类别"] != "伤害" {
		t.Fatal("hero-skill-effect chain broken")
	}
	item := yunlanCSV(t, root, "content.item")
	if item["item_material_01"]["效果ID"] != "effect_10" || item["item_potion_01"]["可交易"] != "true" {
		t.Fatal("ordinary defaults did not resolve")
	}
	if item["item_potion_01"]["装备槽位"] != "" || item["item_potion_01"]["额外效果ID"] != "[]" {
		t.Fatal("nullable/empty-array semantics changed")
	}
	parameters, err := s.GetTable(yunlanPath(root, "system.parameter"))
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range parameters.Rows {
		if row.ID == "level_cap" && row.Values["数值"].Display != "60" {
			t.Fatal("parameter cap and growth cap diverged")
		}
	}
	weights := map[string]float64{}
	for _, drop := range yunlanCSV(t, root, "world.loot_entry") {
		v, err := strconv.ParseFloat(drop["权重"], 64)
		if err != nil {
			t.Fatal(err)
		}
		weights[drop["掉落组ID"]] += v
	}
	for group, sum := range weights {
		if math.Abs(sum-1) > 1e-9 {
			t.Fatalf("drop group %s weight=%v", group, sum)
		}
	}
	materials := map[string]int{}
	for _, row := range yunlanCSV(t, root, "craft.ingredient") {
		materials[row["配方ID"]]++
		if item[row["材料ID"]]["类别"] != "材料" {
			t.Fatal("recipe ingredient is not a material")
		}
	}
	for rid, row := range yunlanCSV(t, root, "craft.recipe") {
		if materials[rid] != 2 || item[row["产物ID"]]["装备槽位"] == "" {
			t.Fatalf("recipe %s no longer has two materials and equipment", rid)
		}
	}
}
func TestYunlanChineseEditsPasteRenameSchemaAndHistory(t *testing.T) {
	s, root := openExample(t, "yunlan-rpg")
	original := sourceHashes(t, root)
	product := yunlanPath(root, "economy.product")
	if _, err := s.Paste(workspace.Paste{Path: product, RowIDs: []string{"product_01_1", "product_01_2"}, Columns: []string{"标价"}, Values: [][]string{{"100"}, {"200"}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RenameColumn(workspace.ColumnRename{Path: product, OldName: "标价", NewName: "售价"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Export(); err != nil {
		v, _ := s.Current()
		t.Fatalf("%v\n%s", err, formatDiagnostics(v.Compile.Diagnostics))
	}
	records := yunlanCSV(t, root, "economy.product")
	if records["product_01_1"]["实售价"] != "80" || records["product_01_2"]["实售价"] != "200" {
		t.Fatal("paste/Chinese rename changed formula semantics")
	}
	after := sourceHashes(t, root)
	if _, err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Redo(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(after, sourceHashes(t, root)) {
		t.Fatal("renamed Chinese formulas/comments did not redo exactly")
	}
	if _, err := s.RenameSchemaDefinition(workspace.SchemaDefinitionRename{Path: filepath.Join(root, "schema", "common.cue"), OldName: "必填文本", NewName: "非空文字"}); err != nil {
		t.Fatal(err)
	}
	view, _ := s.Current()
	requireCleanCompile(t, view)
	if !strings.Contains(readFile(t, yunlanPath(root, "localization.text")), "common.#非空文字") {
		t.Fatal("Chinese schema usage not migrated")
	}
	if _, err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(original, sourceHashes(t, root)) {
		t.Fatal("Undo schema/column/paste did not restore source")
	}
	// Fixed literal references migrate; default-created references refuse safely.
	item := yunlanPath(root, "content.item")
	if _, err := s.RenameRow(workspace.RowRename{Path: item, OldID: "item_weapon_01", NewID: "item_qinglan_blade"}); err != nil {
		t.Fatal(err)
	}
	view, _ = s.Current()
	requireCleanCompile(t, view)
	if _, err := s.DeleteRow(workspace.RowDelete{Path: item, RowID: "item_qinglan_blade"}); err == nil {
		t.Fatal("referenced equipment deletion accepted")
	}
	before := sourceHashes(t, root)
	viewBefore, _ := s.Current()
	effect := yunlanPath(root, "content.effect")
	if _, err := s.RenameRow(workspace.RowRename{Path: effect, OldID: "effect_10", NewID: "effect_guard"}); err == nil {
		t.Fatal("shared/default reference rename accepted")
	}
	viewAfter, _ := s.Current()
	if !reflect.DeepEqual(before, sourceHashes(t, root)) || !reflect.DeepEqual(viewBefore, viewAfter) {
		t.Fatal("rejected rename changed source/view/history")
	}
}
func TestYunlanExcludedTableErrorStillBlocksExport(t *testing.T) {
	s, root := openExample(t, "yunlan-rpg")
	if _, err := s.Export(); err != nil {
		t.Fatal(err)
	}
	good := readFile(t, filepath.Join(root, "exports", "economy.product.csv"))
	parameter := yunlanPath(root, "system.parameter")
	if _, err := s.UpdateCell(workspace.CellUpdate{Path: parameter, RowID: "level_cap", Column: "数值", Expression: "-1"}); err != nil {
		t.Fatal(err)
	}
	view, _ := s.Current()
	if len(view.Compile.Diagnostics) == 0 {
		t.Fatal("excluded table not validated")
	}
	if _, err := s.Export(); err == nil {
		t.Fatal("excluded invalid table allowed export")
	}
	if readFile(t, filepath.Join(root, "exports", "economy.product.csv")) != good {
		t.Fatal("failed compile replaced valid output")
	}
	if _, err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Export(); err != nil {
		t.Fatalf("cannot recover after Undo: %v", err)
	}
}
