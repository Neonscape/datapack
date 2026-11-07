// A reproducible real-Service scenario. Run only against a disposable copy;
// the Python runner supplies an empty output directory and isolated XDG roots.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"datapack/internal/workspace"
)

var root, phase string
var events []map[string]any

func check(err error) {
	if err != nil {
		panic(err)
	}
}
func require(ok bool, message string) {
	if !ok {
		panic(message)
	}
}
func record(name string, value any) {
	events = append(events, map[string]any{"check": name, "result": value})
	fmt.Println("PASS", name)
}
func table(base, name string) string {
	p := strings.Split(name, ".")
	return filepath.Join(base, "tables", p[0], p[1], p[1]+".cue")
}
func hashes(base string) map[string]string {
	out := map[string]string{}
	check(filepath.WalkDir(base, func(path string, e os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if e.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(base, path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		out[rel] = hex.EncodeToString(sum[:])
		return nil
	}))
	return out
}
func save(name string, data any) {
	b, err := json.MarshalIndent(data, "", "  ")
	check(err)
	check(os.WriteFile(filepath.Join(root, name), append(b, '\n'), 0600))
}
func compile(s *workspace.Service) {
	r, err := s.Compile()
	check(err)
	require(len(r.Workspace.Compile.Diagnostics) == 0, fmt.Sprintf("compile diagnostics: %+v", r.Workspace.Compile.Diagnostics))
}
func literal(s *workspace.Service, base, name, row, col, expr string) {
	_, err := s.UpdateCell(workspace.CellUpdate{Path: table(base, name), RowID: row, Column: col, Expression: expr})
	check(err)
}
func display(s *workspace.Service, base, name, row, col string) string {
	t, err := s.GetTable(table(base, name))
	check(err)
	for _, r := range t.Rows {
		if r.ID == row {
			return r.Values[col].Display
		}
	}
	panic("row not found: " + row)
}
func edit() {
	base := filepath.Join(root, "project")
	s := workspace.NewService(filepath.Join(root, "settings-edit"))
	_, err := s.OpenProject(filepath.Join(base, "yunlan-rpg.dtproj"))
	check(err)
	compile(s)
	record("copied Chinese project opens and compiles", nil)
	for _, v := range []string{"100", "125", "140"} {
		literal(s, base, "economy.product", "product_01_1", "list_price", v)
	}
	require(display(s, base, "economy.product", "product_01_1", "list_price") == "140", "continuous edits lost")
	_, err = s.Undo()
	check(err)
	require(display(s, base, "economy.product", "product_01_1", "list_price") == "125", "undo latest edit")
	_, err = s.Redo()
	check(err)
	require(display(s, base, "economy.product", "product_01_1", "list_price") == "140", "redo latest edit")
	record("continuous edits and one-step undo/redo", nil)
	before := hashes(base)
	_, err = s.Paste(workspace.Paste{Path: table(base, "economy.product"), RowIDs: []string{"product_01_1", "product_01_2"}, Columns: []string{"list_price"}, Values: [][]string{{"100"}, {"200"}}})
	check(err)
	after := hashes(base)
	_, err = s.Undo()
	check(err)
	require(reflect.DeepEqual(before, hashes(base)), "paste undo changed bytes")
	_, err = s.Redo()
	check(err)
	require(reflect.DeepEqual(after, hashes(base)), "paste redo changed bytes")
	record("Service batch paste is a single reversible disk edit (no system clipboard)", nil)
	_, err = s.SetCellText(workspace.CellUpdate{Path: table(base, "localization.text"), RowID: "hero_01_name", Column: "zh_cn", Expression: "陆青禾，\"归来\"\n第二行\t终"})
	check(err)
	literal(s, base, "content.hero", "hero_01", "skill_ids", `["skill_02", "skill_01", "skill_03", "skill_04"]`)
	literal(s, base, "content.hero", "hero_01", "spawn_position", `{x: -1.25, y: 2.5}`)
	literal(s, base, "economy.product", "product_01_1", "listed", "false")
	compile(s)
	require(display(s, base, "economy.product", "product_01_1", "sale_price") == "80", "price formula should be 80")
	record("Chinese multiline text, array, struct, boolean and formula edits", nil)
	_, err = s.RenameRow(workspace.RowRename{Path: table(base, "content.item"), OldID: "item_weapon_01", NewID: "item_qinglan_blade"})
	check(err)
	compile(s)
	require(display(s, base, "content.hero", "hero_01", "starting_equipment_id") == "item_qinglan_blade", "hero FK did not rename")
	require(display(s, base, "economy.product", "product_01_2", "item_id") == "item_qinglan_blade", "product FK did not rename")
	require(display(s, base, "craft.recipe", "recipe_01", "output_item_id") == "item_qinglan_blade", "recipe FK did not rename")
	record("row rename migrates dependent tables", nil)
	_, err = s.SaveProjectAs(filepath.Join(root, "saved", "session.dtproj"))
	check(err)
	record("SaveProjectAs writes independent source copy", nil)
	save("edit-events.json", events)
	save("saved-before-reopen-hashes.json", hashes(filepath.Join(root, "saved")))
}
func reopen() {
	base := filepath.Join(root, "saved")
	s := workspace.NewService(filepath.Join(root, "settings-reopen"))
	view, err := s.OpenProject(filepath.Join(base, "session.dtproj"))
	check(err)
	require(!view.CanUndo && !view.CanRedo, "new process should not inherit history")
	before := hashes(base)
	compile(s)
	require(reflect.DeepEqual(before, hashes(base)), "reopen/compile rewrote source")
	require(display(s, base, "economy.product", "product_01_1", "sale_price") == "80", "reopened formula")
	require(display(s, base, "localization.text", "hero_01_name", "zh_cn") == "陆青禾，\"归来\"\n第二行\t终", "reopened text")
	record("fresh process reopens saved sources with empty history", nil)
	exported, err := s.Export()
	check(err)
	require(len(exported.Files) == 15, "expected 15 CSVs")
	old := hashes(exported.Directory)
	save("export-before-failures-hashes.json", old)
	record("clean compile exports exactly 15 configured tables", nil)
	literal(s, base, "system.parameter", "level_cap", "value", "-1")
	_, err = s.Export()
	require(err != nil, "invalid excluded table should block export")
	v, _ := s.Current()
	found := false
	for _, d := range v.Compile.Diagnostics {
		if d.RowID == "level_cap" && d.Column == "value" {
			found = true
		}
	}
	require(found, "diagnostic must identify level_cap.value")
	require(reflect.DeepEqual(old, hashes(exported.Directory)), "failed validation changed prior exports")
	save("blocked-diagnostics.json", v.Compile.Diagnostics)
	record("invalid unexported table blocks export and preserves old CSV", err.Error())
	_, err = s.Undo()
	check(err)
	compile(s)
	// A real filesystem permission failure, using only this disposable export directory.
	check(os.Chmod(exported.Directory, 0500))
	_, exportErr := s.Export()
	check(os.Chmod(exported.Directory, 0755))
	require(exportErr != nil, "expected export staging/rename permission failure")
	require(reflect.DeepEqual(old, hashes(exported.Directory)), "permission failure damaged exports")
	record("export permission failure preserves every old CSV byte", exportErr.Error())
	_, err = s.Export()
	check(err)
	require(reflect.DeepEqual(old, hashes(exported.Directory)), "retry produced changed CSV")
	record("export recovers after permissions restored", nil)
	save("reopen-events.json", events)
	save("final-export-hashes.json", hashes(exported.Directory))
}
func main() {
	flag.StringVar(&root, "root", "", "isolated scenario directory")
	flag.StringVar(&phase, "phase", "", "edit or reopen")
	flag.Parse()
	require(filepath.IsAbs(root), "root must be absolute")
	switch phase {
	case "edit":
		edit()
	case "reopen":
		reopen()
	default:
		panic("unknown phase")
	}
}
