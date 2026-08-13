package document

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/format"
	"cuelang.org/go/cue/parser"
	"cuelang.org/go/cue/token"
)

// SchemaDefinitionUpdate carries the editable metadata and expression of one
// Schema definition. Type applies to type Schemas and AppliesTo applies to
// constraint Schemas.
type SchemaDefinitionUpdate struct {
	Kind             string
	Type             string
	AppliesTo        []string
	SchemaExpression string
}

func validateSchemaDefinitionUpdate(name string, update SchemaDefinitionUpdate) error {
	if !validIdentifier(name) {
		return fmt.Errorf("definition %q must be a valid CUE identifier", name)
	}
	switch update.Kind {
	case "type":
		if strings.TrimSpace(update.Type) == "" {
			return fmt.Errorf("type Schema %q must declare a type", name)
		}
	case "constraint":
		if len(update.AppliesTo) == 0 {
			return fmt.Errorf("constraint Schema %q must declare appliesTo", name)
		}
	default:
		return fmt.Errorf("Schema kind must be type or constraint")
	}
	return nil
}

func parseSchemaDefinitionExpression(path, expression string) (ast.Expr, error) {
	parsed, err := parser.ParseExpr(path, expression)
	if err != nil {
		return nil, fmt.Errorf("parse Schema expression: %w", err)
	}
	return parsed, nil
}

func schemaDefinitionMetadata(update SchemaDefinitionUpdate) *ast.StructLit {
	fields := []ast.Decl{
		&ast.Field{Label: ast.NewIdent("kind"), Value: stringLiteral(update.Kind)},
	}
	if update.Kind == "type" {
		fields = append(fields, &ast.Field{Label: ast.NewIdent("type"), Value: stringLiteral(update.Type)})
	} else {
		appliesTo := &ast.ListLit{}
		for _, target := range update.AppliesTo {
			appliesTo.Elts = append(appliesTo.Elts, stringLiteral(target))
		}
		fields = append(fields, &ast.Field{Label: ast.NewIdent("appliesTo"), Value: appliesTo})
	}
	return &ast.StructLit{Elts: fields}
}

func readSourceFile(path string) (*ast.File, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read schema source: %w", err)
	}
	file, err := parser.ParseFile(path, source, parser.ParseComments, parser.AllErrors)
	if err != nil {
		return nil, fmt.Errorf("parse schema source: %w", err)
	}
	return file, nil
}

// upsertPublicDefinition adds or replaces the #Name declaration and its
// _schema.definitions metadata entry.
func upsertPublicDefinition(file *ast.File, name string, update SchemaDefinitionUpdate, schema ast.Expr, create bool) error {
	fields := fieldsOfFile(file)
	declaration := fields["#"+name]
	if create {
		if declaration != nil {
			return fmt.Errorf("Schema definition #%s already exists", name)
		}
		file.Decls = append(file.Decls, &ast.Field{Label: ast.NewIdent("#" + name), Value: schema})
	} else {
		if declaration == nil {
			return fmt.Errorf("Schema definition #%s does not exist", name)
		}
		declaration.Value = schema
	}
	schemaField := fieldsOfFile(file)["_schema"]
	if schemaField == nil {
		return fmt.Errorf("schema file must declare _schema")
	}
	schemaValue, ok := schemaField.Value.(*ast.StructLit)
	if !ok {
		return fmt.Errorf("_schema must be a struct")
	}
	definitionsField := fieldsOfStruct(schemaField.Value)["definitions"]
	var definitionsValue *ast.StructLit
	if definitionsField == nil {
		if !create {
			return fmt.Errorf("_schema must declare definitions")
		}
		definitionsValue = &ast.StructLit{}
		schemaValue.Elts = append(schemaValue.Elts, &ast.Field{Label: ast.NewIdent("definitions"), Value: definitionsValue})
	} else {
		var ok bool
		definitionsValue, ok = definitionsField.Value.(*ast.StructLit)
		if !ok {
			return fmt.Errorf("_schema.definitions must be a struct")
		}
	}
	registered := fieldsOfStruct(definitionsValue)[name]
	if create {
		if registered != nil {
			return fmt.Errorf("Schema definition %q is already registered", name)
		}
		definitionsValue.Elts = append(definitionsValue.Elts, &ast.Field{Label: ast.NewIdent(name), Value: schemaDefinitionMetadata(update)})
		return nil
	}
	if registered == nil {
		return fmt.Errorf("Schema definition %q is not registered in _schema.definitions", name)
	}
	registered.Value = schemaDefinitionMetadata(update)
	return nil
}

