package workspace

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"datapack/internal/foreignkey"
	"datapack/internal/navigation"
)

func assertReferenceEditRejected(t *testing.T, s *Service, target string) {
	t.Helper()
	// Keep both stacks populated to detect accidental history replacement,
	// not merely a no-op with initially empty history.
	for _, name := range []string{"Hot", "Hotter"} {
		if _, err := s.UpdateCell(CellUpdate{Path: target, RowID: "burn", Column: "name", Expression: strconv.Quote(name)}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	before := fileTree(t, s.current.Project.Root)
	history := func() string {
		data, err := json.Marshal([]any{s.undoStack, s.redoStack})
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	beforeHistory := history()
	for _, operation := range []func() error{
		func() error { _, err := s.DeleteRow(RowDelete{Path: target, RowID: "burn"}); return err },
		func() error {
			_, err := s.RenameRow(RowRename{Path: target, OldID: "burn", NewID: "flame"})
			return err
		},
	} {
		if err := operation(); err == nil {
			t.Fatal("unsafe reference edit succeeded")
		}
		if !reflect.DeepEqual(before, fileTree(t, s.current.Project.Root)) || history() != beforeHistory {
			t.Fatal("rejected operation changed files or history")
		}
	}
}

func TestDefaultForeignKeyProtectsDeletionAndRenameWithoutWrites(t *testing.T) {
	for _, schema := range []string{`{effectID: *"burn" | string}`, `{effectID: *["burn"] | [...string]}`} {
		t.Run(schema, func(t *testing.T) {
			s, target, ref := renameFixture(t)
			writeFile(t, ref, "package item\n#PrimaryKey: string\n#Row: "+schema+"\nrows: {sword:{}}\n_tool:{table:{namespace:\"rpg\",name:\"item\"},foreignKeys:{effectID:{table:\"rpg.effect\"}}}")
			// Remove other literal references so only the default protects the target.
			writeFile(t, filepath.Join(s.current.Project.Root, "tables", "rpg", "skill", "skill.cue"), `package skill
#PrimaryKey:string
#Row:{name:string}
rows:{}
_tool:{table:{namespace:"rpg",name:"skill"}}`)
			if _, err := s.OpenProject(s.current.Project.Path); err != nil {
				t.Fatal(err)
			}
			assertReferenceEditRejected(t, s, target)
		})
	}
}

func TestImportedSharedDefaultProtectsReferencesWithoutSchemaWrites(t *testing.T) {
	for _, array := range []bool{false, true} {
		t.Run(map[bool]string{false: "scalar", true: "array"}[array], func(t *testing.T) {
			s, target, ref := renameFixture(t)
			root := s.current.Project.Root
			writeFile(t, filepath.Join(root, "cue.mod", "module.cue"), "module:\"example.com/game\"\nlanguage:{version:\"v0.9.0\"}")
			definition := `*"burn" | string`
			kind := "string"
			if array {
				definition = `*["burn","burn"] | [...string]`
				kind = "array"
			}
			writeFile(t, filepath.Join(root, "schema", "common.cue"), "package common\n#EffectID: "+definition+"\n_schema:{definitions:{EffectID:{kind:\"type\",type:"+strconv.Quote(kind)+"}}}")
			for _, path := range []string{ref, filepath.Join(root, "tables", "rpg", "skill", "skill.cue")} {
				name := strings.TrimSuffix(filepath.Base(path), ".cue")
				writeFile(t, path, "package "+name+"\nimport \"example.com/game/schema:common\"\n#PrimaryKey:string\n#Row:{effectID:common.#EffectID}\nrows:{one:{}}\n_tool:{table:{namespace:\"rpg\",name:"+strconv.Quote(name)+"},foreignKeys:{effectID:{table:\"rpg.effect\"}}}")
			}
			if _, err := s.OpenProject(s.current.Project.Path); err != nil {
				t.Fatal(err)
			}
			nav, err := s.Navigate(navigation.Request{Kind: navigation.KindPrimaryKey, Path: target, RowID: "burn"})
			if err != nil || len(nav.References) != map[bool]int{false: 2, true: 4}[array] {
				t.Fatalf("reverse usages=%#v err=%v", nav, err)
			}
			assertReferenceEditRejected(t, s, target)
		})
	}
}

func TestUnresolvedReferenceSourcesBlockEditsWithoutWrites(t *testing.T) {
	for _, mode := range []string{"damaged on open", "damaged after open", "invalid fixed input", "invalid array item"} {
		t.Run(mode, func(t *testing.T) {
			s, target, ref := renameFixture(t)
			source := `package item
#PrimaryKey:string
#Row:{effectID:string}
rows:{one:{effectID:"burn"}}
_tool:{table:{namespace:"rpg",name:"item"},foreignKeys:{effectID:{table:"rpg.effect"}}}`
			if strings.HasPrefix(mode, "damaged") {
				source = "package item\n#Row: ("
			}
			if mode == "invalid fixed input" {
				source = strings.Replace(source, `effectID:"burn"`, `effectID:123`, 1)
			}
			if mode == "invalid array item" {
				source = strings.ReplaceAll(source, "effectID:string", "effectID:[...string]")
				source = strings.Replace(source, `effectID:"burn"`, `effectID:["burn",123]`, 1)
			}
			writeFile(t, ref, source)
			if mode != "damaged after open" {
				if _, err := s.OpenProject(s.current.Project.Path); err != nil {
					t.Fatal(err)
				}
			}
			assertReferenceEditRejected(t, s, target)
		})
	}
}

func TestDerivedReferenceIsProtectedButNeverRewrittenAsLiteral(t *testing.T) {
	s, target, ref := renameFixture(t)
	writeFile(t, ref, `package item
#PrimaryKey:string
#Row:{effectID:string}
rows:{one:{effectID:("burn")}}
_tool:{table:{namespace:"rpg",name:"item"},foreignKeys:{effectID:{table:"rpg.effect"}}}`)
	if _, err := s.OpenProject(s.current.Project.Path); err != nil {
		t.Fatal(err)
	}
	index := buildForeignKeyIndex(s.current.Tree)
	if index.Incomplete {
		t.Fatal(index.Diagnostics)
	}
	found := false
	for _, usage := range index.Usages[foreignkey.Target{Table: "rpg.effect", RowID: "burn"}] {
		if usage.Path == ref {
			found = true
			if usage.Source != foreignkey.SourceDerived {
				t.Fatalf("source=%#v", usage)
			}
		}
	}
	if !found {
		t.Fatal("derived reference was not indexed")
	}
	assertReferenceEditRejected(t, s, target)
}

func TestNullableArrayLiteralRenameUsesOriginalIndices(t *testing.T) {
	s, target, ref := renameFixture(t)
	writeFile(t, ref, `package item
#PrimaryKey:string
#Row:{effectIDs:[...(null|string)]}
rows:{one:{effectIDs:[null,"burn",null,"burn"]}}
_tool:{table:{namespace:"rpg",name:"item"},foreignKeys:{effectIDs:{table:"rpg.effect"}}}`)
	if _, err := s.OpenProject(s.current.Project.Path); err != nil {
		t.Fatal(err)
	}
	before := fileTree(t, s.current.Project.Root)
	if _, err := s.RenameRow(RowRename{Path: target, OldID: "burn", NewID: "flame"}); err != nil {
		t.Fatal(err)
	}
	table, err := s.GetTable(ref)
	if err != nil {
		t.Fatal(err)
	}
	if got := table.Rows[0].Values["effectIDs"].Expression; got != `[null, "flame", null, "flame"]` {
		t.Fatal(got)
	}
	after := fileTree(t, s.current.Project.Root)
	if _, err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	assertFileTree(t, s.current.Project.Root, before)
	if _, err := s.Redo(); err != nil {
		t.Fatal(err)
	}
	assertFileTree(t, s.current.Project.Root, after)
}
