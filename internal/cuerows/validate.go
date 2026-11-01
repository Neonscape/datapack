// Package cuerows contains pure CUE row-validation mechanics shared by preview
// and compilation. Loading, diagnostic policy and final decoding stay callers'.
package cuerows

import "cuelang.org/go/cue"

type IssueKind uint8

const (
	Identity IssueKind = iota
	Field
	Row
)

type Issue struct {
	Kind   IssueKind
	Column string
	Err    error
}

// Validate observes identity, supplied fields and the whole row in the existing
// order. The schemas must belong to the caller's loaded CUE context. No source,
// input map, compiler state or package-selection policy is changed.
func Validate(rowSchema, primaryKey cue.Value, rowID string, values map[string]any) (cue.Value, []Issue) {
	context := rowSchema.Context()
	var issues []Issue
	if err := primaryKey.Unify(context.Encode(rowID)).Validate(cue.Concrete(true)); err != nil {
		issues = append(issues, Issue{Kind: Identity, Err: err})
	}
	for column, value := range values {
		schema := rowSchema.LookupPath(cue.MakePath(cue.Str(column)))
		if err := schema.Unify(context.Encode(value)).Validate(cue.Concrete(true)); err != nil {
			issues = append(issues, Issue{Kind: Field, Column: column, Err: err})
		}
	}
	candidate := rowSchema.Unify(context.Encode(values))
	if err := candidate.Validate(cue.Concrete(true)); err != nil {
		issues = append(issues, Issue{Kind: Row, Err: err})
	}
	return candidate, issues
}
