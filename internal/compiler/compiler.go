// Package compiler turns workspace documents into a validated, read-only
// CompiledPackage. All validation problems are reported as diagnostics; a
// Package is only produced when the compile is clean.
package compiler

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"datapack/internal/document"
	"datapack/internal/foreignkey"
	"datapack/internal/formula"
	"datapack/internal/schemaregistry"
)

// Diagnostic describes one compile problem with a source location.
type Diagnostic struct {
	Path    string   `json:"path"`
	Table   string   `json:"table,omitempty"`
	Field   string   `json:"field,omitempty"`
	RowID   string   `json:"rowID,omitempty"`
	Column  string   `json:"column,omitempty"`
	Key     string   `json:"key,omitempty"`
	Args    []string `json:"args,omitempty"`
	Message string   `json:"message"`
}

// Result is the outcome of one full compile.
type Result struct {
	Diagnostics []Diagnostic `json:"diagnostics"`
	Package     *Package     `json:"package,omitempty"`
}

// Package is the read-only compiled data an exporter consumes.
type Package struct {
	Tables []Table `json:"tables"`
}

type Table struct {
	Path      string                             `json:"path"`
	Namespace string                             `json:"namespace"`
	Name      string                             `json:"name"`
	CSV       *CSV                               `json:"csv,omitempty"`
	Exports   map[string]document.ExportSettings `json:"exports,omitempty"`
	Rows      []Row                              `json:"rows"`
}

type CSV struct {
	Columns []string `json:"columns"`
}

type Row struct {
	ID     string         `json:"id"`
	Values map[string]any `json:"values"`
}

// CompileAll validates every table and Schema definition and, when clean,
// returns the compiled package.
func CompileAll(publicSchemas []document.PublicSchema, tables []document.Table) Result {
	result := Result{}
	registry := schemaregistry.Build(publicSchemas, tables)
	for _, diagnostic := range registry.Diagnostics {
		result.Diagnostics = append(result.Diagnostics, Diagnostic{Path: diagnostic.Path, Table: diagnostic.Table, Field: diagnostic.Field, Key: diagnostic.Key, Args: diagnostic.Args, Message: diagnostic.Message})
	}
	fk := foreignkey.Build(tables)
	for _, diagnostic := range fk.Diagnostics {
		result.Diagnostics = append(result.Diagnostics, Diagnostic{Path: diagnostic.Path, Table: diagnostic.Table, RowID: diagnostic.RowID, Column: diagnostic.Column, Key: diagnostic.Key, Args: diagnostic.Args, Message: diagnostic.Message})
	}
	compiled := make([]Table, 0, len(tables))
	for _, table := range tables {
		compiledTable, diagnostics := compileTable(table)
		result.Diagnostics = append(result.Diagnostics, diagnostics...)
		compiled = append(compiled, compiledTable)
	}
	if len(result.Diagnostics) > 0 {
		return result
	}
	sort.Slice(compiled, func(i, j int) bool { return compiledTableName(compiled[i]) < compiledTableName(compiled[j]) })
	result.Package = &Package{Tables: compiled}
	return result
}

func compiledTableName(table Table) string {
	if table.Namespace == "" {
		return table.Name
	}
	return table.Namespace + "." + table.Name
}

func compileTable(table document.Table) (Table, []Diagnostic) {
	compiled := Table{Path: table.Path, Namespace: table.Namespace, Name: table.Name, Exports: table.Exports}
	if table.CSVExport != nil {
		compiled.CSV = &CSV{Columns: table.CSVExport.Columns}
	}
	diagnostics := []Diagnostic{}
	evaluated := formula.Evaluate(table)
	for _, diagnostic := range evaluated.Diagnostics {
		diagnostics = append(diagnostics, Diagnostic{Path: table.Path, Table: tableName(table), RowID: diagnostic.RowID, Column: diagnostic.Column, Message: diagnostic.Message})
	}
	rows, rowDiagnostics := validateRows(table, evaluated.Rows)
	diagnostics = append(diagnostics, rowDiagnostics...)
	for _, row := range table.Rows {
		compiled.Rows = append(compiled.Rows, Row{ID: row.ID, Values: rows[row.ID]})
	}
	return compiled, diagnostics
}

// validateRows unifies the evaluated row values with the table Schema and
// decodes the final concrete values, including defaults.
func validateRows(table document.Table, evaluated map[string]map[string]any) (map[string]map[string]any, []Diagnostic) {
	instances := load.Instances([]string{"."}, &load.Config{Dir: filepath.Dir(table.Path)})
	diagnostics := []Diagnostic{}
	if len(instances) != 1 {
		count := strconv.Itoa(len(instances))
		return nil, append(diagnostics, Diagnostic{Path: table.Path, Table: tableName(table), Key: "compile.expectedOnePackage", Args: []string{count}, Message: fmt.Sprintf("load table package: expected one package, found %s", count)})
	}
	context := cuecontext.New()
	packageValue := context.BuildInstance(instances[0])
	if err := packageValue.Err(); err != nil {
		return nil, append(diagnostics, Diagnostic{Path: table.Path, Table: tableName(table), Message: err.Error()})
	}
	rowSchema := packageValue.LookupPath(cue.MakePath(cue.Def("Row")))
	primaryKey := packageValue.LookupPath(cue.MakePath(cue.Def("PrimaryKey")))
	finalRows := make(map[string]map[string]any, len(table.Rows))
	for _, row := range table.Rows {
		base := len(diagnostics)
		if err := primaryKey.Unify(context.Encode(row.ID)).Validate(cue.Concrete(true)); err != nil {
			diagnostics = append(diagnostics, Diagnostic{Path: table.Path, Table: tableName(table), RowID: row.ID, Key: "compile.invalidID", Args: []string{err.Error()}, Message: "invalid ID: " + err.Error()})
		}
		for column, value := range evaluated[row.ID] {
			fieldSchema := rowSchema.LookupPath(cue.MakePath(cue.Str(column)))
			if err := fieldSchema.Unify(context.Encode(value)).Validate(cue.Concrete(true)); err != nil {
				diagnostics = append(diagnostics, Diagnostic{Path: table.Path, Table: tableName(table), RowID: row.ID, Column: column, Message: err.Error()})
			}
		}
		candidate := rowSchema.Unify(context.Encode(evaluated[row.ID]))
		if err := candidate.Validate(cue.Concrete(true)); err != nil {
			diagnostics = append(diagnostics, Diagnostic{Path: table.Path, Table: tableName(table), RowID: row.ID, Message: err.Error()})
		}
		if len(diagnostics) > base {
			continue
		}
		var decoded map[string]any
		if err := candidate.Decode(&decoded); err != nil {
			diagnostics = append(diagnostics, Diagnostic{Path: table.Path, Table: tableName(table), RowID: row.ID, Message: err.Error()})
			continue
		}
		finalRows[row.ID] = decoded
	}
	return finalRows, diagnostics
}

func tableName(table document.Table) string {
	if table.Namespace == "" {
		return table.Name
	}
	return table.Namespace + "." + table.Name
}
