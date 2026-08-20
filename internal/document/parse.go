package document

import (
	"fmt"
	"os"
	"sort"
	"strconv"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/format"
	"cuelang.org/go/cue/parser"
	"cuelang.org/go/cue/token"
	"datapack/internal/schemamodel"
)

// LoadTable parses a valid restricted table document into its editable model.
// Structural file diagnostics remain the responsibility of cuesource; callers
// use this only after the file has passed that first structural check.
func LoadTable(path string) (Table, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return Table{}, fmt.Errorf("read table source: %w", err)
	}
	file, err := parser.ParseFile(path, source, parser.ParseComments, parser.AllErrors)
	if err != nil {
		return Table{}, fmt.Errorf("parse table source: %w", err)
	}
	fields := fieldsOfFile(file)
	document := Table{
		Path:         path,
		Columns:      []Column{},
		Rows:         []Row{},
		ForeignKeys:  []ForeignKey{},
		Formulas:     []Formula{},
		LocalSchema:  []SchemaDefinition{},
		Exports:      map[string]ExportSettings{},
		CellComments: map[string]map[string]string{},
	}
	displayNames := map[string]string{}
	if editor := fields["_editor"]; editor != nil {
		if columns := fieldsOfStruct(editor.Value)["columns"]; columns != nil {
			for name, metadata := range fieldsOfStruct(columns.Value) {
				if title := fieldsOfStruct(metadata.Value)["title"]; title != nil {
					displayNames[name] = stringOf(title.Value)
				}
			}
		}
	}
	if packageDecl := packageOfFile(file); packageDecl != nil && packageDecl.Name != nil {
		document.Package = packageDecl.Name.Name
	}

	toolFields := fieldsOfStruct(requiredField(fields, "_tool").Value)
	identityFields := fieldsOfStruct(requiredField(toolFields, "table").Value)
	document.Namespace = stringOf(requiredField(identityFields, "namespace").Value)
	document.Name = stringOf(requiredField(identityFields, "name").Value)
	if foreignKeys := toolFields["foreignKeys"]; foreignKeys != nil {
		for _, field := range orderedStructFields(foreignKeys.Value) {
			metadata := fieldsOfStruct(field.Value)
			target := metadata["table"]
			if target == nil {
				return Table{}, fmt.Errorf("parse foreign key %q: table is required", labelOf(field))
			}
			document.ForeignKeys = append(document.ForeignKeys, ForeignKey{Column: labelOf(field), Table: stringOf(target.Value)})
		}
	}
	if formulas := toolFields["formulas"]; formulas != nil {
		for _, row := range orderedStructFields(formulas.Value) {
			for _, formula := range orderedStructFields(row.Value) {
				expression, ok := formulaExpression(formula.Value)
				if !ok {
					return Table{}, fmt.Errorf("parse formula %q.%q: expression must be a string", labelOf(row), labelOf(formula))
				}
				document.Formulas = append(document.Formulas, Formula{RowID: labelOf(row), Column: labelOf(formula), Expression: expression})
			}
		}
	}
	if editor := fields["_editor"]; editor != nil {
		if comments := fieldsOfStruct(editor.Value)["cellComments"]; comments != nil {
			for _, row := range orderedStructFields(comments.Value) {
				values := map[string]string{}
				for _, column := range orderedStructFields(row.Value) {
					if comment, ok := formulaExpression(column.Value); ok && comment != "" {
						values[labelOf(column)] = comment
					}
				}
				if len(values) > 0 {
					document.CellComments[labelOf(row)] = values
				}
			}
		}
	}
	if export := fields["_export"]; export != nil {
		for _, format := range orderedStructFields(export.Value) {
			formatFields := fieldsOfStruct(format.Value)
			enabled := formatFields["enabled"]
			if enabled == nil {
				continue
			}
			columns := stringListOf(formatFields["columns"])
			settings := ExportSettings{Enabled: boolOf(enabled.Value), Columns: columns}
			document.Exports[labelOf(format)] = settings
			if labelOf(format) == "csv" && settings.Enabled {
				if len(columns) == 0 {
					return Table{}, fmt.Errorf("parse CSV export: columns must not be empty")
				}
				document.CSVExport = &CSVExport{Columns: columns}
			}
		}
	}

	primaryKey, err := schemamodel.Parse(requiredField(fields, "#PrimaryKey").Value)
	if err != nil {
		return Table{}, fmt.Errorf("parse #PrimaryKey: %w", err)
	}
	document.PrimaryKey = primaryKey

	rowSchema := requiredField(fields, "#Row")
	for _, field := range orderedStructFields(rowSchema.Value) {
		if field.Constraint == token.OPTION {
			return Table{}, fmt.Errorf("table column %q must not be optional", labelOf(field))
		}
		schema, err := schemamodel.Parse(field.Value)
		if err != nil {
			return Table{}, fmt.Errorf("parse column %q: %w", labelOf(field), err)
		}
		schemaExpression, err := expressionOf(field.Value)
		if err != nil {
			return Table{}, fmt.Errorf("read column %q Schema: %w", labelOf(field), err)
		}
		document.Columns = append(document.Columns, Column{
			Name:             labelOf(field),
			DisplayName:      displayNames[labelOf(field)],
			Schema:           schema,
			SchemaExpression: schemaExpression,
		})
	}

	rows := requiredField(fields, "rows")
	for _, field := range orderedStructFields(rows.Value) {
		if _, ok := field.Value.(*ast.StructLit); !ok {
			return Table{}, fmt.Errorf("parse row %q: row must be a struct", labelOf(field))
		}
		values := fieldsOfStruct(field.Value)
		row := Row{ID: labelOf(field), Values: make(map[string]Cell, len(values))}
		for name, value := range values {
			expression, err := expressionOf(value.Value)
			if err != nil {
				return Table{}, fmt.Errorf("parse row %q column %q: %w", row.ID, name, err)
			}
			row.Values[name] = Cell{Expression: expression, Display: displayOf(value.Value, expression)}
		}
		document.Rows = append(document.Rows, row)
	}

	localDefinitions := localDefinitionKinds(fields)
	if local := fields["_local"]; local != nil {
		for _, field := range orderedStructFields(local.Value) {
			name := labelOf(field)
			if len(name) < 2 || name[0] != '#' {
				continue
			}
			schema, err := schemamodel.Parse(field.Value)
			if err != nil {
				return Table{}, fmt.Errorf("parse local schema %q: %w", name, err)
			}
			expression, err := expressionOf(field.Value)
			if err != nil {
				return Table{}, fmt.Errorf("read local schema %q: %w", name, err)
			}
			document.LocalSchema = append(document.LocalSchema, SchemaDefinition{
				Name:             name[1:],
				Kind:             localDefinitions[name[1:]].Kind,
				Type:             localDefinitions[name[1:]].Type,
				AppliesTo:        localDefinitions[name[1:]].AppliesTo,
				Schema:           schema,
				SchemaExpression: expression,
			})
		}
	}
	sort.Slice(document.LocalSchema, func(i, j int) bool { return document.LocalSchema[i].Name < document.LocalSchema[j].Name })
	return document, nil
}

