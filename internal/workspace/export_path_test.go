package workspace

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestUnsafeExportPathsAreRejectedBeforeMutation(t *testing.T) {
	for _, test := range []struct {
		name, destination, tableRoot, schemaRoot string
	}{
		{name: "table root", destination: "tables"},
		{name: "table descendant", destination: "tables/generated"},
		{name: "table ancestor", destination: "source", tableRoot: "source/tables"},
		{name: "schema root", destination: "schema"},
		{name: "schema descendant", destination: "schema/generated"},
		{name: "schema ancestor", destination: "source", schemaRoot: "source/schema"},
		{name: "cleaned overlap", destination: "output/../tables"},
		{name: "project root", destination: "."},
		{name: "parent", destination: ".."},
		{name: "project file", destination: "game.dtproj"},
		{name: "CUE module", destination: "cue.mod"},
		{name: "CUE module descendant", destination: "cue.mod/generated"},
		{name: "version control", destination: ".git"},
		{name: "existing file", destination: "notes.txt"},
	} {
		t.Run(test.name, func(t *testing.T) {
			project := exportSafetyProject(t, test.tableRoot, test.schemaRoot)
			service := NewService(t.TempDir())
			if _, err := service.OpenProject(project.Path); err != nil {
				t.Fatal(err)
			}
			before := snapshotExportSafetyTree(t, project.Root)
			unsafe := project
			unsafe.Config.Export.Directory = test.destination
			if err := SaveProject(unsafe); err == nil {
				t.Fatal("SaveProject accepted an unsafe destination")
			}
			if _, err := service.SetProjectExportSettings(ProjectExportSettings{Directory: test.destination, Targets: []string{"csv"}}); err == nil {
				t.Fatal("SetProjectExportSettings accepted an unsafe destination")
			}
			if !reflect.DeepEqual(service.current.Project, project) {
				t.Fatal("rejected settings changed the in-memory project")
			}
			if _, err := service.NormalizeProjectExportDirectory(filepath.Join(project.Root, test.destination)); err == nil {
				t.Fatal("NormalizeProjectExportDirectory accepted an unsafe destination")
			}
			// Simulate a project loaded by an older version. Export must enforce
			// the boundary itself, not rely solely on load/settings validation.
			service.current.Project = unsafe
			if _, err := service.Export(); err == nil {
				t.Fatal("Export accepted an unsafe destination")
			}
			assertExportSafetyTree(t, project.Root, before)

			writeExportSafetyConfig(t, unsafe)
			before = snapshotExportSafetyTree(t, project.Root)
			if _, err := LoadProject(project.Path); err == nil {
				t.Fatal("LoadProject accepted an unsafe destination")
			}
			assertExportSafetyTree(t, project.Root, before)
		})
	}
}

