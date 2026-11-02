package formula

import (
	"cuelang.org/go/cue/cuecontext"
	"reflect"
	"testing"
)

func TestResearchScalarFastPathMatchesCUE(t *testing.T) {
	expressions := []string{`"中文,\"引号\"\n反斜杠\\"`, `"\u4e2d"`, `""`, `true`, `false`, `null`, `0`, `-12`, `9007199254740993`, `9223372036854775807`, `1.0`, `0.001`, `1e3`, `1.2345678901234567`}
	for _, expression := range expressions {
		t.Run(expression, func(t *testing.T) {
			actual, ok := decodeScalar(expression)
			if !ok {
				t.Fatal("unexpected fallback")
			}
			var expected any
			err := cuecontext.New().CompileString(expression).Decode(&expected)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("got %#v (%T), CUE %#v (%T)", actual, actual, expected, expected)
			}
		})
	}
}
func TestResearchScalarLeavesCUEExpressionsToCUE(t *testing.T) {
	for _, expression := range []string{`1+2`, `0x20`, `1_000`, `1Ki`, `#"raw"#`, `"""` + "\ntext\n" + `"""`, `"\(1+2)"`, `9223372036854775808`, `-9223372036854775808`, `1e999`, `0.0`, `-0.0`, `1e-324`, `4e-324`, `1.7976931348623158e308`, `{a: 1}`, `[1,2]`, `string`, `"\uD800"`, `"\x41"`, `"\a"`} {
		if _, ok := decodeScalar(expression); ok {
			t.Errorf("unexpected fast decode: %s", expression)
		}
	}
}

func TestResearchScalarBoundaryValuesMatchCUEWhenFast(t *testing.T) {
	for _, expression := range []string{`-0`, `1e-300`, `1e-324`, `1e309`, `1.7976931348623157e308`, `2.2250738585072014e-308`, `-9223372036854775807`, `"\\\"\\\\\\t"`} {
		actual, ok := decodeScalar(expression)
		if !ok {
			continue
		}
		var expected any
		err := cuecontext.New().CompileString(expression).Decode(&expected)
		if err != nil || !reflect.DeepEqual(actual, expected) {
			t.Fatalf("literal %s got %#v (%T), CUE %#v (%T), err=%v", expression, actual, actual, expected, expected, err)
		}
	}
}
