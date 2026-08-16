package document

import (
	"fmt"

	"cuelang.org/go/cue/ast"
)

// TableReference identifies a foreign-key metadata entry that must track a
// renamed table.
type TableReference struct {
	Path     string
	Column   string
	OldTable string
	NewTable string
}

// RenameTable rewrites a table's identity (package and _tool metadata) and
// updates foreign-key references in the supplied referencing files. Every
// candidate source is validated before any file is replaced.
func RenameTable(path, newNamespace, newName string, references []TableReference) (Table, error) {
	if !validIdentifier(newName) {
		return Table{}, fmt.Errorf("table name %q must be a valid CUE identifier", newName)
	}
	paths := []string{path}
	for _, reference := range references {
		paths = append(paths, reference.Path)
	}
	err := editSchemaSourceFiles(paths, func(files map[string]*ast.File) error {
		if err := renameTableIdentity(files[path], newNamespace, newName); err != nil {
			return err
		}
		for _, reference := range references {
			if err := renameForeignKeyTable(files[reference.Path], reference.Column, reference.OldTable, reference.NewTable); err != nil {
				return err
			}
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

func renameTableIdentity(file *ast.File, newNamespace, newName string) error {
	for _, declaration := range file.Decls {
		if pkg, ok := declaration.(*ast.Package); ok {
			pkg.Name = ast.NewIdent(newName)
			break
		}
	}
	fields := fieldsOfFile(file)
	tool := fields["_tool"]
	if tool == nil {
		return fmt.Errorf("table file must declare _tool")
	}
	identityField := fieldsOfStruct(tool.Value)["table"]
	if identityField == nil {
		return fmt.Errorf("_tool must declare table")
	}
	identity := fieldsOfStruct(identityField.Value)
	if namespace := identity["namespace"]; namespace != nil {
		namespace.Value = stringLiteral(newNamespace)
	} else {
		return fmt.Errorf("_tool.table.namespace is missing")
	}
	if name := identity["name"]; name != nil {
		name.Value = stringLiteral(newName)
	} else {
		return fmt.Errorf("_tool.table.name is missing")
	}
	return nil
}

func renameForeignKeyTable(file *ast.File, column, oldTable, newTable string) error {
	fields := fieldsOfFile(file)
	tool := fields["_tool"]
	if tool == nil {
		return fmt.Errorf("table file must declare _tool")
	}
	foreignKeys := fieldsOfStruct(tool.Value)["foreignKeys"]
	if foreignKeys == nil {
		return fmt.Errorf("table has no foreign key metadata")
	}
	key := fieldsOfStruct(foreignKeys.Value)[column]
	if key == nil {
		return fmt.Errorf("foreign key column %q does not exist", column)
	}
	tableField := fieldsOfStruct(key.Value)["table"]
	if tableField == nil {
		return fmt.Errorf("foreign key %q has no table metadata", column)
	}
	if !matchesString(tableField.Value, oldTable) {
		return fmt.Errorf("foreign key %q no longer references %q", column, oldTable)
	}
	tableField.Value = stringLiteral(newTable)
	return nil
}
