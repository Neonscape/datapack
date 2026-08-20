package workspace

import (
	"strings"

	"datapack/internal/document"
	"datapack/internal/projecttree"
	"datapack/internal/schemamodel"
)

func withForeignKeyOptions(table document.Table, tree projecttree.Tree) document.Table {
	targets := map[string][]string{}
	var tables []document.Table
	collectTables(tree.Root, &tables)
	for _, candidate := range tables {
		name := candidate.Name
		if candidate.Namespace != "" {
			name = candidate.Namespace + "." + candidate.Name
		}
		for _, row := range candidate.Rows {
			targets[name] = append(targets[name], row.ID)
		}
	}
	for index := range table.Columns {
		for _, key := range table.ForeignKeys {
			if key.Column == table.Columns[index].Name {
				table.Columns[index].ForeignKeyTable = key.Table
				table.Columns[index].ForeignKeyOptions = targets[key.Table]
				break
			}
		}
	}
	return table
}

// resolveEffectiveSchema expands named type Schema references for UI editors
// while preserving the original reference in the document model.
func resolveEffectiveSchema(node schemamodel.Node, logical string, definitions map[string]schemamodel.Node, resolving map[string]bool) schemamodel.Node {
	if node.Kind == schemamodel.KindReference {
		id := node.Reference
		if strings.HasPrefix(id, "_local.#") {
			id = logical + "." + id
		}
		if resolving[id] {
			return node
		}
		resolved, ok := definitions[id]
		if !ok {
			return node
		}
		resolving[id] = true
		result := resolveEffectiveSchema(resolved, logical, definitions, resolving)
		delete(resolving, id)
		result.Nullable = result.Nullable || node.Nullable
		if result.Default == nil {
			result.Default = node.Default
		}
		return result
	}
	result := node
	if node.Element != nil {
		element := resolveEffectiveSchema(*node.Element, logical, definitions, resolving)
		result.Element = &element
	}
	for index, field := range node.Fields {
		result.Fields[index].Schema = resolveEffectiveSchema(field.Schema, logical, definitions, resolving)
	}
	return result
}
