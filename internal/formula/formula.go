// Package formula evaluates deterministic same-table formulas.
package formula

import (
	"fmt"
	"math"
	"path/filepath"
	"sort"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"github.com/expr-lang/expr"

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

// Evaluate computes formulas into an in-memory result. It never changes the
// fixed source values in document.Table.
type evaluationState uint8

const (
	formulaPending evaluationState = iota
	formulaVisiting
	formulaSucceeded
	formulaFailed
)

type formulaEntry struct {
	formula  document.Formula
	analysis formulaAnalysis
	state    evaluationState
}

func Evaluate(table document.Table) Result {
	result := Result{Rows: map[string]map[string]any{}}
	entries := map[Reference]*formulaEntry{}
	columns := map[string]schemamodel.Kind{}
	foreignKeys := map[string]bool{}
	blocked := map[Reference]bool{}
	rowSchema := loadRowSchema(table)
	for _, column := range table.Columns {
		kind := column.Schema.Kind
		if kind == schemamodel.KindReference && rowSchema.Exists() {
			switch rowSchema.LookupPath(cue.MakePath(cue.Str(column.Name))).IncompleteKind() &^ cue.NullKind {
			case cue.IntKind:
				kind = schemamodel.KindInteger
			case cue.NumberKind, cue.FloatKind:
				kind = schemamodel.KindNumber
			case cue.StringKind:
				kind = schemamodel.KindString
			case cue.BoolKind:
				kind = schemamodel.KindBoolean
			case cue.ListKind:
				kind = schemamodel.KindArray
			case cue.StructKind:
				kind = schemamodel.KindStruct
			}
		}
		columns[column.Name] = kind
	}
	for _, key := range table.ForeignKeys {
		foreignKeys[key.Column] = true
	}
	for _, row := range table.Rows {
		result.Rows[row.ID] = map[string]any{}
	}
	fail := func(entry *formulaEntry, message string) {
		if entry.state != formulaFailed {
			result.add(entry.formula, message)
			entry.state = formulaFailed
		}
	}
	for _, f := range table.Formulas {
		if _, ok := result.Rows[f.RowID]; !ok {
			result.add(f, "formula row does not exist")
			continue
		}
		kind, ok := columns[f.Column]
		if !ok {
			result.add(f, "formula column does not exist")
			continue
		}
		ref := Reference{RowID: f.RowID, Column: f.Column}
		if entry, ok := entries[ref]; ok {
			fail(entry, "duplicate formula for cell")
			continue
		}
		entry := &formulaEntry{formula: f}
		entries[ref] = entry
		blocked[ref] = true
		if kind == schemamodel.KindArray || kind == schemamodel.KindStruct {
			fail(entry, "formulas only support scalar or enum columns")
		}
		if foreignKeys[f.Column] {
			fail(entry, "foreign key columns cannot have formulas")
		}
	}
	for _, row := range table.Rows {
		for column, cell := range row.Values {
			ref := Reference{RowID: row.ID, Column: column}
			if entry, ok := entries[ref]; ok {
				fail(entry, "formula and fixed value cannot coexist")
				continue
			}
			value := cuecontext.New().CompileString(cell.Expression)
			var decoded any
			if err := value.Decode(&decoded); err != nil {
				result.Diagnostics = append(result.Diagnostics, Diagnostic{RowID: row.ID, Column: column, Message: err.Error()})
				blocked[ref] = true
				continue
			}
			result.Rows[row.ID][column] = normalizeValueForColumn(decoded, columns[column])
		}
	}
	applyDefaults(table, result.Rows, blocked, rowSchema, columns)
	keys := make([]Reference, 0, len(entries))
	for ref := range entries {
		keys = append(keys, ref)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].RowID == keys[j].RowID {
			return keys[i].Column < keys[j].Column
		}
		return keys[i].RowID < keys[j].RowID
	})
	for _, ref := range keys {
		entry := entries[ref]
		if entry.state == formulaFailed {
			continue
		}
		analysis, err := analyzeFormula(table, entry.formula, true)
		if err != nil {
			fail(entry, err.Error())
			continue
		}
		entry.analysis = analysis
	}
	var evaluate func(Reference) bool
	evaluate = func(ref Reference) bool {
		entry := entries[ref]
		switch entry.state {
		case formulaSucceeded:
			return true
		case formulaFailed:
			return false
		case formulaVisiting:
			fail(entry, "formula has a circular or missing dependency")
			return false
		}
		entry.state = formulaVisiting
		for _, dependency := range entry.analysis.references {
			if _, ok := entries[dependency]; ok {
				if !evaluate(dependency) {
					fail(entry, fmt.Sprintf("formula dependency %s.%s failed", dependency.RowID, dependency.Column))
					return false
				}
			} else if blocked[dependency] {
				fail(entry, fmt.Sprintf("formula input %s.%s failed", dependency.RowID, dependency.Column))
				return false
			}
			if _, present := result.Rows[dependency.RowID][dependency.Column]; !present {
				fail(entry, fmt.Sprintf("formula input %s.%s is missing", dependency.RowID, dependency.Column))
				return false
			}
		}
		options := []expr.Option{expr.Env(map[string]any{})}
		options = append(options, deterministicFunctions()...)
		options = append(options, expr.Function(entry.analysis.cellFunction, func(params ...any) (any, error) {
			row, column := params[0].(string), params[1].(string)
			value, present := result.Rows[row][column]
			if !present {
				return nil, fmt.Errorf("formula input %s.%s is missing", row, column)
			}
			return value, nil
		}, func(row, column string) any { return nil }))
		program, err := expr.Compile(entry.analysis.expression, options...)
		if err != nil {
			fail(entry, err.Error())
			return false
		}
		value, err := expr.Run(program, map[string]any{})
		if err != nil {
			fail(entry, "evaluate formula: "+err.Error())
			return false
		}
		if err := validateFiniteFormulaValue(value); err != nil {
			fail(entry, err.Error())
			return false
		}
		result.Rows[ref.RowID][ref.Column] = normalizeValueForColumn(value, columns[ref.Column])
		entry.state = formulaSucceeded
		return true
	}
	for _, ref := range keys {
		evaluate(ref)
	}
	return result
}

