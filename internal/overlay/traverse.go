// traverse.go defines the semantic and source-native tree traversal functions
// used by the overlay engine to resolve match-based and source-based patch
// targets: semanticMatch (walks map[string]any depth-first) and sourceNativeMatch.
package overlay

import (
	"fmt"
	"sort"
	"strings"
)

// matchNode is a resolved node from a semantic or source-native traversal.
type matchNode struct {
	// path is the canonical dot-separated path to this node in the document.
	// May include array index notation (e.g. "workflows[0].steps[1]") when
	// the traversal passed through a slice.
	path string
	// value is the resolved node.
	value any
	// set replaces the node value in the parent container. It is set by the
	// traversal and avoids re-resolution from path when array indices are involved.
	set func(newVal any)
}

// semanticMatch traverses the JSON document depth-first in canonical normalized order
// (sorted map keys at every level) and returns all nodes whose fields satisfy all
// declared selector entries.
//
// mode controls cardinality: TargetModeOne requires exactly one match.
// TargetModeAll accepts zero or more.
//
// Reference: overlay spec §9.2, §12, §20
func semanticMatch(doc map[string]any, sel MatchSelector, mode TargetMode, overlayName string, patchIdx int) ([]matchNode, error) {
	var results []matchNode
	collectMatchingNodes(doc, "", sel, &results)

	if mode == "" || mode == TargetModeOne {
		switch len(results) {
		case 0:
			return nil, engineErr(ErrModeOneNoMatch, overlayName, patchIdx,
				fmt.Sprintf("match:%v", sel), "", "mode:one selector matched zero nodes")
		case 1:
			return results, nil
		default:
			return nil, engineErr(ErrModeOneMultipleMatches, overlayName, patchIdx,
				fmt.Sprintf("match:%v", sel), "", fmt.Sprintf("mode:one selector matched %d nodes", len(results)))
		}
	}

	// mode:all — zero or more matches are accepted.
	return results, nil
}

// collectMatchingNodes performs the depth-first traversal. It recurses into maps and
// slices. Traversal order is deterministic: map keys are sorted lexicographically before
// iteration at every level; slice elements are visited in index order.
// Reference: overlay spec §12 (canonical normalized traversal order)
func collectMatchingNodes(v any, currentPath string, sel MatchSelector, results *[]matchNode) {
	collectMatchingNodesWithSetter(v, currentPath, sel, results, nil)
}

func collectMatchingNodesWithSetter(v any, currentPath string, sel MatchSelector, results *[]matchNode, setter func(any)) {
	switch node := v.(type) {
	case map[string]any:
		// Check whether this map node satisfies all selector fields.
		if nodeMatchesSelector(node, sel) {
			*results = append(*results, matchNode{path: currentPath, value: v, set: setter})
			// Do not recurse into a matched node — its children are not independent targets.
			return
		}
		// Recurse into children in sorted key order.
		keys := sortedKeys(node)
		for _, k := range keys {
			childPath := k
			if currentPath != "" {
				childPath = currentPath + "." + k
			}
			kCopy := k
			childSetter := func(newVal any) { node[kCopy] = newVal }
			collectMatchingNodesWithSetter(node[k], childPath, sel, results, childSetter)
		}

	case []any:
		// Recurse into each slice element in index order.
		for i, elem := range node {
			childPath := fmt.Sprintf("%s[%d]", currentPath, i)
			if currentPath == "" {
				childPath = fmt.Sprintf("[%d]", i)
			}
			iCopy := i
			childSetter := func(newVal any) { node[iCopy] = newVal }
			collectMatchingNodesWithSetter(elem, childPath, sel, results, childSetter)
		}
	}
}

// nodeMatchesSelector returns true if m contains every key-value pair in sel.
func nodeMatchesSelector(m map[string]any, sel MatchSelector) bool {
	for k, v := range sel {
		actual, ok := m[k]
		if !ok {
			return false
		}
		// Compare as strings for simplicity; selector values from YAML/JSON are strings.
		if fmt.Sprintf("%v", actual) != fmt.Sprintf("%v", v) {
			return false
		}
	}
	return true
}

