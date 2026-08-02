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
	exprparser "github.com/expr-lang/expr/parser"
)

// SetCellValue replaces one cell with a syntactically valid CUE expression.
// Schema errors are intentionally persisted as diagnostics so users can build a
// row incrementally, like a spreadsheet.
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
	removeFormula(file, rowID, column)
	candidate, err := format.Node(file)
	if err != nil {
		return Table{}, fmt.Errorf("format updated table: %w", err)
	}
	return validateAndReplace(path, candidate)
}

// PasteCells applies a rectangular set of fixed values and formulas as one
// source transaction. A single malformed input leaves the source unchanged.
func PasteCells(path string, rowIDs, columns []string, values [][]string) (Table, error) {
	if len(rowIDs) == 0 || len(columns) == 0 || len(values) != len(rowIDs) {
		return Table{}, fmt.Errorf("paste range dimensions do not match")
	}
	source, err := os.ReadFile(path)
	if err != nil {
		return Table{}, fmt.Errorf("read table source: %w", err)
	}
	file, err := parser.ParseFile(path, source, parser.ParseComments, parser.AllErrors)
	if err != nil {
		return Table{}, fmt.Errorf("parse table source: %w", err)
	}
	for rowIndex, rowID := range rowIDs {
		if len(values[rowIndex]) != len(columns) {
			return Table{}, fmt.Errorf("paste row %d has %d values, want %d", rowIndex+1, len(values[rowIndex]), len(columns))
		}
		for columnIndex, column := range columns {
			input := strings.TrimSpace(values[rowIndex][columnIndex])
			if input == "" {
				continue
			}
			if strings.HasPrefix(input, "=") {
				expression := strings.TrimSpace(strings.TrimPrefix(input, "="))
				if _, err := exprparser.Parse(expression); err != nil {
					return Table{}, fmt.Errorf("paste row %q column %q formula: %w", rowID, column, err)
				}
				if err := setFormula(file, rowID, column, expression); err != nil {
					return Table{}, err
				}
				continue
			}
			value, err := parser.ParseExpr(path, input)
			if err != nil {
				return Table{}, fmt.Errorf("paste row %q column %q: %w", rowID, column, err)
			}
			if err := replaceCell(file, rowID, column, value); err != nil {
				return Table{}, err
			}
			removeFormula(file, rowID, column)
		}
	}
	// Apply clears last so a partially cleared row remains available for later
	// values in the same paste transaction.
	for rowIndex, rowID := range rowIDs {
		var emptyColumns []string
		for columnIndex, column := range columns {
			if strings.TrimSpace(values[rowIndex][columnIndex]) == "" {
				emptyColumns = append(emptyColumns, column)
			}
		}
		if err := clearCells(file, rowID, emptyColumns); err != nil {
			return Table{}, err
		}
	}
	candidate, err := format.Node(file)
	if err != nil {
		return Table{}, fmt.Errorf("format pasted table: %w", err)
	}
	return validateAndReplace(path, candidate)
}

// SetCellFormula stores an expr-lang/expr expression and removes the fixed
// source value. Formula evaluation supplies the value during preview and export.
func SetCellFormula(path, rowID, column, expression string) (Table, error) {
	if strings.TrimSpace(expression) == "" {
		return Table{}, fmt.Errorf("formula expression must not be empty")
	}
	if _, err := exprparser.Parse(expression); err != nil {
		return Table{}, fmt.Errorf("parse formula expression: %w", err)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		return Table{}, fmt.Errorf("read table source: %w", err)
	}
	file, err := parser.ParseFile(path, source, parser.ParseComments, parser.AllErrors)
	if err != nil {
		return Table{}, fmt.Errorf("parse table source: %w", err)
	}
	if err := setFormula(file, rowID, column, expression); err != nil {
		return Table{}, err
	}
	candidate, err := format.Node(file)
	if err != nil {
		return Table{}, fmt.Errorf("format updated table: %w", err)
	}
	return validateAndReplace(path, candidate)
}

