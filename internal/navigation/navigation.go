// Package navigation describes semantic jumps between documents and cells.
package navigation

type Kind string

const (
	KindSchemaReference Kind = "schema-reference"
	KindForeignKey      Kind = "foreign-key"
	KindFormula         Kind = "formula-reference"
	KindPrimaryKey      Kind = "primary-key-references"
)

// Request describes the source of one navigation.
type Request struct {
	Kind      Kind   `json:"kind"`
	Path      string `json:"path"`
	RowID     string `json:"rowID,omitempty"`
	Column    string `json:"column,omitempty"`
	Reference string `json:"reference,omitempty"`
}

// Target describes where navigation lands.
type Target struct {
	Kind       Kind     `json:"kind"`
	Path       string   `json:"path"`
	Local      bool     `json:"local,omitempty"`
	Definition string   `json:"definition,omitempty"`
	RowID      string   `json:"rowID,omitempty"`
	Column     string   `json:"column,omitempty"`
	References []Target `json:"references,omitempty"`
}
