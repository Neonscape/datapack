package workspace

import (
	"datapack/internal/compiler"
	"reflect"
	"testing"
)

func TestLifecyclePublicationSeparatesCachedCompilationAndCurrentMode(t *testing.T) {
	var lifecycle viewLifecycle
	lifecycle.recordCompile("auto", "first", nil)
	lifecycle.markEdited()
	input := View{}
	manual := lifecycle.publish(input, "manual", true, false)
	if manual.Compile.Mode != "manual" || !manual.Compile.Pending || manual.Compile.CompiledAt != "first" || manual.Compile.Diagnostics == nil || !manual.CanUndo || manual.CanRedo {
		t.Fatalf("manual: %#v", manual)
	}
	if lifecycle.compileState.Mode != "auto" || lifecycle.compiledRevision != 0 || input.Compile.Mode != "" {
		t.Fatal("publication mutated cached state or input")
	}
	diagnostics := []compiler.Diagnostic{{Path: "table.cue", Message: "preserve"}}
	lifecycle.recordCompile("manual", "second", diagnostics)
	current := lifecycle.publish(input, "manual", false, true)
	if current.Compile.Pending || !reflect.DeepEqual(current.Compile.Diagnostics, diagnostics) || current.Compile.CompiledAt != "second" || current.CanUndo || !current.CanRedo {
		t.Fatalf("current: %#v", current)
	}
	lifecycle.resetRevisions()
	if lifecycle.revision != 0 || lifecycle.compiledRevision != 0 || lifecycle.compileState.CompiledAt != "second" {
		t.Fatal("reset changed the last compile before the new project's initial compile")
	}
}