// CreateRow writes a partial row after its ID becomes a valid CUE identifier.
// It intentionally accepts incomplete business data; compilation reports that
// state back to the editor as cell diagnostics.
func CreateRow(path, rowID string, values map[string]string, formulas map[string]string) (Table, error) {
	if !validIdentifier(rowID) {
		return Table{}, fmt.Errorf("row ID %q must be a valid CUE identifier", rowID)
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
	if _, exists := fieldsOfStruct(requiredField(fields, "rows").Value)[rowID]; exists {
		return Table{}, fmt.Errorf("row %q already exists", rowID)
	}
	row := &ast.StructLit{}
	if err := insertRow(file, rowID, row); err != nil {
		return Table{}, err
	}
	for column, expression := range values {
		value, err := parser.ParseExpr(path, expression)
		if err != nil {
			return Table{}, fmt.Errorf("parse row %q column %q: %w", rowID, column, err)
		}
		if err := replaceCell(file, rowID, column, value); err != nil {
			return Table{}, err
		}
	}
	for column, expression := range formulas {
		if _, err := exprparser.Parse(expression); err != nil {
			return Table{}, fmt.Errorf("parse formula %q.%q: %w", rowID, column, err)
		}
		if err := setFormula(file, rowID, column, expression); err != nil {
			return Table{}, err
		}
	}
	candidate, err := format.Node(file)
	if err != nil {
		return Table{}, fmt.Errorf("format updated table: %w", err)
	}
	return validateAndReplace(path, candidate)
}

// ClearCell removes both a fixed expression and formula from one cell. Empty
// rows are removed to avoid persisting invisible placeholder objects.
func ClearCell(path, rowID, column string) (Table, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return Table{}, fmt.Errorf("read table source: %w", err)
	}
	file, err := parser.ParseFile(path, source, parser.ParseComments, parser.AllErrors)
	if err != nil {
		return Table{}, fmt.Errorf("parse table source: %w", err)
	}
	fields := fieldsOfFile(file)
	rows := requiredField(fields, "rows")
	row := fieldsOfStruct(rows.Value)[rowID]
	if row == nil {
		return Table{}, fmt.Errorf("row %q does not exist", rowID)
	}
	if err := clearCell(file, rowID, column); err != nil {
		return Table{}, err
	}
	candidate, err := format.Node(file)
	if err != nil {
		return Table{}, fmt.Errorf("format updated table: %w", err)
	}
	return validateAndReplace(path, candidate)
}

func clearCell(file *ast.File, rowID, column string) error {
	return clearCells(file, rowID, []string{column})
}

func clearCells(file *ast.File, rowID string, columns []string) error {
	fields := fieldsOfFile(file)
	rows := requiredField(fields, "rows")
	row := fieldsOfStruct(rows.Value)[rowID]
	if row == nil {
		return fmt.Errorf("row %q does not exist", rowID)
	}
	for _, column := range columns {
		removeNamedField(row.Value, column)
		removeFormula(file, rowID, column)
	}
	if len(orderedStructFields(row.Value)) == 0 && !hasFormulasForRow(file, rowID) {
		return removeRow(file, rowID)
	}
	return nil
}

// RestoreSource atomically replaces a table with a previously recorded valid
// source snapshot. Workspace uses it for undo and redo.
func RestoreSource(path string, source []byte) (Table, error) {
	return validateAndReplace(path, source)
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
	if err := validateRows(path, candidate, []string{rowID}); err != nil {
		return Table{}, err
	}
	return validateAndReplace(path, candidate)
}

// DeleteRow removes an existing table row.
func DeleteRow(path, rowID string) (Table, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return Table{}, fmt.Errorf("read table source: %w", err)
	}
	file, err := parser.ParseFile(path, source, parser.ParseComments, parser.AllErrors)
	if err != nil {
		return Table{}, fmt.Errorf("parse table source: %w", err)
	}
	if err := removeRow(file, rowID); err != nil {
		return Table{}, err
	}
	candidate, err := format.Node(file)
	if err != nil {
		return Table{}, fmt.Errorf("format updated table: %w", err)
	}
	return validateAndReplace(path, candidate)
}

