// scan.go defines the step-reference scanner used by Normalize: the stepRefRE
// regular expression, scanStepRefs (string scanning), and scanRefsInValue
// (recursive value scanning for body content maps and slices).
package normalize

import "regexp"

// stepRefRE matches the step identifier inside a ${steps.<id>...} expression.
//
// The pattern intentionally only matches `${steps.<id>` (with the dot after
// `steps` and an identifier following) so that `${vars.x}`, `${env.x}`, and
// `${runtime.<...>}` are NOT picked up.
//
// The captured identifier is the step ID. \b ensures the identifier ends
// cleanly at a word boundary (e.g. before `.extracts`, `.response`, `}`, etc.).
var stepRefRE = regexp.MustCompile(`\$\{steps\.([a-zA-Z][a-zA-Z0-9_-]*)\b`)

// scanStepRefs returns the set of step IDs referenced by any of the supplied strings.
// Empty strings are skipped. The returned set is deterministic only via the caller's
// sort; this helper itself returns a map for fast dedup.
func scanStepRefs(strs ...string) map[string]struct{} {
	out := make(map[string]struct{})
	for _, s := range strs {
		if s == "" {
			continue
		}
		matches := stepRefRE.FindAllStringSubmatch(s, -1)
		for _, m := range matches {
			if len(m) >= 2 && m[1] != "" {
				out[m[1]] = struct{}{}
			}
		}
	}
	return out
}

// scanRefsInValue extracts step refs from a value of arbitrary shape.
// Supports string, []any, map[string]any, and arbitrary nested structures
// produced by yaml.v3 / encoding/json unmarshal. Non-string scalars are ignored.
func scanRefsInValue(v any, into map[string]struct{}) {
	switch t := v.(type) {
	case string:
		for k := range scanStepRefs(t) {
			into[k] = struct{}{}
		}
	case []any:
		for _, el := range t {
			scanRefsInValue(el, into)
		}
	case map[string]any:
		for _, el := range t {
			scanRefsInValue(el, into)
		}
	case map[any]any:
		for _, el := range t {
			scanRefsInValue(el, into)
		}
	}
}
