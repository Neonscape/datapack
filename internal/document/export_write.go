package document

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/format"
	"cuelang.org/go/cue/parser"
)

// SetTableExports updates the known format entries under _export while
// preserving any unrecognized entries for future exporter modules.
func SetTableExports(path string, exports map[string]ExportSettings) (Table, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return Table{}, fmt.Errorf("read table source: %w", err)
	}
	file, err := parser.ParseFile(path, source, parser.ParseComments, parser.AllErrors)
	if err != nil {
		return Table{}, fmt.Errorf("parse table source: %w", err)
	}
	fields := fieldsOfFile(file)
	exportField := fields["_export"]
	if exportField == nil {
		exportField = &ast.Field{Label: ast.NewIdent("_export"), Value: &ast.StructLit{}}
		file.Decls = append(file.Decls, exportField)
	}
	exportStruct, ok := exportField.Value.(*ast.StructLit)
	if !ok {
		return Table{}, fmt.Errorf("_export must be a struct")
	}
	known := fieldsOfStruct(exportStruct)
	ids := make([]string, 0, len(exports))
	for id := range exports {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		value, err := parser.ParseExpr(path, exportSettingsExpression(exports[id]))
		if err != nil {
			return Table{}, fmt.Errorf("parse %s export settings: %w", id, err)
		}
		if field := known[id]; field != nil {
			field.Value = value
			continue
		}
		exportStruct.Elts = append(exportStruct.Elts, &ast.Field{Label: ast.NewIdent(id), Value: value})
	}
	candidate, err := format.Node(file)
	if err != nil {
		return Table{}, fmt.Errorf("format export settings: %w", err)
	}
	return validateAndReplace(path, candidate)
}

func exportSettingsExpression(settings ExportSettings) string {
	var source strings.Builder
	source.WriteString("{enabled: ")
	source.WriteString(strconv.FormatBool(settings.Enabled))
	if len(settings.Columns) > 0 {
		source.WriteString(", columns: [")
		for index, column := range settings.Columns {
			if index > 0 {
				source.WriteString(", ")
			}
			source.WriteString(strconv.Quote(column))
		}
		source.WriteString("]")
	}
	source.WriteString("}")
	return source.String()
}
