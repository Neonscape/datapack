package cuesource

import (
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/ast/astutil"
	"cuelang.org/go/cue/parser"
	"cuelang.org/go/mod/modfile"
)

// CanonicalImport only normalizes syntax. An omitted major is resolved by the
// enclosing module, not a global v0 default.
func CanonicalImport(value string) string {
	if value == "" {
		return ""
	}
	parts := ast.ParseImportPath(value)
	parts.ExplicitQualifier = true
	return parts.String()
}

type ImportContext struct {
	Package string
	module  *modfile.File
}

func (c ImportContext) Resolve(value string) string {
	if value == "" {
		return ""
	}
	parts := ast.ParseImportPath(value)
	if parts.Version == "" && c.module != nil {
		if target, ok := c.module.ModuleForImportPath(value); ok {
			parts.Version = ast.ParseImportPath(target.Path()).Version
		}
	}
	return CanonicalImport(parts.String())
}

// ReadImportContext reads local module metadata only. Official CUE module
// logic resolves defaults, including the main major and declared dependency
// defaults; it does not consult registries or download a dependency.
func ReadImportContext(path, packageName string) (ImportContext, error) {
	for dir := filepath.Dir(path); ; dir = filepath.Dir(dir) {
		modulePath := filepath.Join(dir, "cue.mod", "module.cue")
		source, err := os.ReadFile(modulePath)
		if err == nil {
			syntax, err := parser.ParseFile(modulePath, source, parser.AllErrors)
			if err != nil {
				return ImportContext{}, err
			}
			modern := false
			for _, decl := range syntax.Decls {
				if field, ok := decl.(*ast.Field); ok {
					name, _, _ := ast.LabelName(field.Label)
					if name == "language" {
						modern = true
					}
				}
			}
			var module *modfile.File
			if modern {
				module, err = modfile.ParseNonStrict(source, modulePath)
			} else {
				module, err = modfile.ParseLegacy(source, modulePath)
				if err == nil {
					err = module.InitNonStrict()
				}
			}
			if err != nil {
				return ImportContext{}, err
			}
			parts := ast.ParseImportPath(module.QualifiedModule())
			relative, err := filepath.Rel(dir, filepath.Dir(path))
			if err != nil {
				return ImportContext{}, err
			}
			if relative != "." {
				parts.Path = strings.TrimSuffix(parts.Path, "/") + "/" + filepath.ToSlash(relative)
			}
			parts.Qualifier = packageName
			parts.ExplicitQualifier = true
			packagePath := ""
			if packageName != "" {
				packagePath = CanonicalImport(parts.String())
			}
			return ImportContext{Package: packagePath, module: module}, nil
		}
		if !os.IsNotExist(err) {
			return ImportContext{}, err
		}
		if filepath.Dir(dir) == dir {
			return ImportContext{}, nil
		}
	}
}
func PackageImportPath(path, packageName string) (string, error) {
	context, err := ReadImportContext(path, packageName)
	return context.Package, err
}
func ImportBindings(file *ast.File) map[string]string {
	result := map[string]string{}
	for _, decl := range file.Decls {
		if imports, ok := decl.(*ast.ImportDecl); ok {
			for _, spec := range imports.Specs {
				if info, err := astutil.ParseImportSpec(spec); err == nil {
					result[info.Ident] = CanonicalImport(info.ID)
				}
			}
		}
	}
	return result
}
func ReadImportBindings(path string) (map[string]string, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	file, err := parser.ParseFile(path, source, parser.AllErrors)
	if err != nil {
		return nil, err
	}
	// Invalid module metadata remains a compiler diagnostic; import-name
	// inspection must not block an otherwise legal source edit.
	context, _ := ReadImportContext(path, "")
	bindings := ImportBindings(file)
	for alias, value := range bindings {
		bindings[alias] = context.Resolve(value)
	}
	return bindings, nil
}
