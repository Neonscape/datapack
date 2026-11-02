package document

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/parser"
	"cuelang.org/go/cue/token"
	"datapack/internal/cueformat"
)

// LiteralSnapshot is an experimental scalar-only source index. The Table rows
// are immutable after publication: an edit copies the row slice and edited map.
// Unsupported expressions/formulas/FKs always keep the existing slow path.
type LiteralSnapshot struct {
	Source   []byte
	Table    Table
	rowIndex map[string]int
	columns  map[string]int
	spans    []literalSpan
}
type literalSpan struct{ start, end int }
type LiteralPatch struct {
	Source      []byte
	Before      string
	After       string
	row, column int
	span        literalSpan
	cell        Cell
}

func scalarAST(v ast.Expr) bool {
	if b, ok := v.(*ast.BasicLit); ok {
		switch b.Kind {
		case token.STRING, token.INT, token.FLOAT, token.TRUE, token.FALSE, token.NULL:
			return true
		}
		return false
	}
	if u, ok := v.(*ast.UnaryExpr); ok && u.Op == token.SUB {
		if b, ok := u.X.(*ast.BasicLit); ok {
			return b.Kind == token.INT || b.Kind == token.FLOAT
		}
	}
	return false
}

func NewLiteralSnapshot(path string, source []byte, table Table) (*LiteralSnapshot, error) {
	sum := sha256.Sum256(source)
	if table.SourceSHA256 != hex.EncodeToString(sum[:]) {
		return nil, fmt.Errorf("source changed during snapshot load")
	}
	if len(table.Formulas) != 0 || len(table.ForeignKeys) != 0 {
		return nil, fmt.Errorf("not a fixed scalar table")
	}
	file, err := parser.ParseFile(path, source, parser.ParseComments, parser.AllErrors)
	if err != nil {
		return nil, err
	}
	fields := fieldsOfFile(file)
	rows, ok := fields["rows"].Value.(*ast.StructLit)
	if !ok {
		return nil, fmt.Errorf("rows is not a literal")
	}
	snapshot := &LiteralSnapshot{Source: source, Table: table, rowIndex: map[string]int{}, columns: map[string]int{}, spans: make([]literalSpan, len(table.Rows)*len(table.Columns))}
	for i, c := range table.Columns {
		snapshot.columns[c.Name] = i
	}
	for i, r := range table.Rows {
		if _, exists := snapshot.rowIndex[r.ID]; exists {
			return nil, fmt.Errorf("duplicate row")
		}
		snapshot.rowIndex[r.ID] = i
	}
	for _, decl := range rows.Elts {
		row, ok := decl.(*ast.Field)
		if !ok {
			return nil, fmt.Errorf("embedded row declaration")
		}
		ri, ok := snapshot.rowIndex[labelOf(row)]
		if !ok {
			return nil, fmt.Errorf("unindexed row")
		}
		value, ok := row.Value.(*ast.StructLit)
		if !ok {
			return nil, fmt.Errorf("nonliteral row")
		}
		seen := map[string]bool{}
		for _, decl := range value.Elts {
			field, ok := decl.(*ast.Field)
			if !ok {
				return nil, fmt.Errorf("embedded cell declaration")
			}
			name := labelOf(field)
			ci, ok := snapshot.columns[name]
			if !ok || seen[name] || !scalarAST(field.Value) {
				return nil, fmt.Errorf("unsupported field %q", name)
			}
			seen[name] = true
			snapshot.spans[ri*len(table.Columns)+ci] = literalSpan{field.Value.Pos().Offset(), field.Value.End().Offset()}
		}
	}
	return snapshot, nil
}

func (s *LiteralSnapshot) Patch(path, rowID, column, expression string, raw bool) (LiteralPatch, error) {
	ri, ok := s.rowIndex[rowID]
	if !ok {
		return LiteralPatch{}, fmt.Errorf("unknown row")
	}
	ci, ok := s.columns[column]
	if !ok {
		return LiteralPatch{}, fmt.Errorf("unknown column")
	}
	span := s.spans[ri*len(s.columns)+ci]
	if span.end <= span.start || span.end > len(s.Source) {
		return LiteralPatch{}, fmt.Errorf("missing cell requires slow path")
	}
	value, err := parser.ParseExpr(path, expression)
	if err != nil {
		return LiteralPatch{}, err
	}
	if !scalarAST(value) {
		return LiteralPatch{}, fmt.Errorf("nonliteral requires slow path")
	}
	formatted, err := cueformat.Node(value)
	if err != nil {
		return LiteralPatch{}, err
	}
	after := strings.TrimSpace(string(formatted))
	if raw {
		after = expression
	}
	candidate := make([]byte, 0, len(s.Source)-(span.end-span.start)+len(after))
	candidate = append(candidate, s.Source[:span.start]...)
	candidate = append(candidate, after...)
	candidate = append(candidate, s.Source[span.end:]...)
	return LiteralPatch{Source: candidate, Before: string(s.Source[span.start:span.end]), After: after, row: ri, column: ci, span: span, cell: Cell{Expression: string(formatted), Display: displayOf(value, string(formatted))}}, nil
}

func (s *LiteralSnapshot) Commit(p LiteralPatch) {
	delta := len(p.After) - (p.span.end - p.span.start)
	for i, span := range s.spans {
		if span.start >= p.span.end {
			s.spans[i] = literalSpan{span.start + delta, span.end + delta}
		}
	}
	s.spans[p.row*len(s.columns)+p.column] = literalSpan{p.span.start, p.span.start + len(p.After)}
	rows := append([]Row(nil), s.Table.Rows...)
	row := rows[p.row]
	values := make(map[string]Cell, len(row.Values))
	for k, v := range row.Values {
		values[k] = v
	}
	values[s.Table.Columns[p.column].Name] = p.cell
	row.Values = values
	rows[p.row] = row
	sum := sha256.Sum256(p.Source)
	s.Table.SourceSHA256 = hex.EncodeToString(sum[:])
	s.Table.Rows = rows
	s.Source = p.Source
}
