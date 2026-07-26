// Package document contains the editor's structured, source-backed documents.
package document

import "datapack/internal/schemamodel"

type Table struct {
	Path        string             `json:"path"`
	Package     string             `json:"package"`
	Namespace   string             `json:"namespace"`
	Name        string             `json:"name"`
	PrimaryKey  schemamodel.Node   `json:"primaryKey"`
	Columns     []Column           `json:"columns"`
	Rows        []Row              `json:"rows"`
	LocalSchema []SchemaDefinition `json:"localSchema"`
}

type Column struct {
	Name     string           `json:"name"`
	Optional bool             `json:"optional"`
	Schema   schemamodel.Node `json:"schema"`
}

type Row struct {
	ID string `json:"id"`
}

type SchemaDefinition struct {
	Name      string           `json:"name"`
	Kind      string           `json:"kind"`
	AppliesTo []string         `json:"appliesTo,omitempty"`
	Schema    schemamodel.Node `json:"schema"`
}

type PublicSchema struct {
	Path        string             `json:"path"`
	Package     string             `json:"package"`
	Definitions []SchemaDefinition `json:"definitions"`
}
