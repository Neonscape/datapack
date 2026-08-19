// Package formula evaluates deterministic same-table formulas.
package formula

import (
	"fmt"
	"math"
	"path/filepath"
	"regexp"
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

// Reference identifies one same-table cell a formula reads.
type Reference struct {
	RowID  string `json:"rowID"`
	Column string `json:"column"`
}

// References returns the cells a formula reads, in source order. Same-row
// references carry the formula's own row ID; cross-row references carry the
// referenced row ID.
func References(table document.Table, formula document.Formula) []Reference {
	tree, err := parser.Parse(formula.Expression)
	if err != nil {
		return nil
	}
	columns := map[string]bool{}
	for _, column := range table.Columns {
		columns[column.Name] = true
	}
	collector := referenceCollector{table: table, columns: columns, formula: formula}
	ast.Walk(&tree.Node, &collector)
	return collector.references
}

type referenceCollector struct {
	table      document.Table
	columns    map[string]bool
	formula    document.Formula
	references []Reference
	seen       map[string]bool
}

func (collector *referenceCollector) add(rowID, column string) {
	if collector.seen == nil {
		collector.seen = map[string]bool{}
	}
	key := rowID + "." + column
	if collector.seen[key] {
		return
	}
	collector.seen[key] = true
	collector.references = append(collector.references, Reference{RowID: rowID, Column: column})
}

func (collector *referenceCollector) Visit(node *ast.Node) {
	switch value := (*node).(type) {
	case *ast.IdentifierNode:
		if collector.columns[value.Value] {
			collector.add(collector.formula.RowID, value.Value)
		}
	case *ast.MemberNode:
		path := memberPath(value)
		if len(path) == 4 && path[0] == collector.table.Namespace && path[1] == collector.table.Name {
			collector.add(path[2], path[3])
		}
	}
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
			kind := columns[formula.Column]
			if err := validateReferences(table, formula, foreignKeys); err != nil {
				result.add(formula, err.Error())
				delete(pending, key)
				continue
			}
			if name := unknownFunction(formula.Expression); name != "" {
				result.add(formula, fmt.Sprintf("unknown formula function %q", name))
				delete(pending, key)
				continue
			}
			options := []expr.Option{expr.Env(environment(table, result.Rows, formula.RowID)), expr.AllowUndefinedVariables()}
			options = append(options, deterministicFunctions()...)
			program, err := expr.Compile(formula.Expression, options...)
			if err != nil {
				result.add(formula, err.Error())
				delete(pending, key)
				continue
			}
			value, err := expr.Run(program, environment(table, result.Rows, formula.RowID))
			if err != nil {
				continue // A formula dependency may be evaluated in a later pass.
			}
			value = normalizeFormulaValue(value, kind)
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

func normalizeFormulaValue(value any, kind schemamodel.Kind) any {
	if kind != schemamodel.KindInteger {
		return value
	}
	number, ok := value.(float64)
	if !ok || math.IsNaN(number) || math.IsInf(number, 0) || math.Trunc(number) != number {
		return value
	}
	return int64(number)
}

func deterministicFunctions() []expr.Option {
	return []expr.Option{
		expr.DisableAllBuiltins(),
		expr.Function("abs", func(params ...any) (any, error) {
			switch value := params[0].(type) {
			case int:
				if value < 0 {
					return -value, nil
				}
				return value, nil
			case int64:
				if value < 0 {
					return -value, nil
				}
				return value, nil
			case float64:
				return math.Abs(value), nil
			default:
				return nil, fmt.Errorf("abs expects a number")
			}
		}, func(value int) int { return value }, func(value int64) int64 { return value }, func(value float64) float64 { return value }),
		expr.Function("min", func(params ...any) (any, error) { return numericMinMax(params, false) }, func(left, right int) int { return left }, func(left, right int64) int64 { return left }, func(left, right float64) float64 { return left }),
		expr.Function("max", func(params ...any) (any, error) { return numericMinMax(params, true) }, func(left, right int) int { return left }, func(left, right int64) int64 { return left }, func(left, right float64) float64 { return left }),
		expr.Function("round", func(params ...any) (any, error) { return math.Round(toFloat(params[0])), nil }, func(value float64) float64 { return value }, func(value int) float64 { return float64(value) }),
		expr.Function("floor", func(params ...any) (any, error) { return math.Floor(toFloat(params[0])), nil }, func(value float64) float64 { return value }, func(value int) float64 { return float64(value) }),
		expr.Function("ceil", func(params ...any) (any, error) { return math.Ceil(toFloat(params[0])), nil }, func(value float64) float64 { return value }, func(value int) float64 { return float64(value) }),
		expr.Function("clamp", func(params ...any) (any, error) {
			value, low, high := toFloat(params[0]), toFloat(params[1]), toFloat(params[2])
			return math.Min(math.Max(value, low), high), nil
		}, func(value, low, high float64) float64 { return value }, func(value, low, high int) float64 { return float64(value) }),
		expr.Function("pow", func(params ...any) (any, error) { return math.Pow(toFloat(params[0]), toFloat(params[1])), nil }, func(left, right float64) float64 { return left }, func(left, right int) float64 { return float64(left) }),
		expr.Function("len", func(params ...any) (any, error) {
			switch value := params[0].(type) {
			case string:
				return len([]rune(value)), nil
			case []any:
				return len(value), nil
			default:
				return nil, fmt.Errorf("len expects a string or array")
			}
		}, func(value string) int { return len(value) }, func(value []any) int { return len(value) }),
	}
}

func toFloat(value any) float64 {
	switch number := value.(type) {
	case int:
		return float64(number)
	case int64:
		return float64(number)
	case float64:
		return number
	default:
		return math.NaN()
	}
}

func numericMinMax(params []any, maximum bool) (any, error) {
	if len(params) == 2 {
		switch left := params[0].(type) {
		case int:
			if right, ok := params[1].(int); ok {
				if maximum && left > right || !maximum && left < right {
					return left, nil
				}
				return right, nil
			}
		case int64:
			if right, ok := params[1].(int64); ok {
				if maximum && left > right || !maximum && left < right {
					return left, nil
				}
				return right, nil
			}
		}
	}
	left, right := toFloat(params[0]), toFloat(params[1])
	if maximum {
		return math.Max(left, right), nil
	}
	return math.Min(left, right), nil
}

var functionCallPattern = regexp.MustCompile(`\b([A-Za-z_][A-Za-z0-9_]*)\s*\(`)

var knownFunctionNames = map[string]bool{
	"abs": true, "min": true, "max": true, "round": true, "floor": true,
	"ceil": true, "clamp": true, "pow": true, "len": true,
}

func unknownFunction(expression string) string {
	for _, match := range functionCallPattern.FindAllStringSubmatch(expression, -1) {
		if !knownFunctionNames[match[1]] {
			return match[1]
		}
	}
	return ""
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
