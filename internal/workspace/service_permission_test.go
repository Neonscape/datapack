package workspace

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func readOnlyDirectory(t *testing.T, directory string) func() {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("requires Unix directory permissions")
	}
	info, err := os.Stat(directory)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(directory, 0500); err != nil {
		t.Fatal(err)
	}
	restore := func() {
		if err := os.Chmod(directory, info.Mode().Perm()); err != nil {
			t.Error(err)
		}
	}
	t.Cleanup(restore)
	probe, err := os.CreateTemp(directory, ".permission-probe-")
	if err == nil {
		probe.Close()
		os.Remove(probe.Name())
		t.Skip("process bypasses directory write restrictions")
	}
	if !os.IsPermission(err) {
		t.Fatalf("unexpected permission probe: %v", err)
	}
	return restore
}

func TestServiceSourcePermissionFailureCanRetry(t *testing.T) {
	for _, fast := range []bool{false, true} {
		name := "snapshot"
		if fast {
			name = "literal"
		}
		t.Run(name, func(t *testing.T) {
			s, target, _ := literalProject(t)
			if !fast {
				t.Setenv("DATAPACK_FAST_EDIT", "0")
			}
			update := CellUpdate{Path: target, RowID: "first", Column: "price", Expression: "7"}
			before, state := fileTree(t, s.current.Project.Root), historyTestState(s)
			restore := readOnlyDirectory(t, filepath.Dir(target))
			if _, err := s.UpdateCell(update); err == nil {
				t.Fatal("edit succeeded without directory write permission")
			}
			assertFileTree(t, s.current.Project.Root, before)
			if historyTestState(s) != state || s.sourceUncertain {
				t.Fatal("failed write changed authority or history")
			}
			restore()
			if _, err := s.UpdateCell(update); err != nil {
				t.Fatalf("retry edit: %v", err)
			}
			if _, err := s.Undo(); err != nil {
				t.Fatalf("undo retry: %v", err)
			}
			assertFileTree(t, s.current.Project.Root, before)
		})
	}
}

func TestServiceSaveAsStagingPermissionFailureCanRetry(t *testing.T) {
	s, _, _ := renameFixture(t)
	parent := t.TempDir()
	target := filepath.Join(parent, "copy", "copy.dtproj")
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		t.Fatal(err)
	}
	before, state := fileTree(t, s.current.Project.Root), historyTestState(s)
	restore := readOnlyDirectory(t, parent)
	if _, err := s.SaveProjectAs(target); err == nil {
		t.Fatal("save-as succeeded without staging permission")
	}
	assertFileTree(t, s.current.Project.Root, before)
	if historyTestState(s) != state || s.operationBusy() {
		t.Fatal("failed save-as changed state or retained operation lock")
	}
	if entries, err := os.ReadDir(filepath.Dir(target)); err != nil || len(entries) != 0 {
		t.Fatalf("partial destination: %v %v", entries, err)
	}
	restore()
	if _, err := s.SaveProjectAs(target); err != nil {
		t.Fatalf("retry save-as: %v", err)
	}
	if s.current.Project.Path != target {
		t.Fatal("successful retry did not switch to saved project")
	}
}
