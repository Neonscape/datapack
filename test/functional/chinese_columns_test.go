package functional

import (
	"encoding/csv"
	"path/filepath"
	"strings"
	"testing"

	"datapack/internal/workspace"
)

// Standalone Unicode-ID compatibility, separate from Yunlan's English IDs and
// Chinese display names. Backend edits/diagnostics/exports use a private TempDir.
func TestChineseColumnsCompileEditRenameDiagnoseAndExport(t *testing.T) {
	root, project := emptyProject(t)
	writeFile(t, filepath.Join(root, "cue.mod", "module.cue"), `module: "example.com/chinese-probe"
language: {version: "v0.9.0"}`)
	writeFile(t, filepath.Join(root, "schema", "common.cue"), `package common
#名称: string & != ""
#非负: >=0
_schema: {version:1,definitions:{名称:{kind:"type",type:"string"},非负:{kind:"constraint",appliesTo:["integer","number"]}}}`)
	effect := filepath.Join(root, "tables", "effect", "effect.cue")
	writeFile(t, effect, `package effect
import "example.com/chinese-probe/schema:common"
#PrimaryKey:string
#Row:{名称:common.#名称,持续回合:int&common.#非负}
rows:{burn:{名称:"灼烧",持续回合:3}}
_tool:{table:{namespace:"probe",name:"effect"}}
_export:{csv:{enabled:true,columns:["id","名称","持续回合"]}}`)
	item := filepath.Join(root, "tables", "item", "item.cue")
	writeFile(t, item, `package item
import "example.com/chinese-probe/schema:common"
#PrimaryKey:string
_local:{#坐标:{横:number,纵:number}}
_schema:{localDefinitions:{坐标:{kind:"type",type:"struct"}}}
#Row:{名称:common.#名称,售价:int&common.#非负,折扣:*1|number&>=0&<=1,实售价:int&common.#非负,效果ID:string|null,效果列表:[...string],坐标:_local.#坐标,策划说明:string|null}
rows:{one:{名称:"赤砂短剑",售价:100,折扣:0.8,效果ID:"burn",效果列表:["burn"],坐标:{横:1,纵:2},策划说明:null},two:{名称:"试作短剑",售价:20,效果ID:null,效果列表:[],坐标:{横:0,纵:0},策划说明:"跨行中文字段探针"}}
_tool:{table:{namespace:"probe",name:"item"},foreignKeys:{效果ID:{table:"probe.effect"},效果列表:{table:"probe.effect"}},formulas:{one:{实售价:"ceil(售价 * 折扣)"},two:{实售价:"probe.item.one.售价 + 2"}}}
_editor:{table:{title:"中文商品探针"},cellComments:{one:{售价:"标价乘折扣后向上取整，避免少收一枚金币。"}}}
_export:{csv:{enabled:true,columns:["id","名称","售价","折扣","实售价","效果ID","效果列表","坐标"],encoding:{效果列表:"json",坐标:"json"}}}`)
	s := openProject(t, project)
	view, _ := s.Current()
	requireCleanCompile(t, view)
	check := func(column string, want string) {
		t.Helper()
		table, err := s.GetTable(item)
		if err != nil {
			t.Fatal(err)
		}
		for _, row := range table.Rows {
			if row.ID == "one" && row.Values[column].Display == want {
				return
			}
		}
		t.Fatalf("Chinese column %s did not show %s", column, want)
	}
	check("实售价", "80")
	if _, err := s.SetCellText(workspace.CellUpdate{Path: item, RowID: "one", Column: "名称", Expression: "霜月长剑"}); err != nil {
		t.Fatal(err)
	}
	check("名称", "霜月长剑")
	if _, err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	check("名称", "赤砂短剑")
	source := readFile(t, item)
	result, err := s.Export()
	if err != nil {
		t.Fatal(err)
	}
	if readFile(t, item) != source {
		t.Fatal("Export rewrote Chinese source")
	}
	rows, err := csv.NewReader(strings.NewReader(readFile(t, filepath.Join(result.Directory, "probe.item.csv")))).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(rows[0], ",") != "id,名称,售价,折扣,实售价,效果ID,效果列表,坐标" || rows[1][4] != "80" || !strings.Contains(rows[1][7], "横") {
		t.Fatalf("Chinese CSV lost field names/JSON/value: %#v", rows)
	}
	if _, err := s.RenameColumn(workspace.ColumnRename{Path: item, OldName: "售价", NewName: "标价"}); err != nil {
		t.Fatal(err)
	}
	check("实售价", "80")
	if _, err := s.UpdateCell(workspace.CellUpdate{Path: item, RowID: "one", Column: "标价", Expression: "200"}); err != nil {
		t.Fatal(err)
	}
	check("实售价", "160")
	if _, err := s.RenameRow(workspace.RowRename{Path: effect, OldID: "burn", NewID: "ignite"}); err != nil {
		t.Fatal(err)
	}
	check("效果ID", "ignite")
	after := readFile(t, item)
	if _, err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	check("效果ID", "burn")
	if _, err := s.Redo(); err != nil {
		t.Fatal(err)
	}
	if readFile(t, item) != after {
		t.Fatal("Chinese edits not restored exactly")
	}
	if _, err := s.Export(); err != nil {
		t.Fatal(err)
	}
	csvBefore := readFile(t, filepath.Join(result.Directory, "probe.item.csv"))
	if _, err := s.UpdateCell(workspace.CellUpdate{Path: item, RowID: "one", Column: "标价", Expression: "-1"}); err != nil {
		t.Fatal(err)
	}
	view, _ = s.Current()
	if len(view.Compile.Diagnostics) == 0 {
		t.Fatal("invalid Chinese column did not produce diagnostic")
	}
	found := false
	for _, d := range view.Compile.Diagnostics {
		if d.Column == "标价" || strings.Contains(d.Message, "标价") {
			found = true
		}
	}
	if !found {
		t.Fatal("diagnostic did not identify Chinese field")
	}
	if _, err := s.Export(); err == nil {
		t.Fatal("invalid Chinese value exported")
	}
	if readFile(t, filepath.Join(result.Directory, "probe.item.csv")) != csvBefore {
		t.Fatal("failed export replaced valid output")
	}
}