// sourceNativeMatch resolves nodes via provenance metadata on the spec's Metadata field.
//
// The document's "metadata.provenance" field maps canonical paths to provenance entries.
// A source selector matches entries whose "source" field's "type" equals the selector type,
// and whose other selector fields match the corresponding provenance entry fields.
//
// Reference: overlay spec §9.3, §18.2
func sourceNativeMatch(doc map[string]any, sel *SourceSelector, mode TargetMode, overlayName string, patchIdx int) ([]matchNode, error) {
	knownSourceTypes := map[string]bool{
		"openapi": true,
		"wsdl":    true,
		"postman": true,
		"har":     true,
		"curl":    true,
		"bruno":   true,
		"toon":    true,
		"yaml":    true,
		"json":    true,
	}

	if !knownSourceTypes[strings.ToLower(sel.Type)] {
		return nil, engineErr(ErrSourceSelectorUnknown, overlayName, patchIdx,
			fmt.Sprintf("source.type=%q", sel.Type), "",
			fmt.Sprintf("source selector type %q is not recognised", sel.Type))
	}

	// Walk metadata.provenance to find matching paths.
	meta, _ := doc["metadata"].(map[string]any)
	if meta == nil {
		if mode == TargetModeAll {
			return nil, nil
		}
		return nil, engineErr(ErrModeOneNoMatch, overlayName, patchIdx,
			fmt.Sprintf("source.type=%q", sel.Type), "", "no metadata found; zero nodes matched")
	}

	provenance, _ := meta["provenance"].(map[string]any)
	if provenance == nil {
		if mode == TargetModeAll {
			return nil, nil
		}
		return nil, engineErr(ErrModeOneNoMatch, overlayName, patchIdx,
			fmt.Sprintf("source.type=%q", sel.Type), "", "no provenance found; zero nodes matched")
	}

	var results []matchNode
	provenancePaths := sortedKeys(provenance)
	for _, canonicalPath := range provenancePaths {
		entry, ok := provenance[canonicalPath].(map[string]any)
		if !ok {
			continue
		}
		if !provenanceMatchesSelector(entry, sel) {
			continue
		}
		node, err := resolvePath(doc, canonicalPath)
		if err != nil {
			continue
		}
		results = append(results, matchNode{path: canonicalPath, value: node.value})
	}

	if mode == "" || mode == TargetModeOne {
		switch len(results) {
		case 0:
			return nil, engineErr(ErrModeOneNoMatch, overlayName, patchIdx,
				fmt.Sprintf("source.type=%q", sel.Type), "", "mode:one source selector matched zero nodes")
		case 1:
			return results, nil
		default:
			return nil, engineErr(ErrModeOneMultipleMatches, overlayName, patchIdx,
				fmt.Sprintf("source.type=%q", sel.Type), "",
				fmt.Sprintf("mode:one source selector matched %d nodes", len(results)))
		}
	}

	return results, nil
}

// provenanceMatchesSelector checks whether a provenance entry satisfies the source selector.
func provenanceMatchesSelector(entry map[string]any, sel *SourceSelector) bool {
	entryType, _ := entry["sourceRef"].(string)
	_ = entryType

	// Match the type field against a "sourceType" field in the provenance entry.
	if st, ok := entry["sourceType"].(string); ok {
		if !strings.EqualFold(st, sel.Type) {
			return false
		}
	}

	// Match additional selector fields against provenance entry fields.
	for k, v := range sel.Fields {
		actual, ok := entry[k]
		if !ok {
			return false
		}
		if fmt.Sprintf("%v", actual) != fmt.Sprintf("%v", v) {
			return false
		}
	}
	return true
}

// sortedKeys returns the keys of a map in lexicographic order.
// This enforces deterministic traversal per overlay spec §12 and §20.
func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
