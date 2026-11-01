package workspace

import (
	"path/filepath"
	"testing"

	"datapack/internal/compiler"
	"datapack/internal/document"
	"datapack/internal/projecttree"
	"datapack/internal/schemaregistry"
)

// BenchmarkYunlan separates filesystem scanning, restricted-model parsing,
// registry construction and full CUE/formula compilation. It intentionally
// does not add caches or alter observable compilation/publication timing.
func BenchmarkYunlan(b *testing.B) {
	path, err := filepath.Abs("../../docs/examples/yunlan-rpg/yunlan-rpg.dtproj")
	if err != nil {
		b.Fatal(err)
	}
	project, err := LoadProject(path)
	if err != nil {
		b.Fatal(err)
	}
	scan := func() projecttree.Tree {
		tree, err := projecttree.Scan(project.Path, project.Config.TableRoots, project.Config.SchemaRoots, project.Config.Exclude)
		if err != nil {
			b.Fatal(err)
		}
		return tree
	}
	tree := scan()
	var schemas []document.PublicSchema
	var tables []document.Table
	if diagnostics := collectDocuments(tree.Root, &schemas, &tables); len(diagnostics) != 0 {
		b.Fatal(diagnostics)
	}
	if len(tables) != 16 {
		b.Fatalf("unexpected sample: %d tables", len(tables))
	}
	b.Run("scan", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			_ = scan()
		}
	})
	b.Run("parse", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			var schemas []document.PublicSchema
			var tables []document.Table
			if diagnostics := collectDocuments(tree.Root, &schemas, &tables); len(diagnostics) != 0 {
				b.Fatal(diagnostics)
			}
		}
	})
	b.Run("index", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			if result := schemaregistry.Build(schemas, tables); len(result.Diagnostics) != 0 {
				b.Fatal(result.Diagnostics)
			}
		}
	})
	b.Run("compile", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			if result := compiler.CompileAll(schemas, tables); result.Package == nil || len(result.Diagnostics) != 0 {
				b.Fatal(result.Diagnostics)
			}
		}
	})
}
