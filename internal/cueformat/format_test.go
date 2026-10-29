package cueformat

import (
	"strings"
	"testing"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/parser"
)

func TestChangedMultilineListKeepsLegacyCommasAndComments(t *testing.T) {
	for _, version := range []string{"v0.9.0", "v0.17.0"} {
		t.Run(version, func(t *testing.T) {
			file, err := parser.ParseFile("input.cue", `package example
values: [
 "first", // keep first
 "middle", // keep middle
 "last",
]
`, parser.ParseComments, parser.Version(version))
			if err != nil {
				t.Fatal(err)
			}
			var list *ast.ListLit
			ast.Walk(file, func(n ast.Node) bool {
				if value, ok := n.(*ast.ListLit); ok {
					list = value
				}
				return true
			}, nil)
			old := list.Elts[1]
			next := ast.NewString("中文替换值")
			ast.SetComments(next, ast.Comments(old))
			list.Elts[1] = next
			position := list.Lbrack
			output, err := Node(file)
			if err != nil {
				t.Fatal(err)
			}
			if list.Lbrack != position {
				t.Fatal("formatter mutated original position annotations")
			}
			if _, err := parser.ParseFile("output.cue", output, parser.Version("v0.9.0")); err != nil {
				t.Fatalf("legacy output invalid: %v\n%s", err, output)
			}
			for _, part := range []string{"first", "中文替换值", "last", "keep first", "keep middle"} {
				if !strings.Contains(string(output), part) {
					t.Fatalf("lost %s: %s", part, output)
				}
			}
		})
	}
}
func TestCommaFreeModernInputIsSafelySeparated(t *testing.T) {
	file, err := parser.ParseFile("modern.cue", `package example
values: [
 "one"
 "two"
]
`, parser.Version("v0.17.0"))
	if err != nil {
		t.Fatal(err)
	}
	out, err := Node(file)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parser.ParseFile("legacy.cue", out, parser.Version("v0.9.0")); err != nil {
		t.Fatalf("explicit separators not inserted: %v\n%s", err, out)
	}
}
