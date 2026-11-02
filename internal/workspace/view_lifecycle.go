package workspace

import "datapack/internal/compiler"

// viewLifecycle owns revision/last-compile publication. It performs no I/O or
// scheduling; Service still owns the mutex, source commands and history stacks.
// Manual views retain the previous compile while exposing Pending separately.
type viewLifecycle struct {
	revision         int
	compiledRevision int
	compileState     CompileState
}

func (l *viewLifecycle) resetRevisions() { l.revision = 0; l.compiledRevision = 0 }
func (l *viewLifecycle) markEdited()     { l.revision++ }
func (l *viewLifecycle) recordCompile(mode, at string, diagnostics []compiler.Diagnostic) CompileState {
	if diagnostics == nil {
		diagnostics = []compiler.Diagnostic{}
	}
	l.compileState = CompileState{InputRevision: l.revision, Mode: mode, CompiledAt: at, Diagnostics: diagnostics}
	l.compiledRevision = l.revision
	return l.compileState
}
func (l viewLifecycle) publish(view View, mode string, canUndo, canRedo bool) View {
	view.Revision = l.revision
	view.Compile = l.compileState
	view.Compile.Mode = mode
	view.Compile.Pending = l.compiledRevision < l.revision
	if view.Compile.Diagnostics == nil {
		view.Compile.Diagnostics = []compiler.Diagnostic{}
	}
	view.CanUndo, view.CanRedo = canUndo, canRedo
	return view
}