// AddColumn adds a required column and initializes every existing row with one
// CUE expression. New columns are intentionally not added to export or foreign
// key metadata; those are explicit follow-up configuration changes.
func AddColumn(path, name, schemaExpression, initialExpression string) (Table, error) {
	if !validIdentifier(name) {
		return Table{}, fmt.Errorf("column %q must be a valid CUE identifier", name)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		return Table{}, fmt.Errorf("read table source: %w", err)
	}
	file, err := parser.ParseFile(path, source, parser.ParseComments, parser.AllErrors)
	if err != nil {
		return Table{}, fmt.Errorf("parse table source: %w", err)
	}
	schema, err := parser.ParseExpr(path, schemaExpression)
	if err != nil {
		return Table{}, fmt.Errorf("parse column Schema: %w", err)
	}
	value, err := parser.ParseExpr(path, initialExpression)
	if err != nil {
		return Table{}, fmt.Errorf("parse initial column value: %w", err)
	}
	if err := insertColumn(file, name, schema, value); err != nil {
		return Table{}, err
	}
	candidate, err := format.Node(file)
	if err != nil {
		return Table{}, fmt.Errorf("format updated table: %w", err)
	}
	if err := validateAllRows(path, candidate); err != nil {
		return Table{}, err
	}
	return validateAndReplace(path, candidate)
}

// DeleteColumn removes a column definition, its values, and same-table tool
// metadata that would otherwise reference the removed column.
func DeleteColumn(path, name string) (Table, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return Table{}, fmt.Errorf("read table source: %w", err)
	}
	file, err := parser.ParseFile(path, source, parser.ParseComments, parser.AllErrors)
	if err != nil {
		return Table{}, fmt.Errorf("parse table source: %w", err)
	}
	if err := removeColumn(file, name); err != nil {
		return Table{}, err
	}
	candidate, err := format.Node(file)
	if err != nil {
		return Table{}, fmt.Errorf("format updated table: %w", err)
	}
	return validateAndReplace(path, candidate)
}

// RenameColumn updates a column and all same-table metadata that identifies it.
func RenameColumn(path, oldName, newName string) (Table, error) {
	if !validIdentifier(newName) {
		return Table{}, fmt.Errorf("column %q must be a valid CUE identifier", newName)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		return Table{}, fmt.Errorf("read table source: %w", err)
	}
	file, err := parser.ParseFile(path, source, parser.ParseComments, parser.AllErrors)
	if err != nil {
		return Table{}, fmt.Errorf("parse table source: %w", err)
	}
	if err := renameColumn(file, oldName, newName); err != nil {
		return Table{}, err
	}
	if err := renameFormulaColumnReferences(file, oldName, newName); err != nil {
		return Table{}, err
	}
	candidate, err := format.Node(file)
	if err != nil {
		return Table{}, fmt.Errorf("format updated table: %w", err)
	}
	return validateAndReplace(path, candidate)
}

// MoveColumn changes a column's zero-based position in #Row and the configured
// CSV column order. Row field order is left intact because it has no semantic
// meaning and changing it would create unnecessary source churn.
func MoveColumn(path, name string, position int) (Table, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return Table{}, fmt.Errorf("read table source: %w", err)
	}
	file, err := parser.ParseFile(path, source, parser.ParseComments, parser.AllErrors)
	if err != nil {
		return Table{}, fmt.Errorf("parse table source: %w", err)
	}
	if err := moveColumn(file, name, position); err != nil {
		return Table{}, err
	}
	candidate, err := format.Node(file)
	if err != nil {
		return Table{}, fmt.Errorf("format updated table: %w", err)
	}
	if err := validateAllRows(path, candidate); err != nil {
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
	rowValues := row.Value.(*ast.StructLit)
	cell, exists := fieldsOfStruct(row.Value)[column]
	if exists {
		cell.Value = value
		return nil
	}
	rowValues.Elts = append(rowValues.Elts, &ast.Field{Label: ast.NewIdent(column), Value: value})
	return nil
}

