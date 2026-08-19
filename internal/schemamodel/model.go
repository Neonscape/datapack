// Package schemamodel describes the restricted Schema shapes supported by the editor.
package schemamodel

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/format"
	"cuelang.org/go/cue/token"
)

var negativeComparisonPattern = regexp.MustCompile(`(<|<=)[\t ]+(-[0-9]+(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?)`)

// NormalizeNegativeComparisons protects valid comparisons such as `< -50`
// from CUE's formatter, which otherwise emits the distinct `<-` token.
// Bare `<-50` remains invalid because this only matches existing whitespace.
func NormalizeNegativeComparisons(expression string) string {
	return negativeComparisonPattern.ReplaceAllString(expression, `$1 ($2)`)
}

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

// Requirement kinds classify a single conjunct in a field Schema.
const (
	RequirementCompare   = "compare"
	RequirementRegex     = "regex"
	RequirementNonEmpty  = "nonEmpty"
	RequirementLength    = "length"
	RequirementReference = "reference"
	RequirementCustom    = "custom"
)

type Requirement struct {
	Kind       string `json:"kind,omitempty"`
	Op         string `json:"op,omitempty"`
	Value      string `json:"value,omitempty"`
	Pattern    string `json:"pattern,omitempty"`
	Min        *int   `json:"min,omitempty"`
	Max        *int   `json:"max,omitempty"`
	Reference  string `json:"reference,omitempty"`
	Expression string `json:"expression,omitempty"`
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
	return Node{Kind: KindUnknown, Requirements: []Requirement{classifyRequirement(expression)}}, nil
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
			result.Requirements = append(result.Requirements, Requirement{Kind: RequirementReference, Expression: node.Reference, Reference: node.Reference})
			continue
		}
		if hasBaseType(node) {
			if result.Kind != KindUnknown {
				result.Requirements = append(result.Requirements, classifyRequirement(part))
				continue
			}
			result = mergeNode(result, node)
			continue
		}
		result.Requirements = append(result.Requirements, classifyRequirement(part))
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
		result.Requirements = append(result.Requirements, classifyRequirement(part))
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

// classifyRequirement turns one conjunct into a typed predicate while keeping
// the raw expression for round-trip fidelity.
func classifyRequirement(expression ast.Expr) Requirement {
	text, err := expressionText(expression)
	if err != nil {
		text = ""
	}
	requirement := Requirement{Kind: RequirementCustom, Expression: text}
	switch value := expression.(type) {
	case *ast.UnaryExpr:
		operand, err := expressionText(value.X)
		if err != nil {
			return requirement
		}
		switch value.Op {
		case token.GEQ, token.GTR, token.LEQ, token.LSS, token.EQL:
			requirement.Kind = RequirementCompare
			requirement.Op = tokenSymbol(value.Op)
			requirement.Value = operand
		case token.NEQ:
			if operand == `""` {
				requirement.Kind = RequirementNonEmpty
				requirement.Op = "!="
				requirement.Value = `""`
				return requirement
			}
			requirement.Kind = RequirementCompare
			requirement.Op = "!="
			requirement.Value = operand
		case token.MAT, token.NMAT:
			if pattern, ok := stringLiteral(value.X); ok {
				requirement.Kind = RequirementRegex
				requirement.Op = tokenSymbol(value.Op)
				requirement.Pattern = pattern
			}
		}
	case *ast.CallExpr:
		if selector, ok := value.Fun.(*ast.SelectorExpr); ok {
			if identifier, ok := selector.X.(*ast.Ident); ok && identifier.Name == "strings" {
				arg, ok := integerLiteral(value.Args)
				if ok {
					name, _, err := ast.LabelName(selector.Sel)
					if err == nil {
						switch name {
						case "MinRunes":
							requirement.Kind = RequirementLength
							requirement.Min = &arg
						case "MaxRunes":
							requirement.Kind = RequirementLength
							requirement.Max = &arg
						}
					}
				}
			}
		}
	}
	return requirement
}

func tokenSymbol(operator token.Token) string {
	switch operator {
	case token.GEQ:
		return ">="
	case token.GTR:
		return ">"
	case token.LEQ:
		return "<="
	case token.LSS:
		return "<"
	case token.EQL:
		return "=="
	case token.NEQ:
		return "!="
	case token.MAT:
		return "=~"
	case token.NMAT:
		return "!~"
	default:
		return ""
	}
}

