package cuerows

import (
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"reflect"
	"testing"
)

func TestValidateKeepsInputAndReturnsDefaultedCandidate(t *testing.T) {
	context := cuecontext.New()
	value := context.CompileString(`#ID:string
#Row:{amount:*7|int,optional:null|int}`)
	row := value.LookupPath(cue.MakePath(cue.Def("Row")))
	id := value.LookupPath(cue.MakePath(cue.Def("ID")))
	input := map[string]any{"optional": nil}
	candidate, issues := Validate(row, id, "one", input)
	if len(issues) != 0 {
		t.Fatal(issues)
	}
	var decoded map[string]any
	if err := candidate.Decode(&decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded, map[string]any{"amount": int64(7), "optional": nil}) || !reflect.DeepEqual(input, map[string]any{"optional": nil}) {
		t.Fatalf("decoded/input: %#v / %#v", decoded, input)
	}
}
func TestValidateReportsIdentityFieldAndRowWithoutDecodingPolicy(t *testing.T) {
	context := cuecontext.New()
	value := context.CompileString(`#ID:string & =~"^[a-z]+$"
#Row:{amount:int & >=10}`)
	_, issues := Validate(value.LookupPath(cue.MakePath(cue.Def("Row"))), value.LookupPath(cue.MakePath(cue.Def("ID"))), "UPPER", map[string]any{"amount": 1})
	if len(issues) != 3 || issues[0].Kind != Identity || issues[1].Kind != Field || issues[1].Column != "amount" || issues[2].Kind != Row {
		t.Fatalf("issues: %#v", issues)
	}
}
