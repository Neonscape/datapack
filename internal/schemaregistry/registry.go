// Package schemaregistry resolves Schema references across structured documents.
package schemaregistry

import (
	"fmt"
	"sort"
	"strings"

	"datapack/internal/document"
	"datapack/internal/schemamodel"
)

const (
	KindType       = "type"
	KindConstraint = "constraint"
)

type ID string

type Definition struct {
	ID        ID               `json:"id"`
	Kind      string           `json:"kind"`
	AppliesTo []string         `json:"appliesTo,omitempty"`
	Schema    schemamodel.Node `json:"schema"`
	Path      string           `json:"path"`
	Table     string           `json:"table,omitempty"`
}

type Usage struct {
	Definition ID     `json:"definition"`
	Path       string `json:"path"`
	Table      string `json:"table,omitempty"`
	Field      string `json:"field"`
}

type Diagnostic struct {
	Path    string `json:"path"`
	Table   string `json:"table,omitempty"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Registry struct {
	Definitions map[ID]Definition
	Usages      map[ID][]Usage
	Diagnostics []Diagnostic
}

// Index is the JSON-friendly Schema navigation and validation result for a
// workspace snapshot. Definitions are ordered by ID for stable UI rendering.
type Index struct {
	Definitions []IndexedDefinition `json:"definitions"`
	Diagnostics []Diagnostic        `json:"diagnostics,omitempty"`
}

type IndexedDefinition struct {
	ID        ID       `json:"id"`
	Kind      string   `json:"kind"`
	AppliesTo []string `json:"appliesTo,omitempty"`
	Path      string   `json:"path"`
	Table     string   `json:"table,omitempty"`
	Usages    []Usage  `json:"usages,omitempty"`
}

func Build(publicSchemas []document.PublicSchema, tables []document.Table) Registry {
	registry := Registry{
		Definitions: map[ID]Definition{},
		Usages:      map[ID][]Usage{},
	}
	for _, schema := range publicSchemas {
		for _, definition := range schema.Definitions {
			id := publicID(schema.Package, definition.Name)
			registry.addDefinition(Definition{
				ID:        id,
				Kind:      definition.Kind,
				AppliesTo: definition.AppliesTo,
				Schema:    definition.Schema,
				Path:      schema.Path,
			})
		}
	}
	for _, table := range tables {
		logicalName := tableName(table)
		for _, definition := range table.LocalSchema {
			id := localID(logicalName, definition.Name)
			registry.addDefinition(Definition{
				ID:        id,
				Kind:      definition.Kind,
				AppliesTo: definition.AppliesTo,
				Schema:    definition.Schema,
				Path:      table.Path,
				Table:     logicalName,
			})
		}
	}
	for _, schema := range publicSchemas {
		for _, definition := range schema.Definitions {
			registry.validateDefinition(publicID(schema.Package, definition.Name), schema.Path, "")
		}
	}
	for _, table := range tables {
		logicalName := tableName(table)
		registry.validateNode(table.PrimaryKey, table.Path, logicalName, "#PrimaryKey", logicalName, "", false)
		for _, column := range table.Columns {
			registry.validateNode(column.Schema, table.Path, logicalName, column.Name, logicalName, "", false)
		}
		for _, definition := range table.LocalSchema {
			registry.validateDefinition(localID(logicalName, definition.Name), table.Path, logicalName)
		}
	}
	registry.detectCycles()
	for id := range registry.Usages {
		sort.Slice(registry.Usages[id], func(i, j int) bool {
			left, right := registry.Usages[id][i], registry.Usages[id][j]
			if left.Path == right.Path {
				return left.Field < right.Field
			}
			return left.Path < right.Path
		})
	}
	return registry
}

// Index returns a stable, UI-safe view of the definitions and their usages.
func (r Registry) Index() Index {
	ids := make([]ID, 0, len(r.Definitions))
	for id := range r.Definitions {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	definitions := make([]IndexedDefinition, 0, len(ids))
	for _, id := range ids {
		definition := r.Definitions[id]
		definitions = append(definitions, IndexedDefinition{
			ID:        id,
			Kind:      definition.Kind,
			AppliesTo: definition.AppliesTo,
			Path:      definition.Path,
			Table:     definition.Table,
			Usages:    r.Usages[id],
		})
	}
	return Index{Definitions: definitions, Diagnostics: r.Diagnostics}
}

func (r *Registry) addDefinition(definition Definition) {
	if definition.Kind != KindType && definition.Kind != KindConstraint {
		r.Diagnostics = append(r.Diagnostics, Diagnostic{Path: definition.Path, Table: definition.Table, Field: string(definition.ID), Message: "Schema kind must be type or constraint"})
		return
	}
	if _, exists := r.Definitions[definition.ID]; exists {
		r.Diagnostics = append(r.Diagnostics, Diagnostic{Path: definition.Path, Table: definition.Table, Field: string(definition.ID), Message: "duplicate Schema definition"})
		return
	}
	r.Definitions[definition.ID] = definition
}

func (r *Registry) validateDefinition(id ID, path, table string) {
	definition, exists := r.Definitions[id]
	if !exists {
		return
	}
	r.validateNode(definition.Schema, path, table, string(id), table, packageOf(id), definition.Kind == KindConstraint)
}

func (r *Registry) validateNode(node schemamodel.Node, path, table, field, currentTable, currentPackage string, constraintContext bool) {
	if node.Kind == schemamodel.KindReference {
		r.validateReference(node.Reference, path, table, field, currentTable, currentPackage, constraintContext)
	}
	for _, requirement := range node.Requirements {
		if requirement.Reference != "" {
			r.validateReference(requirement.Reference, path, table, field, currentTable, currentPackage, true, node.Kind)
		}
	}
	if node.Element != nil {
		r.validateNode(*node.Element, path, table, field+"[]", currentTable, currentPackage, constraintContext)
	}
	for _, structField := range node.Fields {
		r.validateNode(structField.Schema, path, table, field+"."+structField.Name, currentTable, currentPackage, constraintContext)
	}
}

func (r *Registry) validateReference(reference, path, table, field, currentTable, currentPackage string, constraintContext bool, sourceKind ...schemamodel.Kind) {
	id, err := resolve(reference, currentTable, currentPackage)
	if err != nil {
		r.error(path, table, field, err.Error())
		return
	}
	definition, exists := r.Definitions[id]
	if !exists {
		r.error(path, table, field, "Schema reference "+reference+" does not exist")
		return
	}
	if constraintContext && definition.Kind != KindConstraint {
		r.error(path, table, field, "Schema reference "+reference+" must be a constraint Schema")
		return
	}
	if constraintContext && len(sourceKind) > 0 && !appliesTo(definition.AppliesTo, sourceKind[0]) {
		r.error(path, table, field, "constraint Schema "+reference+" does not apply to "+string(sourceKind[0]))
		return
	}
	if !constraintContext && definition.Kind != KindType {
		r.error(path, table, field, "Schema reference "+reference+" must be a type Schema")
		return
	}
	r.Usages[id] = append(r.Usages[id], Usage{Definition: id, Path: path, Table: table, Field: field})
}

func appliesTo(applicable []string, kind schemamodel.Kind) bool {
	if kind == schemamodel.KindUnknown || kind == schemamodel.KindReference {
		return true
	}
	for _, candidate := range applicable {
		if candidate == string(kind) || candidate == "number" && kind == schemamodel.KindInteger {
			return true
		}
	}
	return false
}

func (r *Registry) detectCycles() {
	graph := map[ID][]ID{}
	for id, definition := range r.Definitions {
		graph[id] = references(definition.Schema, definition.Table, packageOf(id))
	}
	state := map[ID]int{}
	stack := []ID{}
	for id := range graph {
		r.visit(id, graph, state, &stack)
	}
}

func (r *Registry) visit(id ID, graph map[ID][]ID, state map[ID]int, stack *[]ID) {
	if state[id] == 2 {
		return
	}
	if state[id] == 1 {
		cycle := append(append([]ID{}, *stack...), id)
		definition := r.Definitions[id]
		r.error(definition.Path, definition.Table, string(id), "Schema cycle: "+joinIDs(cycle))
		return
	}
	state[id] = 1
	*stack = append(*stack, id)
	for _, target := range graph[id] {
		if _, exists := r.Definitions[target]; exists {
			r.visit(target, graph, state, stack)
		}
	}
	*stack = (*stack)[:len(*stack)-1]
	state[id] = 2
}

func references(node schemamodel.Node, table, packageName string) []ID {
	references := []ID{}
	add := func(reference string) {
		if id, err := resolve(reference, table, packageName); err == nil {
			references = append(references, id)
		}
	}
	if node.Kind == schemamodel.KindReference {
		add(node.Reference)
	}
	for _, requirement := range node.Requirements {
		if requirement.Reference != "" {
			add(requirement.Reference)
		}
	}
	if node.Element != nil {
		references = append(references, referencesForNode(*node.Element, table, packageName)...)
	}
	for _, field := range node.Fields {
		references = append(references, referencesForNode(field.Schema, table, packageName)...)
	}
	return references
}

func referencesForNode(node schemamodel.Node, table, packageName string) []ID {
	return references(node, table, packageName)
}

func (r *Registry) error(path, table, field, message string) {
	r.Diagnostics = append(r.Diagnostics, Diagnostic{Path: path, Table: table, Field: field, Message: message})
}

func publicID(packageName, name string) ID {
	return ID(packageName + ".#" + name)
}

func localID(table, name string) ID {
	return ID(table + "._local.#" + name)
}

func tableName(table document.Table) string {
	if table.Namespace == "" {
		return table.Name
	}
	return table.Namespace + "." + table.Name
}

func packageOf(id ID) string {
	value := string(id)
	if index := strings.Index(value, ".#"); index >= 0 {
		return value[:index]
	}
	return ""
}

func resolve(reference, table, packageName string) (ID, error) {
	if strings.HasPrefix(reference, "_local.#") {
		if table == "" {
			return "", fmt.Errorf("local Schema reference %s has no table scope", reference)
		}
		return ID(table + "." + reference), nil
	}
	if strings.HasPrefix(reference, "#") {
		if packageName == "" {
			return "", fmt.Errorf("Schema reference %s has no package scope", reference)
		}
		return publicID(packageName, strings.TrimPrefix(reference, "#")), nil
	}
	marker := strings.LastIndex(reference, ".#")
	if marker < 1 || marker == len(reference)-2 {
		return "", fmt.Errorf("unsupported Schema reference %s", reference)
	}
	return publicID(reference[:marker], reference[marker+2:]), nil
}

func joinIDs(ids []ID) string {
	parts := make([]string, len(ids))
	for index, id := range ids {
		parts[index] = string(id)
	}
	return strings.Join(parts, " -> ")
}
