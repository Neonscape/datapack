package document

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/parser"
	"datapack/internal/cueformat"
)

func SetColumnWidth(path, column string, width int) (Table, error) {
	if width < 80 || width > 600 {
		return Table{}, fmt.Errorf("column width must be between 80 and 600")
	}
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
	columnField := fieldsOfStruct(columnsStruct)[column]
	if columnField == nil {
		columnField = &ast.Field{Label: ast.NewIdent(column), Value: &ast.StructLit{}}
		columnsStruct.Elts = append(columnsStruct.Elts, columnField)
	}
	metadata, ok := columnField.Value.(*ast.StructLit)
	if !ok {
		return Table{}, fmt.Errorf("_editor.columns.%s must be a struct", column)
	}
	metadataFields := fieldsOfStruct(metadata)
	widthValue, err := parser.ParseExpr(path, strconv.Itoa(width))
	if err != nil {
		return Table{}, fmt.Errorf("parse column width: %w", err)
	}
	if existing := metadataFields["width"]; existing != nil {
		existing.Value = widthValue
	} else {
		metadata.Elts = append(metadata.Elts, &ast.Field{Label: ast.NewIdent("width"), Value: widthValue})
	}
	candidate, err := cueformat.Node(file)
	if err != nil {
		return Table{}, fmt.Errorf("format column width: %w", err)
	}
	return validateAndReplace(path, candidate)
}

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
	candidate, err := cueformat.Node(file)
	if err != nil {
		return Table{}, fmt.Errorf("format column display name: %w", err)
	}
	return validateAndReplace(path, candidate)
}
