// provenance.go defines the provenance tracking functions responsible for
// recording which overlay and patch index last modified each field, writing
// overlay refs and patch records into the spec candidate's metadata section.
package overlay

import (
	"fmt"
	"time"
)

// provenanceKey builds the key used in metadata.provenance for a given resolved path.
func provenanceKey(resolvedPath string) string {
	return resolvedPath
}

// recordProvenance writes a provenance entry and an overlay ref into the candidate's
// metadata map (the map[string]any representation of *spec.Metadata).
//
// Provenance is informational only. No merge decision in the engine may depend on it.
// This function is called after a patch is successfully applied.
//
// Reference: overlay spec §13, ADR-003 §8
func recordProvenance(doc map[string]any, overlayName string, overlayIdx, patchIdx int, targetPath, action string) {
	meta := ensureMetadataMap(doc)

	// Update overlayRefs.
	ref := overlayRefEntry(overlayName, overlayIdx)
	overlayRefs, _ := meta["overlayRefs"].([]any)
	if !containsRef(overlayRefs, ref) {
		meta["overlayRefs"] = append(overlayRefs, ref)
	}

	// Update provenance map.
	provenance, _ := meta["provenance"].(map[string]any)
	if provenance == nil {
		provenance = make(map[string]any)
		meta["provenance"] = provenance
	}

	key := provenanceKey(targetPath)
	provenance[key] = map[string]any{
		"source":     ref,
		"patchIndex": patchIdx,
		"target":     targetPath,
		"action":     action,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}
}

// ensureMetadataMap returns the metadata sub-map, creating it if absent.
func ensureMetadataMap(doc map[string]any) map[string]any {
	if m, ok := doc["metadata"].(map[string]any); ok {
		return m
	}
	m := make(map[string]any)
	doc["metadata"] = m
	return m
}

// overlayRefEntry returns a stable string identifier for an overlay document.
func overlayRefEntry(overlayName string, overlayIdx int) string {
	if overlayName != "" {
		return overlayName
	}
	return fmt.Sprintf("overlay[%d]", overlayIdx)
}

// containsRef returns true if ref is already present in overlayRefs.
func containsRef(refs []any, ref string) bool {
	for _, r := range refs {
		if s, ok := r.(string); ok && s == ref {
			return true
		}
	}
	return false
}
