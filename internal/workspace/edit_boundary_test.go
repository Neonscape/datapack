package workspace

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestRenameAndDeleteRefreshFailurePreservesSourcesAndHistory(t *testing.T) {
	for _, operation := range []string{"row rename", "public rename", "local rename", "public delete", "local delete"} {
		t.Run(operation, func(t *testing.T) {
			s, table, rules := sourceEditFixture(t)
			data, _ := os.ReadFile(table)
			writeFile(t, table, strings.ReplaceAll(string(data), "Existing", "LocalOnly"))
			// A config changed on disk since the last view must not publish a
			// successful source edit when refreshing the view fails.
			writeFile(t, s.current.Project.Path, "invalid project config")
			writeFile(t, filepath.Join(filepath.Dir(table), "attachment.bin"), "binary\x00attachment")
			beforeModes := map[string]os.FileMode{}
			filepath.WalkDir(s.current.Project.Root, func(path string, entry os.DirEntry, err error) error {
				if err != nil {
					t.Fatal(err)
				}
				info, err := entry.Info()
				if err != nil {
					t.Fatal(err)
				}
				beforeModes[path] = info.Mode()
				return nil
			})
			before := fileTree(t, s.current.Project.Root)
			state := sourceEditState(s)
			var err error
			switch operation {
			case "row rename":
				_, err = s.RenameRow(RowRename{Path: table, OldID: "one", NewID: "renamed"})
			case "public rename":
				_, err = s.RenameSchemaDefinition(SchemaDefinitionRename{Path: rules, OldName: "Existing", NewName: "Renamed"})
			case "local rename":
				_, err = s.RenameSchemaDefinition(SchemaDefinitionRename{Path: table, OldName: "LocalOnly", NewName: "Renamed"})
			case "public delete":
				_, err = s.DeleteSchemaDefinition(SchemaDefinitionDelete{Path: rules, Name: "Existing"})
			case "local delete":
				_, err = s.DeleteSchemaDefinition(SchemaDefinitionDelete{Path: table, Name: "LocalOnly"})
			}
			if err == nil {
				t.Fatal("refresh unexpectedly succeeded")
			}
			assertFileTree(t, s.current.Project.Root, before)
			afterModes := map[string]os.FileMode{}
			filepath.WalkDir(s.current.Project.Root, func(path string, entry os.DirEntry, err error) error {
				if err != nil {
					t.Fatal(err)
				}
				info, err := entry.Info()
				if err != nil {
					t.Fatal(err)
				}
				afterModes[path] = info.Mode()
				return nil
			})
			if !reflect.DeepEqual(beforeModes, afterModes) {
				t.Fatal("failed command changed file/directory permissions")
			}
			if sourceEditState(s) != state {
				t.Fatal("failed operation changed history/revision/compile state")
			}
		})
	}
}

func TestCreateSchemaRefreshFailureRemovesOnlyItsNewDirectories(t *testing.T) {
	s, _, _ := sourceEditFixture(t)
	writeFile(t, s.current.Project.Path, "invalid project config")
	before := fileTree(t, s.current.Project.Root)
	state := sourceEditState(s)
	if _, err := s.CreateSchemaFile("fresh"); err == nil {
		t.Fatal("refresh unexpectedly succeeded")
	}
	assertFileTree(t, s.current.Project.Root, before)
	if sourceEditState(s) != state {
		t.Fatal("failed create changed state")
	}
	if _, err := os.Stat(s.current.Project.Root + "/schema/fresh"); !os.IsNotExist(err) {
		t.Fatalf("left created directory: %v", err)
	}
}

func TestSnapshotSourcesDistinguishesEmptyMissingAndUnreadable(t *testing.T) {
	root := t.TempDir()
	empty, missing := root+"/empty.cue", root+"/missing.cue"
	writeFile(t, empty, "")
	snapshot, err := snapshotRenameFiles(empty, missing)
	if err != nil || snapshot[empty] == nil || len(snapshot[empty]) != 0 || snapshot[missing] != nil {
		t.Fatalf("empty/missing snapshot: %#v %v", snapshot, err)
	}
	if _, err := snapshotRenameFiles(root); err == nil {
		t.Fatal("unreadable source directory accepted")
	}
	link := root + "/dangling.cue"
	if err := os.Symlink(missing, link); err == nil {
		if _, err := snapshotRenameFiles(link); err == nil {
			t.Fatal("existing unreadable symlink treated as missing")
		}
	}
}