func setFormula(file *ast.File, rowID, column, expression string) error {
	fields := fieldsOfFile(file)
	if _, exists := fieldsOfStruct(requiredField(fields, "#Row").Value)[column]; !exists {
		return fmt.Errorf("column %q does not exist", column)
	}
	rows := requiredField(fields, "rows")
	row, exists := fieldsOfStruct(rows.Value)[rowID]
	if !exists {
		return fmt.Errorf("row %q does not exist", rowID)
	}
	removeNamedField(row.Value, column)
	tool := requiredField(fields, "_tool")
	toolValue, ok := tool.Value.(*ast.StructLit)
	if !ok {
		return fmt.Errorf("_tool must be a struct")
	}
	toolFields := fieldsOfStruct(tool.Value)
	formulas := toolFields["formulas"]
	if formulas == nil {
		formulas = &ast.Field{Label: ast.NewIdent("formulas"), Value: &ast.StructLit{}}
		toolValue.Elts = append(toolValue.Elts, formulas)
	}
	formulaValues, ok := formulas.Value.(*ast.StructLit)
	if !ok {
		return fmt.Errorf("_tool.formulas must be a struct")
	}
	formulaRow := fieldsOfStruct(formulas.Value)[rowID]
	if formulaRow == nil {
		formulaRow = &ast.Field{Label: ast.NewIdent(rowID), Value: &ast.StructLit{}}
		formulaValues.Elts = append(formulaValues.Elts, formulaRow)
	}
	formulaRowValues, ok := formulaRow.Value.(*ast.StructLit)
	if !ok {
		return fmt.Errorf("formula row %q must be a struct", rowID)
	}
	if existing := fieldsOfStruct(formulaRow.Value)[column]; existing != nil {
		existing.Value = stringLiteral(expression)
		return nil
	}
	formulaRowValues.Elts = append(formulaRowValues.Elts, &ast.Field{Label: ast.NewIdent(column), Value: stringLiteral(expression)})
	return nil
}

func removeFormula(file *ast.File, rowID, column string) {
	tool := fieldsOfFile(file)["_tool"]
	if tool == nil {
		return
	}
	formulas := fieldsOfStruct(tool.Value)["formulas"]
	if formulas == nil {
		return
	}
	formulaRow := fieldsOfStruct(formulas.Value)[rowID]
	if formulaRow == nil {
		return
	}
	removeNamedField(formulaRow.Value, column)
	if len(orderedStructFields(formulaRow.Value)) == 0 {
		removeNamedField(formulas.Value, rowID)
	}
}

