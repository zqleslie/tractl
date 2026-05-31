package yaml

import (
	goyaml "gopkg.in/yaml.v3"

	"github.com/tractl/tractl/internal/validation/common"
)

// walkDocument validates the top-level document node and dispatches to the
// recursive node walker. It expects doc to be a yaml.DocumentNode.
func walkDocument(doc *goyaml.Node) []ValidationError {
	if doc.Kind != goyaml.DocumentNode || len(doc.Content) == 0 {
		// An empty document is structurally valid; nothing to check.
		return nil
	}
	return walkNode(doc.Content[0])
}

// walkNode recursively validates a YAML AST node and all its descendants,
// enforcing the traCtl YAML subset restrictions.
//
// Go concept — recursion: this function calls itself on child nodes. This is
// the standard way to traverse tree structures in Go (and most languages).
// Think of it like a recursive JSON traversal in JavaScript.
func walkNode(node *goyaml.Node) []ValidationError {
	var errs []ValidationError

	// ── Alias check ────────────────────────────────────────────────────────
	// Aliases have Kind == AliasNode and carry no Tag, Value, or Content.
	// We must handle them first to avoid dereferencing nil Content.
	if node.Kind == goyaml.AliasNode {
		errs = append(errs, ValidationError{
			Code:    ErrAlias,
			Message: "alias references (*) are not permitted in traCtl YAML documents; eliminate anchors and expand values inline",
			Line:    node.Line,
			Column:  node.Column,
		})
		// Aliases have no useful Content to recurse into.
		return errs
	}

	// ── Anchor check ───────────────────────────────────────────────────────
	// node.Anchor holds the raw anchor name (without the & sigil).
	// An empty string means no anchor is declared.
	if node.Anchor != "" {
		errs = append(errs, ValidationError{
			Code:    ErrAnchor,
			Message: "anchor declaration &" + node.Anchor + " is not permitted in traCtl YAML documents",
			Line:    node.Line,
			Column:  node.Column,
		})
		// Continue validation; anchor presence alone doesn't prevent further
		// useful checks on the same node.
	}

	// ── Explicit tag check ─────────────────────────────────────────────────
	// yaml.v3 sets the TaggedStyle flag only when the author explicitly wrote
	// a tag (e.g. !!str, !!int, !custom). Implicitly resolved tags (e.g. the
	// library inferring !!bool from the value "true") do NOT set this flag.
	//
	// Go concept — bitmask: node.Style is a uint32 bitmask. The & operator
	// checks whether a specific bit is set. This is identical to JS:
	//   (node.Style & yaml.TaggedStyle) !== 0
	if node.Style&goyaml.TaggedStyle != 0 {
		errs = append(errs, ValidationError{
			Code:    ErrExplicitTag,
			Message: "explicit YAML tag " + node.Tag + " is not permitted; remove the tag and let the value speak for itself",
			Line:    node.Line,
			Column:  node.Column,
		})
	}

	// ── Kind-specific validation ───────────────────────────────────────────
	switch node.Kind {
	case goyaml.ScalarNode:
		errs = append(errs, validateScalarNode(node)...)

	case goyaml.MappingNode:
		errs = append(errs, validateMappingNode(node)...)

	case goyaml.SequenceNode:
		errs = append(errs, validateSequenceNode(node)...)

	case goyaml.DocumentNode, goyaml.AliasNode:
		// document/alias nodes are not expected at this level; no-op
	}

	return errs
}

// validateScalarNode checks a scalar node for quoting style violations and
// then delegates to the value-level scalar validator in scalar.go.
func validateScalarNode(node *goyaml.Node) []ValidationError {
	var errs []ValidationError

	// Single-quoted string check.
	// yaml.v3 sets SingleQuotedStyle only when the author used '...' quoting.
	if node.Style&goyaml.SingleQuotedStyle != 0 {
		errs = append(errs, ValidationError{
			Code:    ErrSingleQuote,
			Message: "single-quoted string literals are not permitted; use plain scalars or double-quoted strings",
			Line:    node.Line,
			Column:  node.Column,
		})
	}

	errs = append(errs, validateScalar(node)...)
	return errs
}