// CreatePublicSchemaDefinition appends one #Name declaration and registers it
// in _schema.definitions.
func CreatePublicSchemaDefinition(path, name string, update SchemaDefinitionUpdate) (PublicSchema, error) {
	if err := validateSchemaDefinitionUpdate(name, update); err != nil {
		return PublicSchema{}, err
	}
	schema, err := parseSchemaDefinitionExpression(path, update.SchemaExpression)
	if err != nil {
		return PublicSchema{}, err
	}
	file, err := readSourceFile(path)
	if err != nil {
		return PublicSchema{}, err
	}
	if err := upsertPublicDefinition(file, name, update, schema, true); err != nil {
		return PublicSchema{}, err
	}
	candidate, err := format.Node(file)
	if err != nil {
		return PublicSchema{}, fmt.Errorf("format updated schema: %w", err)
	}
	return validatePublicSchemaAndReplace(path, candidate)
}

// SetPublicSchemaDefinition replaces one definition expression and metadata.
func SetPublicSchemaDefinition(path, name string, update SchemaDefinitionUpdate) (PublicSchema, error) {
	if err := validateSchemaDefinitionUpdate(name, update); err != nil {
		return PublicSchema{}, err
	}
	schema, err := parseSchemaDefinitionExpression(path, update.SchemaExpression)
	if err != nil {
		return PublicSchema{}, err
	}
	file, err := readSourceFile(path)
	if err != nil {
		return PublicSchema{}, err
	}
	if err := upsertPublicDefinition(file, name, update, schema, false); err != nil {
		return PublicSchema{}, err
	}
	candidate, err := format.Node(file)
	if err != nil {
		return PublicSchema{}, fmt.Errorf("format updated schema: %w", err)
	}
	return validatePublicSchemaAndReplace(path, candidate)
}

// DeletePublicSchemaDefinition removes the declaration and its registration.
func DeletePublicSchemaDefinition(path, name string) (PublicSchema, error) {
	file, err := readSourceFile(path)
	if err != nil {
		return PublicSchema{}, err
	}
	fields := fieldsOfFile(file)
	if fields["#"+name] == nil {
		return PublicSchema{}, fmt.Errorf("Schema definition #%s does not exist", name)
	}
	for index, declaration := range file.Decls {
		field, ok := declaration.(*ast.Field)
		if ok && labelOf(field) == "#"+name {
			file.Decls = append(file.Decls[:index], file.Decls[index+1:]...)
			break
		}
	}
	schemaField := fieldsOfFile(file)["_schema"]
	if schemaField == nil {
		return PublicSchema{}, fmt.Errorf("schema file must declare _schema")
	}
	definitionsField := fieldsOfStruct(schemaField.Value)["definitions"]
	if definitionsField == nil {
		return PublicSchema{}, fmt.Errorf("_schema must declare definitions")
	}
	if !removeNamedField(definitionsField.Value, name) {
		return PublicSchema{}, fmt.Errorf("Schema definition %q is not registered in _schema.definitions", name)
	}
	candidate, err := format.Node(file)
	if err != nil {
		return PublicSchema{}, fmt.Errorf("format updated schema: %w", err)
	}
	return validatePublicSchemaAndReplace(path, candidate)
}

// RenamePublicSchemaDefinition renames the declaration and registration and
// rewrites references across the supplied reference files plus the schema file.
func RenamePublicSchemaDefinition(path, oldName, newName string, referencePaths []string) (PublicSchema, error) {
	if !validIdentifier(newName) {
		return PublicSchema{}, fmt.Errorf("definition %q must be a valid CUE identifier", newName)
	}
	err := editSchemaSourceFiles(append(referencePaths, path), func(files map[string]*ast.File) error {
		target := files[path]
		fields := fieldsOfFile(target)
		if fields["#"+oldName] == nil {
			return fmt.Errorf("Schema definition #%s does not exist", oldName)
		}
		if fields["#"+newName] != nil {
			return fmt.Errorf("Schema definition #%s already exists", newName)
		}
		for _, declaration := range target.Decls {
			if field, ok := declaration.(*ast.Field); ok && labelOf(field) == "#"+oldName {
				field.Label = ast.NewIdent("#" + newName)
				break
			}
		}
		schemaField := fields["_schema"]
		if schemaField == nil {
			return fmt.Errorf("schema file must declare _schema")
		}
		definitionsField := fieldsOfStruct(schemaField.Value)["definitions"]
		if definitionsField == nil {
			return fmt.Errorf("_schema must declare definitions")
		}
		if !renameNamedField(definitionsField.Value, oldName, newName) {
			return fmt.Errorf("Schema definition %q is not registered in _schema.definitions", oldName)
		}
		for _, file := range files {
			renameSchemaReferences(file, oldName, newName)
		}
		return nil
	})
	if err != nil {
		return PublicSchema{}, err
	}
	schema, err := LoadPublicSchema(path)
	if err != nil {
		return PublicSchema{}, fmt.Errorf("reload renamed schema: %w", err)
	}
	return schema, nil
}

