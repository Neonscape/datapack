// Package formula evaluates deterministic same-table formulas.
package formula

import (
	"fmt"
	"path/filepath"
	"sort"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/parser"

	"datapack/internal/document"
	"datapack/internal/schemamodel"
)

type Result struct {
	Rows        map[string]map[string]any
	Diagnostics []Diagnostic
}

type Diagnostic struct {
	RowID   string `json:"rowID"`
	Column  string `json:"column"`
	Message string `json:"message"`
}

// Evaluate computes formulas into an in-memory result. It never changes the
// fixed source values in document.Table.
func Evaluate(table document.Table) Result {
	result := Result{Rows: map[string]map[string]any{}}
	formulaByCell := map[string]document.Formula{}
	columns := map[string]schemamodel.Kind{}
	foreignKeys := map[string]bool{}
	rowExists := map[string]bool{}
	for _, column := range table.Columns {
		columns[column.Name] = column.Schema.Kind
	}
	for _, foreignKey := range table.ForeignKeys {
		foreignKeys[foreignKey.Column] = true
	}
	for _, row := range table.Rows {
		rowExists[row.ID] = true
	}
	for _, formula := range table.Formulas {
		key := formula.RowID + "." + formula.Column
		if !rowExists[formula.RowID] {
			result.add(formula, "formula row does not exist")
			continue
		}
		kind, exists := columns[formula.Column]
		if !exists {
			result.add(formula, "formula column does not exist")
			continue
		}
		if kind == schemamodel.KindArray || kind == schemamodel.KindStruct {
			result.add(formula, "formulas only support scalar or enum columns")
			continue
		}
		if foreignKeys[formula.Column] {
			result.add(formula, "foreign key columns cannot have formulas")
			continue
		}
		formulaByCell[key] = formula
	}
	for _, row := range table.Rows {
		values := map[string]any{}
		for column, cell := range row.Values {
			if _, exists := formulaByCell[row.ID+"."+column]; exists {
				result.Diagnostics = append(result.Diagnostics, Diagnostic{RowID: row.ID, Column: column, Message: "formula and fixed value cannot coexist"})
				continue
			}
			value := cuecontext.New().CompileString(cell.Expression)
			var decoded any
			if err := value.Decode(&decoded); err != nil {
				result.Diagnostics = append(result.Diagnostics, Diagnostic{RowID: row.ID, Column: column, Message: err.Error()})
				continue
			}
			values[column] = decoded
		}
		result.Rows[row.ID] = values
	}
	applyDefaults(table, result.Rows)

	pending := formulaByCell
	for len(pending) > 0 {
		progress := false
		for _, key := range sortedFormulaKeys(pending) {
			formula := pending[key]
			if err := validateReferences(table, formula, foreignKeys); err != nil {
				result.add(formula, err.Error())
				delete(pending, key)
				continue
			}
			program, err := expr.Compile(formula.Expression, expr.Env(environment(table, result.Rows, formula.RowID)), expr.AllowUndefinedVariables(), expr.DisableAllBuiltins())
			if err != nil {
				result.add(formula, err.Error())
				delete(pending, key)
				continue
			}
			value, err := expr.Run(program, environment(table, result.Rows, formula.RowID))
			if err != nil {
				continue // A formula dependency may be evaluated in a later pass.
			}
			result.Rows[formula.RowID][formula.Column] = value
			delete(pending, key)
			progress = true
		}
		if progress {
			continue
		}
		for _, key := range sortedFormulaKeys(pending) {
			result.add(pending[key], "formula has a circular or missing dependency")
		}
		break
	}
	return result
}

func applyDefaults(table document.Table, rows map[string]map[string]any) {
	instances := load.Instances([]string{"."}, &load.Config{Dir: filepath.Dir(table.Path)})
	if len(instances) != 1 {
		return
	}
	context := cuecontext.New()
	value := context.BuildInstance(instances[0])
	if value.Err() != nil {
		return
	}
	rowSchema := value.LookupPath(cue.MakePath(cue.Def("Row")))
	for _, row := range rows {
		for _, column := range table.Columns {
			if _, exists := row[column.Name]; exists {
				continue
			}
			defaultValue, hasDefault := rowSchema.LookupPath(cue.MakePath(cue.Str(column.Name))).Default()
			if !hasDefault || defaultValue.Validate(cue.Concrete(true)) != nil {
				continue
			}
			var decoded any
			if defaultValue.Decode(&decoded) == nil {
				row[column.Name] = decoded
			}
		}
	}
}

func (result *Result) add(formula document.Formula, message string) {
	result.Diagnostics = append(result.Diagnostics, Diagnostic{RowID: formula.RowID, Column: formula.Column, Message: message})
}

func environment(table document.Table, rows map[string]map[string]any, rowID string) map[string]any {
	environment := map[string]any{}
	for column, value := range rows[rowID] {
		environment[column] = value
	}
	environment[table.Namespace] = map[string]any{table.Name: rows}
	return environment
}

func validateReferences(table document.Table, formula document.Formula, foreignKeys map[string]bool) error {
	tree, err := parser.Parse(formula.Expression)
	if err != nil {
		return err
	}
	var validationError error
	ast.Walk(&tree.Node, formulaReferenceVisitor{table: table, foreignKeys: foreignKeys, err: &validationError})
	return validationError
}

type formulaReferenceVisitor struct {
	table       document.Table
	foreignKeys map[string]bool
	err         *error
}

func (visitor formulaReferenceVisitor) Visit(node *ast.Node) {
	if *visitor.err != nil {
		return
	}
	if identifier, ok := (*node).(*ast.IdentifierNode); ok && visitor.foreignKeys[identifier.Value] {
		*visitor.err = fmt.Errorf("formulas cannot read foreign key column %q", identifier.Value)
		return
	}
	member, ok := (*node).(*ast.MemberNode)
	if !ok {
		return
	}
	path := memberPath(member)
	if len(path) == 0 {
		return
	}
	if visitor.foreignKeys[path[0]] {
		*visitor.err = fmt.Errorf("formulas cannot read foreign key column %q", path[0])
		return
	}
	if len(path) < 4 {
		return
	}
	if len(path) != 4 || path[0] != visitor.table.Namespace || path[1] != visitor.table.Name {
		*visitor.err = fmt.Errorf("formula reference %q must target the current table", joinPath(path))
	}
}

func memberPath(node ast.Node) []string {
	switch value := node.(type) {
	case *ast.IdentifierNode:
		return []string{value.Value}
	case *ast.MemberNode:
		path := memberPath(value.Node)
		property, ok := value.Property.(*ast.StringNode)
		if !ok {
			return nil
		}
		return append(path, property.Value)
	default:
		return nil
	}
}

func joinPath(path []string) string {
	result := ""
	for index, value := range path {
		if index > 0 {
			result += "."
		}
		result += value
	}
	return result
}

func sortedFormulaKeys(formulas map[string]document.Formula) []string {
	keys := make([]string, 0, len(formulas))
	for key := range formulas {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
