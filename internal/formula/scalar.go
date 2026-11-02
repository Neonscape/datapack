package formula

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
)

// decodeScalar avoids a fresh CUE runtime only for unambiguous concrete scalar
// literals. All CUE expressions, raw/multiline strings, non-JSON number syntax,
// huge integers and parse/range failures keep the original CUE evaluator.
func decodeScalar(expression string) (any, bool) {
	switch expression {
	case "true":
		return true, true
	case "false":
		return false, true
	case "null":
		return nil, true
	}
	if strings.HasPrefix(expression, `"`) && !strings.HasPrefix(expression, `"""`) && json.Valid([]byte(expression)) {
		v, err := strconv.Unquote(expression)
		if err == nil {
			return v, true
		}
		return nil, false
	}
	if expression == "" || !((expression[0] >= '0' && expression[0] <= '9') || expression[0] == '-') || !json.Valid([]byte(expression)) {
		return nil, false
	}
	if !strings.ContainsAny(expression, ".eE") {
		v, err := strconv.ParseInt(expression, 10, 64)
		if err == nil && v != math.MinInt64 {
			return v, true
		}
		return nil, false
	}
	v, err := strconv.ParseFloat(expression, 64)
	// CUE returns big.Float outside its float64 interval, including values
	// that strconv may round onto the smallest/largest float64 boundary.
	if err == nil && math.Abs(v) >= 0x1p-1022 && math.Abs(v) < math.MaxFloat64 {
		return v, true
	}
	return nil, false
}
