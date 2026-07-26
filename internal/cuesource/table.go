// Package cuesource parses and prints the editor's restricted CUE source form.
package cuesource

import (
	"fmt"
	"os"
	"sort"
	"strconv"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/parser"
	"cuelang.org/go/cue/token"
)

type Severity string

const SeverityError Severity = "error"

type Diagnostic struct {
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
	Line     int      `json:"line,omitempty"`
	Column   int      `json:"column,omitempty"`
}

type Table struct {
	Path             string       `json:"path"`
	Package          string       `json:"package"`
	Namespace        string       `json:"namespace"`
	Name             string       `json:"name"`
	Columns          []string     `json:"columns"`
	RowIDs           []string     `json:"rowIDs"`
	LocalDefinitions []string     `json:"localDefinitions"`
	Diagnostics      []Diagnostic `json:"diagnostics,omitempty"`
}

func (t Table) LogicalName() string {
	if t.Namespace == "" {
		return t.Name
	}
	if t.Name == "" {
		return t.Namespace
	}
	return t.Namespace + "." + t.Name
}

// ParseTableFile reads the first restricted-table shape from a CUE file. The
// returned table remains useful when the source is structurally incomplete;
// callers can show diagnostics without hiding the file from the project tree.
func ParseTableFile(path string) Table {
	source, err := os.ReadFile(path)
	if err != nil {
		return Table{Path: path, Diagnostics: []Diagnostic{{Severity: SeverityError, Message: fmt.Sprintf("read source: %v", err)}}}
	}
	return ParseTable(path, source)
}

func ParseTable(path string, source []byte) Table {
	table := Table{Path: path}
	file, err := parseFile(path, source)
	if err != nil {
		table.Diagnostics = append(table.Diagnostics, Diagnostic{Severity: SeverityError, Message: err.Error()})
		return table
	}

	fields := topLevelFields(file)
	packageDecl := packageDeclaration(file)
	if packageDecl == nil || packageDecl.Name == nil {
		table.errorAt(token.NoPos, "table file must declare a package")
	} else {
		table.Package = packageDecl.Name.Name
	}

	tool := fields["_tool"]
	if tool == nil {
		table.errorAt(token.NoPos, "table file must declare _tool")
	} else {
		table.readIdentity(tool)
	}

	rowSchema := fields["#Row"]
	if rowSchema == nil {
		table.errorAt(token.NoPos, "table file must declare #Row")
	} else {
		table.Columns = structFieldNames(rowSchema.Value)
		if table.Columns == nil {
			table.errorAt(rowSchema.Pos(), "#Row must be a struct")
		}
	}
	if fields["#PrimaryKey"] == nil {
		table.errorAt(token.NoPos, "table file must declare #PrimaryKey")
	}

	rows := fields["rows"]
	if rows == nil {
		table.errorAt(token.NoPos, "table file must declare rows")
	} else {
		table.RowIDs = structFieldNames(rows.Value)
		if table.RowIDs == nil {
			table.errorAt(rows.Pos(), "rows must be a struct")
		}
	}

	if local := fields["_local"]; local != nil {
		table.LocalDefinitions = definitionNames(local.Value)
		if table.LocalDefinitions == nil {
			table.errorAt(local.Pos(), "_local must be a struct")
		}
	}

	sort.Strings(table.Columns)
	sort.Strings(table.RowIDs)
	sort.Strings(table.LocalDefinitions)
	return table
}

func parseFile(path string, source []byte) (*ast.File, error) {
	return parser.ParseFile(path, source, parser.ParseComments, parser.AllErrors)
}

func (t *Table) readIdentity(tool *ast.Field) {
	toolFields := structFields(tool.Value)
	if toolFields == nil {
		t.errorAt(tool.Pos(), "_tool must be a struct")
		return
	}
	table := toolFields["table"]
	if table == nil {
		t.errorAt(tool.Pos(), "_tool must declare table")
		return
	}
	identityFields := structFields(table.Value)
	if identityFields == nil {
		t.errorAt(table.Pos(), "_tool.table must be a struct")
		return
	}
	t.Namespace = requiredString(t, identityFields["namespace"], "_tool.table.namespace")
	t.Name = requiredString(t, identityFields["name"], "_tool.table.name")
}

func requiredString(table *Table, field *ast.Field, path string) string {
	if field == nil {
		table.errorAt(token.NoPos, path+" must be a string")
		return ""
	}
	value, ok := stringLiteral(field.Value)
	if !ok {
		table.errorAt(field.Pos(), path+" must be a string")
		return ""
	}
	if value == "" {
		table.errorAt(field.Pos(), path+" must not be empty")
	}
	return value
}

func unquote(value string) (string, error) {
	return strconv.Unquote(value)
}

func topLevelFields(file *ast.File) map[string]*ast.Field {
	fields := make(map[string]*ast.Field)
	for _, declaration := range file.Decls {
		field, ok := declaration.(*ast.Field)
		if !ok {
			continue
		}
		name, _, err := ast.LabelName(field.Label)
		if err == nil {
			fields[name] = field
		}
	}
	return fields
}

func packageDeclaration(file *ast.File) *ast.Package {
	for _, declaration := range file.Decls {
		if packageDecl, ok := declaration.(*ast.Package); ok {
			return packageDecl
		}
	}
	return nil
}

func structFields(expression ast.Expr) map[string]*ast.Field {
	structLiteral, ok := expression.(*ast.StructLit)
	if !ok {
		return nil
	}
	fields := make(map[string]*ast.Field)
	for _, declaration := range structLiteral.Elts {
		field, ok := declaration.(*ast.Field)
		if !ok {
			continue
		}
		name, _, err := ast.LabelName(field.Label)
		if err == nil {
			fields[name] = field
		}
	}
	return fields
}

func structFieldNames(expression ast.Expr) []string {
	fields := structFields(expression)
	if fields == nil {
		return nil
	}
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	return names
}

func definitionNames(expression ast.Expr) []string {
	fields := structFields(expression)
	if fields == nil {
		return nil
	}
	names := make([]string, 0, len(fields))
	for name := range fields {
		if len(name) > 1 && name[0] == '#' {
			names = append(names, name[1:])
		}
	}
	return names
}

func (t *Table) errorAt(pos token.Pos, message string) {
	diagnostic := Diagnostic{Severity: SeverityError, Message: message}
	if pos.IsValid() {
		diagnostic.Line = pos.Line()
		diagnostic.Column = pos.Column()
	}
	t.Diagnostics = append(t.Diagnostics, diagnostic)
}
