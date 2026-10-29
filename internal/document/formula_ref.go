package document

import (
	"fmt"
	"strings"

	"cuelang.org/go/cue/ast"
	exprast "github.com/expr-lang/expr/ast"
	exprparser "github.com/expr-lang/expr/parser"
)

func renameFormulaRowReferences(file *ast.File, oldID, newID string) error {
	formulas := formulaFields(file)
	if formulas == nil {
		return nil
	}
	if fieldsOfStruct(formulas.Value)[newID] != nil {
		return fmt.Errorf("formulas for row %q already exist", newID)
	}
	if row := fieldsOfStruct(formulas.Value)[oldID]; row != nil {
		row.Label = ast.NewIdent(newID)
	}
	return rewriteFormulaExpressions(formulas.Value, formulaRename{identity: formulaIdentity(file), rowRename: true, oldRow: oldID, newRow: newID})
}

func renameFormulaColumnReferences(file *ast.File, oldName, newName string) error {
	formulas := formulaFields(file)
	if formulas == nil {
		return nil
	}
	for _, row := range orderedStructFields(formulas.Value) {
		if fieldsOfStruct(row.Value)[newName] != nil {
			return fmt.Errorf("formula for %s.%s already exists", labelOf(row), newName)
		}
		renameNamedField(row.Value, oldName, newName)
	}
	return rewriteFormulaExpressions(formulas.Value, formulaRename{identity: formulaIdentity(file), columnRename: true, oldColumn: oldName, newColumn: newName})
}

func renameFormulaTableReferences(file *ast.File, newNamespace, newName string) error {
	formulas := formulaFields(file)
	if formulas == nil {
		return nil
	}
	next := tableIdentity{namespace: newNamespace, name: newName}
	columns := map[string]bool{}
	for name := range fieldsOfStruct(requiredField(fieldsOfFile(file), "#Row").Value) {
		columns[name] = true
	}
	return rewriteFormulaExpressions(formulas.Value, formulaRename{identity: formulaIdentity(file), nextIdentity: &next, columns: columns})
}

func formulaFields(file *ast.File) *ast.Field {
	tool := fieldsOfFile(file)["_tool"]
	if tool == nil {
		return nil
	}
	return fieldsOfStruct(tool.Value)["formulas"]
}

type tableIdentity struct{ namespace, name string }

func formulaIdentity(file *ast.File) tableIdentity {
	tool := requiredField(fieldsOfFile(file), "_tool")
	table := requiredField(fieldsOfStruct(tool.Value), "table")
	fields := fieldsOfStruct(table.Value)
	return tableIdentity{namespace: stringOf(requiredField(fields, "namespace").Value), name: stringOf(requiredField(fields, "name").Value)}
}
func (identity tableIdentity) prefix() []string {
	var prefix []string
	if identity.namespace != "" {
		prefix = strings.Split(identity.namespace, ".")
	}
	return append(prefix, identity.name)
}

type formulaRename struct {
	identity                tableIdentity
	nextIdentity            *tableIdentity
	columns                 map[string]bool
	rowRename, columnRename bool
	oldRow, newRow          string
	oldColumn, newColumn    string
}

func rewriteFormulaExpressions(expression ast.Expr, rename formulaRename) error {
	if _, ok := expression.(*ast.StructLit); !ok {
		return fmt.Errorf("formulas must be a struct")
	}
	for _, row := range orderedStructFields(expression) {
		if _, ok := row.Value.(*ast.StructLit); !ok {
			return fmt.Errorf("formula row %q must be a struct", labelOf(row))
		}
		for _, formula := range orderedStructFields(row.Value) {
			text, ok := formulaExpression(formula.Value)
			if !ok {
				return fmt.Errorf("formula %q.%q must be a string", labelOf(row), labelOf(formula))
			}
			tree, err := exprparser.Parse(text)
			if err != nil {
				return fmt.Errorf("parse formula %q.%q: %w", labelOf(row), labelOf(formula), err)
			}
			if err := rename.walk(&tree.Node, labelOf(row), map[string]bool{}); err != nil {
				return fmt.Errorf("rewrite formula %q.%q: %w", labelOf(row), labelOf(formula), err)
			}
			rewritten := tree.Node.String()
			if _, err := exprparser.Parse(rewritten); err != nil {
				return fmt.Errorf("renamed formula %q.%q is not valid syntax: %w", labelOf(row), labelOf(formula), err)
			}
			formula.Value = stringLiteral(rewritten)
		}
	}
	return nil
}

