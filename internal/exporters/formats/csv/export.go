// Package csv writes compiled table data as CSV files. The exporter receives
// a compiler.Package and never re-evaluates formulas or re-validates data.
package csv

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"datapack/internal/compiler"
)

type Result struct {
	Files []string `json:"files"`
}

// Export writes all compiled tables with CSV configuration to destination.
// The destination is replaced as a directory, so callers never observe a
// partially written export set.
func Export(destination string, pkg *compiler.Package) (Result, error) {
	if pkg == nil {
		return Result{}, fmt.Errorf("export requires a compiled package")
	}
	temporary, err := os.MkdirTemp(filepath.Dir(destination), ".datapack-export-*")
	if err != nil {
		return Result{}, fmt.Errorf("create export directory: %w", err)
	}
	defer os.RemoveAll(temporary)
	result := Result{}
	for _, table := range pkg.Tables {
		if table.CSV == nil {
			continue
		}
		file := filepath.Join(temporary, tableName(table)+".csv")
		if err := writeTable(file, table); err != nil {
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

func writeTable(path string, table compiler.Table) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create CSV %s: %w", tableName(table), err)
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	knownColumns := make(map[string]bool, len(table.Columns))
	for _, column := range table.Columns {
		knownColumns[column] = true
	}
	if err := writer.Write(table.CSV.Columns); err != nil {
		return fmt.Errorf("write CSV header: %w", err)
	}
	for _, row := range table.Rows {
		record := make([]string, 0, len(table.CSV.Columns))
		for _, column := range table.CSV.Columns {
			if column == "id" {
				record = append(record, row.ID)
				continue
			}
			if !knownColumns[column] {
				return fmt.Errorf("export table %s: column %q does not exist", tableName(table), column)
			}
			value, exists := row.Values[column]
			if !exists {
				record = append(record, "")
				continue
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

func tableName(table compiler.Table) string {
	if table.Namespace == "" {
		return table.Name
	}
	return table.Namespace + "." + table.Name
}
