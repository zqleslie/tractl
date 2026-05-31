package overlay

import (
	"fmt"
)

// appendAccumulatingArrays is the set of array field names whose default merge action
// is append (semantically accumulating). Reference: overlay spec §10.3
var appendAccumulatingArrays = map[string]bool{
	"assertions": true,
	"extracts":   true,
	"tags":       true,
}

// replaceSetArrays is the set of array field names whose default merge action is replace
// (set-identity membership). Reference: overlay spec §10.3
var replaceSetArrays = map[string]bool{
	"dependsOn":    true,
	"capabilities": true,
	"invariants":   true,
}

// sliceIdentityFields maps identity-contract array names to their identity field name.
// Reference: overlay spec §10.5
var sliceIdentityFields = map[string]string{
	"assertions": "id",
	"extracts":   "id",
}

// applyMerge applies a merge action to an existing value, producing a new value.
// fieldName is the canonical JSON field name of the target, used to determine
// schema-aware defaults for arrays.
//
// Returns the merged value or an error.
// Reference: overlay spec §10, §11
func applyMerge(existing, patch any, action MergeAction, fieldName string, overlayName string, patchIdx int, targetDesc string) (any, error) {
	// Determine the effective action.
	effective := action
	if effective == "" {
		var err error
		effective, err = defaultAction(existing, fieldName, overlayName, patchIdx, targetDesc)
		if err != nil {
			return nil, err
		}
	}

	switch effective {
	case ActionReplace:
		return patch, nil

	case ActionDeepMerge:
		return deepMerge(existing, patch), nil

	case ActionAppend:
		return appendSlice(existing, patch), nil

	case ActionAppendUnique:
		idField, ok := sliceIdentityFields[fieldName]
		if !ok {
			return nil, engineErr(ErrNoIdentityContract, overlayName, patchIdx, targetDesc, string(effective),
				fmt.Sprintf("appendUnique is not permitted on %q: no identity contract declared", fieldName))
		}
		return appendUniqueSlice(existing, patch, idField), nil

	case ActionRemove:
		// Remove is handled at the call site (engine), not here.
		// If we reach here it means an internal routing error.
		return nil, engineErr(ErrInvalidAction, overlayName, patchIdx, targetDesc, string(effective),
			"remove action must not reach applyMerge; handled by engine")

	default:
		return nil, engineErr(ErrInvalidAction, overlayName, patchIdx, targetDesc, string(effective),
			fmt.Sprintf("unrecognised action %q", effective))
	}
}

// defaultAction returns the schema-aware default merge action for a field.
// Reference: overlay spec §10.1–§10.3
func defaultAction(existing any, fieldName, overlayName string, patchIdx int, targetDesc string) (MergeAction, error) {
	switch {
	case appendAccumulatingArrays[fieldName]:
		return ActionAppend, nil
	case replaceSetArrays[fieldName]:
		return ActionReplace, nil
	}

	// Determine by value type.
	switch existing.(type) {
	case map[string]any:
		return ActionDeepMerge, nil
	case []any:
		// Array with no schema default and no explicit action — error.
		return "", engineErr(ErrInvalidAction, overlayName, patchIdx, targetDesc, "",
			fmt.Sprintf("no canonical merge semantic for array field %q; declare an explicit action", fieldName))
	default:
		// Scalar.
		return ActionReplace, nil
	}
}

// deepMerge merges patch into existing recursively.
// For objects: keys in patch are applied over existing; keys only in existing are preserved.
// For non-objects: patch replaces existing.
// Reference: overlay spec §10.2
func deepMerge(existing, patch any) any {
	existingMap, existingIsMap := existing.(map[string]any)
	patchMap, patchIsMap := patch.(map[string]any)

	if existingIsMap && patchIsMap {
		result := make(map[string]any, len(existingMap))
		for k, v := range existingMap {
			result[k] = v
		}
		for k, v := range patchMap {
			if existingVal, ok := result[k]; ok {
				result[k] = deepMerge(existingVal, v)
			} else {
				result[k] = v
			}
		}
		return result
	}

	// Non-map existing or patch: patch replaces.
	return patch
}

// appendSlice appends patch entries to existing slice.
// Reference: overlay spec §10.3 (append)
func appendSlice(existing, patch any) any {
	existingSlice, _ := toSlice(existing)
	patchSlice, _ := toSlice(patch)
	result := make([]any, len(existingSlice), len(existingSlice)+len(patchSlice))
	copy(result, existingSlice)
	return append(result, patchSlice...)
}

// appendUniqueSlice appends patch entries to existing slice using identity-aware deduplication.
// If an incoming entry's identity field matches an existing entry, the existing entry is
// replaced by the incoming one (last-wins). If no collision, the entry is appended.
// Reference: overlay spec §10.3, §10.5
func appendUniqueSlice(existing, patch any, idField string) any {
	existingSlice, _ := toSlice(existing)
	patchSlice, _ := toSlice(patch)

	// Build index: idValue -> position in result.
	result := make([]any, len(existingSlice))
	copy(result, existingSlice)

	idIndex := make(map[any]int, len(result))
	for i, entry := range result {
		if m, ok := entry.(map[string]any); ok {
			if id, ok := m[idField]; ok {
				idIndex[id] = i
			}
		}
	}

	for _, incoming := range patchSlice {
		m, isMap := incoming.(map[string]any)
		if !isMap {
			result = append(result, incoming)
			continue
		}
		id, hasID := m[idField]
		if !hasID {
			result = append(result, incoming)
			continue
		}
		if pos, exists := idIndex[id]; exists {
			result[pos] = incoming
		} else {
			idIndex[id] = len(result)
			result = append(result, incoming)
		}
	}
	return result
}

// toSlice coerces any to []any. Returns nil, false if not a slice.
func toSlice(v any) ([]any, bool) {
	s, ok := v.([]any)
	return s, ok
}
