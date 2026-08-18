package functional

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"datapack/internal/compiler"
	"datapack/internal/projecttree"
	"datapack/internal/workspace"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q): %v", path, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile(%q): %v", path, err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", path, err)
	}
	return string(contents)
}

func copyTree(t *testing.T, source, destination string) {
	t.Helper()
	err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		input, err := os.Open(path)
		if err != nil {
			return err
		}
		defer input.Close()
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		output, err := os.Create(target)
		if err != nil {
			return err
		}
		defer output.Close()
		_, err = io.Copy(output, input)
		return err
	})
	if err != nil {
		t.Fatalf("copy %s -> %s: %v", source, destination, err)
	}
}

func copyExample(t *testing.T, name string) string {
	t.Helper()
	destination := t.TempDir()
	copyTree(t, filepath.Join(repoRoot(t), "docs", "examples", name), destination)
	return destination
}

func openProject(t *testing.T, projectPath string) *workspace.Service {
	t.Helper()
	service := workspace.NewService(t.TempDir())
	if _, err := service.OpenProject(projectPath); err != nil {
		t.Fatalf("OpenProject(%q): %v", projectPath, err)
	}
	return service
}

func openExample(t *testing.T, name string) (*workspace.Service, string) {
	t.Helper()
	root := copyExample(t, name)
	matches, err := filepath.Glob(filepath.Join(root, "*.dtproj"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("example %s dtproj = %v, %v", name, matches, err)
	}
	return openProject(t, matches[0]), root
}

func emptyProject(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "game.dtproj"), `{
  "version": 1,
  "tableRoots": ["tables"],
  "schemaRoots": ["schema"],
  "export": {"directory": "exports", "targets": ["csv"]}
}`)
	writeFile(t, filepath.Join(root, "cue.mod", "module.cue"), "module: \"example.com/game\"\nlanguage: {version: \"v0.9.0\"}\n")
	if err := os.MkdirAll(filepath.Join(root, "tables"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root, filepath.Join(root, "game.dtproj")
}

func enableCSVExport(t *testing.T, tablePath string, columns []string) {
	t.Helper()
	encoded, err := json.Marshal(columns)
	if err != nil {
		t.Fatal(err)
	}
	source := strings.TrimRight(readFile(t, tablePath), "\n")
	if strings.Contains(source, "_export:") {
		t.Fatalf("%s already has _export", tablePath)
	}
	writeFile(t, tablePath, source+"\n\n_export: {\n\tcsv: {\n\t\tenabled: true\n\t\tcolumns: "+string(encoded)+"\n\t}\n}\n")
}

func requireCleanCompile(t *testing.T, view workspace.View) {
	t.Helper()
	if len(view.Compile.Diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %s", formatDiagnostics(view.Compile.Diagnostics))
	}
}

func formatDiagnostics(diagnostics []compiler.Diagnostic) string {
	parts := make([]string, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		parts = append(parts, fmt.Sprintf("%s %s %s %s: %s", diagnostic.Path, diagnostic.Table, diagnostic.RowID, diagnostic.Column, diagnostic.Message))
	}
	return strings.Join(parts, "\n")
}

func hasDiagnosticAt(diagnostics []compiler.Diagnostic, rowID, column string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.RowID == rowID && diagnostic.Column == column {
			return true
		}
	}
	return false
}

func findNode(node projecttree.Node, name string) *projecttree.Node {
	if node.Name == name {
		return &node
	}
	for _, child := range node.Children {
		if found := findNode(child, name); found != nil {
			return found
		}
	}
	return nil
}

func goldenPath(t *testing.T, example, file string) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "test", "testdata", "golden", example, file)
}