func normalizeValueForColumn(value any, kind schemamodel.Kind) any {
	if kind != schemamodel.KindInteger {
		return value
	}
	number, ok := value.(float64)
	if !ok || math.IsNaN(number) || math.IsInf(number, 0) || math.Trunc(number) != number || number < math.MinInt64 || number >= float64(math.MaxInt64) {
		return value
	}
	return int64(number)
}

func validateFiniteFormulaValue(value any) error {
	number, ok := value.(float64)
	if !ok {
		return nil
	}
	if math.IsNaN(number) {
		return fmt.Errorf("formula result is NaN (possibly caused by division by zero)")
	}
	if math.IsInf(number, 0) {
		return fmt.Errorf("formula result is infinite (possibly caused by division by zero or overflow)")
	}
	return nil
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

var knownFunctionNames = map[string]bool{
	"abs": true, "min": true, "max": true, "round": true, "floor": true,
	"ceil": true, "clamp": true, "pow": true, "len": true,
}

func loadRowSchema(table document.Table) cue.Value {
	if table.Path == "" {
		return cue.Value{}
	}
	instances := load.Instances([]string{"."}, &load.Config{Dir: filepath.Dir(table.Path)})
	if len(instances) != 1 {
		return cue.Value{}
	}
	context := cuecontext.New()
	value := context.BuildInstance(instances[0])
	if value.Err() != nil {
		return cue.Value{}
	}
	return value.LookupPath(cue.MakePath(cue.Def("Row")))
}

func applyDefaults(table document.Table, rows map[string]map[string]any, blocked map[Reference]bool, rowSchema cue.Value, kinds map[string]schemamodel.Kind) {
	if !rowSchema.Exists() {
		return
	}
	for rowID, row := range rows {
		for _, column := range table.Columns {
			if blocked[Reference{RowID: rowID, Column: column.Name}] {
				continue
			}
			if _, exists := row[column.Name]; exists {
				continue
			}
			defaultValue, hasDefault := rowSchema.LookupPath(cue.MakePath(cue.Str(column.Name))).Default()
			if !hasDefault || defaultValue.Validate(cue.Concrete(true)) != nil {
				continue
			}
			var decoded any
			if defaultValue.Decode(&decoded) == nil {
				row[column.Name] = normalizeValueForColumn(decoded, kinds[column.Name])
			}
		}
	}
}

func (result *Result) add(formula document.Formula, message string) {
	result.Diagnostics = append(result.Diagnostics, Diagnostic{RowID: formula.RowID, Column: formula.Column, Message: message})
}
