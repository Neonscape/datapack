package document

import (
	"fmt"

	"cuelang.org/go/cue/ast"
	exprast "github.com/expr-lang/expr/ast"
	exprparser "github.com/expr-lang/expr/parser"
)

func renameFormulaRowReferences(file *ast.File, oldID, newID string) error {
	identity := formulaIdentity(file)
	formulas := formulaFields(file)
	if formulas == nil {
		return nil
	}
	if row := fieldsOfStruct(formulas.Value)[oldID]; row != nil {
		row.Label = ast.NewIdent(newID)
	}
	return rewriteFormulaExpressions(formulas.Value, func(node *exprast.Node) {
		member, ok := (*node).(*exprast.MemberNode)
		if !ok || !matchesFormulaPath(member, identity, oldID, "") {
			return
		}
		setFormulaPathPart(member, 2, newID)
	})
}

func renameFormulaColumnReferences(file *ast.File, oldName, newName string) error {
	identity := formulaIdentity(file)
	formulas := formulaFields(file)
	if formulas == nil {
		return nil
	}
	for _, row := range orderedStructFields(formulas.Value) {
		renameNamedField(row.Value, oldName, newName)
	}
	return rewriteFormulaExpressions(formulas.Value, func(node *exprast.Node) {
		switch value := (*node).(type) {
		case *exprast.IdentifierNode:
			if value.Value == oldName {
				value.Value = newName
			}
		case *exprast.MemberNode:
			if matchesFormulaPath(value, identity, "", oldName) {
				setFormulaPathPart(value, 3, newName)
			}
		}
	})
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

func rewriteFormulaExpressions(expression ast.Expr, rewrite func(*exprast.Node)) error {
	for _, row := range orderedStructFields(expression) {
		for _, formula := range orderedStructFields(row.Value) {
			text, ok := formulaExpression(formula.Value)
			if !ok {
				return fmt.Errorf("formula %q.%q must be a string", labelOf(row), labelOf(formula))
			}
			tree, err := exprparser.Parse(text)
			if err != nil {
				return fmt.Errorf("parse formula %q.%q: %w", labelOf(row), labelOf(formula), err)
			}
			exprast.Walk(&tree.Node, exprVisitor(rewrite))
			formula.Value = stringLiteral(tree.Node.String())
		}
	}
	return nil
}

type exprVisitor func(*exprast.Node)

func (visitor exprVisitor) Visit(node *exprast.Node) { visitor(node) }

func matchesFormulaPath(member *exprast.MemberNode, identity tableIdentity, rowID, column string) bool {
	path := formulaPath(member)
	if len(path) != 4 || path[0] != identity.namespace || path[1] != identity.name {
		return false
	}
	return (rowID == "" || path[2] == rowID) && (column == "" || path[3] == column)
}

func formulaPath(node exprast.Node) []string {
	switch value := node.(type) {
	case *exprast.IdentifierNode:
		return []string{value.Value}
	case *exprast.MemberNode:
		path := formulaPath(value.Node)
		property, ok := value.Property.(*exprast.StringNode)
		if !ok {
			return nil
		}
		return append(path, property.Value)
	default:
		return nil
	}
}

func setFormulaPathPart(node *exprast.MemberNode, index int, replacement string) {
	parts := formulaMembers(node)
	if index == 0 {
		if identifier, ok := parts[0].(*exprast.IdentifierNode); ok {
			identifier.Value = replacement
		}
		return
	}
	if index < len(parts) {
		if property, ok := parts[index].(*exprast.StringNode); ok {
			property.Value = replacement
		}
	}
}

func formulaMembers(node *exprast.MemberNode) []exprast.Node {
	if parent, ok := node.Node.(*exprast.MemberNode); ok {
		return append(formulaMembers(parent), node.Property)
	}
	return []exprast.Node{node.Node, node.Property}
}
