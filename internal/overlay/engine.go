// Package overlay implements the overlay engine per overlay spec §3.
//
// Pipeline position:
//
//	Parser output → *spec.TraCtlSpec → [ Overlay Engine ] → *spec.TraCtlSpec candidate → SpecValidator
//
// The engine operates on the spec struct through a JSON round-trip clone and map[string]any
// traversal. It does not import internal/spec, maintaining the import constraint established
// in Phase 1 milestone 1.2.
package overlay

import (
	"encoding/json"
	"fmt"
)

// Engine applies a stack of validated overlay documents to a parsed spec.
// The engine receives any value that can be marshaled to/from JSON and treats it
// as the canonical spec document model.
//
// Callers must pass a *spec.TraCtlSpec (or any JSON-serializable type that represents it).
// The engine does not import internal/spec; the caller is responsible for type assertion
// after Apply returns.
//
// Reference: overlay spec §3, §12, ADR-003 §6
type Engine struct{}

// NewEngine returns a ready-to-use Engine.
func NewEngine() *Engine { return &Engine{} }

// Apply applies overlays in stack order to src and returns a post-overlay candidate.
//
// Contract:
//  1. src is deep-cloned via JSON round-trip; the original is never mutated.
//  2. Overlays are applied in stack order (first overlay first — overlay spec §12).
//  3. Within each overlay, patches are applied in authored order.
//  4. For each patch, the target is resolved and the merge action is applied.
//  5. Provenance is recorded after each successful patch.
//  6. The candidate is returned; the caller must run SpecValidator.Validate on it.
//
// Apply does NOT call SpecValidator.Validate or OverlayValidator.Validate.
// Overlays passed to Apply are assumed to have already passed OverlayValidator.Validate.
//
// Reference: overlay spec §12, §20; ADR-003 §2
func (e *Engine) Apply(src any, overlays []*OverlayDocument) (any, error) {
	// Step 1: deep-clone src so we never mutate the input.
	var cloned map[string]any
	if err := cloneSpec(src, &cloned); err != nil {
		return nil, err
	}

	// Step 2: apply each overlay in stack order.
	for overlayIdx, doc := range overlays {
		overlayName := ""
		if doc.Metadata != nil {
			overlayName = doc.Metadata.Name
		}

		for patchIdx, patch := range doc.Patches {
			if err := applyPatch(cloned, patch, overlayName, overlayIdx, patchIdx); err != nil {
				return nil, err
			}
		}
	}

	return cloned, nil
}

// ApplyToJSON is a convenience wrapper that accepts and returns JSON bytes.
// It is useful for callers that work with raw JSON rather than typed structs.
func (e *Engine) ApplyToJSON(srcJSON []byte, overlays []*OverlayDocument) ([]byte, error) {
	var src map[string]any
	if err := json.Unmarshal(srcJSON, &src); err != nil {
		return nil, fmt.Errorf("engine: failed to unmarshal source JSON: %w", err)
	}

	result, err := e.Apply(src, overlays)
	if err != nil {
		return nil, err
	}

	out, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("engine: failed to marshal result: %w", err)
	}
	return out, nil
}

// applyPatch resolves the target and applies the merge action for a single patch.
func applyPatch(doc map[string]any, patch Patch, overlayName string, overlayIdx, patchIdx int) error {
	target := patch.Target
	action := patch.Action

	// Guard: remove with patch body is forbidden per overlay spec §11.
	if action == ActionRemove && len(patch.Data) > 0 {
		return engineErr(ErrRemoveWithPatch, overlayName, patchIdx,
			describeTarget(target), string(action),
			"action=remove must not include a data field")
	}

	// Guard: unrecognised explicit action.
	if action != "" && !validActions[action] {
		return engineErr(ErrInvalidAction, overlayName, patchIdx,
			describeTarget(target), string(action),
			fmt.Sprintf("unrecognised action %q", action))
	}

	switch {
	case target.Path != "":
		return applyPathPatch(doc, patch, overlayName, overlayIdx, patchIdx)
	case len(target.Match) > 0:
		return applyMatchPatch(doc, patch, overlayName, overlayIdx, patchIdx)
	case target.Source != nil:
		return applySourcePatch(doc, patch, overlayName, overlayIdx, patchIdx)
	default:
		return engineErr(ErrPathNotFound, overlayName, patchIdx, "", string(action),
			"target has no selector (path, match, or source)")
	}
}

