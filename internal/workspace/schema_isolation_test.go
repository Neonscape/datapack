package workspace

import (
	"datapack/internal/schemamodel"
	"reflect"
	"sync"
	"testing"
)

func TestEffectiveSchemaDoesNotRewriteSourceOrRegistryFields(t *testing.T) {
	reference := schemamodel.Node{Kind: schemamodel.KindReference, Reference: "common.#ID"}
	source := schemamodel.Node{Kind: schemamodel.KindStruct, Fields: []schemamodel.Field{{Name: "nested", Schema: reference}}}
	definitions := map[string]schemamodel.Node{"common.#ID": {Kind: schemamodel.KindString}, "common.#Shape": source}
	effective := resolveEffectiveSchema(schemamodel.Node{Kind: schemamodel.KindReference, Reference: "common.#Shape"}, "game.item", definitions, map[string]bool{})
	if effective.Fields[0].Schema.Kind != schemamodel.KindString {
		t.Fatal("effective type did not expand")
	}
	if !reflect.DeepEqual(source.Fields[0].Schema, reference) || !reflect.DeepEqual(definitions["common.#Shape"].Fields[0].Schema, reference) {
		t.Fatal("effective expansion rewrote the original source/registry reference")
	}
}

func TestConcurrentEffectiveSchemaReadsKeepReferencesStable(t *testing.T) {
	source := schemamodel.Node{Kind: schemamodel.KindStruct, Fields: []schemamodel.Field{{Name: "nested", Schema: schemamodel.Node{Kind: schemamodel.KindReference, Reference: "common.#ID"}}}}
	definitions := map[string]schemamodel.Node{"common.#ID": {Kind: schemamodel.KindString}, "common.#Shape": source}
	var group sync.WaitGroup
	for range 8 {
		group.Go(func() {
			for range 10 {
				_ = resolveEffectiveSchema(schemamodel.Node{Kind: schemamodel.KindReference, Reference: "common.#Shape"}, "game.item", definitions, map[string]bool{})
			}
		})
	}
	group.Wait()
	if source.Fields[0].Schema.Reference != "common.#ID" {
		t.Fatal("shared definitions changed after concurrent reads")
	}
}