// LoadPublicSchema parses a public Schema document. Its metadata defines the
// kind; the actual CUE definition determines the structured SchemaNode.
func LoadPublicSchema(path string) (PublicSchema, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return PublicSchema{}, fmt.Errorf("read schema source: %w", err)
	}
	file, err := parser.ParseFile(path, source, parser.ParseComments, parser.AllErrors)
	if err != nil {
		return PublicSchema{}, fmt.Errorf("parse schema source: %w", err)
	}
	fields := fieldsOfFile(file)
	document := PublicSchema{Path: path, Definitions: []SchemaDefinition{}}
	if packageDecl := packageOfFile(file); packageDecl != nil && packageDecl.Name != nil {
		document.Package = packageDecl.Name.Name
	}

	metadata := fieldsOfStruct(requiredField(fields, "_schema").Value)
	definitions := fieldsOfStruct(requiredField(metadata, "definitions").Value)
	for _, metadataField := range orderedFields(definitions) {
		name := labelOf(metadataField)
		definition := requiredField(fields, "#"+name)
		schema, err := schemamodel.Parse(definition.Value)
		if err != nil {
			return PublicSchema{}, fmt.Errorf("parse schema %q: %w", name, err)
		}
		expression, err := expressionOf(definition.Value)
		if err != nil {
			return PublicSchema{}, fmt.Errorf("read schema %q: %w", name, err)
		}
		definitionMetadata := fieldsOfStruct(metadataField.Value)
		document.Definitions = append(document.Definitions, SchemaDefinition{
			Name:             name,
			Kind:             stringOf(requiredField(definitionMetadata, "kind").Value),
			Type:             optionalString(definitionMetadata["type"]),
			AppliesTo:        stringListOf(definitionMetadata["appliesTo"]),
			Schema:           schema,
			SchemaExpression: expression,
		})
	}
	return document, nil
}

