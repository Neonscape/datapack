package document

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/ast/astutil"
	"cuelang.org/go/cue/token"
)

// Bind before changing labels: equal spellings in another definition, lexical
// scope or imported package are different references. CUE's resolver handles
// local bindings; package-level unresolved names may live in another file.
func renameBoundSchemaReferences(files map[string]*ast.File, path, oldName, newName string, local bool) error {
	target := files[path]
	fields := fieldsOfFile(target)
	declaration := fields["#"+oldName]
	var localValue ast.Expr
	if local {
		localValue = fields["_local"].Value
		declaration = fieldsOfStruct(localValue)["#"+oldName]
	}
	declarations := target.Decls
	if local {
		declarations = localValue.(*ast.StructLit).Elts
		localDeclarations := 0
		for _, decl := range target.Decls {
			if field, ok := decl.(*ast.Field); ok && labelOf(field) == "_local" {
				localDeclarations++
			}
		}
		if localDeclarations != 1 {
			return fmt.Errorf("cannot rename a local Schema in a split _local declaration")
		}
	}
	declarationCount := 0
	for _, decl := range declarations {
		if field, ok := decl.(*ast.Field); ok && labelOf(field) == "#"+oldName {
			declarationCount++
		}
	}
	if declarationCount != 1 {
		return fmt.Errorf("cannot rename a Schema with split declarations")
	}
	packageDecl := packageOfFile(target)
	if packageDecl == nil || packageDecl.Name == nil {
		return fmt.Errorf("Schema source must declare a package")
	}
	packageName := packageDecl.Name.Name
	samePackage := func(p string, file *ast.File) bool {
		pkg := packageOfFile(file)
		return filepath.Clean(filepath.Dir(p)) == filepath.Clean(filepath.Dir(path)) && pkg != nil && pkg.Name.Name == packageName
	}
	// A split declaration would require migrating every registration as a unit.
	// Reject it rather than changing one file and leaving the package ambiguous.
	for p, file := range files {
		if !samePackage(p, file) || p == path {
			continue
		}
		otherFields := fieldsOfFile(file)
		if local {
			if otherFields["_local"] != nil {
				return fmt.Errorf("cannot rename local Schema split across package files: %s", p)
			}
			continue
		}
		if otherFields["#"+oldName] != nil || otherFields["#"+newName] != nil {
			return fmt.Errorf("cannot rename Schema split across package files: %s", p)
		}
	}
	var importPath ast.ImportPath
	if !local {
		var err error
		importPath, err = schemaPackageImportPath(path, packageName)
		if err != nil {
			return err
		}
	}
	uniquePackage := true
	for p, file := range files {
		if pkg := packageOfFile(file); pkg != nil && pkg.Name.Name == packageName && !samePackage(p, file) {
			uniquePackage = false
		}
	}
	var bare []*ast.Ident
	for p, file := range files {
		if err := resolveSchemaBindings(file); err != nil {
			return err
		}
		var binding func(ast.Expr, map[ast.Node]bool) (bool, error)
		binding = func(expression ast.Expr, seen map[ast.Node]bool) (bool, error) {
			if seen[expression] {
				return false, nil
			}
			seen[expression] = true
			switch value := expression.(type) {
			case *ast.ParenExpr:
				return binding(value.X, seen)
			case *ast.Ident:
				switch node := value.Node.(type) {
				case *ast.LetClause:
					return binding(node.Expr, seen)
				case *ast.Alias:
					return binding(node.Expr, seen)
				}
				if local && (value.Node == localValue || value.Node == nil && value.Name == "_local" && samePackage(p, file)) {
					return true, nil
				}
				switch node := value.Node.(type) {
				case ast.Expr:
					return binding(node, seen)
				case *ast.ImportSpec:
					if local {
						return false, nil
					}
					info, err := astutil.ParseImportSpec(node)
					if err != nil {
						return false, err
					}
					parts := ast.ParseImportPath(info.ID)
					if importPath.Path == "" && parts.Qualifier == packageName {
						return false, fmt.Errorf("cannot resolve imported Schema package %q without an enclosing CUE module", info.ID)
					}
					return importPath.Path != "" && parts.Path == importPath.Path && parts.Qualifier == packageName && (parts.Version == "" || parts.Version == importPath.Version || parts.Version == "v0" && importPath.Version == ""), nil
				}
				// Legacy projects may contain unimported package selectors. Keep
				// that syntax only when the supplied project has one such package.
				if !local && value.Node == nil && value.Name == packageName {
					if !uniquePackage {
						return false, fmt.Errorf("ambiguous unimported Schema package %q", packageName)
					}
					return true, nil
				}
			default:
				// The editor cannot prove the selected definition's origin after
				// arbitrary composition/indexing. Refuse a package-derived access
				// instead of silently leaving a reference to the removed name.
				var compositeErr error
				astutil.Apply(expression, func(cursor astutil.Cursor) bool {
					id, ok := cursor.Node().(*ast.Ident)
					if !ok {
						return true
					}
					if cursor.Parent() != nil {
						switch parent := cursor.Parent().Node().(type) {
						case *ast.Field:
							if parent.Label == id {
								return false
							}
						case *ast.SelectorExpr:
							if parent.Sel == id {
								return false
							}
						}
					}
					matches, err := binding(id, seen)
					if err != nil {
						compositeErr = err
					}
					if matches {
						compositeErr = fmt.Errorf("cannot safely rename Schema accessed through a composite expression")
					}
					return false
				}, nil)
				if compositeErr != nil {
					return false, compositeErr
				}
			}
			return false, nil
		}
		var rewriteErr error
		astutil.Apply(file, func(cursor astutil.Cursor) bool {
			if rewriteErr != nil {
				return false
			}
			id, ok := cursor.Node().(*ast.Ident)
			if !ok || id.Name != "#"+oldName || cursor.Parent() == nil {
				return true
			}
			switch parent := cursor.Parent().Node().(type) {
			case *ast.Field:
				if parent.Label == id {
					return false
				}
			case *ast.SelectorExpr:
				if parent.Sel == id {
					matches, err := binding(parent.X, map[ast.Node]bool{})
					rewriteErr = err
					if matches {
						id.Name = "#" + newName
					}
					return false
				}
			case *ast.Alias:
				if parent.Ident == id {
					return false
				}
			case *ast.PostfixAlias, *ast.ForClause:
				return false
			case *ast.LetClause:
				if parent.Ident == id {
					return false
				}
			}
			if id.Node == declaration.Value || !local && id.Node == nil && samePackage(p, file) {
				id.Name = "#" + newName
				bare = append(bare, id)
			}
			return true
		}, nil)
		if rewriteErr != nil {
			return rewriteErr
		}
	}
	label, ok := declaration.Label.(*ast.Ident)
	if !ok {
		return fmt.Errorf("Schema declaration must use an identifier label")
	}
	label.Name = "#" + newName
	// A nested #New would capture a renamed bare #Old. Resolving again detects
	// that change before any source is staged or written.
	for _, file := range files {
		ast.Walk(file, func(node ast.Node) bool {
			if id, ok := node.(*ast.Ident); ok {
				id.Node, id.Scope = nil, nil
			}
			return true
		}, nil)
		file.Unresolved = nil
		if err := resolveSchemaBindings(file); err != nil {
			return err
		}
	}
	for _, reference := range bare {
		if reference.Node != declaration.Value && (local || reference.Node != nil) {
			return fmt.Errorf("renaming Schema to #%s would capture a reference in a nested scope", newName)
		}
	}
	return nil
}

