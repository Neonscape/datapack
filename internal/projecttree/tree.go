// Package projecttree builds the source-file tree shown by the editor.
package projecttree

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type NodeKind string

const (
	NodeProject    NodeKind = "project"
	NodeGroup      NodeKind = "group"
	NodeDirectory  NodeKind = "directory"
	NodeTableFile  NodeKind = "tableFile"
	NodeSchemaFile NodeKind = "schemaFile"
)

// Node is a physical source-tree node. Table and Schema files are refined into
// logical views after CUE parsing is available.
type Node struct {
	Kind     NodeKind `json:"kind"`
	Name     string   `json:"name"`
	Path     string   `json:"path,omitempty"`
	Children []Node   `json:"children,omitempty"`
}

type Tree struct {
	Root Node `json:"root"`
}

// Scan discovers CUE files declared by the project. It deliberately does not
// infer table metadata from CUE yet; that belongs to the source parser.
func Scan(projectPath string, tableRoots, schemaRoots, excludes []string) (Tree, error) {
	root := filepath.Dir(projectPath)
	projectName := strings.TrimSuffix(filepath.Base(projectPath), filepath.Ext(projectPath))

	tableNodes, err := scanRoots(root, tableRoots, excludes, NodeTableFile)
	if err != nil {
		return Tree{}, err
	}
	schemaNodes, err := scanRoots(root, schemaRoots, excludes, NodeSchemaFile)
	if err != nil {
		return Tree{}, err
	}

	return Tree{Root: Node{
		Kind: NodeProject,
		Name: projectName,
		Path: root,
		Children: []Node{
			{Kind: NodeGroup, Name: "Tables", Children: tableNodes},
			{Kind: NodeGroup, Name: "Schemas", Children: schemaNodes},
		},
	}}, nil
}

func scanRoots(projectRoot string, roots, excludes []string, fileKind NodeKind) ([]Node, error) {
	result := make([]Node, 0, len(roots))
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

		node, err := scanDirectory(projectRoot, root, excludes, fileKind)
		if err != nil {
			return nil, err
		}
		result = append(result, node)
	}
	sortNodes(result)
	return result, nil
}

func scanDirectory(projectRoot, directory string, excludes []string, fileKind NodeKind) (Node, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return Node{}, err
	}

	node := Node{Kind: NodeDirectory, Name: filepath.Base(directory), Path: directory}
	for _, entry := range entries {
		path := filepath.Join(directory, entry.Name())
		relative, err := filepath.Rel(projectRoot, path)
		if err != nil {
			return Node{}, err
		}
		if isExcluded(filepath.ToSlash(relative), excludes) {
			continue
		}
		if entry.IsDir() {
			child, err := scanDirectory(projectRoot, path, excludes, fileKind)
			if err != nil {
				return Node{}, err
			}
			node.Children = append(node.Children, child)
			continue
		}
		if filepath.Ext(entry.Name()) == ".cue" {
			node.Children = append(node.Children, Node{Kind: fileKind, Name: entry.Name(), Path: path})
		}
	}
	sortNodes(node.Children)
	return node, nil
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
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Kind == NodeDirectory && nodes[j].Kind != NodeDirectory {
			return true
		}
		if nodes[i].Kind != NodeDirectory && nodes[j].Kind == NodeDirectory {
			return false
		}
		return nodes[i].Name < nodes[j].Name
	})
}