type definitionMetadata struct {
	Kind      string
	Type      string
	AppliesTo []string
}

func localDefinitionKinds(fields map[string]*ast.Field) map[string]definitionMetadata {
	result := map[string]definitionMetadata{}
	metadata := fields["_schema"]
	if metadata == nil {
		return result
	}
	metadataFields := fieldsOfStruct(metadata.Value)
	localDefinitions := metadataFields["localDefinitions"]
	if localDefinitions == nil {
		return result
	}
	for _, field := range orderedStructFields(localDefinitions.Value) {
		definition := fieldsOfStruct(field.Value)
		if kind := definition["kind"]; kind != nil {
			result[labelOf(field)] = definitionMetadata{
				Kind:      stringOf(kind.Value),
				Type:      optionalString(definition["type"]),
				AppliesTo: stringListOf(definition["appliesTo"]),
			}
		}
	}
	return result
}

func stringListOf(field *ast.Field) []string {
	if field == nil {
		return nil
	}
	list, ok := field.Value.(*ast.ListLit)
	if !ok {
		return nil
	}
	values := make([]string, 0, len(list.Elts))
	for _, element := range list.Elts {
		values = append(values, stringOf(element))
	}
	return values
}

func optionalString(field *ast.Field) string {
	if field == nil {
		return ""
	}
	return stringOf(field.Value)
}

func requiredField(fields map[string]*ast.Field, name string) *ast.Field {
	field := fields[name]
	if field == nil {
		panic("restricted CUE source was not structurally validated: missing " + name)
	}
	return field
}

func fieldsOfFile(file *ast.File) map[string]*ast.Field {
	fields := map[string]*ast.Field{}
	for _, declaration := range file.Decls {
		if field, ok := declaration.(*ast.Field); ok {
			fields[labelOf(field)] = field
		}
	}
	return fields
}

func fieldsOfStruct(expression ast.Expr) map[string]*ast.Field {
	fields := map[string]*ast.Field{}
	structLiteral, ok := expression.(*ast.StructLit)
	if !ok {
		return fields
	}
	for _, declaration := range structLiteral.Elts {
		if field, ok := declaration.(*ast.Field); ok {
			fields[labelOf(field)] = field
		}
	}
	return fields
}

func orderedStructFields(expression ast.Expr) []*ast.Field {
	structLiteral, ok := expression.(*ast.StructLit)
	if !ok {
		return nil
	}
	fields := make([]*ast.Field, 0, len(structLiteral.Elts))
	for _, declaration := range structLiteral.Elts {
		if field, ok := declaration.(*ast.Field); ok {
			fields = append(fields, field)
		}
	}
	return fields
}

func orderedFields(fields map[string]*ast.Field) []*ast.Field {
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	sort.Strings(names)
	result := make([]*ast.Field, 0, len(names))
	for _, name := range names {
		result = append(result, fields[name])
	}
	return result
}

func packageOfFile(file *ast.File) *ast.Package {
	for _, declaration := range file.Decls {
		if packageDecl, ok := declaration.(*ast.Package); ok {
			return packageDecl
		}
	}
	return nil
}

func labelOf(field *ast.Field) string {
	name, _, err := ast.LabelName(field.Label)
	if err != nil {
		panic(fmt.Sprintf("unsupported restricted CUE label: %v", err))
	}
	return name
}

func stringOf(expression ast.Expr) string {
	literal, ok := expression.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		panic("restricted CUE source was not structurally validated: expected string")
	}
	value, err := strconv.Unquote(literal.Value)
	if err != nil {
		panic(fmt.Sprintf("invalid CUE string: %v", err))
	}
	return value
}

func formulaExpression(expression ast.Expr) (string, bool) {
	literal, ok := expression.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return "", false
	}
	value, err := strconv.Unquote(literal.Value)
	return value, err == nil
}

func boolOf(expression ast.Expr) bool {
	literal, ok := expression.(*ast.BasicLit)
	return ok && literal.Kind == token.TRUE
}

func expressionOf(expression ast.Expr) (string, error) {
	formatted, err := format.Node(expression)
	if err != nil {
		return "", fmt.Errorf("format expression: %w", err)
	}
	return string(formatted), nil
}

func displayOf(expression ast.Expr, fallback string) string {
	if literal, ok := expression.(*ast.BasicLit); ok && literal.Kind == token.STRING {
		if value, ok := displayStringExpression(literal.Value); ok {
			return value
		}
	}
	return fallback
}
