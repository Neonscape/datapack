// Package schemamodel describes the restricted Schema shapes supported by the editor.
package schemamodel

import (
	"fmt"
	"strconv"
	"strings"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/format"
	"cuelang.org/go/cue/token"
)

type Kind string

const (
	KindString    Kind = "string"
	KindInteger   Kind = "integer"
	KindNumber    Kind = "number"
	KindBoolean   Kind = "boolean"
	KindEnum      Kind = "enum"
	KindArray     Kind = "array"
	KindStruct    Kind = "struct"
	KindReference Kind = "reference"
	KindUnknown   Kind = "unknown"
)

type Value struct {
	Expression string `json:"expression"`
}

type Requirement struct {
	Expression string `json:"expression"`
	Reference  string `json:"reference,omitempty"`
}

type Field struct {
	Name     string `json:"name"`
	Optional bool   `json:"optional"`
	Schema   Node   `json:"schema"`
}

type Node struct {
	Kind         Kind          `json:"kind"`
	Nullable     bool          `json:"nullable"`
	Default      *Value        `json:"default,omitempty"`
	Reference    string        `json:"reference,omitempty"`
	EnumValues   []Value       `json:"enumValues,omitempty"`
	Element      *Node         `json:"element,omitempty"`
	Fields       []Field       `json:"fields,omitempty"`
	Requirements []Requirement `json:"requirements,omitempty"`
}

// Parse converts the supported CUE Schema subset into an editor model. Unknown
// expressions remain requirements so the caller can report or preserve them
// without discarding source meaning.
func Parse(expression ast.Expr) (Node, error) {
	if expression == nil {
		return Node{}, fmt.Errorf("schema expression is missing")
	}
	return parse(expression)
}

func parse(expression ast.Expr) (Node, error) {
	if paren, ok := expression.(*ast.ParenExpr); ok {
		return parse(paren.X)
	}
	if binary, ok := expression.(*ast.BinaryExpr); ok {
		switch binary.Op {
		case token.AND:
			return parseAnd(binary)
		case token.OR:
			return parseOr(binary)
		}
	}
	if unary, ok := expression.(*ast.UnaryExpr); ok && unary.Op == token.MUL {
		value, err := expressionText(unary.X)
		if err != nil {
			return Node{}, err
		}
		return Node{Kind: KindUnknown, Default: &Value{Expression: value}}, nil
	}
	if identifier, ok := expression.(*ast.Ident); ok {
		if kind, ok := builtinKind(identifier.Name); ok {
			return Node{Kind: kind}, nil
		}
		return Node{Kind: KindReference, Reference: identifier.Name}, nil
	}
	if _, ok := expression.(*ast.SelectorExpr); ok {
		reference, err := expressionText(expression)
		if err != nil {
			return Node{}, err
		}
		return Node{Kind: KindReference, Reference: reference}, nil
	}
	if structLiteral, ok := expression.(*ast.StructLit); ok {
		return parseStruct(structLiteral)
	}
	if list, ok := expression.(*ast.ListLit); ok {
		return parseArray(list)
	}
	if literal, ok := expression.(*ast.BasicLit); ok {
		return Node{Kind: KindEnum, EnumValues: []Value{{Expression: literal.Value}}}, nil
	}
	text, err := expressionText(expression)
	if err != nil {
		return Node{}, err
	}
	return Node{Kind: KindUnknown, Requirements: []Requirement{{Expression: text}}}, nil
}

func parseAnd(expression *ast.BinaryExpr) (Node, error) {
	parts := flatten(expression, token.AND)
	result := Node{Kind: KindUnknown}
	for _, part := range parts {
		node, err := parse(part)
		if err != nil {
			return Node{}, err
		}
		if node.Kind == KindReference && result.Kind != KindUnknown {
			result.Requirements = append(result.Requirements, Requirement{Expression: node.Reference, Reference: node.Reference})
			continue
		}
		if hasBaseType(node) {
			if result.Kind != KindUnknown {
				text, err := expressionText(part)
				if err != nil {
					return Node{}, err
				}
				result.Requirements = append(result.Requirements, Requirement{Expression: text})
				continue
			}
			result = mergeNode(result, node)
			continue
		}
		text, err := expressionText(part)
		if err != nil {
			return Node{}, err
		}
		result.Requirements = append(result.Requirements, Requirement{Expression: text})
	}
	return result, nil
}