// ensureLocalDefinition adds or replaces one _local.#Name declaration and its
// _schema.localDefinitions metadata entry.
func ensureLocalDefinition(file *ast.File, name string, update SchemaDefinitionUpdate, schema ast.Expr, create bool) error {
	fields := fieldsOfFile(file)
	var localValue *ast.StructLit
	if local := fields["_local"]; local != nil {
		var ok bool
		localValue, ok = local.Value.(*ast.StructLit)
		if !ok {
			return fmt.Errorf("_local must be a struct")
		}
	} else {
		localValue = &ast.StructLit{}
		file.Decls = append(file.Decls, &ast.Field{Label: ast.NewIdent("_local"), Value: localValue})
	}
	declaration := fieldsOfStruct(localValue)["#"+name]
	if create {
		if declaration != nil {
			return fmt.Errorf("local Schema definition #%s already exists", name)
		}
		localValue.Elts = append(localValue.Elts, &ast.Field{Label: ast.NewIdent("#" + name), Value: schema})
	} else {
		if declaration == nil {
			return fmt.Errorf("local Schema definition #%s does not exist", name)
		}
		declaration.Value = schema
	}
	schemaField := fieldsOfFile(file)["_schema"]
	var schemaValue *ast.StructLit
	if schemaField == nil {
		if !create {
			return fmt.Errorf("table _schema metadata is missing")
		}
		schemaValue = &ast.StructLit{}
		file.Decls = append(file.Decls, &ast.Field{Label: ast.NewIdent("_schema"), Value: schemaValue})
	} else {
		var ok bool
		schemaValue, ok = schemaField.Value.(*ast.StructLit)
		if !ok {
			return fmt.Errorf("_schema must be a struct")
		}
	}
	localDefinitions := fieldsOfStruct(schemaValue)["localDefinitions"]
	var localDefinitionsValue *ast.StructLit
	if localDefinitions == nil {
		if !create {
			return fmt.Errorf("table _schema.localDefinitions is missing")
		}
		localDefinitionsValue = &ast.StructLit{}
		schemaValue.Elts = append(schemaValue.Elts, &ast.Field{Label: ast.NewIdent("localDefinitions"), Value: localDefinitionsValue})
	} else {
		var ok bool
		localDefinitionsValue, ok = localDefinitions.Value.(*ast.StructLit)
		if !ok {
			return fmt.Errorf("_schema.localDefinitions must be a struct")
		}
	}
	registered := fieldsOfStruct(localDefinitionsValue)[name]
	if create {
		if registered != nil {
			return fmt.Errorf("local Schema definition %q is already registered", name)
		}
		localDefinitionsValue.Elts = append(localDefinitionsValue.Elts, &ast.Field{Label: ast.NewIdent(name), Value: schemaDefinitionMetadata(update)})
		return nil
	}
	if registered == nil {
		return fmt.Errorf("local Schema definition %q is not registered in _schema.localDefinitions", name)
	}
	registered.Value = schemaDefinitionMetadata(update)
	return nil
}

// CreateLocalSchemaDefinition adds one _local definition to a table file.
func CreateLocalSchemaDefinition(path, name string, update SchemaDefinitionUpdate) (Table, error) {
	if err := validateSchemaDefinitionUpdate(name, update); err != nil {
		return Table{}, err
	}
	schema, err := parseSchemaDefinitionExpression(path, update.SchemaExpression)
	if err != nil {
		return Table{}, err
	}
	file, err := readSourceFile(path)
	if err != nil {
		return Table{}, err
	}
	if err := ensureLocalDefinition(file, name, update, schema, true); err != nil {
		return Table{}, err
	}
	candidate, err := format.Node(file)
	if err != nil {
		return Table{}, fmt.Errorf("format updated table: %w", err)
	}
	return validateAndReplace(path, candidate)
}