func stringLiteral(expression ast.Expr) (string, bool) {
	literal, ok := expression.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return "", false
	}
	value, err := strconv.Unquote(literal.Value)
	return value, err == nil
}

func integerLiteral(arguments []ast.Expr) (int, bool) {
	if len(arguments) != 1 {
		return 0, false
	}
	literal, ok := arguments[0].(*ast.BasicLit)
	if !ok || literal.Kind != token.INT {
		return 0, false
	}
	value, err := strconv.Atoi(literal.Value)
	return value, err == nil
}

// Print serializes a Node back into its canonical CUE expression.
func Print(node Node) (string, error) {
	if node.Kind == KindEnum {
		values := make([]string, 0, len(node.EnumValues)+2)
		for _, value := range node.EnumValues {
			values = append(values, value.Expression)
		}
		if node.Default != nil {
			values = append(values, "*"+node.Default.Expression)
		}
		if node.Nullable {
			values = append(values, "null")
		}
		if len(values) == 0 {
			return "", fmt.Errorf("enum Schema has no values")
		}
		return strings.Join(values, " | "), nil
	}
	conjunction := []string{}
	base, err := baseExpression(node)
	if err != nil {
		return "", err
	}
	if base != "" {
		conjunction = append(conjunction, base)
	}
	for _, requirement := range node.Requirements {
		conjunction = append(conjunction, requirementExpression(requirement))
	}
	if len(conjunction) == 0 {
		return "", fmt.Errorf("Schema has no type or requirements")
	}
	combined := strings.Join(conjunction, " & ")
	disjunction := []string{}
	if node.Default != nil {
		disjunction = append(disjunction, "*"+node.Default.Expression)
	}
	if node.Nullable {
		disjunction = append(disjunction, "null")
	}
	if len(disjunction) > 0 {
		if len(conjunction) > 1 {
			return "(" + combined + ") | " + strings.Join(disjunction, " | "), nil
		}
		return combined + " | " + strings.Join(disjunction, " | "), nil
	}
	return combined, nil
}

func baseExpression(node Node) (string, error) {
	switch node.Kind {
	case KindString:
		return "string", nil
	case KindInteger:
		return "int", nil
	case KindNumber:
		return "number", nil
	case KindBoolean:
		return "bool", nil
	case KindReference:
		if node.Reference == "" {
			return "", fmt.Errorf("reference Schema is missing its target")
		}
		return node.Reference, nil
	case KindArray:
		if node.Element == nil {
			return "", fmt.Errorf("array Schema is missing its element type")
		}
		element, err := Print(*node.Element)
		if err != nil {
			return "", err
		}
		return "[..." + element + "]", nil
	case KindStruct:
		fields := make([]string, 0, len(node.Fields))
		for _, field := range node.Fields {
			schema, err := Print(field.Schema)
			if err != nil {
				return "", err
			}
			if field.Optional {
				fields = append(fields, field.Name+"?: "+schema)
			} else {
				fields = append(fields, field.Name+": "+schema)
			}
		}
		return "{" + strings.Join(fields, ", ") + "}", nil
	case KindUnknown:
		return "", nil
	default:
		return "", fmt.Errorf("unsupported Schema kind %q", node.Kind)
	}
}

func requirementExpression(requirement Requirement) string {
	switch requirement.Kind {
	case RequirementCompare:
		value := strings.TrimSpace(requirement.Value)
		if strings.HasPrefix(value, "-") {
			value = "(" + value + ")"
		}
		return requirement.Op + " " + value
	case RequirementRegex:
		return requirement.Op + " " + strconv.Quote(requirement.Pattern)
	case RequirementNonEmpty:
		return `!= ""`
	case RequirementLength:
		if requirement.Min != nil {
			return fmt.Sprintf("strings.MinRunes(%d)", *requirement.Min)
		}
		if requirement.Max != nil {
			return fmt.Sprintf("strings.MaxRunes(%d)", *requirement.Max)
		}
		return requirement.Expression
	case RequirementReference:
		if requirement.Reference != "" {
			return requirement.Reference
		}
		return requirement.Expression
	default:
		return requirement.Expression
	}
}
