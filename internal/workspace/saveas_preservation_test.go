package workspace

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestSaveAsServicePipe is an opt-in real Service boundary, without GUI or Wails.
func TestSaveAsServicePipe(t *testing.T) {
	if os.Getenv("SAVEAS_PIPE") != "1" {
		t.Skip("explicit Service pipe only")
	}
	s := NewService(os.Getenv("SAVEAS_SETTINGS"))
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var q struct {
			Method string
			Path   string
			Update CellUpdate
		}
		if err := json.Unmarshal(scanner.Bytes(), &q); err != nil {
			t.Fatal(err)
		}
		var v any
		var err error
		switch q.Method {
		case "OpenProject":
			v, err = s.OpenProject(q.Path)
		case "SaveProjectAs":
			v, err = s.SaveProjectAs(q.Path)
		case "CreateProject":
			v, err = s.CreateProject(q.Path)
		case "UpdateCell":
			v, err = s.UpdateCell(q.Update)
		case "Undo":
			v, err = s.Undo()
		case "Redo":
			v, err = s.Redo()
		case "Compile":
			v, err = s.Compile()
		case "Export":
			v, err = s.Export()
		case "Current":
			v, _ = s.Current()
		default:
			err = fmt.Errorf("unsupported method %q", q.Method)
		}
		r := map[string]any{"value": v, "error": "", "undo": len(s.undoStack), "redo": len(s.redoStack), "busy": s.operationBusy()}
		if err != nil {
			r["error"] = err.Error()
		}
		if s.current != nil {
			r["current"] = s.current
			r["historyState"] = historyTestState(s)
		}
		data, e := json.Marshal(r)
		if e != nil {
			t.Fatal(e)
		}
		fmt.Printf("SAVEAS_JSON %s\n", data)
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
}

func TestSaveAsPreservesDanglingDestination(t *testing.T) {
	for _, name := range []string{"copy.dtproj", "notes.txt", "tables", "schema", "cue.mod"} {
		t.Run(name, func(t *testing.T) {
			s, path, _ := literalProject(t)
			if _, err := s.UpdateCell(CellUpdate{Path: path, RowID: "first", Column: "price", Expression: "7"}); err != nil {
				t.Fatal(err)
			}
			writeFile(t, filepath.Join(s.current.Project.Root, "notes.txt"), "source sidecar")
			target := t.TempDir()
			link := filepath.Join(target, name)
			text := filepath.Join(t.TempDir(), "missing")
			if err := os.Symlink(text, link); err != nil {
				t.Fatal(err)
			}
			sourceRoot := s.current.Project.Root
			before, state := fileTree(t, sourceRoot), historyTestState(s)
			_, err := s.SaveProjectAs(filepath.Join(target, "copy.dtproj"))
			t.Logf("SaveProjectAs(%s): %v", name, err)
			if err == nil {
				t.Error("SaveProjectAs accepted an existing dangling destination node")
			}
			got, e := os.Readlink(link)
			if e != nil || got != text {
				t.Errorf("destination link changed: %q %v", got, e)
			}
			assertFileTree(t, sourceRoot, before)
			if historyTestState(s) != state || s.operationBusy() {
				t.Error("failure changed workspace/history or retained operation lock")
			}
		})
	}
}

func TestProjectSlotNodeExistence(t *testing.T) {
	for _, name := range []string{"copy.dtproj", "tables", "schema", "cue.mod"} {
		for _, live := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/live=%t", name, live), func(t *testing.T) {
				root := t.TempDir()
				outside := filepath.Join(t.TempDir(), "node")
				if live {
					if err := os.Mkdir(outside, 0755); err != nil {
						t.Fatal(err)
					}
				}
				if err := os.Symlink(outside, filepath.Join(root, name)); err != nil {
					t.Fatal(err)
				}
				if err := validateProjectSlots(root, filepath.Join(root, "copy.dtproj")); err == nil {
					t.Fatal("accepted occupied slot symlink")
				}
			})
		}
	}
}
