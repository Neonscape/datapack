package document

import (
	"fmt"
	"os"
	"strings"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/format"
	"cuelang.org/go/cue/parser"
)

// SetColumnDisplayName updates the optional user-facing title in _editor.columns.
func SetColumnDisplayName(path, column, displayName string) (Table, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return Table{}, fmt.Errorf("read table source: %w", err)
	}
	file, err := parser.ParseFile(path, source, parser.ParseComments, parser.AllErrors)
	if err != nil {
		return Table{}, fmt.Errorf("parse table source: %w", err)
	}
	fields := fieldsOfFile(file)
	editor := fields["_editor"]
	if editor == nil {
		editor = &ast.Field{Label: ast.NewIdent("_editor"), Value: &ast.StructLit{}}
		file.Decls = append(file.Decls, editor)
	}
	editorStruct, ok := editor.Value.(*ast.StructLit)
	if !ok {
		return Table{}, fmt.Errorf("_editor must be a struct")
	}
	editorFields := fieldsOfStruct(editorStruct)
	columns := editorFields["columns"]
	if columns == nil {
		columns = &ast.Field{Label: ast.NewIdent("columns"), Value: &ast.StructLit{}}
		editorStruct.Elts = append(editorStruct.Elts, columns)
	}
	columnsStruct, ok := columns.Value.(*ast.StructLit)
	if !ok {
		return Table{}, fmt.Errorf("_editor.columns must be a struct")
	}
	columnFields := fieldsOfStruct(columnsStruct)
	if strings.TrimSpace(displayName) == "" {
		if existing := columnFields[column]; existing != nil {
			removeNamedField(columnsStruct, column)
		}
	} else if existing := columnFields[column]; existing != nil {
		metadata, ok := existing.Value.(*ast.StructLit)
		if !ok {
			return Table{}, fmt.Errorf("_editor.columns.%s must be a struct", column)
		}
		metadataFields := fieldsOfStruct(metadata)
		if title := metadataFields["title"]; title != nil {
			title.Value = stringLiteral(displayName)
		} else {
			metadata.Elts = append(metadata.Elts, &ast.Field{Label: ast.NewIdent("title"), Value: stringLiteral(displayName)})
		}
	} else {
		columnsStruct.Elts = append(columnsStruct.Elts, &ast.Field{Label: ast.NewIdent(column), Value: &ast.StructLit{Elts: []ast.Decl{&ast.Field{Label: ast.NewIdent("title"), Value: stringLiteral(displayName)}}}})
	}
	if len(orderedStructFields(columnsStruct)) == 0 {
		removeNamedField(editorStruct, "columns")
	}
	candidate, err := format.Node(file)
	if err != nil {
		return Table{}, fmt.Errorf("format column display name: %w", err)
	}
	return validateAndReplace(path, candidate)
}
