package foreignkey

import (
	"fmt"
	"path/filepath"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"datapack/internal/document"
	"datapack/internal/formula"
)

// Workspace indexing uses the final CUE row context, validating fields
// separately where possible. An unresolved field makes edits fail closed.
func resolveRows(tables []document.Table) (ResolvedRows, []Diagnostic) {
	resolved := ResolvedRows{}
	var diagnostics []Diagnostic
	targets := map[string]bool{}
	for _, table := range tables {
		for _, key := range table.ForeignKeys {
			targets[key.Table] = true
		}
	}
	for _, table := range tables {
		if len(table.ForeignKeys) == 0 && !targets[tableName(table)] {
			continue
		}
		add := func(row, column string, err error) {
			diagnostics = append(diagnostics, Diagnostic{Path: table.Path, Table: tableName(table), RowID: row, Column: column, Message: err.Error()})
		}
		instances := load.Instances([]string{"."}, &load.Config{Dir: filepath.Dir(table.Path), Package: table.Package})
		if len(instances) != 1 {
			add("", "", fmt.Errorf("resolve foreign keys: expected one CUE package"))
			continue
		}
		context := cuecontext.New()
		value := context.BuildInstance(instances[0])
		if err := value.Err(); err != nil {
			add("", "", fmt.Errorf("resolve foreign keys: %w", err))
			continue
		}
		schema := value.LookupPath(cue.MakePath(cue.Def("Row")))
		if targets[tableName(table)] {
			primary := value.LookupPath(cue.MakePath(cue.Def("PrimaryKey")))
			if !primary.Exists() || primary.IncompleteKind()&^cue.NullKind != cue.StringKind {
				add("", "#PrimaryKey", fmt.Errorf("foreign key target primary key must be a string schema"))
			}
		}
		invalidColumns := map[string]bool{}
		for _, key := range table.ForeignKeys {
			field := schema.LookupPath(cue.MakePath(cue.Str(key.Column)))
			if !field.Exists() {
				continue
			} // Missing columns are diagnosed by BuildResolved.
			if err := ValidateColumnSchema(field); err != nil {
				add("", key.Column, err)
				invalidColumns[key.Column] = true
			}
		}
		evaluated := formula.Evaluate(table)
		failed := map[formula.Reference]bool{}
		for _, d := range evaluated.Diagnostics {
			failed[formula.Reference{RowID: d.RowID, Column: d.Column}] = true
		}
		resolved[table.Path] = map[string]map[string]any{}
		for _, row := range table.Rows {
			resolved[table.Path][row.ID] = map[string]any{}
			candidate := schema.Unify(context.Encode(evaluated.Rows[row.ID]))
			for _, key := range table.ForeignKeys {
				if invalidColumns[key.Column] {
					continue
				}
				if failed[formula.Reference{RowID: row.ID, Column: key.Column}] {
					add(row.ID, key.Column, fmt.Errorf("foreign key input failed evaluation"))
					continue
				}
				field := candidate.LookupPath(cue.MakePath(cue.Str(key.Column)))
				if !field.Exists() {
					continue
				} // BuildResolved diagnoses the declaration.
				if err := field.Validate(cue.Concrete(true)); err != nil {
					add(row.ID, key.Column, fmt.Errorf("resolve foreign key: %w", err))
					continue
				}
				var decoded any
				if err := field.Decode(&decoded); err != nil {
					add(row.ID, key.Column, err)
					continue
				}
				resolved[table.Path][row.ID][key.Column] = decoded
			}
		}
	}
	return resolved, diagnostics
}

// ValidateColumnSchema checks the reference shape even for an empty table.
// Actual values still undergo full CUE validation before they enter the index.
func ValidateColumnSchema(field cue.Value) error {
	kind := field.IncompleteKind() &^ cue.NullKind
	if kind == cue.BottomKind && field.IncompleteKind() == cue.NullKind {
		return nil
	}
	if kind == cue.StringKind {
		return nil
	}
	if kind != cue.ListKind {
		return fmt.Errorf("foreign key column must have a string or list-of-strings schema")
	}
	check := func(value cue.Value) error {
		if value.IncompleteKind() != cue.StringKind {
			return fmt.Errorf("foreign key list must have string elements")
		}
		return nil
	}
	if element := field.LookupPath(cue.MakePath(cue.AnyIndex)); element.Exists() {
		if err := check(element); err != nil {
			return err
		}
	}
	if elements, err := field.List(); err == nil {
		for elements.Next() {
			if err := check(elements.Value()); err != nil {
				return err
			}
		}
	}
	return nil
}