// Walk with lexical scope and handle complete cell addresses before visiting
// identifiers. Callees, namespace roots, static keys and locals are not columns.
func (rename formulaRename) walk(node *exprast.Node, rowID string, locals map[string]bool) error {
	if *node == nil {
		return nil
	}
	walk := func(n *exprast.Node) error { return rename.walk(n, rowID, locals) }
	switch n := (*node).(type) {
	case *exprast.IdentifierNode:
		if locals[n.Value] {
			return nil
		}
		if rename.nextIdentity != nil && rename.columns[n.Value] {
			prefix := rename.nextIdentity.prefix()
			if n.Value == prefix[0] {
				// A bare object column may otherwise become indistinguishable
				// from the new namespace/table prefix after a table rename.
				*node = makeFormulaPath(append(prefix, rowID, n.Value))
				return nil
			}
		}
		if !rename.columnRename || n.Value != rename.oldColumn {
			return nil
		}
		prefix := rename.identity.prefix()
		if !locals[rename.newColumn] && rename.newColumn != prefix[0] && bareFormulaIdentifier(rename.newColumn) {
			n.Value = rename.newColumn
			return nil
		}
		if locals[prefix[0]] {
			return fmt.Errorf("new column %q and namespace root %q are hidden by local bindings", rename.newColumn, prefix[0])
		}
		if !bareFormulaIdentifier(prefix[0]) {
			return fmt.Errorf("namespace root %q cannot be represented as a formula identifier", prefix[0])
		}
		// A new bare name could capture a let binding. Use the explicit cell address.
		*node = makeFormulaPath(append(prefix, rowID, rename.newColumn))
		return nil
	case *exprast.NilNode, *exprast.IntegerNode, *exprast.FloatNode, *exprast.BoolNode, *exprast.StringNode, *exprast.ConstantNode:
		return nil
	case *exprast.MemberNode:
		if matched, err := rename.member(node, n, locals); matched || err != nil {
			return err
		}
		if _, ok := n.Property.(*exprast.StringNode); !ok {
			return fmt.Errorf("dynamic or array element formula references cannot be renamed")
		}
		if err := walk(&n.Node); err != nil {
			return err
		}
		return walk(&n.Property)
	case *exprast.ChainNode:
		return walk(&n.Node)
	case *exprast.UnaryNode:
		return walk(&n.Node)
	case *exprast.BinaryNode:
		if err := walk(&n.Left); err != nil {
			return err
		}
		return walk(&n.Right)
	case *exprast.CallNode:
		for i := range n.Arguments {
			if err := walk(&n.Arguments[i]); err != nil {
				return err
			}
		}
		return nil
	case *exprast.BuiltinNode:
		for i := range n.Arguments {
			if err := walk(&n.Arguments[i]); err != nil {
				return err
			}
		}
		return nil
	case *exprast.VariableDeclaratorNode:
		if err := walk(&n.Value); err != nil {
			return err
		}
		next := make(map[string]bool, len(locals)+1)
		for key, value := range locals {
			next[key] = value
		}
		next[n.Name] = true
		return rename.walk(&n.Expr, rowID, next)
	case *exprast.ConditionalNode:
		for _, child := range []*exprast.Node{&n.Cond, &n.Exp1, &n.Exp2} {
			if err := walk(child); err != nil {
				return err
			}
		}
		return nil
	case *exprast.SequenceNode:
		for i := range n.Nodes {
			if err := walk(&n.Nodes[i]); err != nil {
				return err
			}
		}
		return nil
	case *exprast.ArrayNode:
		for i := range n.Nodes {
			if err := walk(&n.Nodes[i]); err != nil {
				return err
			}
		}
		return nil
	case *exprast.MapNode:
		for i := range n.Pairs {
			if err := walk(&n.Pairs[i]); err != nil {
				return err
			}
		}
		return nil
	case *exprast.PairNode:
		if err := walk(&n.Key); err != nil {
			return err
		}
		return walk(&n.Value)
	default:
		return fmt.Errorf("unsupported formula expression %T during rename", n)
	}
}