// validateMappingNode enforces all mapping-level subset restrictions:
// non-empty flow style, complex keys, merge keys, duplicate keys, reserved
// key names, and recursively validates all key and value nodes.
func validateMappingNode(node *goyaml.Node) []ValidationError {
	var errs []ValidationError

	// Non-empty flow-style mapping check.
	// {} (empty) is explicitly permitted by spec §7.11.
	// {key: value} (non-empty) is not.
	if node.Style&goyaml.FlowStyle != 0 && len(node.Content) > 0 {
		errs = append(errs, ValidationError{
			Code:    ErrFlowStyle,
			Message: "non-empty flow-style mappings {…} are not permitted; use block style",
			Line:    node.Line,
			Column:  node.Column,
		})
		// Do not recurse: the entire flow mapping is invalid and recursing
		// would produce confusing secondary errors on its contents.
		return errs
	}

	// A MappingNode's Content is a flat list of alternating key/value nodes:
	//   [key0, val0, key1, val1, ...]
	// Iterating in steps of 2 is the idiomatic yaml.v3 pattern.
	//
	// Go concept — make with capacity: make(map[string]bool, n) pre-allocates
	// the map to avoid repeated internal resizing as we add entries.
	// This is like new Map() in JS but with a size hint.
	seen := make(map[string]bool, len(node.Content)/2)

	for i := 0; i+1 < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valNode := node.Content[i+1]

		// Complex key check.
		// Only scalar string keys are permitted. Sequence-as-key and
		// mapping-as-key forms (YAML's "? key" notation) are rejected.
		if keyNode.Kind != goyaml.ScalarNode {
			errs = append(errs, ValidationError{
				Code:    ErrComplexKey,
				Message: "complex (non-scalar) mapping keys are not permitted; all keys must be plain string scalars",
				Line:    keyNode.Line,
				Column:  keyNode.Column,
			})
			// Still validate the value node so we collect all errors.
			errs = append(errs, walkNode(valNode)...)
			continue
		}

		key := keyNode.Value

		// Merge key check.
		// "<<" as a key is the YAML 1.1 merge key extension. It is not part
		// of YAML 1.2 core and is prohibited unconditionally.
		if key == "<<" {
			errs = append(errs, ValidationError{
				Code:    ErrMergeKey,
				Message: "YAML merge keys (<<:) are not permitted; inline the merged values explicitly",
				Line:    keyNode.Line,
				Column:  keyNode.Column,
			})
			// Skip the value node to avoid cascading alias errors.
			continue
		}

		// _ulid and reserved-prefix key checks (shared rule, YAML-specific codes).
		for _, ce := range common.CheckKey(key, keyNode.Line, keyNode.Column) {
			switch ce.Code {
			case common.ErrULIDKey:
				ce.Code = ErrULIDKey
			case common.ErrReservedPrefix:
				ce.Code = ErrReservedPrefix
			case common.ErrEncoding, common.ErrDuplicateKey:
				// CheckKey never produces these codes; handled elsewhere
			}
			errs = append(errs, ce)
		}

		// Duplicate key check.
		// YAML 1.2 leaves duplicate-key behaviour implementation-defined.
		// traCtl makes it an explicit error.
		if seen[key] {
			errs = append(errs, ValidationError{
				Code:    ErrDuplicateKey,
				Message: "duplicate mapping key \"" + key + "\"",
				Line:    keyNode.Line,
				Column:  keyNode.Column,
			})
		}
		seen[key] = true

		// Recurse into the key node itself (it may carry an anchor or tag).
		errs = append(errs, walkNode(keyNode)...)
		// Recurse into the value node.
		errs = append(errs, walkNode(valNode)...)
	}

	return errs
}

// validateSequenceNode enforces sequence-level subset restrictions:
// non-empty flow style, and recursively validates all element nodes.
func validateSequenceNode(node *goyaml.Node) []ValidationError {
	var errs []ValidationError

	// Non-empty flow-style sequence check.
	// [] (empty) is explicitly permitted by spec §7.11.
	// [a, b, c] (non-empty) is not.
	if node.Style&goyaml.FlowStyle != 0 && len(node.Content) > 0 {
		errs = append(errs, ValidationError{
			Code:    ErrFlowStyle,
			Message: "non-empty flow-style sequences […] are not permitted; use block style",
			Line:    node.Line,
			Column:  node.Column,
		})
		return errs
	}

	for _, child := range node.Content {
		errs = append(errs, walkNode(child)...)
	}

	return errs
}