func TestExportRechecksFilesystemAliasesBeforeMutation(t *testing.T) {
	for _, test := range []struct {
		name, destination string
		change            func(*testing.T, Project)
	}{
		{name: "destination aliases tables", destination: "exports", change: func(t *testing.T, p Project) {
			if err := os.RemoveAll(filepath.Join(p.Root, "exports")); err != nil {
				t.Fatal(err)
			}
			exportSafetySymlink(t, "tables", filepath.Join(p.Root, "exports"))
		}},
		{name: "destination aliases project root", destination: "alias", change: func(t *testing.T, p Project) {
			exportSafetySymlink(t, ".", filepath.Join(p.Root, "alias"))
		}},
		{name: "parent aliases tables", destination: "alias/generated/csv", change: func(t *testing.T, p Project) {
			exportSafetySymlink(t, "tables", filepath.Join(p.Root, "alias"))
		}},
		{name: "parent aliases schema", destination: "alias/generated", change: func(t *testing.T, p Project) {
			exportSafetySymlink(t, "schema", filepath.Join(p.Root, "alias"))
		}},
		{name: "parent escapes project", destination: "alias/new/csv", change: func(t *testing.T, p Project) {
			outside := t.TempDir()
			writeFile(t, filepath.Join(outside, "sentinel.txt"), "outside data")
			before := snapshotExportSafetyTree(t, outside)
			t.Cleanup(func() { assertExportSafetyTree(t, outside, before) })
			exportSafetySymlink(t, outside, filepath.Join(p.Root, "alias"))
		}},
		{name: "dangling destination", destination: "alias", change: func(t *testing.T, p Project) {
			exportSafetySymlink(t, "missing", filepath.Join(p.Root, "alias"))
		}},
		{name: "dangling parent", destination: "alias/new/csv", change: func(t *testing.T, p Project) {
			exportSafetySymlink(t, "missing", filepath.Join(p.Root, "alias"))
		}},
		{name: "cyclic alias", destination: "alias/new", change: func(t *testing.T, p Project) {
			exportSafetySymlink(t, "alias", filepath.Join(p.Root, "alias"))
		}},
		{name: "source root aliases output", destination: "exports", change: func(t *testing.T, p Project) {
			if err := os.Rename(filepath.Join(p.Root, "tables"), filepath.Join(p.Root, "exports", "tables")); err != nil {
				t.Fatal(err)
			}
			exportSafetySymlink(t, "exports/tables", filepath.Join(p.Root, "tables"))
		}},
		{name: "table source aliases output", destination: "exports", change: func(t *testing.T, p Project) {
			exportSafetyMoveAndLink(t, p, "tables/item.cue", "exports/item.cue")
		}},
		{name: "schema source aliases output", destination: "exports", change: func(t *testing.T, p Project) {
			exportSafetyMoveAndLink(t, p, "schema/common.cue", "exports/common.cue")
		}},
		{name: "project file aliases output", destination: "exports", change: func(t *testing.T, p Project) {
			exportSafetyMoveAndLink(t, p, "game.dtproj", "exports/game.dtproj")
		}},
		{name: "excluded source aliases output", destination: "exports", change: func(t *testing.T, p Project) {
			writeFile(t, filepath.Join(p.Root, "exports", "ignored.cue"), "package ignored\n")
			exportSafetySymlink(t, "../exports/ignored.cue", filepath.Join(p.Root, "tables", "ignored.cue"))
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			project := exportSafetyProject(t, "", "")
			service := NewService(t.TempDir())
			if _, err := service.OpenProject(project.Path); err != nil {
				t.Fatal(err)
			}
			if _, err := service.SetProjectExportSettings(ProjectExportSettings{Directory: test.destination, Targets: []string{"csv"}}); err != nil {
				t.Fatal(err)
			}
			test.change(t, project)
			before := snapshotExportSafetyTree(t, project.Root)
			if _, err := service.Export(); err == nil {
				t.Fatal("Export accepted a changed unsafe filesystem alias")
			}
			if _, err := LoadProject(project.Path); err == nil {
				t.Fatal("LoadProject accepted an unsafe filesystem alias")
			}
			assertExportSafetyTree(t, project.Root, before)
		})
	}
}

func TestSafeExportPathsPreserveProjectSources(t *testing.T) {
	for _, directory := range []string{"exports", "build/generated/csv", "tables-output", "schema-output"} {
		t.Run(directory, func(t *testing.T) {
			project := exportSafetyProject(t, "", "")
			project.Config.Export.Directory = directory
			if err := SaveProject(project); err != nil {
				t.Fatal(err)
			}
			destination := filepath.Join(project.Root, directory)
			writeFile(t, filepath.Join(destination, "obsolete.csv"), "old output")
			before := snapshotExportSafetyTree(t, project.Root)
			service := NewService(t.TempDir())
			if _, err := service.OpenProject(project.Path); err != nil {
				t.Fatal(err)
			}
			for attempt := 0; attempt < 2; attempt++ {
				result, err := service.Export()
				if err != nil || len(result.Files) != 1 {
					t.Fatalf("Export() = %#v, %v", result, err)
				}
				if got, err := os.ReadFile(result.Files[0]); err != nil || string(got) != "id,name\nsword,Sword\n" {
					t.Fatalf("exported CSV = %q, error = %v", got, err)
				}
			}
			after := snapshotExportSafetyTree(t, project.Root)
			for path, original := range before {
				if !isWithin(path, filepath.Clean(directory)) && after[path] != original {
					t.Errorf("export changed %q outside its destination", path)
				}
			}
			if _, err := os.Stat(filepath.Join(destination, "obsolete.csv")); !os.IsNotExist(err) {
				t.Fatalf("obsolete output was not replaced: %v", err)
			}
		})
	}
}

