package compiler

import (
	"encoding/json"
	"math/big"
	"strconv"

	"cuelang.org/go/cue"
	"datapack/internal/document"
)

// exactValidationValues keeps the formula engine's computed results, but takes
// fixed inputs and defaults from CUE. Routing those values through the formula
// engine's float64 representation would round decimals before validation/export.
func exactValidationValues(row document.Row, evaluated map[string]any, formulaColumns map[string]bool, rowSchema cue.Value) map[string]any {
	values := make(map[string]any, len(evaluated))
	for column, value := range evaluated {
		values[column] = value
		if formulaColumns[column] {
			continue
		}
		cell, fixed := row.Values[column]
		// Inspect the source as well: the evaluator may already have rounded
		// an integral decimal and normalized it to int64.
		needsExact := containsDecimal(value)
		if _, integer := value.(int64); integer {
			needsExact = !fixed || !plainInteger(cell.Expression)
		}
		if !needsExact {
			continue
		}
		var source cue.Value
		if fixed {
			source = rowSchema.Context().CompileString(cell.Expression)
		}
		if !source.Exists() || source.Kind() == cue.NullKind {
			// Defaults have already been selected by the evaluator according
			// to the column's existing nullable/foreign-key policy.
			source, _ = rowSchema.LookupPath(cue.MakePath(cue.Str(column))).Default()
		}
		if !source.Exists() {
			continue
		}
		// Freeze the value selected by the source's defaults, recursively.
		// Unifying an unresolved disjunction with #Row could otherwise select
		// a different branch instead of rejecting the original concrete value.
		encoded, err := source.MarshalJSON()
		if err != nil {
			values[column] = source
			continue
		}
		source = rowSchema.Context().CompileBytes(encoded)
		if _, integer := value.(int64); integer && source.Kind() == cue.FloatKind {
			// Preserve the existing whole-decimal-to-integer policy, but make
			// the integrality check before binary floating-point rounding.
			if exact, ok := new(big.Rat).SetString(string(encoded)); ok && exact.IsInt() {
				source = rowSchema.Context().Encode(exact.Num())
			}
		}
		values[column] = source
	}
	return values
}

// decodeExportValue retains CUE decimal digits in JSON numbers, including in
// arrays and structures. Other scalar types retain CUE Decode's existing Go
// representation (notably int64/big.Int, strings and nullable foreign keys).
func decodeExportValue(value cue.Value) (any, error) {
	// Missing cells can leave field-level defaults in the validated candidate.
	// Select each default before inspecting Kind: an unresolved disjunction
	// would otherwise fall through to Decode and round decimal numbers.
	value, _ = value.Default()
	switch value.Kind() {
	case cue.FloatKind:
		encoded, err := value.MarshalJSON()
		if err != nil {
			return nil, err
		}
		// Keep the existing float64 representation when its JSON decimal
		// round-trip is exact, preserving established CSV number formatting.
		if number, err := strconv.ParseFloat(string(encoded), 64); err == nil {
			original, ok := new(big.Rat).SetString(string(encoded))
			roundTrip, roundOK := new(big.Rat).SetString(strconv.FormatFloat(number, 'g', -1, 64))
			if ok && roundOK && original.Cmp(roundTrip) == 0 {
				return number, nil
			}
		}
		return json.Number(encoded), nil
	case cue.StructKind:
		fields, err := value.Fields()
		if err != nil {
			return nil, err
		}
		decoded := make(map[string]any)
		for fields.Next() {
			field, err := decodeExportValue(fields.Value())
			if err != nil {
				return nil, err
			}
			decoded[fields.Selector().Unquoted()] = field
		}
		return decoded, nil
	case cue.ListKind:
		items, err := value.List()
		if err != nil {
			return nil, err
		}
		decoded := []any{}
		for items.Next() {
			item, err := decodeExportValue(items.Value())
			if err != nil {
				return nil, err
			}
			decoded = append(decoded, item)
		}
		return decoded, nil
	default:
		var decoded any
		if err := value.Decode(&decoded); err != nil {
			return nil, err
		}
		return decoded, nil
	}
}

// Integer and nonnumeric inputs are already lossless. Avoid reparsing those
// common cells merely to recover decimal precision in other columns.
func containsDecimal(value any) bool {
	switch value := value.(type) {
	case float64, *big.Float:
		return true
	case []any:
		for _, item := range value {
			if containsDecimal(item) {
				return true
			}
		}
	case map[string]any:
		for _, field := range value {
			if containsDecimal(field) {
				return true
			}
		}
	}
	return false
}

// Plain integer literals already travel through int64/big.Int without loss.
func plainInteger(expression string) bool {
	if expression == "" {
		return false
	}
	if expression[0] == '-' {
		expression = expression[1:]
	}
	if expression == "" {
		return false
	}
	for _, digit := range expression {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	return true
}