// SetLocalSchemaDefinition replaces one _local definition and its metadata.
func SetLocalSchemaDefinition(path, name string, update SchemaDefinitionUpdate) (Table, error) {
	if err := validateSchemaDefinitionUpdate(name, update); err != nil {
		return Table{}, err
	}
	schema, err := parseSchemaDefinitionExpression(path, update.SchemaExpression)
	if err != nil {
		return Table{}, err
	}
	file, err := readSourceFile(path)
	if err != nil {
		return Table{}, err
	}
	if err := ensureLocalDefinition(file, name, update, schema, false); err != nil {
		return Table{}, err
	}
	candidate, err := format.Node(file)
	if err != nil {
		return Table{}, fmt.Errorf("format updated table: %w", err)
	}
	return validateAndReplace(path, candidate)
}

// DeleteLocalSchemaDefinition removes one _local definition and its metadata.
func DeleteLocalSchemaDefinition(path, name string) (Table, error) {
	file, err := readSourceFile(path)
	if err != nil {
		return Table{}, err
	}
	fields := fieldsOfFile(file)
	local := fields["_local"]
	if local == nil {
		return Table{}, fmt.Errorf("table has no local Schema")
	}
	if !removeNamedField(local.Value, "#"+name) {
		return Table{}, fmt.Errorf("local Schema definition #%s does not exist", name)
	}
	if schemaField := fields["_schema"]; schemaField != nil {
		if localDefinitions := fieldsOfStruct(schemaField.Value)["localDefinitions"]; localDefinitions != nil {
			removeNamedField(localDefinitions.Value, name)
		}
	}
	candidate, err := format.Node(file)
	if err != nil {
		return Table{}, fmt.Errorf("format updated table: %w", err)
	}
	return validateAndReplace(path, candidate)
}

// RenameLocalSchemaDefinition renames one _local definition and rewrites
// references across the supplied reference files plus the table file.
func RenameLocalSchemaDefinition(path, oldName, newName string, referencePaths []string) (Table, error) {
	if !validIdentifier(newName) {
		return Table{}, fmt.Errorf("definition %q must be a valid CUE identifier", newName)
	}
	err := editSchemaSourceFiles(append(referencePaths, path), func(files map[string]*ast.File) error {
		target := files[path]
		fields := fieldsOfFile(target)
		local := fields["_local"]
		if local == nil {
			return fmt.Errorf("table has no local Schema")
		}
		if fieldsOfStruct(local.Value)["#"+newName] != nil {
			return fmt.Errorf("local Schema definition #%s already exists", newName)
		}
		if !renameNamedField(local.Value, "#"+oldName, "#"+newName) {
			return fmt.Errorf("local Schema definition #%s does not exist", oldName)
		}
		if schemaField := fields["_schema"]; schemaField != nil {
			if localDefinitions := fieldsOfStruct(schemaField.Value)["localDefinitions"]; localDefinitions != nil {
				renameNamedField(localDefinitions.Value, oldName, newName)
			}
		}
		for _, file := range files {
			renameSchemaReferences(file, oldName, newName)
		}
		return nil
	})
	if err != nil {
		return Table{}, err
	}
	table, err := LoadTable(path)
	if err != nil {
		return Table{}, fmt.Errorf("reload renamed table: %w", err)
	}
	return table, nil
}

// editSchemaSourceFiles loads, mutates, validates, and atomically replaces a
// set of table and Schema source files. Any validation failure leaves every
// file untouched.
func editSchemaSourceFiles(paths []string, mutate func(map[string]*ast.File) error) error {
	files := map[string]*ast.File{}
	original := map[string][]byte{}
	for _, path := range paths {
		if _, exists := files[path]; exists {
			continue
		}
		file, err := readSourceFile(path)
		if err != nil {
			return err
		}
		source, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read schema source: %w", err)
		}
		files[path] = file
		original[path] = source
	}
	if err := mutate(files); err != nil {
		return err
	}
	candidates := map[string][]byte{}
	ordered := make([]string, 0, len(files))
	for path, file := range files {
		candidate, err := format.Node(file)
		if err != nil {
			return fmt.Errorf("format schema source: %w", err)
		}
		if err := validateSchemaSourceCandidate(path, candidate); err != nil {
			return err
		}
		candidates[path] = candidate
		ordered = append(ordered, path)
	}
	sort.Strings(ordered)
	return replaceAll(ordered, candidates, original)
}