func hasFormulasForRow(file *ast.File, rowID string) bool {
	tool := fieldsOfFile(file)["_tool"]
	if tool == nil {
		return false
	}
	formulas := fieldsOfStruct(tool.Value)["formulas"]
	if formulas == nil {
		return false
	}
	row := fieldsOfStruct(formulas.Value)[rowID]
	return row != nil && len(orderedStructFields(row.Value)) > 0
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

func removeRow(file *ast.File, rowID string) error {
	fields := fieldsOfFile(file)
	rows := requiredField(fields, "rows")
	rowsValue, ok := rows.Value.(*ast.StructLit)
	if !ok {
		return fmt.Errorf("rows must be a struct")
	}
	for index, element := range rowsValue.Elts {
		field, ok := element.(*ast.Field)
		if !ok || labelOf(field) != rowID {
			continue
		}
		rowsValue.Elts = append(rowsValue.Elts[:index], rowsValue.Elts[index+1:]...)
		removeFormulaRow(file, rowID)
		return nil
	}
	return fmt.Errorf("row %q does not exist", rowID)
}

func removeFormulaRow(file *ast.File, rowID string) {
	tool := fieldsOfFile(file)["_tool"]
	if tool == nil {
		return
	}
	formulas := fieldsOfStruct(tool.Value)["formulas"]
	if formulas != nil {
		removeNamedField(formulas.Value, rowID)
	}
}

func insertColumn(file *ast.File, name string, schema, value ast.Expr) error {
	fields := fieldsOfFile(file)
	rowSchema := requiredField(fields, "#Row")
	rowSchemaValue, ok := rowSchema.Value.(*ast.StructLit)
	if !ok {
		return fmt.Errorf("#Row must be a struct")
	}
	if _, exists := fieldsOfStruct(rowSchema.Value)[name]; exists {
		return fmt.Errorf("column %q already exists", name)
	}
	rows := requiredField(fields, "rows")
	for _, row := range orderedStructFields(rows.Value) {
		rowValues, ok := row.Value.(*ast.StructLit)
		if !ok {
			return fmt.Errorf("row %q is not a struct", labelOf(row))
		}
		rowValues.Elts = append(rowValues.Elts, &ast.Field{Label: ast.NewIdent(name), Value: cloneExpression(value)})
	}
	rowSchemaValue.Elts = append(rowSchemaValue.Elts, &ast.Field{Label: ast.NewIdent(name), Value: schema})
	return nil
}

func removeColumn(file *ast.File, name string) error {
	fields := fieldsOfFile(file)
	rowSchema := requiredField(fields, "#Row")
	if !removeNamedField(rowSchema.Value, name) {
		return fmt.Errorf("column %q does not exist", name)
	}
	rows := requiredField(fields, "rows")
	for _, row := range orderedStructFields(rows.Value) {
		removeNamedField(row.Value, name)
	}
	tool := fields["_tool"]
	if tool != nil {
		toolFields := fieldsOfStruct(tool.Value)
		if foreignKeys := toolFields["foreignKeys"]; foreignKeys != nil {
			removeNamedField(foreignKeys.Value, name)
		}
		if formulas := toolFields["formulas"]; formulas != nil {
			removeNamedField(formulas.Value, name)
		}
	}
	if export := fields["_export"]; export != nil {
		exportFields := fieldsOfStruct(export.Value)
		if csv := exportFields["csv"]; csv != nil {
			csvFields := fieldsOfStruct(csv.Value)
			if columns := csvFields["columns"]; columns != nil {
				removeStringListValue(columns.Value, name)
			}
		}
	}
	return nil
}

func renameColumn(file *ast.File, oldName, newName string) error {
	fields := fieldsOfFile(file)
	rowSchema := requiredField(fields, "#Row")
	if _, exists := fieldsOfStruct(rowSchema.Value)[newName]; exists {
		return fmt.Errorf("column %q already exists", newName)
	}
	if !renameNamedField(rowSchema.Value, oldName, newName) {
		return fmt.Errorf("column %q does not exist", oldName)
	}
	rows := requiredField(fields, "rows")
	for _, row := range orderedStructFields(rows.Value) {
		renameNamedField(row.Value, oldName, newName)
	}
	tool := fields["_tool"]
	if tool != nil {
		toolFields := fieldsOfStruct(tool.Value)
		if foreignKeys := toolFields["foreignKeys"]; foreignKeys != nil {
			renameNamedField(foreignKeys.Value, oldName, newName)
		}
		if formulas := toolFields["formulas"]; formulas != nil {
			renameNamedField(formulas.Value, oldName, newName)
		}
	}
	if export := fields["_export"]; export != nil {
		exportFields := fieldsOfStruct(export.Value)
		if csv := exportFields["csv"]; csv != nil {
			csvFields := fieldsOfStruct(csv.Value)
			if columns := csvFields["columns"]; columns != nil {
				renameStringListValue(columns.Value, oldName, newName)
			}
		}
	}
	return nil
}

func moveColumn(file *ast.File, name string, position int) error {
	fields := fieldsOfFile(file)
	rowSchema := requiredField(fields, "#Row")
	rowSchemaValue, ok := rowSchema.Value.(*ast.StructLit)
	if !ok {
		return fmt.Errorf("#Row must be a struct")
	}
	index := indexOfNamedField(rowSchemaValue, name)
	if index < 0 {
		return fmt.Errorf("column %q does not exist", name)
	}
	if position < 0 || position >= len(rowSchemaValue.Elts) {
		return fmt.Errorf("column position %d is out of range", position)
	}
	moveElement(rowSchemaValue.Elts, index, position)
	if export := fields["_export"]; export != nil {
		exportFields := fieldsOfStruct(export.Value)
		if csv := exportFields["csv"]; csv != nil {
			csvFields := fieldsOfStruct(csv.Value)
			if columns := csvFields["columns"]; columns != nil {
				moveStringListValue(columns.Value, name, position)
			}
		}
	}
	return nil
}

func indexOfNamedField(structLiteral *ast.StructLit, name string) int {
	for index, element := range structLiteral.Elts {
		field, ok := element.(*ast.Field)
		if ok && labelOf(field) == name {
			return index
		}
	}
	return -1
}

func moveElement(elements []ast.Decl, from, to int) {
	if from == to {
		return
	}
	element := elements[from]
	if from < to {
		copy(elements[from:to], elements[from+1:to+1])
	} else {
		copy(elements[to+1:from+1], elements[to:from])
	}
	elements[to] = element
}

func removeNamedField(expression ast.Expr, name string) bool {
	structLiteral, ok := expression.(*ast.StructLit)
	if !ok {
		return false
	}
	for index, element := range structLiteral.Elts {
		field, ok := element.(*ast.Field)
		if ok && labelOf(field) == name {
			structLiteral.Elts = append(structLiteral.Elts[:index], structLiteral.Elts[index+1:]...)
			return true
		}
	}
	return false
}

func renameNamedField(expression ast.Expr, oldName, newName string) bool {
	structLiteral, ok := expression.(*ast.StructLit)
	if !ok {
		return false
	}
	for _, element := range structLiteral.Elts {
		field, ok := element.(*ast.Field)
		if ok && labelOf(field) == oldName {
			field.Label = ast.NewIdent(newName)
			return true
		}
	}
	return false
}

func removeStringListValue(expression ast.Expr, value string) {
	list, ok := expression.(*ast.ListLit)
	if !ok {
		return
	}
	for index := 0; index < len(list.Elts); index++ {
		if matchesString(list.Elts[index], value) {
			list.Elts = append(list.Elts[:index], list.Elts[index+1:]...)
			return
		}
	}
}

func renameStringListValue(expression ast.Expr, oldValue, newValue string) {
	list, ok := expression.(*ast.ListLit)
	if !ok {
		return
	}
	for index, element := range list.Elts {
		if matchesString(element, oldValue) {
			list.Elts[index] = stringLiteral(newValue)
			return
		}
	}
}

func moveStringListValue(expression ast.Expr, value string, position int) {
	list, ok := expression.(*ast.ListLit)
	if !ok {
		return
	}
	from := -1
	for index, element := range list.Elts {
		if matchesString(element, value) {
			from = index
			break
		}
	}
	if from < 0 {
		return
	}
	// The export list may omit columns. Preserve its special id column at the
	// start when present, then clamp the table position to this list.
	if len(list.Elts) > 0 && matchesString(list.Elts[0], "id") {
		position++
	}
	if position >= len(list.Elts) {
		position = len(list.Elts) - 1
	}
	moveExpression(list.Elts, from, position)
}

func moveExpression(elements []ast.Expr, from, to int) {
	if from == to {
		return
	}
	element := elements[from]
	if from < to {
		copy(elements[from:to], elements[from+1:to+1])
	} else {
		copy(elements[to+1:from+1], elements[to:from])
	}
	elements[to] = element
}

func cloneExpression(expression ast.Expr) ast.Expr {
	formatted, err := format.Node(expression)
	if err != nil {
		return expression
	}
	clone, err := parser.ParseExpr("column-value.cue", formatted)
	if err != nil {
		return expression
	}
	return clone
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

func validateCellValue(path string, source []byte, rowID, column string) error {
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
	fieldSchema := packageValue.LookupPath(cue.MakePath(cue.Def("Row"), cue.Str(column)))
	value := packageValue.LookupPath(cue.MakePath(cue.Str("rows"), cue.Str(rowID), cue.Str(column)))
	if err := fieldSchema.Unify(value).Validate(cue.Concrete(true)); err != nil {
		return fmt.Errorf("validate row %q column %q: %w", rowID, column, err)
	}
	return nil
}

func validateAllRows(path string, source []byte) error {
	file, err := parser.ParseFile(path, source, parser.ParseComments, parser.AllErrors)
	if err != nil {
		return fmt.Errorf("parse updated table: %w", err)
	}
	rows := requiredField(fieldsOfFile(file), "rows")
	rowIDs := make([]string, 0, len(orderedStructFields(rows.Value)))
	for _, row := range orderedStructFields(rows.Value) {
		rowIDs = append(rowIDs, labelOf(row))
	}
	return validateRows(path, source, rowIDs)
}

func validateRows(path string, source []byte, rowIDs []string) error {
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
	for _, rowID := range rowIDs {
		row := packageValue.LookupPath(cue.MakePath(cue.Str("rows"), cue.Str(rowID)))
		if err := rowSchema.Unify(row).Validate(cue.Concrete(true)); err != nil {
			return fmt.Errorf("validate row %q: %w", rowID, err)
		}
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
