package cuesource

import (
	"fmt"
	"os"
	"sort"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/token"
)

type SchemaDefinition struct {
	Name      string   `json:"name"`
	Kind      string   `json:"kind"`
	Type      string   `json:"type,omitempty"`
	AppliesTo []string `json:"appliesTo,omitempty"`
}

type SchemaFile struct {
	Path        string             `json:"path"`
	Package     string             `json:"package"`
	Definitions []SchemaDefinition `json:"definitions"`
	Diagnostics []Diagnostic       `json:"diagnostics,omitempty"`
}

// ParseSchemaFile reads a public Schema file. Public Schema files have a
// package, one CUE definition per registered Schema, and _schema metadata that
// explicitly distinguishes type and constraint Schemas.
func ParseSchemaFile(path string) SchemaFile {
	source, err := os.ReadFile(path)
	if err != nil {
		return SchemaFile{Path: path, Diagnostics: []Diagnostic{{Severity: SeverityError, Message: fmt.Sprintf("read source: %v", err)}}}
	}
	return ParseSchema(path, source)
}

func ParseSchema(path string, source []byte) SchemaFile {
	schema := SchemaFile{Path: path}
	file, err := parseFile(path, source)
	if err != nil {
		schema.Diagnostics = append(schema.Diagnostics, Diagnostic{Severity: SeverityError, Message: err.Error()})
		return schema
	}

	packageDecl := packageDeclaration(file)
	if packageDecl == nil || packageDecl.Name == nil {
		schema.errorAt(token.NoPos, "schema file must declare a package")
	} else {
		schema.Package = packageDecl.Name.Name
	}

	fields := topLevelFields(file)
	if fields["rows"] != nil || fields["_tool"] != nil {
		schema.errorAt(token.NoPos, "public schema file cannot declare rows or _tool")
	}
	metadata := fields["_schema"]
	if metadata == nil {
		schema.errorAt(token.NoPos, "schema file must declare _schema")
		return schema
	}
	metadataFields := structFields(metadata.Value)
	if metadataFields == nil {
		schema.errorAt(metadata.Pos(), "_schema must be a struct")
		return schema
	}
	definitionsField := metadataFields["definitions"]
	if definitionsField == nil {
		schema.errorAt(metadata.Pos(), "_schema must declare definitions")
		return schema
	}
	registered := structFields(definitionsField.Value)
	if registered == nil {
		schema.errorAt(definitionsField.Pos(), "_schema.definitions must be a struct")
		return schema
	}

	declarations := topLevelDefinitions(file)
	for name, declaration := range registered {
		definition := schema.readDefinition(name, declaration)
		if _, exists := declarations[name]; !exists {
			schema.errorAt(declaration.Pos(), "_schema.definitions."+name+" has no matching #"+name+" definition")
			continue
		}
		schema.Definitions = append(schema.Definitions, definition)
	}
	for name, declaration := range declarations {
		if _, exists := registered[name]; !exists {
			schema.errorAt(declaration.Pos(), "#"+name+" must be registered in _schema.definitions")
		}
	}
	sort.Slice(schema.Definitions, func(i, j int) bool { return schema.Definitions[i].Name < schema.Definitions[j].Name })
	return schema
}

func (s *SchemaFile) readDefinition(name string, field *ast.Field) SchemaDefinition {
	definition := SchemaDefinition{Name: name}
	metadata := structFields(field.Value)
	if metadata == nil {
		s.errorAt(field.Pos(), "_schema.definitions."+name+" must be a struct")
		return definition
	}
	definition.Kind = requiredStringForSchema(s, metadata["kind"], "_schema.definitions."+name+".kind")
	switch definition.Kind {
	case "type":
		definition.Type = requiredStringForSchema(s, metadata["type"], "_schema.definitions."+name+".type")
	case "constraint":
		definition.AppliesTo = requiredStringList(s, metadata["appliesTo"], "_schema.definitions."+name+".appliesTo")
		if len(definition.AppliesTo) == 0 {
			s.errorAt(field.Pos(), "_schema.definitions."+name+".appliesTo must not be empty")
		}
	default:
		s.errorAt(field.Pos(), "_schema.definitions."+name+".kind must be type or constraint")
	}
	return definition
}

func requiredStringForSchema(schema *SchemaFile, field *ast.Field, path string) string {
	if field == nil {
		schema.errorAt(token.NoPos, path+" must be a string")
		return ""
	}
	value, ok := stringLiteral(field.Value)
	if !ok || value == "" {
		schema.errorAt(field.Pos(), path+" must be a non-empty string")
		return ""
	}
	return value
}

func requiredStringList(schema *SchemaFile, field *ast.Field, path string) []string {
	if field == nil {
		schema.errorAt(token.NoPos, path+" must be a list of strings")
		return nil
	}
	list, ok := field.Value.(*ast.ListLit)
	if !ok {
		schema.errorAt(field.Pos(), path+" must be a list of strings")
		return nil
	}
	values := make([]string, 0, len(list.Elts))
	for _, element := range list.Elts {
		value, ok := stringLiteral(element)
		if !ok || value == "" {
			schema.errorAt(element.Pos(), path+" must only contain non-empty strings")
			continue
		}
		values = append(values, value)
	}
	return values
}

func stringLiteral(expression ast.Expr) (string, bool) {
	literal, ok := expression.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return "", false
	}
	value, err := unquote(literal.Value)
	return value, err == nil
}

func topLevelDefinitions(file *ast.File) map[string]*ast.Field {
	definitions := make(map[string]*ast.Field)
	for name, field := range topLevelFields(file) {
		if len(name) > 1 && name[0] == '#' {
			definitions[name[1:]] = field
		}
	}
	return definitions
}

func (s *SchemaFile) errorAt(pos token.Pos, message string) {
	diagnostic := Diagnostic{Severity: SeverityError, Message: message}
	if pos.IsValid() {
		diagnostic.Line = pos.Line()
		diagnostic.Column = pos.Column()
	}
	s.Diagnostics = append(s.Diagnostics, diagnostic)
}
