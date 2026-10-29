// Package cueformat emits source compatible with the editor's supported older
// CUE modules, independent of the installed CUE formatter's default syntax.
package cueformat

import (
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/format"
	"cuelang.org/go/cue/token"
)

// Node preserves explicit list separators. CUE v0.17's formatter infers a
// comma-free list when an element's position lacks HasComma (e.g. a new AST
// literal replacing an old one). Such output is invalid for modules declaring
// v0.9.0. Setting format.Version alone is insufficient: list style is inferred
// from the scanned opening bracket and element positions in this dependency.
// Treat lists as constructed ASTs for this call, retaining all actual positions
// and comments. Restore the annotation afterwards, like format.Node's read-only
// contract; callers own their local AST while formatting it.
func Node(node ast.Node) ([]byte, error) {
	type opening struct {
		list     *ast.ListLit
		position token.Pos
	}
	var originals []opening
	ast.Walk(node, func(n ast.Node) bool {
		if list, ok := n.(*ast.ListLit); ok {
			originals = append(originals, opening{list, list.Lbrack})
			list.Lbrack = list.Lbrack.WithScanned(false)
		}
		return true
	}, nil)
	defer func() {
		for _, original := range originals {
			original.list.Lbrack = original.position
		}
	}()
	return format.Node(node)
}
