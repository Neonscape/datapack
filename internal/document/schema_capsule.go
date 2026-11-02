package document

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/parser"
	"cuelang.org/go/cue/token"
)

// SchemaCapsule omits rows only when every data value is an independent scalar
// literal, IDs/fields are unique, no imports/formulas/FKs exist, and no other
// top-level expression references rows. Remaining schema/metadata source bytes
// are unchanged. Every original row is still fully validated by the compiler.
// Everything outside this deliberately narrow proof keeps full CUE loading.
func SchemaCapsule(table Table) ([]byte, bool) {
	if len(table.Formulas) != 0 || len(table.ForeignKeys) != 0 {
		return nil, false
	}
	source, err := os.ReadFile(table.Path)
	if err != nil {
		return nil, false
	}
	file, err := parser.ParseFile(table.Path, source, parser.ParseComments, parser.AllErrors)
	if err != nil {
		return nil, false
	}
	// Sibling and ancestor package files could add unseen dependencies on rows.
	for directory := filepath.Dir(table.Path); ; directory = filepath.Dir(directory) {
		candidates, _ := filepath.Glob(filepath.Join(directory, "*.cue"))
		for _, candidate := range candidates {
			if filepath.Clean(candidate) != filepath.Clean(table.Path) {
				return nil, false
			}
		}
		if _, err := os.Stat(filepath.Join(directory, "cue.mod")); err == nil {
			break
		}
		if filepath.Dir(directory) == directory {
			break
		}
	}
	fields := fieldsOfFile(file)
	rowsField := fields["rows"]
	if rowsField == nil {
		return nil, false
	}
	rows, ok := rowsField.Value.(*ast.StructLit)
	if !ok {
		return nil, false
	}
	seenRows := map[string]bool{}
	for _, decl := range rows.Elts {
		row, ok := decl.(*ast.Field)
		if !ok {
			return nil, false
		}
		name := labelOf(row)
		if name == "" || seenRows[name] || row.Constraint != token.ILLEGAL {
			return nil, false
		}
		seenRows[name] = true
		values, ok := row.Value.(*ast.StructLit)
		if !ok {
			return nil, false
		}
		seenColumns := map[string]bool{}
		for _, decl := range values.Elts {
			field, ok := decl.(*ast.Field)
			if !ok {
				return nil, false
			}
			name := labelOf(field)
			if name == "" || seenColumns[name] || field.Constraint != token.ILLEGAL || !scalarAST(field.Value) {
				return nil, false
			}
			seenColumns[name] = true
		}
	}
	if len(seenRows) != len(table.Rows) {
		return nil, false
	}
	for _, decl := range file.Decls {
		if decl == rowsField {
			continue
		}
		switch decl.(type) {
		case *ast.Field, *ast.Package:
		default:
			return nil, false
		}
		referencesRows := false
		ast.Walk(decl, func(n ast.Node) bool {
			if ident, ok := n.(*ast.Ident); ok && ident.Name == "rows" {
				referencesRows = true
			}
			return !referencesRows
		}, nil)
		if referencesRows {
			return nil, false
		}
	}
	start, end := rows.Pos().Offset(), rows.End().Offset()
	// Do not use a document snapshot from a different filesystem revision.
	snapshotHash := shaSource(source)
	if snapshotHash != table.SourceSHA256 {
		return nil, false
	}
	var capsule bytes.Buffer
	capsule.Write(source[:start])
	capsule.WriteString("{}")
	capsule.Write(source[end:])
	return capsule.Bytes(), true
}

func shaSource(source []byte) string { sum := sha256.Sum256(source); return hex.EncodeToString(sum[:]) }
