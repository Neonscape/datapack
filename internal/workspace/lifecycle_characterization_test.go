package workspace

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestManualPublicationRetainsLastCompileUntilExplicitCompile(t *testing.T) {
	s, table, _ := sourceEditFixture(t)
	if _, err := s.SetCompileMode("manual"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetColumnSchema(ColumnSchemaSet{Path: table, Name: "price", SchemaExpression: "int & >=99"}); err != nil {
		t.Fatal(err)
	}
	compiled, err := s.Compile()
	if err != nil || len(compiled.Workspace.Compile.Diagnostics) == 0 {
		t.Fatalf("expected initial constraint diagnostics: %#v %v", compiled, err)
	}
	previous := compiled.Workspace.Compile
	revision := s.revision
	compiledRevision := s.compiledRevision
	result, err := s.SetColumnSchema(ColumnSchemaSet{Path: table, Name: "price", SchemaExpression: "int"})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Workspace.Compile.Pending || result.Workspace.Compile.Mode != "manual" || result.Workspace.Compile.CompiledAt != previous.CompiledAt || !reflect.DeepEqual(result.Workspace.Compile.Diagnostics, previous.Diagnostics) {
		t.Fatalf("manual edit changed cached compilation: %#v", result.Workspace.Compile)
	}
	if s.revision != revision+1 || s.compiledRevision != compiledRevision || !result.Workspace.CanUndo || result.Workspace.CanRedo {
		t.Fatal("manual publication changed revision/history protocol")
	}
	again, err := s.Compile()
	if err != nil {
		t.Fatal(err)
	}
	if again.Workspace.Compile.Pending || len(again.Workspace.Compile.Diagnostics) != 0 || s.compiledRevision != s.revision {
		t.Fatalf("explicit compile did not publish repaired data: %#v", again)
	}
}

func TestAutoPublicationCompilesEachSuccessfulEdit(t *testing.T) {
	s, table, _ := sourceEditFixture(t)
	if _, err := s.SetCompileMode("auto"); err != nil {
		t.Fatal(err)
	}
	before := s.revision
	result, err := s.UpdateCell(CellUpdate{Path: table, RowID: "one", Column: "name", Expression: `"Auto"`})
	if err != nil {
		t.Fatal(err)
	}
	if result.Workspace.Compile.Mode != "auto" || result.Workspace.Compile.Pending || result.Workspace.Compile.CompiledAt == "" || s.revision != before+1 || s.compiledRevision != s.revision {
		t.Fatalf("auto publication: %#v", result.Workspace.Compile)
	}
}

func TestFailedRefreshDoesNotPublishLifecycleAndReopenResetsHistory(t *testing.T) {
	s, table, _ := sourceEditFixture(t)
	if _, err := s.SetCompileMode("manual"); err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(s.current.Project.Path)
	if err != nil {
		t.Fatal(err)
	}
	state := sourceEditState(s)
	writeFile(t, s.current.Project.Path, "invalid configuration")
	if _, err := s.refreshCurrentView(); err == nil {
		t.Fatal("invalid refresh succeeded")
	}
	if sourceEditState(s) != state {
		t.Fatal("failed refresh published lifecycle state")
	}
	if err := os.WriteFile(s.current.Project.Path, original, 0644); err != nil {
		t.Fatal(err)
	}
	// Change a real source outside the Service before reopening. Nonempty
	// timestamps alone cannot distinguish a compile from a cached result.
	data, err := os.ReadFile(table)
	if err != nil {
		t.Fatal(err)
	}
	changed := strings.Replace(string(data), "price: 10", "price: \"wrong\"", 1)
	if changed == string(data) || len(s.compileState.Diagnostics) != 0 {
		t.Fatal("reopen fixture needs a newly introduced source error")
	}
	writeFile(t, table, changed)
	result, err := s.OpenProject(s.current.Project.Path)
	if err != nil {
		t.Fatal(err)
	}
	if s.revision != 0 || s.compiledRevision != 0 || result.CanUndo || result.CanRedo || result.Compile.Pending || result.Compile.Mode != "manual" || result.Compile.CompiledAt == "" {
		t.Fatalf("reopen protocol: %#v", result.Compile)
	}
	if len(result.Compile.Diagnostics) == 0 {
		t.Fatal("reopen reused the old successful compilation after external source change")
	}
}
