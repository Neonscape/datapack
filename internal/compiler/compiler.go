// Package compiler turns workspace documents into a validated, read-only
// CompiledPackage. All validation problems are reported as diagnostics; a
// Package is only produced when the compile is clean.
package compiler

import (
	"datapack/internal/perftrace"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"datapack/internal/cuerows"
	"datapack/internal/cueschema"
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
	InputFingerprint string       `json:"-"`
	Diagnostics      []Diagnostic `json:"diagnostics"`
	Package          *Package     `json:"package,omitempty"`
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
	Columns   []string                           `json:"columns"`
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
	return CompileAllProgress(publicSchemas, tables, nil)
}
func CompileAllProgress(publicSchemas []document.PublicSchema, tables []document.Table, progress func(string, int, int)) Result {
	defer perftrace.Start("compiler.CompileAll")()
	result := Result{}
	registry := schemaregistry.Build(publicSchemas, tables)
	for _, diagnostic := range registry.Diagnostics {
		result.Diagnostics = append(result.Diagnostics, Diagnostic{Path: diagnostic.Path, Table: diagnostic.Table, Field: diagnostic.Field, Key: diagnostic.Key, Args: diagnostic.Args, Message: diagnostic.Message})
	}
	compiled := make([]Table, 0, len(tables))
	resolved := foreignkey.ResolvedRows{}
	total, done := 0, 0
	for _, table := range tables {
		total += len(table.Rows)
	}
	for _, table := range tables {
		if progress != nil {
			progress("schema", done, total)
		}
		compiledTable, diagnostics := compileTableProgress(table, func(n int) {
			if progress != nil {
				progress("validate", done+n, total)
			}
		})
		done += len(table.Rows)
		result.Diagnostics = append(result.Diagnostics, diagnostics...)
		compiled = append(compiled, compiledTable)
		resolved[table.Path] = map[string]map[string]any{}
		for _, row := range compiledTable.Rows {
			resolved[table.Path][row.ID] = row.Values
		}
	}
	fk := foreignkey.BuildResolved(tables, resolved)
	for _, diagnostic := range fk.Diagnostics {
		result.Diagnostics = append(result.Diagnostics, Diagnostic{Path: diagnostic.Path, Table: diagnostic.Table, RowID: diagnostic.RowID, Column: diagnostic.Column, Key: diagnostic.Key, Args: diagnostic.Args, Message: diagnostic.Message})
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
	return compileTableProgress(table, nil)
}
func compileTableProgress(table document.Table, progress func(int)) (Table, []Diagnostic) {
	columns := []string{"id"}
	for _, column := range table.Columns {
		columns = append(columns, column.Name)
	}
	compiled := Table{Path: table.Path, Namespace: table.Namespace, Name: table.Name, Exports: table.Exports, Columns: columns}
	if table.CSVExport != nil {
		compiled.CSV = &CSV{Columns: table.CSVExport.Columns}
	}
	diagnostics := []Diagnostic{}
	loaded, loadErr := cueschema.Load(table, table.Package)
	var evaluated formula.Result
	if loadErr == nil {
		evaluated = formula.EvaluateWithSchema(table, loaded.LookupPath(cue.MakePath(cue.Def("Row"))))
	} else {
		evaluated = formula.Evaluate(table)
	}
	for _, diagnostic := range evaluated.Diagnostics {
		diagnostics = append(diagnostics, Diagnostic{Path: table.Path, Table: tableName(table), RowID: diagnostic.RowID, Column: diagnostic.Column, Message: diagnostic.Message})
	}
	var rows map[string]map[string]any
	var rowDiagnostics []Diagnostic
	if loadErr == nil {
		rows, rowDiagnostics = validateRowsProgress(table, evaluated.Rows, progress, loaded)
	} else {
		rows, rowDiagnostics = validateRowsProgress(table, evaluated.Rows, progress)
	}
	diagnostics = append(diagnostics, rowDiagnostics...)
	for _, row := range table.Rows {
		compiled.Rows = append(compiled.Rows, Row{ID: row.ID, Values: rows[row.ID]})
	}
	return compiled, diagnostics
}

// validateRows unifies the evaluated row values with the table Schema and
// decodes the final concrete values, including defaults.
func validateRows(table document.Table, evaluated map[string]map[string]any, provided ...cue.Value) (map[string]map[string]any, []Diagnostic) {
	return validateRowsProgress(table, evaluated, nil, provided...)
}
func validateRowsProgress(table document.Table, evaluated map[string]map[string]any, progress func(int), provided ...cue.Value) (map[string]map[string]any, []Diagnostic) {
	defer perftrace.Start("compiler.validateRows")()
	diagnostics := []Diagnostic{}
	var packageValue cue.Value
	if len(provided) > 0 {
		packageValue = provided[0]
	} else {
		instances := load.Instances([]string{"."}, &load.Config{Dir: filepath.Dir(table.Path), Package: table.Package})
		if len(instances) != 1 {
			count := strconv.Itoa(len(instances))
			return nil, append(diagnostics, Diagnostic{Path: table.Path, Table: tableName(table), Key: "compile.expectedOnePackage", Args: []string{count}, Message: fmt.Sprintf("load table package: expected one package, found %s", count)})
		}
		context := cuecontext.New()
		packageValue = context.BuildInstance(instances[0])
		if err := packageValue.Err(); err != nil {
			return nil, append(diagnostics, Diagnostic{Path: table.Path, Table: tableName(table), Message: err.Error()})
		}

	}
	rowSchema := packageValue.LookupPath(cue.MakePath(cue.Def("Row")))
	primaryKey := packageValue.LookupPath(cue.MakePath(cue.Def("PrimaryKey")))
	if !primaryKey.Exists() || primaryKey.IncompleteKind()&^cue.NullKind != cue.StringKind {
		diagnostics = append(diagnostics, Diagnostic{Path: table.Path, Table: tableName(table), Field: "#PrimaryKey", Message: "primary key must be a string schema"})
	}
	for _, key := range table.ForeignKeys {
		field := rowSchema.LookupPath(cue.MakePath(cue.Str(key.Column)))
		if field.Exists() {
			if err := foreignkey.ValidateColumnSchema(field); err != nil {
				diagnostics = append(diagnostics, Diagnostic{Path: table.Path, Table: tableName(table), Column: key.Column, Message: err.Error()})
			}
		}
	}
	formulaColumns := make(map[string]map[string]bool)
	for _, formula := range table.Formulas {
		if formulaColumns[formula.RowID] == nil {
			formulaColumns[formula.RowID] = make(map[string]bool)
		}
		formulaColumns[formula.RowID][formula.Column] = true
	}
	finalRows := make(map[string]map[string]any, len(table.Rows))
	for index, row := range table.Rows {
		if progress != nil && index%1000 == 0 {
			progress(index)
		}
		base := len(diagnostics)
		values := exactValidationValues(row, evaluated[row.ID], formulaColumns[row.ID], rowSchema)
		candidate, issues := cuerows.Validate(rowSchema, primaryKey, row.ID, values)
		for _, issue := range issues {
			if issue.Kind == cuerows.Identity {
				diagnostics = append(diagnostics, Diagnostic{Path: table.Path, Table: tableName(table), RowID: row.ID, Key: "compile.invalidID", Args: []string{issue.Err.Error()}, Message: "invalid ID: " + issue.Err.Error()})
			} else {
				diagnostics = append(diagnostics, Diagnostic{Path: table.Path, Table: tableName(table), RowID: row.ID, Column: issue.Column, Message: issue.Err.Error()})
			}
		}
		if len(diagnostics) > base {
			continue
		}
		decoded, err := decodeExportValue(candidate)
		if err != nil {
			diagnostics = append(diagnostics, Diagnostic{Path: table.Path, Table: tableName(table), RowID: row.ID, Message: err.Error()})
			continue
		}
		finalRows[row.ID] = decoded.(map[string]any)
	}
	if progress != nil {
		progress(len(table.Rows))
	}
	return finalRows, diagnostics
}

func tableName(table document.Table) string {
	if table.Namespace == "" {
		return table.Name
	}
	return table.Namespace + "." + table.Name
}
