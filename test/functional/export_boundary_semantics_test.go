package functional

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"datapack/internal/workspace"
)

// TestExportBoundarySemantics exercises the public service and a separate Python
// CSV/JSON consumer. Expected values are literal data requirements, not values
// decoded by the Go compiler under test. Set DATAPACK_EXPORT_BOUNDARY_ARTIFACTS
// to retain the isolated input project and actual output for review.
func TestExportBoundarySemantics(t *testing.T) {
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("Python 3 is required for the independent CSV consumer")
	}
	root := t.TempDir()
	if base := os.Getenv("DATAPACK_EXPORT_BOUNDARY_ARTIFACTS"); base != "" {
		if err := os.MkdirAll(base, 0755); err != nil {
			t.Fatal(err)
		}
		root, err = os.MkdirTemp(base, "fixture-")
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("Retained fixture: %s", root)
	}
	project := filepath.Join(root, "boundary.dtproj")
	writeFile(t, project, `{"version":1,"tableRoots":["tables"],"schemaRoots":["schema"],"export":{"directory":"exports","targets":["csv"]}}`)
	writeFile(t, filepath.Join(root, "cue.mod", "module.cue"), "module: \"example.com/export-boundary\"\nlanguage: {version: \"v0.9.0\"}\n")
	if err := os.MkdirAll(filepath.Join(root, "schema"), 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "tables", "text", "text.cue"), `package text
#PrimaryKey: string
#Row: {
    文本: string
    空值: string | null
    空串: string
    数组: [...string]
    结构: {说明: string, 数量: int, 开启: bool}
    默认值: *"默认中文" | string
    原值: int
    计算值: int
}
rows: {
    "中文行": {
        文本: "第一行,\"引号\"\n第二行\r\n第三行\t尾部"
        空值: null
        空串: ""
        数组: ["中文", "逗号,引号\"", "换\n行", ""]
        结构: {说明: "结构,\"值\"\n后行", 数量: 7, 开启: true}
        原值: 21
    }
    "空集合": {
        文本: "  首尾空白  "
        空值: "非空"
        空串: ""
        数组: []
        结构: {说明: "", 数量: 0, 开启: false}
        原值: 0
    }
}
_tool: {table: {namespace: "boundary", name: "text"}, formulas: {"中文行": {计算值: "原值 * 2"}, "空集合": {计算值: "原值 * 2"}}}
_export: {csv: {enabled: true, columns: ["id","文本","空值","空串","数组","结构","默认值","原值","计算值"], encoding: {数组: "json", 结构: "json"}}}
`)
	writeFile(t, filepath.Join(root, "tables", "numbers", "numbers.cue"), `package numbers
#PrimaryKey: string
#Row: {
    exactInteger: int
    exactDecimal: number
    integerArray: [...int]
    decimalStruct: {value: number}
}
rows: {
    beyond_binary64: {exactInteger: 9007199254740993, exactDecimal: 0.12345678901234567890123456789, integerArray: [9007199254740993, -9007199254740993], decimalStruct: {value: 0.12345678901234567890123456789}}
    beyond_int64: {exactInteger: 18446744073709551615, exactDecimal: 12345678901234567890.123456789, integerArray: [18446744073709551615], decimalStruct: {value: 12345678901234567890.123456789}}
    large_integer: {exactInteger: 123456789012345678901234567890, exactDecimal: 1e-100, integerArray: [123456789012345678901234567890], decimalStruct: {value: 1e-100}}
    tiny_decimal: {exactInteger: 1, exactDecimal: 1e-400, integerArray: [1], decimalStruct: {value: 1e-400}}
    huge_decimal: {exactInteger: 1, exactDecimal: 1e400, integerArray: [1], decimalStruct: {value: 1e400}}
}
_tool: {table: {namespace: "boundary", name: "numbers"}}
_export: {csv: {enabled: true, columns: ["id","exactInteger","exactDecimal","integerArray","decimalStruct"], encoding: {integerArray: "json", decimalStruct: "json"}}}
`)
	settingsRoot := filepath.Join(root, "settings")
	settings := workspace.DefaultEditorSettings()
	settings.Build.Mode = "auto"
	if err := workspace.SaveEditorSettings(settingsRoot, settings); err != nil {
		t.Fatal(err)
	}
	service := workspace.NewService(settingsRoot)
	view, err := service.OpenProject(project)
	if err != nil {
		t.Fatal(err)
	}
	requireCleanCompile(t, view)
	result, err := service.Export()
	if err != nil {
		t.Fatal(err)
	}
	consumer := filepath.Join(repoRoot(t), "test", "functional", "export_boundary_consumer.py")
	command := exec.Command(python, consumer, result.Directory)
	output, err := command.CombinedOutput()
	t.Logf("Independent consumer:\n%s", output)
	if err != nil {
		t.Fatalf("independent CSV consumer: %v", err)
	}
}
