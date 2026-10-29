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
	"strings"

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
	// Table identities are user-authored. Validate the whole set before even
	// creating a staging directory, so no filename can escape it or overwrite
	// another table's CSV. Check portable names regardless of the host OS.
	seen := map[string]bool{}
	for _, table := range pkg.Tables {
		if table.CSV == nil {
			continue
		}
		name := tableName(table)
		if err := validateFileName(name); err != nil {
			return Result{}, err
		}
		key := strings.ToLower(name)
		if seen[key] {
			return Result{}, fmt.Errorf("duplicate CSV output name %q", name)
		}
		seen[key] = true
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

func validateFileName(name string) error {
	if name == "" || strings.ContainsAny(name, "/\\:<>\"|?*\x00") || strings.HasSuffix(name, ".") || strings.HasSuffix(name, " ") {
		return fmt.Errorf("unsafe CSV output name %q", name)
	}
	for _, character := range name {
		if character < 32 {
			return fmt.Errorf("unsafe CSV output name %q", name)
		}
	}
	// Windows device names remain reserved even with a .csv extension.
	base := strings.ToUpper(strings.SplitN(name, ".", 2)[0])
	if base == "CON" || base == "PRN" || base == "AUX" || base == "NUL" ||
		(len(base) == 4 && (strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT")) && base[3] >= '1' && base[3] <= '9') {
		return fmt.Errorf("unsafe CSV output name %q", name)
	}
	return nil
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
	if value == nil {
		return "", nil
	}
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
	var backupDirectory, backup string
	if info, err := os.Lstat(destination); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("export destination %q is not a directory", destination)
		}
		// Never remove a predictable sibling: it may contain project sources
		// or unrelated user data, not a backup created by this export.
		backupDirectory, err = os.MkdirTemp(filepath.Dir(destination), ".datapack-export-backup-*")
		if err != nil {
			return fmt.Errorf("create export backup directory: %w", err)
		}
		backup = filepath.Join(backupDirectory, "previous")
		if err := os.Rename(destination, backup); err != nil {
			_ = os.RemoveAll(backupDirectory)
			return fmt.Errorf("stage existing export directory: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect existing export directory: %w", err)
	}
	if err := os.Rename(temporary, destination); err != nil {
		if backup != "" {
			if restoreErr := os.Rename(backup, destination); restoreErr != nil {
				// Preserve the only remaining copy when restoration fails.
				return fmt.Errorf("replace export directory: %w; restore failed: %v; previous export retained at %q", err, restoreErr, backup)
			}
			_ = os.RemoveAll(backupDirectory)
		}
		return fmt.Errorf("replace export directory: %w", err)
	}
	if backupDirectory != "" {
		if err := os.RemoveAll(backupDirectory); err != nil {
			return fmt.Errorf("remove previous export directory: %w", err)
		}
	}
	return nil
}

func tableName(table compiler.Table) string {
	if table.Namespace == "" {
		return table.Name
	}
	return table.Namespace + "." + table.Name
}
