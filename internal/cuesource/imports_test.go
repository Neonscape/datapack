package cuesource

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCanonicalImportPreservesPackageAndMajorVersion(t *testing.T) {
	for _, test := range []struct{ in, want string }{
		{"example.com/game/schema/common", "example.com/game/schema/common@v0:common"},
		{"example.com/game/schema/shared:common", "example.com/game/schema/shared@v0:common"},
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
