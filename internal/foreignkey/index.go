// Package foreignkey indexes table row references declared in _tool.foreignKeys.
package foreignkey

import (
	"fmt"
	"sort"
	"strconv"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/parser"
	"cuelang.org/go/cue/token"
	"datapack/internal/document"
)

type Target struct {
	Table string
	RowID string
}

type Usage struct {
	Path   string `json:"path"`
	Table  string `json:"table"`
	RowID  string `json:"rowID"`
	Column string `json:"column"`
	Index  int    `json:"index,omitempty"`
	Array  bool   `json:"array,omitempty"`
	Source string `json:"source"`
}

const (
	SourceLiteral = "literal"
	SourceDefault = "default"
	SourceDerived = "derived"
)

// ResolvedRows contains the final, schema-validated row values keyed by path.
type ResolvedRows map[string]map[string]map[string]any

type Diagnostic struct {
	Path    string   `json:"path"`
	Table   string   `json:"table"`
	RowID   string   `json:"rowID"`
	Column  string   `json:"column"`
	Key     string   `json:"key,omitempty"`
	Args    []string `json:"args,omitempty"`
	Message string   `json:"message"`
}

type Index struct {
	Usages      map[Target][]Usage
	Diagnostics []Diagnostic
	// Incomplete means some values could not be resolved; destructive edits
	// must not treat the absence of an indexed usage as proof of no reference.
	Incomplete bool
}

// Summary is the JSON-safe foreign-key status included in a workspace view.
type Summary struct {
	Diagnostics []Diagnostic `json:"diagnostics,omitempty"`
}

func (i Index) Summary() Summary {
	diagnostics := i.Diagnostics
	if diagnostics == nil {
		diagnostics = []Diagnostic{}
	}
	return Summary{Diagnostics: diagnostics}
}

func Build(tables []document.Table) Index {
	values, diagnostics := resolveRows(tables)
	index := BuildResolved(tables, values)
	index.Diagnostics = append(diagnostics, index.Diagnostics...)
	index.Incomplete = index.Incomplete || len(diagnostics) > 0
	return index
}

// BuildResolved validates references after the compiler has produced final
// values. Source documents are retained separately for safe rewrite provenance.
func BuildResolved(tables []document.Table, resolved ResolvedRows) Index {
	index := Index{Usages: map[Target][]Usage{}}
	known := map[string]map[string]bool{}
	for _, table := range tables {
		rows := map[string]bool{}
		for _, row := range table.Rows {
			rows[row.ID] = true
		}
		known[tableName(table)] = rows
	}
	for _, table := range tables {
		name := tableName(table)
		columns := map[string]bool{}
		for _, column := range table.Columns {
			columns[column.Name] = true
		}
		for _, key := range table.ForeignKeys {
			if !columns[key.Column] {
				index.Diagnostics = append(index.Diagnostics, Diagnostic{Path: table.Path, Table: name, Column: key.Column, Message: "foreign key column does not exist"})
				index.Incomplete = true
				continue
			}
			if _, exists := known[key.Table]; !exists {
				index.Diagnostics = append(index.Diagnostics, Diagnostic{Path: table.Path, Table: name, Column: key.Column, Message: "foreign key target table does not exist: " + key.Table})
				index.Incomplete = true
				continue
			}
			for _, row := range table.Rows {
				value, present := resolved[table.Path][row.ID][key.Column]
				if !present {
					index.Diagnostics = append(index.Diagnostics, Diagnostic{Path: table.Path, Table: name, RowID: row.ID, Column: key.Column, Message: "final foreign key value is missing or invalid"})
					index.Incomplete = true
					continue
				}
				values, array, err := resolvedReferences(value)
				if err != nil {
					index.Diagnostics = append(index.Diagnostics, Diagnostic{Path: table.Path, Table: name, RowID: row.ID, Column: key.Column, Message: err.Error()})
					index.Incomplete = true
					continue
				}
				source := referenceSource(row.Values[key.Column].Expression, values, array)
				for position, value := range values {
					usage := Usage{Path: table.Path, Table: name, RowID: row.ID, Column: key.Column, Index: position, Array: array, Source: source}
					target := Target{Table: key.Table, RowID: value}
					index.Usages[target] = append(index.Usages[target], usage)
					if rows, exists := known[key.Table]; !exists || !rows[value] {
						location := key.Table + "." + value
						if array {
							location += fmt.Sprintf(" (%s[%d])", key.Column, position)
						}
						index.Diagnostics = append(index.Diagnostics, Diagnostic{Path: table.Path, Table: name, RowID: row.ID, Column: key.Column, Key: "foreignkey.missingRow", Args: []string{location}, Message: "foreign key references missing row " + location})
					}
				}
			}
		}
	}
	for target := range index.Usages {
		sort.Slice(index.Usages[target], func(i, j int) bool {
			left, right := index.Usages[target][i], index.Usages[target][j]
			if left.Path == right.Path {
				if left.RowID == right.RowID {
					if left.Column == right.Column {
						return left.Index < right.Index
					}
					return left.Column < right.Column
				}
				return left.RowID < right.RowID
			}
			return left.Path < right.Path
		})
	}
	return index
}

func resolvedReferences(value any) ([]string, bool, error) {
	switch v := value.(type) {
	case nil:
		return nil, false, nil
	case string:
		return []string{v}, false, nil
	case []any:
		values := make([]string, len(v))
		for i, item := range v {
			text, ok := item.(string)
			if !ok {
				return nil, true, fmt.Errorf("foreign key list item %d must be a string", i)
			}
			values[i] = text
		}
		return values, true, nil
	default:
		return nil, false, fmt.Errorf("foreign key value must be null, a string or list of strings")
	}
}

func referenceSource(expression string, values []string, array bool) string {
	if expression == "" || expression == "null" {
		return SourceDefault
	}
	literals, err := referenceValues(expression)
	if err != nil || isListExpression(expression) != array || len(literals) != len(values) {
		return SourceDerived
	}
	for i, value := range values {
		if literals[i] != value {
			return SourceDerived
		}
	}
	return SourceLiteral
}

func isListExpression(expression string) bool {
	value, err := parser.ParseExpr("foreign-key.cue", expression)
	if err != nil {
		return false
	}
	_, ok := value.(*ast.ListLit)
	return ok
}

func referenceValues(expression string) ([]string, error) {
	if expression == "" {
		return nil, nil
	}
	value, err := parser.ParseExpr("foreign-key.cue", expression)
	if err != nil {
		return nil, fmt.Errorf("parse foreign key value: %w", err)
	}
	if stringValue, ok := stringValue(value); ok {
		return []string{stringValue}, nil
	}
	list, ok := value.(*ast.ListLit)
	if !ok {
		return nil, fmt.Errorf("foreign key value must be a string or list of strings")
	}
	values := make([]string, 0, len(list.Elts))
	for _, element := range list.Elts {
		value, ok := stringValue(element)
		if !ok {
			return nil, fmt.Errorf("foreign key list must contain only strings")
		}
		values = append(values, value)
	}
	return values, nil
}

func stringValue(expression ast.Expr) (string, bool) {
	literal, ok := expression.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return "", false
	}
	value, err := strconv.Unquote(literal.Value)
	return value, err == nil
}

func tableName(table document.Table) string {
	if table.Namespace == "" {
		return table.Name
	}
	return table.Namespace + "." + table.Name
}
