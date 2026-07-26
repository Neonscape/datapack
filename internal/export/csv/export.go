// Package csv writes validated table data as CSV files.
package csv

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"datapack/internal/document"
	"datapack/internal/formula"
)

type Result struct {
	Files []string `json:"files"`
}

// Export writes all enabled CSV tables to destination only after every provided
// table passes its row Schema validation. The destination is replaced as a
// directory, so callers never observe a partially written export set.
func Export(destination string, tables []document.Table) (Result, error) {
	evaluatedRows := make(map[string]map[string]map[string]any, len(tables))
	for _, table := range tables {
		evaluated := formula.Evaluate(table)
		if len(evaluated.Diagnostics) > 0 {
			diagnostic := evaluated.Diagnostics[0]
			return Result{}, fmt.Errorf("validate table %s formula %q.%q: %s", tableName(table), diagnostic.RowID, diagnostic.Column, diagnostic.Message)
		}
		rows, err := validateTable(table, evaluated.Rows)
		if err != nil {
			return Result{}, err
		}
		evaluatedRows[table.Path] = rows
	}
	temporary, err := os.MkdirTemp(filepath.Dir(destination), ".datapack-export-*")
	if err != nil {
		return Result{}, fmt.Errorf("create export directory: %w", err)
	}
	defer os.RemoveAll(temporary)
	result := Result{}
	for _, table := range tables {
		if table.CSVExport == nil {
			continue
		}
		file := filepath.Join(temporary, tableName(table)+".csv")
		if err := writeTable(file, table, evaluatedRows[table.Path]); err != nil {
			return Result{}, err
		}
		result.Files = append(result.Files, filepath.Join(destination, filepath.Base(file)))
	}
	sort.Strings(result.Files)
	if err := replaceDirectory(destination, temporary); err != nil {
		return Result{}, err
	}
	return result, nil
}

func validateTable(table document.Table, rows map[string]map[string]any) (map[string]map[string]any, error) {
	instances := load.Instances([]string{"."}, &load.Config{Dir: filepath.Dir(table.Path)})
	if len(instances) != 1 {
		return nil, fmt.Errorf("validate table %s: expected one package, found %d", tableName(table), len(instances))
	}
	context := cuecontext.New()
	value := context.BuildInstance(instances[0])
	if err := value.Err(); err != nil {
		return nil, fmt.Errorf("validate table %s: %w", tableName(table), err)
	}
	rowSchema := value.LookupPath(cue.MakePath(cue.Def("Row")))
	finalRows := make(map[string]map[string]any, len(table.Rows))
	for _, row := range table.Rows {
		candidate := rowSchema.Unify(context.Encode(rows[row.ID]))
		if err := candidate.Validate(cue.Concrete(true)); err != nil {
			return nil, fmt.Errorf("validate table %s row %q: %w", tableName(table), row.ID, err)
		}
		var decoded map[string]any
		if err := candidate.Decode(&decoded); err != nil {
			return nil, fmt.Errorf("decode table %s row %q: %w", tableName(table), row.ID, err)
		}
		finalRows[row.ID] = decoded
	}
	return finalRows, nil
}

func writeTable(path string, table document.Table, rows map[string]map[string]any) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create CSV %s: %w", tableName(table), err)
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	if err := writer.Write(table.CSVExport.Columns); err != nil {
		return fmt.Errorf("write CSV header: %w", err)
	}
	for _, row := range table.Rows {
		record := make([]string, 0, len(table.CSVExport.Columns))
		for _, column := range table.CSVExport.Columns {
			if column == "id" {
				record = append(record, row.ID)
				continue
			}
			value, exists := rows[row.ID][column]
			if !exists {
				return fmt.Errorf("export table %s row %q: column %q has no value", tableName(table), row.ID, column)
			}
			encoded, err := encodeValue(value)
			if err != nil {
				return fmt.Errorf("export table %s row %q column %q: %w", tableName(table), row.ID, column, err)
			}
			record = append(record, encoded)
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("write CSV row: %w", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("flush CSV: %w", err)
	}
	return nil
}

func encodeValue(value any) (string, error) {
	if stringValue, ok := value.(string); ok {
		return stringValue, nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func replaceDirectory(destination, temporary string) error {
	backup := destination + ".datapack-backup"
	_ = os.RemoveAll(backup)
	if _, err := os.Stat(destination); err == nil {
		if err := os.Rename(destination, backup); err != nil {
			return fmt.Errorf("stage existing export directory: %w", err)
		}
	}
	if err := os.Rename(temporary, destination); err != nil {
		_ = os.Rename(backup, destination)
		return fmt.Errorf("replace export directory: %w", err)
	}
	if err := os.RemoveAll(backup); err != nil {
		return fmt.Errorf("remove previous export directory: %w", err)
	}
	return nil
}

func tableName(table document.Table) string {
	if table.Namespace == "" {
		return table.Name
	}
	return table.Namespace + "." + table.Name
}
