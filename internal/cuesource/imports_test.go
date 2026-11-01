package cuesource

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCanonicalImportPreservesPackageAndMajorVersion(t *testing.T) {
	for _, test := range []struct{ in, want string }{
		{"example.com/game/schema/common", "example.com/game/schema/common:common"},
		{"example.com/game/schema/shared:common", "example.com/game/schema/shared:common"},
		{"example.com/game/schema/shared@v1:common", "example.com/game/schema/shared@v1:common"},
		{"", ""},
	} {
		if got := CanonicalImport(test.in); got != test.want {
			t.Fatalf("%s: %s", test.in, got)
		}
	}
}
func TestPackageImportPathUsesEnclosingModuleAndAllowsStandalone(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "schema", "shared", "rules.cue")
	if got, err := PackageImportPath(path, "common"); err != nil || got != "" {
		t.Fatalf("standalone: %q %v", got, err)
	}
	if err := os.MkdirAll(filepath.Join(root, "cue.mod"), 0755); err != nil {
		t.Fatal(err)
	}
	module := filepath.Join(root, "cue.mod", "module.cue")
	if err := os.WriteFile(module, []byte(`module:"example.com/game@v1"`), 0644); err != nil {
		t.Fatal(err)
	}
	if got, err := PackageImportPath(path, "common"); err != nil || got != "example.com/game/schema/shared@v1:common" {
		t.Fatalf("module path: %q %v", got, err)
	}
	if err := os.WriteFile(module, []byte(`module: {bad:`), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := PackageImportPath(path, "common"); err == nil {
		t.Fatal("bad module was silently accepted")
	}
}

func TestImportContextUsesMainAndDependencyMajorDefaults(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "cue.mod"), 0755); err != nil {
		t.Fatal(err)
	}
	source := `module:"example.com/game@v1"
language:{version:"v0.9.0"}
deps:{"dep.example/rules@v2":{v:"v2.1.0",default:true}}`
	if err := os.WriteFile(filepath.Join(root, "cue.mod/module.cue"), []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	context, err := ReadImportContext(filepath.Join(root, "tables/item.cue"), "item")
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct{ in, want string }{
		{"example.com/game/schema/rules", "example.com/game/schema/rules@v1:rules"},
		{"example.com/game/schema/rules@v1", "example.com/game/schema/rules@v1:rules"},
		{"dep.example/rules/common", "dep.example/rules/common@v2:common"},
		{"other.example/unknown", "other.example/unknown:unknown"},
	} {
		if got := context.Resolve(test.in); got != test.want {
			t.Fatalf("%s: %s", test.in, got)
		}
	}
}