func formulaMemberChain(node exprast.Node) (exprast.Node, []*exprast.MemberNode) {
	if member, ok := node.(*exprast.MemberNode); ok {
		root, parts := formulaMemberChain(member.Node)
		return root, append(parts, member)
	}
	return node, nil
}

func (rename formulaRename) member(node *exprast.Node, member *exprast.MemberNode, locals map[string]bool) (bool, error) {
	root, parts := formulaMemberChain(member)
	identifier, ok := root.(*exprast.IdentifierNode)
	prefix := rename.identity.prefix()
	if !ok || locals[identifier.Value] || identifier.Value != prefix[0] {
		return false, nil
	}
	for i := 1; i < len(prefix); i++ {
		if i-1 >= len(parts) {
			return false, nil
		}
		property, ok := parts[i-1].Property.(*exprast.StringNode)
		if !ok {
			return false, fmt.Errorf("dynamic formula addresses cannot be renamed")
		}
		if property.Value != prefix[i] {
			return false, nil
		}
	}
	if len(parts) < len(prefix)+1 {
		return true, fmt.Errorf("formula address must include a static row and column")
	}
	row, rok := parts[len(prefix)-1].Property.(*exprast.StringNode)
	column, cok := parts[len(prefix)].Property.(*exprast.StringNode)
	if !rok || !cok {
		return true, fmt.Errorf("dynamic formula addresses cannot be renamed")
	}
	for _, part := range parts[len(prefix)+1:] {
		if _, ok := part.Property.(*exprast.StringNode); !ok {
			return true, fmt.Errorf("dynamic or array element formula references cannot be renamed")
		}
	}
	rowValue, columnValue := row.Value, column.Value
	if rename.rowRename && rowValue == rename.oldRow {
		rowValue = rename.newRow
	}
	if rename.columnRename && columnValue == rename.oldColumn {
		columnValue = rename.newColumn
	}
	nextPrefix := prefix
	if rename.nextIdentity != nil {
		nextPrefix = rename.nextIdentity.prefix()
		if !bareFormulaIdentifier(nextPrefix[0]) {
			return true, fmt.Errorf("new namespace root %q cannot be represented as a formula identifier", nextPrefix[0])
		}
		if locals[nextPrefix[0]] {
			return true, fmt.Errorf("new namespace root %q is hidden by a local binding", nextPrefix[0])
		}
	}
	// Preserve optional/member metadata on row, column and field tails. Prefix
	// namespaces identify the current table and do not represent optional data.
	var replacement exprast.Node = &exprast.IdentifierNode{Value: nextPrefix[0]}
	for i := 1; i < len(nextPrefix); i++ {
		next := &exprast.MemberNode{Node: replacement, Property: &exprast.StringNode{Value: nextPrefix[i]}}
		if len(nextPrefix) == len(prefix) {
			copy := *parts[i-1]
			copy.Node = replacement
			copy.Property = next.Property
			next = &copy
		}
		replacement = next
	}
	for i, part := range parts[len(prefix)-1:] {
		copy := *part
		copy.Node = replacement
		if i == 0 {
			copy.Property = &exprast.StringNode{Value: rowValue}
		}
		if i == 1 {
			copy.Property = &exprast.StringNode{Value: columnValue}
		}
		replacement = &copy
	}
	*node = replacement
	return true, nil
}

func makeFormulaPath(path []string) exprast.Node {
	var node exprast.Node = &exprast.IdentifierNode{Value: path[0]}
	for _, part := range path[1:] {
		node = &exprast.MemberNode{Node: node, Property: &exprast.StringNode{Value: part}}
	}
	return node
}

// CUE identifiers such as nil and and are literals/operators in Expr. Do not
// serialize a column read into an expression with a different AST meaning.
func bareFormulaIdentifier(name string) bool {
	tree, err := exprparser.Parse(name)
	if err != nil {
		return false
	}
	identifier, ok := tree.Node.(*exprast.IdentifierNode)
	return ok && identifier.Value == name
}