func parseOr(expression *ast.BinaryExpr) (Node, error) {
	parts := flatten(expression, token.OR)
	result := Node{Kind: KindUnknown}
	for _, part := range parts {
		if literal, ok := part.(*ast.BasicLit); ok && literal.Kind == token.NULL {
			result.Nullable = true
			continue
		}
		if defaultValue, ok := part.(*ast.UnaryExpr); ok && defaultValue.Op == token.MUL {
			text, err := expressionText(defaultValue.X)
			if err != nil {
				return Node{}, err
			}
			result.Default = &Value{Expression: text}
			continue
		}
		node, err := parse(part)
		if err != nil {
			return Node{}, err
		}
		if node.Kind == KindEnum {
			if result.Kind == KindUnknown || result.Kind == KindEnum {
				result.Kind = KindEnum
				result.EnumValues = append(result.EnumValues, node.EnumValues...)
				continue
			}
		}
		if hasBaseType(node) && result.Kind == KindUnknown {
			result = mergeNode(result, node)
			continue
		}
		text, err := expressionText(part)
		if err != nil {
			return Node{}, err
		}
		result.Requirements = append(result.Requirements, Requirement{Expression: text})
	}
	return result, nil
}

func parseArray(list *ast.ListLit) (Node, error) {
	if len(list.Elts) != 1 {
		text, err := expressionText(list)
		if err != nil {
			return Node{}, err
		}
		return Node{Kind: KindUnknown, Requirements: []Requirement{{Expression: text}}}, nil
	}
	ellipsis, ok := list.Elts[0].(*ast.Ellipsis)
	if !ok || ellipsis.Type == nil {
		text, err := expressionText(list)
		if err != nil {
			return Node{}, err
		}
		return Node{Kind: KindUnknown, Requirements: []Requirement{{Expression: text}}}, nil
	}
	element, err := parse(ellipsis.Type)
	if err != nil {
		return Node{}, err
	}
	return Node{Kind: KindArray, Element: &element}, nil
}

func parseStruct(structLiteral *ast.StructLit) (Node, error) {
	result := Node{Kind: KindStruct}
	for _, declaration := range structLiteral.Elts {
		field, ok := declaration.(*ast.Field)
		if !ok {
			continue
		}
		name, _, err := ast.LabelName(field.Label)
		if err != nil {
			return Node{}, fmt.Errorf("read struct field: %w", err)
		}
		schema, err := parse(field.Value)
		if err != nil {
			return Node{}, err
		}
		result.Fields = append(result.Fields, Field{Name: name, Optional: field.Constraint == token.OPTION, Schema: schema})
	}
	return result, nil
}

func flatten(expression ast.Expr, operator token.Token) []ast.Expr {
	binary, ok := expression.(*ast.BinaryExpr)
	if !ok || binary.Op != operator {
		return []ast.Expr{expression}
	}
	parts := flatten(binary.X, operator)
	return append(parts, flatten(binary.Y, operator)...)
}

func hasBaseType(node Node) bool {
	return node.Kind != KindUnknown && node.Kind != KindEnum || node.Kind == KindEnum
}

func mergeNode(existing, next Node) Node {
	next.Nullable = next.Nullable || existing.Nullable
	if next.Default == nil {
		next.Default = existing.Default
	}
	next.Requirements = append(existing.Requirements, next.Requirements...)
	return next
}

func builtinKind(name string) (Kind, bool) {
	switch name {
	case "string":
		return KindString, true
	case "int":
		return KindInteger, true
	case "number":
		return KindNumber, true
	case "bool":
		return KindBoolean, true
	default:
		return "", false
	}
}

func expressionText(expression ast.Expr) (string, error) {
	bytes, err := format.Node(expression)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(bytes)), nil
}

func (v Value) String() string {
	if unquoted, err := strconv.Unquote(v.Expression); err == nil {
		return unquoted
	}
	return v.Expression
}
