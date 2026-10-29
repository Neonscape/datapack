package formula

import (
	"fmt"
	"strings"

	"datapack/internal/document"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/parser"
)

// All cells compile to a private callback, leaving expr's environment empty.
// Its name is fresh for each expression so local bindings cannot shadow it.
const cellFunction = "__datapack_cell"

type formulaAnalysis struct {
	expression   string
	references   []Reference
	cellFunction string
}

type expressionNames map[string]bool

func (names expressionNames) Visit(node *ast.Node) {
	switch n := (*node).(type) {
	case *ast.IdentifierNode:
		names[n.Value] = true
	case *ast.VariableDeclaratorNode:
		names[n.Name] = true
	}
}

// References discovers static cell inputs in source order, respecting lexical
// bindings and treating a qualified address as one cell rather than its roots.
func References(table document.Table, f document.Formula) []Reference {
	analysis, _ := analyzeFormula(table, f, false)
	return analysis.references
}

type referenceAnalysis struct {
	table        document.Table
	formula      document.Formula
	validate     bool
	columns      map[string]bool
	foreignKeys  map[string]bool
	rows         map[string]bool
	seen         map[Reference]bool
	references   []Reference
	prefix       []string
	cellFunction string
}

func analyzeFormula(table document.Table, f document.Formula, validate bool) (formulaAnalysis, error) {
	tree, err := parser.Parse(f.Expression)
	if err != nil {
		return formulaAnalysis{}, err
	}
	a := referenceAnalysis{table: table, formula: f, validate: validate, columns: map[string]bool{}, foreignKeys: map[string]bool{}, rows: map[string]bool{}, seen: map[Reference]bool{}}
	names := expressionNames{}
	ast.Walk(&tree.Node, names)
	a.cellFunction = cellFunction
	for names[a.cellFunction] {
		a.cellFunction += "_"
	}
	for _, c := range table.Columns {
		a.columns[c.Name] = true
	}
	for _, key := range table.ForeignKeys {
		a.foreignKeys[key.Column] = true
	}
	for _, row := range table.Rows {
		a.rows[row.ID] = true
	}
	if table.Namespace != "" {
		a.prefix = strings.Split(table.Namespace, ".")
	}
	if table.Name != "" {
		a.prefix = append(a.prefix, table.Name)
	}
	err = a.walk(&tree.Node, map[string]bool{})
	result := formulaAnalysis{references: a.references, cellFunction: a.cellFunction}
	if err == nil {
		result.expression = tree.Node.String()
	}
	return result, err
}

func (a *referenceAnalysis) add(row, column string) error {
	if !a.columns[column] {
		return fmt.Errorf("formula column %q does not exist", column)
	}
	if a.validate && !a.rows[row] {
		return fmt.Errorf("formula row %q does not exist", row)
	}
	if a.foreignKeys[column] {
		return fmt.Errorf("formulas cannot read foreign key column %q", column)
	}
	r := Reference{RowID: row, Column: column}
	if !a.seen[r] {
		a.seen[r] = true
		a.references = append(a.references, r)
	}
	return nil
}