func validateSchemaSourceCandidate(path string, source []byte) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), ".datapack-validate-*.cue")
	if err != nil {
		return fmt.Errorf("create validation source: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err := temporary.Write(source); err != nil {
		temporary.Close()
		return fmt.Errorf("write validation source: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close validation source: %w", err)
	}
	file, err := parser.ParseFile(temporaryPath, source, parser.ParseComments, parser.AllErrors)
	if err != nil {
		return fmt.Errorf("parse updated source: %w", err)
	}
	if fieldsOfFile(file)["rows"] != nil {
		if _, err := LoadTable(temporaryPath); err != nil {
			return fmt.Errorf("validate updated table: %w", err)
		}
		return nil
	}
	if _, err := LoadPublicSchema(temporaryPath); err != nil {
		return fmt.Errorf("validate updated schema: %w", err)
	}
	return nil
}

func renameSchemaReferences(file *ast.File, oldName, newName string) {
	ast.Walk(file, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.SelectorExpr:
			if identifier, ok := value.Sel.(*ast.Ident); ok && identifier.Name == "#"+oldName {
				value.Sel = ast.NewIdent("#" + newName)
			}
		case *ast.Ident:
			if value.Name == "#"+oldName {
				value.Name = "#" + newName
			}
		}
		return true
	}, nil)
}

func validatePublicSchemaAndReplace(path string, source []byte) (PublicSchema, error) {
	info, err := os.Stat(path)
	if err != nil {
		return PublicSchema{}, fmt.Errorf("stat schema source: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".datapack-*.cue")
	if err != nil {
		return PublicSchema{}, fmt.Errorf("create temporary schema source: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(info.Mode()); err != nil {
		temporary.Close()
		return PublicSchema{}, fmt.Errorf("set temporary schema permissions: %w", err)
	}
	if _, err := temporary.Write(source); err != nil {
		temporary.Close()
		return PublicSchema{}, fmt.Errorf("write temporary schema source: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return PublicSchema{}, fmt.Errorf("sync temporary schema source: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return PublicSchema{}, fmt.Errorf("close temporary schema source: %w", err)
	}
	schema, err := LoadPublicSchema(temporaryPath)
	if err != nil {
		return PublicSchema{}, fmt.Errorf("validate updated schema: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return PublicSchema{}, fmt.Errorf("replace schema source: %w", err)
	}
	schema.Path = path
	return schema, nil
}

// EnsureImport appends one import declaration when the file does not already
// import the same path. Table and Schema files are both supported.
func EnsureImport(path, importPath string) error {
	source, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read source: %w", err)
	}
	file, err := parser.ParseFile(path, source, parser.ParseComments, parser.AllErrors)
	if err != nil {
		return fmt.Errorf("parse source: %w", err)
	}
	for _, declaration := range file.Decls {
		imports, ok := declaration.(*ast.ImportDecl)
		if !ok {
			continue
		}
		for _, spec := range imports.Specs {
			literal := spec.Path
			if literal.Kind != token.STRING {
				continue
			}
			value, err := strconv.Unquote(literal.Value)
			if err == nil && value == importPath {
				return nil
			}
		}
	}
	insertAt := 0
	for index, declaration := range file.Decls {
		switch declaration.(type) {
		case *ast.Package, *ast.ImportDecl:
			insertAt = index + 1
		}
	}
	importDecl := &ast.ImportDecl{Specs: []*ast.ImportSpec{{Path: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(importPath)}}}}
	file.Decls = append(file.Decls[:insertAt], append([]ast.Decl{importDecl}, file.Decls[insertAt:]...)...)
	candidate, err := format.Node(file)
	if err != nil {
		return fmt.Errorf("format source with import: %w", err)
	}
	return replaceSource(path, candidate)
}

// replaceSource validates the candidate with the same rules as other Schema
// edits and atomically replaces the file.
func replaceSource(path string, candidate []byte) error {
	if err := validateSchemaSourceCandidate(path, candidate); err != nil {
		return err
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat source: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".datapack-*.cue")
	if err != nil {
		return fmt.Errorf("create temporary source: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(info.Mode()); err != nil {
		temporary.Close()
		return fmt.Errorf("set temporary source permissions: %w", err)
	}
	if _, err := temporary.Write(candidate); err != nil {
		temporary.Close()
		return fmt.Errorf("write temporary source: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return fmt.Errorf("sync temporary source: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary source: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace source: %w", err)
	}
	return nil
}
