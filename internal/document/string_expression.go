package document

import (
	"strconv"
	"strings"

	"cuelang.org/go/cue/format"
	"cuelang.org/go/cue/parser"
)

func EncodeStringExpression(value string) string {
	if strings.ContainsAny(value, "\r\n") && !strings.Contains(value, `"""`) {
		value = strings.ReplaceAll(strings.ReplaceAll(value, "\r\n", "\n"), "\r", "\n")
		literal := "\"\"\"\n" + value + "\n\"\"\""
		if expression, err := parser.ParseExpr("string-value.cue", literal); err == nil {
			if formatted, err := format.Node(expression); err == nil {
				return strings.TrimSpace(string(formatted))
			}
		}
	}
	literal := strconv.Quote(value)
	expression, err := parser.ParseExpr("string-value.cue", literal)
	if err == nil {
		if formatted, err := format.Node(expression); err == nil {
			return strings.TrimSpace(string(formatted))
		}
	}
	return literal
}

// SetCellText writes a plain user-entered string as a CUE string expression.
func SetCellText(path, rowID, column, value string) (Table, error) {
	return SetCellValue(path, rowID, column, EncodeStringExpression(value))
}

func displayStringExpression(raw string) (string, bool) {
	if strings.HasPrefix(raw, `"""`) && strings.HasSuffix(raw, `"""`) {
		value := raw[3 : len(raw)-3]
		value = strings.TrimPrefix(value, "\n")
		value = strings.TrimRight(value, " \t")
		value = strings.TrimSuffix(value, "\n")
		lines := strings.Split(value, "\n")
		prefix := ""
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}
			prefix = line[:len(line)-len(strings.TrimLeft(line, " \t"))]
			break
		}
		if prefix != "" {
			for index, line := range lines {
				lines[index] = strings.TrimPrefix(line, prefix)
			}
		}
		value = strings.Join(lines, "\n")
		return value, true
	}
	value, err := strconv.Unquote(raw)
	return value, err == nil
}
