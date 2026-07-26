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
}

type Diagnostic struct {
	Path    string `json:"path"`
	Table   string `json:"table"`
	RowID   string `json:"rowID"`
	Column  string `json:"column"`
	Message string `json:"message"`
}

type Index struct {
	Usages      map[Target][]Usage
	Diagnostics []Diagnostic
}

func Build(tables []document.Table) Index {
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
		for _, key := range table.ForeignKeys {
			for _, row := range table.Rows {
				values, err := referenceValues(row.Values[key.Column].Expression)
				if err != nil {
					index.Diagnostics = append(index.Diagnostics, Diagnostic{Path: table.Path, Table: name, RowID: row.ID, Column: key.Column, Message: err.Error()})
					continue
				}
				for position, value := range values {
					usage := Usage{Path: table.Path, Table: name, RowID: row.ID, Column: key.Column, Index: position}
					target := Target{Table: key.Table, RowID: value}
					index.Usages[target] = append(index.Usages[target], usage)
					if rows, exists := known[key.Table]; !exists || !rows[value] {
						index.Diagnostics = append(index.Diagnostics, Diagnostic{Path: table.Path, Table: name, RowID: row.ID, Column: key.Column, Message: "foreign key references missing row " + key.Table + "." + value})
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
					return left.Column < right.Column
				}
				return left.RowID < right.RowID
			}
			return left.Path < right.Path
		})
	}
	return index
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