func resolveSchemaBindings(file *ast.File) error {
	var result error
	astutil.Resolve(file, func(_ token.Pos, message string, args ...interface{}) {
		if result == nil {
			result = fmt.Errorf("resolve Schema source: "+message, args...)
		}
	})
	return result
}

// Resolve only the enclosing local module; never load/download an import.
// A project without a module can still use the legacy unimported selectors.
func schemaPackageImportPath(path, packageName string) (ast.ImportPath, error) {
	for directory := filepath.Dir(path); ; directory = filepath.Dir(directory) {
		modulePath := filepath.Join(directory, "cue.mod", "module.cue")
		if _, err := os.Stat(modulePath); err == nil {
			file, err := readSourceFile(modulePath)
			if err != nil {
				return ast.ImportPath{}, err
			}
			field := fieldsOfFile(file)["module"]
			if field == nil {
				return ast.ImportPath{}, fmt.Errorf("CUE module must declare module")
			}
			literal, ok := field.Value.(*ast.BasicLit)
			if !ok {
				return ast.ImportPath{}, fmt.Errorf("CUE module name must be a string")
			}
			module, err := strconv.Unquote(literal.Value)
			if err != nil {
				return ast.ImportPath{}, err
			}
			relative, err := filepath.Rel(directory, filepath.Dir(path))
			if err != nil {
				return ast.ImportPath{}, err
			}
			parts := ast.ParseImportPath(module)
			parts.Path = strings.TrimSuffix(parts.Path, "/")
			if relative != "." {
				parts.Path += "/" + filepath.ToSlash(relative)
			}
			parts.Qualifier = packageName
			return parts, nil
		} else if !os.IsNotExist(err) {
			return ast.ImportPath{}, fmt.Errorf("read enclosing CUE module: %w", err)
		}
		if parent := filepath.Dir(directory); parent == directory {
			return ast.ImportPath{}, nil
		}
	}
}
