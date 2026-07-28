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
	ForeignKeys []ForeignKey       `json:"foreignKeys,omitempty"`
	Formulas    []Formula          `json:"formulas,omitempty"`
	Diagnostics []Diagnostic       `json:"diagnostics,omitempty"`
	CSVExport   *CSVExport         `json:"csvExport,omitempty"`
	LocalSchema []SchemaDefinition `json:"localSchema"`
}

type Column struct {
	Name     string           `json:"name"`
	Optional bool             `json:"optional"`
	Schema   schemamodel.Node `json:"schema"`
}

type Row struct {
	ID     string          `json:"id"`
	Values map[string]Cell `json:"values"`
}

// Cell preserves the source expression while exposing a compact display value
// for table views. Writes will use Expression rather than Display.
type Cell struct {
	Expression string `json:"expression"`
	Display    string `json:"display"`
	Formula    bool   `json:"formula,omitempty"`
}

type ForeignKey struct {
	Column string `json:"column"`
	Table  string `json:"table"`
}

type Formula struct {
	RowID      string `json:"rowID"`
	Column     string `json:"column"`
	Expression string `json:"expression"`
}

// Diagnostic describes a persisted business error that is shown in the table
// without preventing users from continuing to complete the row.
type Diagnostic struct {
	RowID   string `json:"rowID"`
	Column  string `json:"column,omitempty"`
	Message string `json:"message"`
}

type CSVExport struct {
	Columns []string `json:"columns"`
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