// applyPathPatch applies a patch using canonical path targeting (overlay spec §9.1).
func applyPathPatch(doc map[string]any, patch Patch, overlayName string, overlayIdx, patchIdx int) error {
	path := patch.Target.Path
	targetDesc := "path:" + path

	if patch.Action == ActionRemove {
		if err := deleteAtPath(doc, path); err != nil {
			return wrapEngineErr(err, overlayName, patchIdx, targetDesc, string(patch.Action))
		}
		recordProvenance(doc, overlayName, overlayIdx, patchIdx, path, string(patch.Action))
		return nil
	}

	resolved, err := resolvePath(doc, path)
	if err != nil {
		return wrapEngineErr(err, overlayName, patchIdx, targetDesc, string(patch.Action))
	}

	fieldName := resolved.key
	merged, err := applyMerge(resolved.value, any(patch.Data), patch.Action, fieldName, overlayName, patchIdx, targetDesc)
	if err != nil {
		return err
	}

	if err := setAtPath(doc, path, merged); err != nil {
		return wrapEngineErr(err, overlayName, patchIdx, targetDesc, string(patch.Action))
	}

	recordProvenance(doc, overlayName, overlayIdx, patchIdx, path, string(patch.Action))
	return nil
}

// applyMatchPatch applies a patch using semantic match targeting (overlay spec §9.2).
func applyMatchPatch(doc map[string]any, patch Patch, overlayName string, overlayIdx, patchIdx int) error {
	mode := patch.Target.Mode
	targetDesc := fmt.Sprintf("match:%v", patch.Target.Match)

	nodes, err := semanticMatch(doc, patch.Target.Match, mode, overlayName, patchIdx)
	if err != nil {
		return err
	}

	// mode:all with zero matches: no-op, no error (overlay spec §9.2).
	for _, node := range nodes {
		if err := applyToNode(doc, node, patch, overlayName, overlayIdx, patchIdx, targetDesc); err != nil {
			return err
		}
	}
	return nil
}

// applySourcePatch applies a patch using source-native targeting (overlay spec §9.3).
func applySourcePatch(doc map[string]any, patch Patch, overlayName string, overlayIdx, patchIdx int) error {
	mode := patch.Target.Mode
	targetDesc := fmt.Sprintf("source.type:%s", patch.Target.Source.Type)

	nodes, err := sourceNativeMatch(doc, patch.Target.Source, mode, overlayName, patchIdx)
	if err != nil {
		return err
	}

	for _, node := range nodes {
		if err := applyToNode(doc, node, patch, overlayName, overlayIdx, patchIdx, targetDesc); err != nil {
			return err
		}
	}
	return nil
}

// applyToNode applies a patch to a single resolved matchNode.
// If node.set is non-nil (from semantic traversal through slices), it is used to
// update the value in place without re-resolving from the path string.
func applyToNode(doc map[string]any, node matchNode, patch Patch, overlayName string, overlayIdx, patchIdx int, targetDesc string) error {
	path := node.path

	if patch.Action == ActionRemove {
		if len(patch.Data) > 0 {
			return engineErr(ErrRemoveWithPatch, overlayName, patchIdx, targetDesc, string(patch.Action),
				"action=remove must not include a data field")
		}
		if node.set != nil {
			// For nodes reached through slices we cannot delete from a slice without
			// rewriting the parent; zero the value instead, consistent with removal semantics.
			node.set(nil)
		} else {
			if err := deleteAtPath(doc, path); err != nil {
				return wrapEngineErr(err, overlayName, patchIdx, targetDesc, string(patch.Action))
			}
		}
		recordProvenance(doc, overlayName, overlayIdx, patchIdx, path, string(patch.Action))
		return nil
	}

	// Determine field name from the last segment of the path (without array suffix).
	fieldName := lastSegment(path)

	merged, err := applyMerge(node.value, any(patch.Data), patch.Action, fieldName, overlayName, patchIdx, targetDesc)
	if err != nil {
		return err
	}

	if node.set != nil {
		node.set(merged)
	} else {
		if err := setAtPath(doc, path, merged); err != nil {
			return wrapEngineErr(err, overlayName, patchIdx, targetDesc, string(patch.Action))
		}
	}

	recordProvenance(doc, overlayName, overlayIdx, patchIdx, path, string(patch.Action))
	return nil
}

// validActions is the complete set of recognised action values.
var validActions = map[MergeAction]bool{
	ActionReplace:      true,
	ActionDeepMerge:    true,
	ActionAppend:       true,
	ActionAppendUnique: true,
	ActionRemove:       true,
}

// describeTarget builds a short human-readable target description.
func describeTarget(t Target) string {
	if t.Path != "" {
		return "path:" + t.Path
	}
	if len(t.Match) > 0 {
		return fmt.Sprintf("match:%v", t.Match)
	}
	if t.Source != nil {
		return fmt.Sprintf("source.type:%s", t.Source.Type)
	}
	return "<empty>"
}

// wrapEngineErr enriches a path error (which may already be an *EngineError) with
// overlay and patch context.
func wrapEngineErr(err error, overlayName string, patchIdx int, target, action string) error {
	if ee, ok := err.(*EngineError); ok {
		ee.OverlayName = overlayName
		ee.PatchIndex = patchIdx
		if ee.Target == "" {
			ee.Target = target
		}
		if ee.Action == "" {
			ee.Action = action
		}
		return ee
	}
	return engineErr(ErrPathNotFound, overlayName, patchIdx, target, action, err.Error())
}

// lastSegment returns the last dot-separated segment of a path.
func lastSegment(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			return path[i+1:]
		}
	}
	return path
}
