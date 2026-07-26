package document

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/format"
	"cuelang.org/go/cue/parser"
	"cuelang.org/go/cue/token"
)

// RowReference identifies a scalar or array foreign-key value that must track
// a renamed row ID.
type RowReference struct {
	Path   string
	RowID  string
	Column string
	Index  int
	Array  bool
}

// RenameRow changes a row ID and updates all supplied foreign-key references.
// Every candidate source is parsed and structurally reloaded before any source
// is replaced; failed replacement restores already written files.
func RenameRow(path, oldID, newID string, references []RowReference) (Table, error) {
	if !validIdentifier(newID) {
		return Table{}, fmt.Errorf("row ID %q must be a valid CUE identifier", newID)
	}
	files := map[string]*ast.File{}
	original := map[string][]byte{}
	loadFile := func(filePath string) (*ast.File, error) {
		if file, exists := files[filePath]; exists {
			return file, nil
		}
		source, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("read table source: %w", err)
		}
		file, err := parser.ParseFile(filePath, source, parser.ParseComments, parser.AllErrors)
		if err != nil {
			return nil, fmt.Errorf("parse table source: %w", err)
		}
		files[filePath] = file
		original[filePath] = source
		return file, nil
	}

	target, err := loadFile(path)
	if err != nil {
		return Table{}, err
	}
	if err := renameRowID(target, oldID, newID); err != nil {
		return Table{}, err
	}
	for _, reference := range references {
		file, err := loadFile(reference.Path)
		if err != nil {
			return Table{}, err
		}
		if err := renameReference(file, reference, oldID, newID); err != nil {
			return Table{}, err
		}
	}

	candidates := map[string][]byte{}
	paths := make([]string, 0, len(files))
	for filePath, file := range files {
		candidate, err := format.Node(file)
		if err != nil {
			return Table{}, fmt.Errorf("format updated table: %w", err)
		}
		if _, err := validateCandidate(filePath, candidate); err != nil {
			return Table{}, err
		}
		candidates[filePath] = candidate
		paths = append(paths, filePath)
	}
	sort.Strings(paths)
	if err := replaceAll(paths, candidates, original); err != nil {
		return Table{}, err
	}
	table, err := LoadTable(path)
	if err != nil {
		return Table{}, fmt.Errorf("reload renamed table: %w", err)
	}
	return table, nil
}

func renameRowID(file *ast.File, oldID, newID string) error {
	rows := requiredField(fieldsOfFile(file), "rows")
	rowsValue, ok := rows.Value.(*ast.StructLit)
	if !ok {
		return fmt.Errorf("rows must be a struct")
	}
	if _, exists := fieldsOfStruct(rows.Value)[newID]; exists {
		return fmt.Errorf("row %q already exists", newID)
	}
	for _, element := range rowsValue.Elts {
		field, ok := element.(*ast.Field)
		if ok && labelOf(field) == oldID {
			field.Label = ast.NewIdent(newID)
			return nil
		}
	}
	return fmt.Errorf("row %q does not exist", oldID)
}

func renameReference(file *ast.File, reference RowReference, oldID, newID string) error {
	rows := requiredField(fieldsOfFile(file), "rows")
	row, exists := fieldsOfStruct(rows.Value)[reference.RowID]
	if !exists {
		return fmt.Errorf("referencing row %q does not exist", reference.RowID)
	}
	cell, exists := fieldsOfStruct(row.Value)[reference.Column]
	if !exists {
		return fmt.Errorf("referencing column %q does not exist", reference.Column)
	}
	if reference.Array {
		list, ok := cell.Value.(*ast.ListLit)
		if !ok || reference.Index < 0 || reference.Index >= len(list.Elts) {
			return fmt.Errorf("foreign key array item %s.%s[%d] no longer exists", reference.RowID, reference.Column, reference.Index)
		}
		if !matchesString(list.Elts[reference.Index], oldID) {
			return fmt.Errorf("foreign key array item %s.%s[%d] no longer references %q", reference.RowID, reference.Column, reference.Index, oldID)
		}
		list.Elts[reference.Index] = stringLiteral(newID)
		return nil
	}
	if !matchesString(cell.Value, oldID) {
		return fmt.Errorf("foreign key %s.%s no longer references %q", reference.RowID, reference.Column, oldID)
	}
	cell.Value = stringLiteral(newID)
	return nil
}

func matchesString(expression ast.Expr, expected string) bool {
	literal, ok := expression.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return false
	}
	value, err := strconv.Unquote(literal.Value)
	return err == nil && value == expected
}

func stringLiteral(value string) *ast.BasicLit {
	return &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(value)}
}

func validateCandidate(path string, source []byte) (Table, error) {
	temporary, err := os.CreateTemp(filepath.Dir(path), ".datapack-validate-*.cue")
	if err != nil {
		return Table{}, fmt.Errorf("create validation table source: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err := temporary.Write(source); err != nil {
		temporary.Close()
		return Table{}, fmt.Errorf("write validation table source: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return Table{}, fmt.Errorf("close validation table source: %w", err)
	}
	table, err := LoadTable(temporaryPath)
	if err != nil {
		return Table{}, fmt.Errorf("validate updated table: %w", err)
	}
	return table, nil
}

func replaceAll(paths []string, candidates, originals map[string][]byte) error {
	temporary := map[string]string{}
	for _, path := range paths {
		file, err := os.CreateTemp(filepath.Dir(path), ".datapack-rename-*.cue")
		if err != nil {
			return fmt.Errorf("stage renamed table: %w", err)
		}
		if _, err := file.Write(candidates[path]); err != nil {
			file.Close()
			os.Remove(file.Name())
			return fmt.Errorf("write renamed table: %w", err)
		}
		if err := file.Sync(); err != nil {
			file.Close()
			os.Remove(file.Name())
			return fmt.Errorf("sync renamed table: %w", err)
		}
		if err := file.Close(); err != nil {
			os.Remove(file.Name())
			return fmt.Errorf("close renamed table: %w", err)
		}
		temporary[path] = file.Name()
	}
	committed := []string{}
	for _, path := range paths {
		if err := os.Rename(temporary[path], path); err != nil {
			for _, committedPath := range committed {
				_ = os.WriteFile(committedPath, originals[committedPath], 0o644)
			}
			return fmt.Errorf("replace renamed table: %w", err)
		}
		committed = append(committed, path)
	}
	return nil
}
