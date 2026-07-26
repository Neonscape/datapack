package document

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/format"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/cue/parser"
)

// SetCellValue replaces one existing table cell with a valid CUE expression.
// The candidate is reloaded from a sibling temporary file before an atomic
// rename makes it visible to the rest of the project.
func SetCellValue(path, rowID, column, expression string) (Table, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return Table{}, fmt.Errorf("read table source: %w", err)
	}
	file, err := parser.ParseFile(path, source, parser.ParseComments, parser.AllErrors)
	if err != nil {
		return Table{}, fmt.Errorf("parse table source: %w", err)
	}
	value, err := parser.ParseExpr(path, expression)
	if err != nil {
		return Table{}, fmt.Errorf("parse cell expression: %w", err)
	}
	if err := replaceCell(file, rowID, column, value); err != nil {
		return Table{}, err
	}
	candidate, err := format.Node(file)
	if err != nil {
		return Table{}, fmt.Errorf("format updated table: %w", err)
	}
	if err := validateCellValue(path, candidate, rowID); err != nil {
		return Table{}, err
	}
	return validateAndReplace(path, candidate)
}

// AddRow inserts a new table row using a complete CUE struct expression.
func AddRow(path, rowID, expression string) (Table, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return Table{}, fmt.Errorf("read table source: %w", err)
	}
	file, err := parser.ParseFile(path, source, parser.ParseComments, parser.AllErrors)
	if err != nil {
		return Table{}, fmt.Errorf("parse table source: %w", err)
	}
	value, err := parser.ParseExpr(path, expression)
	if err != nil {
		return Table{}, fmt.Errorf("parse row expression: %w", err)
	}
	rowValue, ok := value.(*ast.StructLit)
	if !ok {
		return Table{}, fmt.Errorf("row expression must be a struct")
	}
	if err := insertRow(file, rowID, rowValue); err != nil {
		return Table{}, err
	}
	candidate, err := format.Node(file)
	if err != nil {
		return Table{}, fmt.Errorf("format updated table: %w", err)
	}
	if err := validateCellValue(path, candidate, rowID); err != nil {
		return Table{}, err
	}
	return validateAndReplace(path, candidate)
}

func replaceCell(file *ast.File, rowID, column string, value ast.Expr) error {
	fields := fieldsOfFile(file)
	rowSchema := requiredField(fields, "#Row")
	if _, exists := fieldsOfStruct(rowSchema.Value)[column]; !exists {
		return fmt.Errorf("column %q does not exist", column)
	}
	rows := requiredField(fields, "rows")
	row, exists := fieldsOfStruct(rows.Value)[rowID]
	if !exists {
		return fmt.Errorf("row %q does not exist", rowID)
	}
	if _, ok := row.Value.(*ast.StructLit); !ok {
		return fmt.Errorf("row %q is not a struct", rowID)
	}
	cell, exists := fieldsOfStruct(row.Value)[column]
	if !exists {
		return fmt.Errorf("row %q has no value for column %q", rowID, column)
	}
	cell.Value = value
	return nil
}

func insertRow(file *ast.File, rowID string, value *ast.StructLit) error {
	rowID = strings.TrimSpace(rowID)
	if !validIdentifier(rowID) {
		return fmt.Errorf("row ID %q must be a valid CUE identifier", rowID)
	}
	fields := fieldsOfFile(file)
	rows := requiredField(fields, "rows")
	rowsValue, ok := rows.Value.(*ast.StructLit)
	if !ok {
		return fmt.Errorf("rows must be a struct")
	}
	if _, exists := fieldsOfStruct(rows.Value)[rowID]; exists {
		return fmt.Errorf("row %q already exists", rowID)
	}
	rowsValue.Elts = append(rowsValue.Elts, &ast.Field{Label: ast.NewIdent(rowID), Value: value})
	return nil
}

func validIdentifier(value string) bool {
	if value == "" {
		return false
	}
	for index, r := range value {
		if index == 0 {
			if r != '_' && !unicode.IsLetter(r) {
				return false
			}
			continue
		}
		if r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func validateCellValue(path string, source []byte, rowID string) error {
	instances := load.Instances([]string{"."}, &load.Config{
		Dir:     filepath.Dir(path),
		Overlay: map[string]load.Source{path: load.FromBytes(source)},
	})
	if len(instances) != 1 {
		return fmt.Errorf("load updated table package: expected one package, found %d", len(instances))
	}
	context := cuecontext.New()
	packageValue := context.BuildInstance(instances[0])
	if err := packageValue.Err(); err != nil {
		return fmt.Errorf("load updated table package: %w", err)
	}
	rowSchema := packageValue.LookupPath(cue.MakePath(cue.Def("Row")))
	row := packageValue.LookupPath(cue.MakePath(cue.Str("rows"), cue.Str(rowID)))
	if err := rowSchema.Unify(row).Validate(cue.Concrete(true)); err != nil {
		return fmt.Errorf("validate row %q: %w", rowID, err)
	}
	return nil
}

func validateAndReplace(path string, source []byte) (Table, error) {
	info, err := os.Stat(path)
	if err != nil {
		return Table{}, fmt.Errorf("stat table source: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".datapack-*.cue")
	if err != nil {
		return Table{}, fmt.Errorf("create temporary table source: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(info.Mode()); err != nil {
		temporary.Close()
		return Table{}, fmt.Errorf("set temporary table permissions: %w", err)
	}
	if _, err := temporary.Write(source); err != nil {
		temporary.Close()
		return Table{}, fmt.Errorf("write temporary table source: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return Table{}, fmt.Errorf("sync temporary table source: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return Table{}, fmt.Errorf("close temporary table source: %w", err)
	}
	table, err := LoadTable(temporaryPath)
	if err != nil {
		return Table{}, fmt.Errorf("validate updated table: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return Table{}, fmt.Errorf("replace table source: %w", err)
	}
	table.Path = path
	return table, nil
}
