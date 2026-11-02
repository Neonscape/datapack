package workspace

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportReloadsExternalDestinationBeforePublishing(t *testing.T) {
	for _, destination := range []string{"exports-B", "tables/item", "project.dtproj/child"} {
		t.Run(strings.ReplaceAll(destination, "/", "_"), func(t *testing.T) {
			s, root, table := dependencyFixture(t, false, false)
			if _, err := s.Export(); err != nil {
				t.Fatal(err)
			}
			oldPath := filepath.Join(root, "exports", "rpg.item.csv")
			before, err := os.ReadFile(oldPath)
			if err != nil {
				t.Fatal(err)
			}
			configuration, err := os.ReadFile(s.current.Project.Path)
			if err != nil {
				t.Fatal(err)
			}
			var values map[string]any
			if err := json.Unmarshal(configuration, &values); err != nil {
				t.Fatal(err)
			}
			values["export"].(map[string]any)["directory"] = destination
			configuration, err = json.Marshal(values)
			if err != nil {
				t.Fatal(err)
			}
			writeFile(t, s.current.Project.Path, string(configuration))
			raw, err := os.ReadFile(table)
			if err != nil {
				t.Fatal(err)
			}
			writeFile(t, table, strings.Replace(string(raw), "price: 10", "price: 11", 1))
			result, err := s.Export()
			after, readErr := os.ReadFile(oldPath)
			if readErr != nil || !bytes.Equal(before, after) {
				t.Fatal("external destination change overwrote old A")
			}
			if destination != "exports-B" {
				if err == nil {
					t.Fatal("unsafe or invalid new destination accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if result.Directory != filepath.Join(root, destination) {
				t.Fatalf("wrong authoritative destination: %q", result.Directory)
			}
			data, err := os.ReadFile(filepath.Join(result.Directory, "rpg.item.csv"))
			if err != nil {
				t.Fatal(err)
			}
			rows, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
			if err != nil {
				t.Fatal(err)
			}
			if len(rows) != 2 || strings.Join(rows[0], ",") != "id,price" || strings.Join(rows[1], ",") != "one,11" {
				t.Fatalf("new B output: %#v", rows)
			}
		})
	}
}

func TestExportReturnsCurrentValidationWorkspace(t *testing.T) {
	s, path, _ := literalProject(t)
	if !s.current.Compile.Pending {
		t.Fatal("fixture must start unvalidated")
	}
	first, err := s.Export()
	if err != nil {
		t.Fatal(err)
	}
	if first.Workspace.Compile.Pending || first.Workspace.Compile.InputRevision != first.Workspace.Revision || first.Workspace.ProjectEpoch != s.current.ProjectEpoch {
		t.Fatal("export omitted current validation authority")
	}
	if _, err := s.UpdateCell(CellUpdate{Path: path, RowID: "first", Column: "price", Expression: `"bad"`}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Compile(); err != nil {
		t.Fatal(err)
	}
	if len(s.current.Compile.Diagnostics) == 0 {
		t.Fatal("fixture needs old validation errors")
	}
	if _, err := s.UpdateCell(CellUpdate{Path: path, RowID: "first", Column: "price", Expression: "2"}); err != nil {
		t.Fatal(err)
	}
	if !s.current.Compile.Pending || len(s.current.Compile.Diagnostics) == 0 {
		t.Fatal("manual repair should retain stale errors")
	}
	exported, err := s.Export()
	if err != nil {
		t.Fatal(err)
	}
	if exported.Workspace.Compile.Pending || len(exported.Workspace.Compile.Diagnostics) != 0 || exported.Workspace.Compile.InputRevision != exported.Workspace.Revision {
		t.Fatal("successful export returned stale validation")
	}
}
