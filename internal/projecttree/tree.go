// Package projecttree builds the source-file tree shown by the editor.
package projecttree

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"datapack/internal/cuesource"
)

type NodeKind string

const (
	NodeProject          NodeKind = "project"
	NodeNamespace        NodeKind = "namespace"
	NodeGroup            NodeKind = "group"
	NodeTable            NodeKind = "table"
	NodeSchemaFile       NodeKind = "schemaFile"
	NodeLocalSchema      NodeKind = "localSchema"
	NodeSchemaDefinition NodeKind = "schemaDefinition"
	NodeDamaged          NodeKind = "damaged"
)

// Node is a logical source-tree node. Tables and Schemas are refined into
// logical views after CUE parsing is available.
type Node struct {
	Kind        NodeKind               `json:"kind"`
	Name        string                 `json:"name"`
	Path        string                 `json:"path,omitempty"`
	Diagnostics []cuesource.Diagnostic `json:"diagnostics,omitempty"`
	Children    []Node                 `json:"children,omitempty"`
}

type Tree struct {
	Root Node `json:"root"`
}

type namespaceGroup struct {
	tables  []Node
	schemas []Node
}

// Scan discovers CUE files declared by the project and builds a namespace-
// grouped logical tree. Tables use _tool.table.namespace and public Schema
// files use package; both are treated as one namespace concept.
func Scan(projectPath string, tableRoots, schemaRoots, excludes []string) (Tree, error) {
	root := filepath.Dir(projectPath)
	projectName := strings.TrimSuffix(filepath.Base(projectPath), filepath.Ext(projectPath))

	tableFiles, err := collectRootFiles(root, tableRoots, excludes)
	if err != nil {
		return Tree{}, err
	}
	schemaFiles, err := collectRootFiles(root, schemaRoots, excludes)
	if err != nil {
		return Tree{}, err
	}

	groups := map[string]*namespaceGroup{}
	order := []string{}
	ensureNamespace := func(name string) *namespaceGroup {
		if group, exists := groups[name]; exists {
			return group
		}
		group := &namespaceGroup{}
		groups[name] = group
		order = append(order, name)
		return group
	}

	damagedTables := []Node{}
	damagedSchemas := []Node{}

	for _, path := range tableFiles {
		table := cuesource.ParseTableFile(path)
		if len(table.Diagnostics) > 0 {
			damagedTables = append(damagedTables, damagedNode(path, table.Diagnostics))
			continue
		}
		group := ensureNamespace(table.Namespace)
		group.tables = append(group.tables, tableNode(table))
	}
	for _, path := range schemaFiles {
		schema := cuesource.ParseSchemaFile(path)
		if len(schema.Diagnostics) > 0 {
			damagedSchemas = append(damagedSchemas, damagedNode(path, schema.Diagnostics))
			continue
		}
		group := ensureNamespace(schema.Package)
		group.schemas = append(group.schemas, Node{Kind: NodeSchemaFile, Name: filepath.Base(schema.Path), Path: schema.Path, Children: schemaDefinitionNodes(schema)})
	}

	sort.Strings(order)
	children := make([]Node, 0, len(order)+2)
	for _, name := range order {
		group := groups[name]
		node := Node{Kind: NodeNamespace, Name: name}
		if len(group.tables) > 0 {
			sortNodes(group.tables)
			node.Children = append(node.Children, Node{Kind: NodeGroup, Name: "Tables", Children: group.tables})
		}
		if len(group.schemas) > 0 {
			sortNodes(group.schemas)
			node.Children = append(node.Children, Node{Kind: NodeGroup, Name: "Schemas", Children: group.schemas})
		}
		children = append(children, node)
	}
	if len(damagedTables) > 0 {
		sortNodes(damagedTables)
		children = append(children, Node{Kind: NodeGroup, Name: "Tables", Children: damagedTables})
	}
	if len(damagedSchemas) > 0 {
		sortNodes(damagedSchemas)
		children = append(children, Node{Kind: NodeGroup, Name: "Schemas", Children: damagedSchemas})
	}

	return Tree{Root: Node{
		Kind:     NodeProject,
		Name:     projectName,
		Path:     root,
		Children: children,
	}}, nil
}

func collectRootFiles(projectRoot string, roots, excludes []string) ([]string, error) {
	var files []string
	for _, configuredRoot := range roots {
		root, err := resolveProjectPath(projectRoot, configuredRoot)
		if err != nil {
			return nil, err
		}
		info, err := os.Stat(root)
		if err != nil {
			return nil, fmt.Errorf("source root %q: %w", configuredRoot, err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("source root %q is not a directory", configuredRoot)
		}
		if err := walkDirectory(projectRoot, root, excludes, &files); err != nil {
			return nil, err
		}
	}
	return files, nil
}

func walkDirectory(projectRoot, directory string, excludes []string, files *[]string) error {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		path := filepath.Join(directory, entry.Name())
		relative, err := filepath.Rel(projectRoot, path)
		if err != nil {
			return err
		}
		if isExcluded(filepath.ToSlash(relative), excludes) {
			continue
		}
		if entry.IsDir() {
			if err := walkDirectory(projectRoot, path, excludes, files); err != nil {
				return err
			}
			continue
		}
		if filepath.Ext(entry.Name()) == ".cue" {
			*files = append(*files, path)
		}
	}
	return nil
}

func tableNode(table cuesource.Table) Node {
	return Node{
		Kind: NodeTable,
		Name: table.Name,
		Path: table.Path,
		Children: []Node{
			{
				Kind:     NodeLocalSchema,
				Name:     "local_schema",
				Path:     table.Path,
				Children: localDefinitionNodes(table.Path, table.LocalDefinitions),
			},
		},
	}
}

func localDefinitionNodes(path string, definitions []string) []Node {
	nodes := make([]Node, 0, len(definitions))
	for _, definition := range definitions {
		nodes = append(nodes, Node{Kind: NodeLocalSchema, Name: "#" + definition, Path: path})
	}
	return nodes
}

func schemaDefinitionNodes(schema cuesource.SchemaFile) []Node {
	nodes := make([]Node, 0, len(schema.Definitions))
	for _, definition := range schema.Definitions {
		nodes = append(nodes, Node{Kind: NodeSchemaDefinition, Name: "#" + definition.Name, Path: schema.Path})
	}
	return nodes
}

func damagedNode(path string, diagnostics []cuesource.Diagnostic) Node {
	return Node{Kind: NodeDamaged, Name: filepath.Base(path), Path: path, Diagnostics: diagnostics}
}

func resolveProjectPath(projectRoot, configuredPath string) (string, error) {
	if configuredPath == "" || filepath.IsAbs(configuredPath) {
		return "", fmt.Errorf("source root must be a non-empty relative path")
	}
	path := filepath.Clean(filepath.Join(projectRoot, configuredPath))
	relative, err := filepath.Rel(projectRoot, path)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("source root %q escapes the project directory", configuredPath)
	}
	return path, nil
}

func isExcluded(relative string, excludes []string) bool {
	for _, excluded := range excludes {
		excluded = strings.Trim(strings.ReplaceAll(excluded, "\\", "/"), "/")
		if excluded != "" && (relative == excluded || strings.HasPrefix(relative, excluded+"/")) {
			return true
		}
	}
	return false
}

func sortNodes(nodes []Node) {
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Name < nodes[j].Name })
}
