package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"datapack/internal/document"
)

func sourceEditFixture(t *testing.T) (*Service, string, string) {
	t.Helper()
	root := t.TempDir()
	project := filepath.Join(root, "game.dtproj")
	writeFile(t, project, `{"version":1,"tableRoots":["tables"],"schemaRoots":["schema"],"export":{"directory":"exports","targets":["csv"]}}`)
	writeFile(t, filepath.Join(root, "cue.mod", "module.cue"), `module:"example.com/game"
language:{version:"v0.9.0"}`)
	writeFile(t, filepath.Join(root, "schema", "common", "common.cue"), `package common
#Positive:int & >=0
_schema:{definitions:{Positive:{kind:"type",type:"integer"}}}`)
	rules := filepath.Join(root, "schema", "rules", "rules.cue")
	writeFile(t, rules, `package rules
#Existing:int
_schema:{definitions:{Existing:{kind:"type",type:"integer"}}}`)
	table := filepath.Join(root, "tables", "item", "item.cue")
	writeFile(t, table, `package item
// Keep the original source and metadata on every rejected operation.
#PrimaryKey:string
#Row:{name:string,price:int}
rows:{one:{name:"Original",price:10}}
_local:{#Existing:int}
_schema:{localDefinitions:{Existing:{kind:"type",type:"integer"}}}
_tool:{table:{namespace:"rpg",name:"item"}}`)
	s := NewService(t.TempDir())
	if _, err := s.OpenProject(project); err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{`"First"`, `"Second"`} {
		if _, err := s.UpdateCell(CellUpdate{Path: table, RowID: "one", Column: "name", Expression: value}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	return s, table, rules
}

func sourceEditState(s *Service) string {
	return fmt.Sprintf("%#v|%#v|%#v|%d|%d|%#v", s.undoStack, s.redoStack, *s.current, s.revision, s.compiledRevision, s.compileState)
}

func TestRejectedSourceEditsRestoreImportsAndHistory(t *testing.T) {
	for _, name := range []string{"duplicate column", "missing column", "invalid column name", "invalid initial value", "duplicate local", "missing local", "duplicate public", "missing public", "invalid definition kind"} {
		t.Run(name, func(t *testing.T) {
			s, table, rules := sourceEditFixture(t)
			if err := os.Chmod(table, 0640); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(rules, 0640); err != nil {
				t.Fatal(err)
			}
			before, state := fileTree(t, s.current.Project.Root), sourceEditState(s)
			var err error
			switch name {
			case "duplicate column":
				_, err = s.AddColumn(ColumnAdd{Path: table, Name: "price", SchemaExpression: "common.#Positive", InitialExpression: "0"})
			case "missing column":
				_, err = s.SetColumnSchema(ColumnSchemaSet{Path: table, Name: "absent", SchemaExpression: "common.#Positive"})
			case "invalid column name":
				_, err = s.AddColumn(ColumnAdd{Path: table, Name: "bad name", SchemaExpression: "common.#Positive", InitialExpression: "0"})
			case "invalid initial value":
				_, err = s.AddColumn(ColumnAdd{Path: table, Name: "newPrice", SchemaExpression: "common.#Positive", InitialExpression: "("})
			case "duplicate local":
				_, err = s.CreateSchemaDefinition(SchemaDefinitionCreate{Path: table, Name: "Existing", Kind: "type", Type: "integer", SchemaExpression: "common.#Positive"})
			case "missing local":
				_, err = s.SetSchemaDefinition(SchemaDefinitionSet{Path: table, Name: "Absent", Kind: "type", Type: "integer", SchemaExpression: "common.#Positive"})
			case "duplicate public":
				_, err = s.CreateSchemaDefinition(SchemaDefinitionCreate{Path: rules, Name: "Existing", Kind: "type", Type: "integer", SchemaExpression: "common.#Positive"})
			case "missing public":
				_, err = s.SetSchemaDefinition(SchemaDefinitionSet{Path: rules, Name: "Absent", Kind: "type", Type: "integer", SchemaExpression: "common.#Positive"})
			case "invalid definition kind":
				_, err = s.CreateSchemaDefinition(SchemaDefinitionCreate{Path: rules, Name: "Fresh", Kind: "bad", Type: "integer", SchemaExpression: "common.#Positive"})
			}
			if err == nil {
				t.Fatal("invalid edit succeeded")
			}
			assertFileTree(t, s.current.Project.Root, before)
			if sourceEditState(s) != state {
				t.Fatal("failed edit changed history or workspace")
			}
			for _, path := range []string{table, rules} {
				if info, err := os.Stat(path); err != nil || runtime.GOOS != "windows" && info.Mode().Perm() != 0640 {
					t.Fatalf("failed edit changed source permissions: %v", err)
				}
			}
		})
	}
}

func TestSourceEditUndoRedoIncludesAutomaticallyAddedImports(t *testing.T) {
	for _, name := range []string{"add column", "set column", "local definition", "public definition"} {
		t.Run(name, func(t *testing.T) {
			s, table, rules := sourceEditFixture(t)
			if err := os.Chmod(table, 0640); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(rules, 0640); err != nil {
				t.Fatal(err)
			}
			before := fileTree(t, s.current.Project.Root)
			path := table
			var err error
			switch name {
			case "add column":
				_, err = s.AddColumn(ColumnAdd{Path: table, Name: "newPrice", SchemaExpression: "common.#Positive", InitialExpression: "0"})
			case "set column":
				_, err = s.SetColumnSchema(ColumnSchemaSet{Path: table, Name: "price", SchemaExpression: "common.#Positive"})
			case "local definition":
				_, err = s.CreateSchemaDefinition(SchemaDefinitionCreate{Path: table, Name: "Fresh", Kind: "type", Type: "integer", SchemaExpression: "common.#Positive"})
			case "public definition":
				path = rules
				_, err = s.CreateSchemaDefinition(SchemaDefinitionCreate{Path: rules, Name: "Fresh", Kind: "type", Type: "integer", SchemaExpression: "common.#Positive"})
			}
			if err != nil {
				t.Fatal(err)
			}
			if info, err := os.Stat(path); err != nil || runtime.GOOS != "windows" && info.Mode().Perm() != 0640 {
				t.Fatalf("successful edit changed source permissions: %v", err)
			}
			bytes, err := os.ReadFile(path)
			if err != nil || !strings.Contains(string(bytes), `import "example.com/game/schema/common:common"`) {
				t.Fatalf("expected import missing: %s %v", bytes, err)
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
		})
	}
}

func TestSourceEditRefreshFailureRestoresImportsAndCanRetry(t *testing.T) {
	s, table, _ := sourceEditFixture(t)
	schemaRoot := filepath.Join(s.current.Project.Root, "schema")
	if err := os.Rename(schemaRoot, schemaRoot+".away"); err != nil {
		t.Fatal(err)
	}
	before, state := fileTree(t, s.current.Project.Root), sourceEditState(s)
	request := ColumnSchemaSet{Path: table, Name: "price", SchemaExpression: "common.#Positive"}
	if _, err := s.SetColumnSchema(request); err == nil {
		t.Fatal("edit succeeded without configured source root")
	}
	assertFileTree(t, s.current.Project.Root, before)
	if sourceEditState(s) != state {
		t.Fatal("failed refresh changed workspace or history")
	}
	if err := os.Rename(schemaRoot+".away", schemaRoot); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetColumnSchema(request); err != nil {
		t.Fatalf("cannot retry source edit: %v", err)
	}
}

func TestSourceEditReportsOriginalAndRollbackFailures(t *testing.T) {
	s, table, _ := sourceEditFixture(t)
	state := sourceEditState(s)
	injected := errors.New("operation failed after partial write")
	_, _, err := editSource(s, table, func() (document.Table, error) {
		if err := os.Remove(table); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(table, "preserve.bin"), "only copy")
		return document.Table{}, injected
	})
	if !errors.Is(err, injected) || !strings.Contains(err.Error(), "rollback edit source") {
		t.Fatalf("missing error evidence: %v", err)
	}
	if sourceEditState(s) != state {
		t.Fatal("failed source rollback recorded history")
	}
	if content, err := os.ReadFile(filepath.Join(table, "preserve.bin")); err != nil || string(content) != "only copy" {
		t.Fatal("rollback removed conflicting content")
	}
}
