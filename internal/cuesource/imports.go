package cuesource

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/ast/astutil"
	"cuelang.org/go/cue/parser"
)

// CanonicalImport identifies a package independently of its source alias and
// implicit/default major version. It never resolves or downloads dependencies.
func CanonicalImport(value string) string {
	if value == "" {
		return ""
	}
	parts := ast.ParseImportPath(value)
	if parts.Version == "" {
		parts.Version = "v0"
	}
	parts.ExplicitQualifier = true
	return parts.String()
}

// PackageImportPath describes a package in its enclosing local CUE module.
// Absence of a module is allowed for the editor's legacy standalone documents.
func PackageImportPath(path, packageName string) (string, error) {
	for dir := filepath.Dir(path); ; dir = filepath.Dir(dir) {
		source, err := os.ReadFile(filepath.Join(dir, "cue.mod", "module.cue"))
		if err == nil {
			file, err := parser.ParseFile("module.cue", source, parser.AllErrors)
			if err != nil {
				return "", err
			}
			for _, decl := range file.Decls {
				field, ok := decl.(*ast.Field)
				if !ok {
					continue
				}
				name, _, _ := ast.LabelName(field.Label)
				if name != "module" {
					continue
				}
				literal, ok := field.Value.(*ast.BasicLit)
				if !ok {
					return "", fmt.Errorf("module name must be a string")
				}
				value, err := strconv.Unquote(literal.Value)
				if err != nil {
					return "", err
				}
				parts := ast.ParseImportPath(value)
				relative, err := filepath.Rel(dir, filepath.Dir(path))
				if err != nil {
					return "", err
				}
				if relative != "." {
					parts.Path = strings.TrimSuffix(parts.Path, "/") + "/" + filepath.ToSlash(relative)
				}
				parts.Qualifier = packageName
				parts.ExplicitQualifier = true
				return CanonicalImport(parts.String()), nil
			}
			return "", fmt.Errorf("CUE module has no module declaration")
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		if filepath.Dir(dir) == dir {
			return "", nil
		}
	}
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
	return ImportBindings(file), nil
}
