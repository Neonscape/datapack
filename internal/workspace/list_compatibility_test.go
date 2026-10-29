package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMultilineExportArrayRenameRespectsModuleVersion(t *testing.T) {
	for _, column := range []string{"value", "数值"} {
		t.Run(column, func(t *testing.T) {
			s, _, table := renameFixture(t)
			writeFile(t, filepath.Join(s.current.Project.Root, "cue.mod", "module.cue"), `module:"example.com/array-test"
language:{version:"v0.9.0"}`)
			writeFile(t, table, fmt.Sprintf(`package item
#PrimaryKey:string
#Row:{name:string,%s:int,tags:[...string]}
rows:{one:{name:"one",%s:1,tags:["a","b"]}}
_tool:{table:{namespace:"rpg",name:"item"}}
_export:{csv:{enabled:true,columns:[
"id",
"name",
%q,
"tags",
],encoding:{tags:"json"}}}
`, column, column, column))
			project := s.current.Project
			project.Config.Export.Tables = map[string]bool{"rpg.item": true, "rpg.effect": false, "rpg.skill": false}
			if err := SaveProject(project); err != nil {
				t.Fatal(err)
			}
			if _, err := s.OpenProject(s.current.Project.Path); err != nil {
				t.Fatal(err)
			}
			if len(s.current.Compile.Diagnostics) != 0 {
				t.Fatalf("initial diagnostics: %#v", s.current.Compile.Diagnostics)
			}
			before, err := os.ReadFile(table)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := s.RenameColumn(ColumnRename{Path: table, OldName: column, NewName: "new_value"}); err != nil {
				t.Fatal(err)
			}
			if len(s.current.Compile.Diagnostics) != 0 {
				t.Fatalf("rename invalidated old module: %#v", s.current.Compile.Diagnostics)
			}
			if _, err := s.Export(); err != nil {
				t.Fatal(err)
			}
			after, err := os.ReadFile(table)
			if err != nil {
				t.Fatal(err)
			}
			reopened := NewService(t.TempDir())
			v, err := reopened.OpenProject(s.current.Project.Path)
			if err != nil || len(v.Compile.Diagnostics) != 0 {
				t.Fatalf("reopen: %v %#v", err, v.Compile.Diagnostics)
			}
			if _, err := s.Undo(); err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile(table)
			if err != nil || string(got) != string(before) {
				t.Fatal("Undo did not restore source", err)
			}
			if _, err := s.Redo(); err != nil {
				t.Fatal(err)
			}
			got, err = os.ReadFile(table)
			if err != nil || string(got) != string(after) {
				t.Fatal("Redo did not restore source", err)
			}
			module, err := os.ReadFile(filepath.Join(s.current.Project.Root, "cue.mod", "module.cue"))
			if err != nil || !strings.Contains(string(module), "v0.9.0") {
				t.Fatal("edit changed module language", err)
			}
		})
	}
}

func TestMultilineForeignKeyArrayRenameRespectsModuleVersion(t *testing.T) {
	s, target, ref := renameFixture(t)
	writeFile(t, filepath.Join(s.current.Project.Root, "cue.mod", "module.cue"), `module:"example.com/array-fk-test"
language:{version:"v0.9.0"}`)
	source, err := os.ReadFile(ref)
	if err != nil {
		t.Fatal(err)
	}
	source = []byte(strings.Replace(string(source), `["burn", "burn"]`, "[\n\"burn\",\n\"burn\",\n]", 1))
	if err := os.WriteFile(ref, source, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.OpenProject(s.current.Project.Path); err != nil {
		t.Fatal(err)
	}
	before := fileTree(t, s.current.Project.Root)
	if _, err := s.RenameRow(RowRename{Path: target, OldID: "burn", NewID: "flame"}); err != nil {
		t.Fatal(err)
	}
	if len(s.current.Compile.Diagnostics) != 0 {
		t.Fatalf("array FK rename broke compilation: %#v", s.current.Compile.Diagnostics)
	}
	reopened := NewService(t.TempDir())
	v, err := reopened.OpenProject(s.current.Project.Path)
	if err != nil || len(v.Compile.Diagnostics) != 0 {
		t.Fatalf("reopen: %v %#v", err, v.Compile.Diagnostics)
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