func TestExportAllowsProjectAndSafeParentAliases(t *testing.T) {
	project := exportSafetyProject(t, "", "")
	alias := filepath.Join(t.TempDir(), "project")
	exportSafetySymlink(t, project.Root, alias)
	project.Path = filepath.Join(alias, "game.dtproj")
	project.Root = alias
	exportSafetySymlink(t, "exports", filepath.Join(project.Root, "output-alias"))
	project.Config.Export.Directory = "output-alias/csv"
	if err := SaveProject(project); err != nil {
		t.Fatal(err)
	}
	service := NewService(t.TempDir())
	if _, err := service.OpenProject(project.Path); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Export(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(project.Root, "exports", "csv", "rpg.item.csv")); err != nil {
		t.Fatal(err)
	}
}

func TestExportPreservesSourceRootNamedLikeLegacyBackup(t *testing.T) {
	project := exportSafetyProject(t, "exports.datapack-backup", "")
	before := snapshotExportSafetyTree(t, filepath.Join(project.Root, "exports.datapack-backup"))
	service := NewService(t.TempDir())
	if _, err := service.OpenProject(project.Path); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Export(); err != nil {
		t.Fatal(err)
	}
	assertExportSafetyTree(t, filepath.Join(project.Root, "exports.datapack-backup"), before)
}

func exportSafetyProject(t *testing.T, tableRoot, schemaRoot string) Project {
	t.Helper()
	if tableRoot == "" {
		tableRoot = "tables"
	}
	if schemaRoot == "" {
		schemaRoot = "schema"
	}
	root := t.TempDir()
	project := Project{Path: filepath.Join(root, "game.dtproj"), Root: root, Config: ProjectConfig{
		Version: 1, TableRoots: []string{tableRoot}, SchemaRoots: []string{schemaRoot},
		Exclude: []string{"tables/ignored.cue"}, Export: ExportConfig{Directory: "exports", Targets: []string{"csv"}},
	}}
	writeExportSafetyConfig(t, project)
	writeFile(t, filepath.Join(root, tableRoot, "item.cue"), `package item
#PrimaryKey: string
#Row: {name: string}
rows: {sword: {name: "Sword"}}
_tool: {table: {namespace: "rpg", name: "item"}}
_export: {csv: {enabled: true, columns: ["id", "name"]}}`)
	writeFile(t, filepath.Join(root, schemaRoot, "common.cue"), "package common\n#Name: string\n_schema: {definitions: {Name: {kind: \"type\", type: \"string\"}}}\n")
	writeFile(t, filepath.Join(root, "cue.mod", "module.cue"), "module: \"example.com/test\"\nlanguage: {version: \"v0.9.0\"}\n")
	writeFile(t, filepath.Join(root, "exports", "previous.csv"), "previous output")
	writeFile(t, filepath.Join(root, "notes.txt"), "user notes")
	return project
}

func writeExportSafetyConfig(t *testing.T, project Project) {
	t.Helper()
	data, err := json.Marshal(project.Config)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, project.Path, string(data))
}

func exportSafetySymlink(t *testing.T, target, path string) {
	t.Helper()
	if err := os.Symlink(target, path); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
}

func exportSafetyMoveAndLink(t *testing.T, project Project, source, target string) {
	t.Helper()
	sourcePath, targetPath := filepath.Join(project.Root, source), filepath.Join(project.Root, target)
	if err := os.Rename(sourcePath, targetPath); err != nil {
		t.Fatal(err)
	}
	relative, err := filepath.Rel(filepath.Dir(sourcePath), targetPath)
	if err != nil {
		t.Fatal(err)
	}
	exportSafetySymlink(t, relative, sourcePath)
}

// Capture contents, links, directories and modes, without following symlinks.
// Rejected operations must not even leave staging directories behind.
func snapshotExportSafetyTree(t *testing.T, root string) map[string]string {
	t.Helper()
	result := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		value := info.Mode().String()
		if entry.Type()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			value += target
		} else if !entry.IsDir() {
			contents, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			value += string(contents)
		}
		result[relative] = value
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func assertExportSafetyTree(t *testing.T, root string, before map[string]string) {
	t.Helper()
	if after := snapshotExportSafetyTree(t, root); !reflect.DeepEqual(after, before) {
		t.Fatalf("project tree changed:\nbefore: %#v\nafter: %#v", before, after)
	}
}
