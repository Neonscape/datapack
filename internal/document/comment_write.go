package document

import (
	"fmt"
	"os"
	"strings"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/parser"
	"datapack/internal/cueformat"
)

// SetCellComment creates, updates, or removes one _editor.cellComments entry.
func SetCellComment(path, rowID, column, comment string) (Table, error) {
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
	comments := editorFields["cellComments"]
	if comments == nil {
		comments = &ast.Field{Label: ast.NewIdent("cellComments"), Value: &ast.StructLit{}}
		editorStruct.Elts = append(editorStruct.Elts, comments)
	}
	commentsStruct, ok := comments.Value.(*ast.StructLit)
	if !ok {
		return Table{}, fmt.Errorf("_editor.cellComments must be a struct")
	}
	rows := fieldsOfStruct(commentsStruct)
	row := rows[rowID]
	if row == nil {
		row = &ast.Field{Label: ast.NewIdent(rowID), Value: &ast.StructLit{}}
		commentsStruct.Elts = append(commentsStruct.Elts, row)
	}
	rowStruct, ok := row.Value.(*ast.StructLit)
	if !ok {
		return Table{}, fmt.Errorf("cell comments for row %q must be a struct", rowID)
	}
	if strings.TrimSpace(comment) == "" {
		removeNamedField(rowStruct, column)
		if len(orderedStructFields(rowStruct)) == 0 {
			removeNamedField(commentsStruct, rowID)
		}
		if len(orderedStructFields(commentsStruct)) == 0 {
			removeNamedField(editorStruct, "cellComments")
		}
		if len(orderedStructFields(editorStruct)) == 0 {
			removeNamedFieldFromFile(file, "_editor")
		}
	} else if existing := fieldsOfStruct(rowStruct)[column]; existing != nil {
		existing.Value = stringLiteral(comment)
	} else {
		rowStruct.Elts = append(rowStruct.Elts, &ast.Field{Label: ast.NewIdent(column), Value: stringLiteral(comment)})
	}
	candidate, err := cueformat.Node(file)
	if err != nil {
		return Table{}, fmt.Errorf("format cell comment: %w", err)
	}
	return validateAndReplace(path, candidate)
}

func removeNamedFieldFromFile(file *ast.File, name string) bool {
	for index, declaration := range file.Decls {
		field, ok := declaration.(*ast.Field)
		if ok && labelOf(field) == name {
			file.Decls = append(file.Decls[:index], file.Decls[index+1:]...)
			return true
		}
	}
	return false
}
