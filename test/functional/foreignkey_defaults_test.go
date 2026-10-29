package functional

import (
	"encoding/csv"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"datapack/internal/workspace"
)

func TestImportedForeignKeyDefaultsExportAndProtectSource(t *testing.T) {
	root, projectPath := emptyProject(t)
	schemaPath := filepath.Join(root, "schema", "common.cue")
	schema := `package common
#EffectID: *"burn" | string | null
#EffectIDs: *["burn","burn"] | [...string] | null
_schema: {definitions:{EffectID:{kind:"type",type:"string"},EffectIDs:{kind:"type",type:"array"}}}
`
	writeFile(t, schemaPath, schema)
	target := filepath.Join(root, "tables", "effect", "effect.cue")
	writeFile(t, target, `package effect
#PrimaryKey:string
#Row:{name:string}
rows:{burn:{name:"Burn"}}
_tool:{table:{namespace:"rpg",name:"effect"}}
`)
	item := filepath.Join(root, "tables", "item", "item.cue")
	writeFile(t, item, `package item
import "example.com/game/schema:common"
#PrimaryKey:string
#Row:{effectID:common.#EffectID,effectIDs:common.#EffectIDs}
rows:{omitted:{},null_input:{effectID:null,effectIDs:null}}
_tool:{table:{namespace:"rpg",name:"item"},foreignKeys:{effectID:{table:"rpg.effect"},effectIDs:{table:"rpg.effect"}}}
_export:{csv:{enabled:true,columns:["id","effectID","effectIDs"]}}
`)
	sources := func() map[string]string {
		m := map[string]string{}
		for _, p := range []string{projectPath, schemaPath, target, item} {
			m[p] = readFile(t, p)
		}
		return m
	}
	before := sources()
	s := openProject(t, projectPath)
	exported, err := s.Export()
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(exported.Directory, "rpg.item.csv")
	prior := readFile(t, output)
	rows, err := csv.NewReader(strings.NewReader(prior)).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{{"id", "effectID", "effectIDs"}, {"omitted", "burn", `["burn","burn"]`}, {"null_input", "burn", `["burn","burn"]`}}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("CSV=%#v want %#v", rows, want)
	}
	if _, err := s.DeleteRow(workspace.RowDelete{Path: target, RowID: "burn"}); err == nil {
		t.Fatal("deleted default-referenced target")
	}
	if _, err := s.RenameRow(workspace.RowRename{Path: target, OldID: "burn", NewID: "flame"}); err == nil {
		t.Fatal("renamed target without shared Schema rewrite")
	}
	if !reflect.DeepEqual(before, sources()) || readFile(t, output) != prior {
		t.Fatal("export/protected edit changed sources or previous CSV")
	}
	writeFile(t, schemaPath, strings.Replace(schema, `*"burn" | string`, `*"missing" | string`, 1))
	s = openProject(t, projectPath)
	before = sources()
	if _, err := s.Export(); err == nil {
		t.Fatal("export accepted invalid imported default")
	}
	if !reflect.DeepEqual(before, sources()) || readFile(t, output) != prior {
		t.Fatal("blocked export changed sources or previous CSV")
	}
}
