package cuesource

import (
	"cuelang.org/go/cue/ast"
	"strconv"
	"strings"
)

// EditorTitles contains optional presentation metadata. It is never data identity
// or a Schema expression; old documents simply produce empty titles.
type EditorTitles struct {
	Table       string
	Schema      string
	Definitions map[string]string
}

func ReadEditorTitles(file *ast.File) EditorTitles {
	result := EditorTitles{Definitions: map[string]string{}}
	editor := topLevelFields(file)["_editor"]
	if editor == nil {
		return result
	}
	fields := structFields(editor.Value)
	title := func(field *ast.Field) string {
		if field == nil {
			return ""
		}
		value := structFields(field.Value)["title"]
		if value == nil {
			return ""
		}
		literal, ok := value.Value.(*ast.BasicLit)
		if !ok {
			return ""
		}
		text, err := strconv.Unquote(literal.Value)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(text)
	}
	result.Table = title(fields["table"])
	result.Schema = title(fields["schema"])
	if definitions := fields["schemas"]; definitions != nil {
		for name, field := range structFields(definitions.Value) {
			result.Definitions[name] = title(field)
		}
	}
	return result
}