// Walk explicitly carries lexical scope and handles complete member chains.
// A post-order identifier visitor cannot distinguish callees, map keys,
// namespace roots and let-bound values from same-row columns.
func (a *referenceAnalysis) walk(node *ast.Node, locals map[string]bool) error {
	if *node == nil {
		return nil
	}
	walk := func(n *ast.Node) error { return a.walk(n, locals) }
	switch n := (*node).(type) {
	case *ast.IdentifierNode:
		if locals[n.Value] {
			return nil
		}
		if !a.columns[n.Value] {
			return fmt.Errorf("unknown formula input %q; dynamic environment access is not supported", n.Value)
		}
		if err := a.add(a.formula.RowID, n.Value); err != nil {
			return err
		}
		*node = a.cell(a.formula.RowID, n.Value)
		return nil
	case *ast.NilNode, *ast.IntegerNode, *ast.FloatNode, *ast.BoolNode, *ast.StringNode, *ast.ConstantNode:
		return nil
	case *ast.UnaryNode:
		return walk(&n.Node)
	case *ast.BinaryNode:
		if err := walk(&n.Left); err != nil {
			return err
		}
		return walk(&n.Right)
	case *ast.ChainNode:
		return walk(&n.Node)
	case *ast.MemberNode:
		return a.member(node, n, locals)
	case *ast.SliceNode:
		return fmt.Errorf("formula slices and array element references are not supported")
	case *ast.CallNode:
		callee, ok := n.Callee.(*ast.IdentifierNode)
		if !ok {
			return fmt.Errorf("formulas only support named deterministic functions")
		}
		if !knownFunctionNames[callee.Value] {
			return fmt.Errorf("unknown formula function %q", callee.Value)
		}
		if locals[callee.Value] {
			return fmt.Errorf("formula function %q cannot be rebound", callee.Value)
		}
		for i := range n.Arguments {
			if err := walk(&n.Arguments[i]); err != nil {
				return err
			}
		}
		return nil
	case *ast.BuiltinNode:
		if !knownFunctionNames[n.Name] {
			return fmt.Errorf("unknown formula function %q", n.Name)
		}
		for i := range n.Arguments {
			if err := walk(&n.Arguments[i]); err != nil {
				return err
			}
		}
		return nil
	case *ast.VariableDeclaratorNode:
		if err := walk(&n.Value); err != nil {
			return err
		}
		next := make(map[string]bool, len(locals)+1)
		for k, v := range locals {
			next[k] = v
		}
		next[n.Name] = true
		return a.walk(&n.Expr, next)
	case *ast.ConditionalNode:
		for _, p := range []*ast.Node{&n.Cond, &n.Exp1, &n.Exp2} {
			if err := walk(p); err != nil {
				return err
			}
		}
		return nil
	case *ast.SequenceNode:
		for i := range n.Nodes {
			if err := walk(&n.Nodes[i]); err != nil {
				return err
			}
		}
		return nil
	case *ast.ArrayNode:
		for i := range n.Nodes {
			if err := walk(&n.Nodes[i]); err != nil {
				return err
			}
		}
		return nil
	case *ast.MapNode:
		for i := range n.Pairs {
			if err := walk(&n.Pairs[i]); err != nil {
				return err
			}
		}
		return nil
	case *ast.PairNode:
		if err := walk(&n.Key); err != nil {
			return err
		}
		return walk(&n.Value)
	default:
		return fmt.Errorf("unsupported formula expression %T", n)
	}
}

func (a *referenceAnalysis) cell(row, column string) ast.Node {
	return &ast.CallNode{Callee: &ast.IdentifierNode{Value: a.cellFunction}, Arguments: []ast.Node{&ast.StringNode{Value: row}, &ast.StringNode{Value: column}}}
}

func memberChain(node ast.Node) (ast.Node, []*ast.MemberNode) {
	if m, ok := node.(*ast.MemberNode); ok {
		root, chain := memberChain(m.Node)
		return root, append(chain, m)
	}
	return node, nil
}

func (a *referenceAnalysis) member(node *ast.Node, member *ast.MemberNode, locals map[string]bool) error {
	root, chain := memberChain(member)
	id, identifier := root.(*ast.IdentifierNode)
	if identifier && !locals[id.Value] && len(a.prefix) > 0 && id.Value == a.prefix[0] {
		matches := len(chain) >= len(a.prefix)-1
		for index := 1; matches && index < len(a.prefix); index++ {
			property, ok := chain[index-1].Property.(*ast.StringNode)
			matches = ok && property.Value == a.prefix[index]
		}
		if matches {
			if len(chain) < len(a.prefix)+1 {
				return fmt.Errorf("formula reference must identify a static row and column in the current table")
			}
			row, rok := chain[len(a.prefix)-1].Property.(*ast.StringNode)
			column, cok := chain[len(a.prefix)].Property.(*ast.StringNode)
			if !rok || !cok {
				return fmt.Errorf("dynamic table references are not supported; use static row and column names")
			}
			if err := a.add(row.Value, column.Value); err != nil {
				return err
			}
			replacement := a.cell(row.Value, column.Value)
			for _, tail := range chain[len(a.prefix)+1:] {
				if _, ok := tail.Property.(*ast.StringNode); !ok {
					return fmt.Errorf("formula references require static field names; array element references are not supported")
				}
				if err := a.walk(&tail.Property, locals); err != nil {
					return err
				}
				copy := *tail
				copy.Node = replacement
				replacement = &copy
			}
			*node = replacement
			return nil
		}
	}
	if identifier && !locals[id.Value] && !a.columns[id.Value] {
		return fmt.Errorf("formula reference %q must target the current table using static names", member.String())
	}
	if _, ok := member.Property.(*ast.StringNode); !ok {
		return fmt.Errorf("formula references require static field names; dynamic and array element references are not supported")
	}
	if err := a.walk(&member.Node, locals); err != nil {
		return err
	}
	return a.walk(&member.Property, locals)
}
