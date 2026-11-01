package document

import (
	"fmt"
	"os"
	"strings"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/parser"
	"datapack/internal/cueformat"
)

// SetDisplayName edits presentation metadata only. An empty definition selects
// the table/schema file; otherwise it selects an existing public/local definition.
func SetDisplayName(path, definition, displayName string) error {
	source, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read display name source: %w", err)
	}
	file, err := parser.ParseFile(path, source, parser.ParseComments, parser.AllErrors)
	if err != nil {
		return fmt.Errorf("parse display name source: %w", err)
	}
	fields := fieldsOfFile(file)
	isTable := fields["_tool"] != nil
	metadataPath := []string{"schema"}
	if isTable {
		metadataPath[0] = "table"
	}
	if definition != "" {
		if !validIdentifier(definition) {
			return fmt.Errorf("invalid Schema definition name %q", definition)
		}
		declarations := fields
		if isTable {
			local := fields["_local"]
			if local == nil {
				return fmt.Errorf("table has no local Schema")
			}
			declarations = fieldsOfStruct(local.Value)
		}
		if declarations["#"+definition] == nil {
			return fmt.Errorf("Schema definition #%s does not exist", definition)
		}
		metadataPath = []string{"schemas", definition}
	}
	editor, err := editorStruct(file, true)
	if err != nil {
		return err
	}
	current := editor
	for _, name := range metadataPath {
		field := fieldsOfStruct(current)[name]
		if field == nil {
			field = &ast.Field{Label: ast.NewIdent(name), Value: &ast.StructLit{}}
			current.Elts = append(current.Elts, field)
		}
		nested, ok := field.Value.(*ast.StructLit)
		if !ok {
			return fmt.Errorf("display name metadata %q must be a struct", name)
		}
		current = nested
	}
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		removeNamedField(current, "title")
	} else if title := fieldsOfStruct(current)["title"]; title != nil {
		title.Value = stringLiteral(displayName)
	} else {
		current.Elts = append(current.Elts, &ast.Field{Label: ast.NewIdent("title"), Value: stringLiteral(displayName)})
	}
	candidate, err := cueformat.Node(file)
	if err != nil {
		return fmt.Errorf("format display name: %w", err)
	}
	return replaceSource(path, candidate)
}

func editorStruct(file *ast.File, create bool) (*ast.StructLit, error) {
	field := fieldsOfFile(file)["_editor"]
	if field == nil {
		if !create {
			return nil, nil
		}
		field = &ast.Field{Label: ast.NewIdent("_editor"), Value: &ast.StructLit{}}
		file.Decls = append(file.Decls, field)
	}
	value, ok := field.Value.(*ast.StructLit)
	if !ok {
		return nil, fmt.Errorf("_editor must be a struct")
	}
	return value, nil
}

// Definition metadata follows its stable internal ID on rename, never its title.
// Orphan destination metadata must not silently become another definition's title.
func renameDefinitionTitle(file *ast.File, oldName, newName string) error {
	definitions, err := definitionTitles(file)
	if err != nil || definitions == nil {
		return err
	}
	if fieldsOfStruct(definitions)[newName] != nil {
		return fmt.Errorf("display metadata already exists for #%s", newName)
	}
	renameNamedField(definitions, oldName, newName)
	return nil
}

func deleteDefinitionTitle(file *ast.File, name string) error {
	definitions, err := definitionTitles(file)
	if err != nil || definitions == nil {
		return err
	}
	removeNamedField(definitions, name)
	return nil
}

func definitionTitles(file *ast.File) (*ast.StructLit, error) {
	editor, err := editorStruct(file, false)
	if err != nil || editor == nil {
		return nil, err
	}
	field := fieldsOfStruct(editor)["schemas"]
	if field == nil {
		return nil, nil
	}
	value, ok := field.Value.(*ast.StructLit)
	if !ok {
		return nil, fmt.Errorf("_editor.schemas must be a struct")
	}
	return value, nil
}
