// walker.go implements the gopkg.in/yaml.v3 AST walker for TOON documents,
// enforcing TOON-specific subset restrictions: anchors, aliases, explicit tags,
// single-quoted strings, merge keys, complex keys, and duplicate keys.
package toon

import (
	goyaml "gopkg.in/yaml.v3"

	"github.com/tractl/tractl/internal/validation/common"
)

// walkDocument validates the top-level document node.
func walkDocument(doc *goyaml.Node) []ValidationError {
	if doc.Kind != goyaml.DocumentNode || len(doc.Content) == 0 {
		return nil
	}
	return walkNode(doc.Content[0])
}

// walkNode recursively validates a TOON AST node and all its descendants.
func walkNode(node *goyaml.Node) []ValidationError {
	var errs []ValidationError

	// Alias check — TOON has no hidden-inheritance mechanism.
	if node.Kind == goyaml.AliasNode {
		errs = append(errs, ValidationError{
			Code:    ErrAlias,
			Message: "alias references (*) are not permitted in TOON documents; expand values inline",
			Line:    node.Line,
			Column:  node.Column,
		})
		return errs
	}

	// Anchor check.
	if node.Anchor != "" {
		errs = append(errs, ValidationError{
			Code:    ErrAnchor,
			Message: "anchor declaration &" + node.Anchor + " is not permitted in TOON documents",
			Line:    node.Line,
			Column:  node.Column,
		})
	}

	// Explicit tag check.
	if node.Style&goyaml.TaggedStyle != 0 {
		errs = append(errs, ValidationError{
			Code:    ErrExplicitTag,
			Message: "explicit YAML tag " + node.Tag + " is not permitted in TOON documents",
			Line:    node.Line,
			Column:  node.Column,
		})
	}

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

// validateScalarNode checks a TOON scalar for prohibited quoting styles.
// Note: unlike YAML, TOON does NOT prohibit yes/no/on/off — those are bare
// strings in TOON (spec §7.2). We do prohibit single-quoted strings (§7.1).
func validateScalarNode(node *goyaml.Node) []ValidationError {
	var errs []ValidationError

	if node.Style&goyaml.SingleQuotedStyle != 0 {
		errs = append(errs, ValidationError{
			Code:    ErrSingleQuote,
			Message: "single-quoted string literals are not permitted in TOON; use plain scalars or double-quoted strings",
			Line:    node.Line,
			Column:  node.Column,
		})
	}

	return errs
}

// validateMappingNode enforces all mapping-level TOON restrictions.
func validateMappingNode(node *goyaml.Node) []ValidationError {
	var errs []ValidationError

	// Non-empty flow-style mapping check. {} (empty) is permitted (§8.3).
	if node.Style&goyaml.FlowStyle != 0 && len(node.Content) > 0 {
		errs = append(errs, ValidationError{
			Code:    ErrFlowStyle,
			Message: "non-empty inline mappings {…} are not permitted in TOON; use block form",
			Line:    node.Line,
			Column:  node.Column,
		})
		return errs
	}

	seen := make(map[string]bool, len(node.Content)/2)

	for i := 0; i+1 < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valNode := node.Content[i+1]

		if keyNode.Kind != goyaml.ScalarNode {
			errs = append(errs, ValidationError{
				Code:    ErrComplexKey,
				Message: "complex (non-scalar) mapping keys are not permitted in TOON",
				Line:    keyNode.Line,
				Column:  keyNode.Column,
			})
			errs = append(errs, walkNode(valNode)...)
			continue
		}

		key := keyNode.Value

		// Merge key check.
		if key == "<<" {
			errs = append(errs, ValidationError{
				Code:    ErrMergeKey,
				Message: "YAML merge keys (<<:) are not permitted in TOON; inline the merged values explicitly",
				Line:    keyNode.Line,
				Column:  keyNode.Column,
			})
			continue
		}

		// _ulid / reserved-prefix checks (remap generic codes to TOON codes).
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
		if seen[key] {
			errs = append(errs, ValidationError{
				Code:    ErrDuplicateKey,
				Message: "duplicate mapping key \"" + key + "\"",
				Line:    keyNode.Line,
				Column:  keyNode.Column,
			})
		}
		seen[key] = true

		errs = append(errs, walkNode(keyNode)...)
		errs = append(errs, walkNode(valNode)...)
	}

	return errs
}

// validateSequenceNode enforces sequence-level TOON restrictions.
func validateSequenceNode(node *goyaml.Node) []ValidationError {
	var errs []ValidationError

	// Non-empty flow-style sequence check. [] (empty) is permitted (§8.3).
	if node.Style&goyaml.FlowStyle != 0 && len(node.Content) > 0 {
		errs = append(errs, ValidationError{
			Code:    ErrFlowStyle,
			Message: "non-empty inline sequences […] are not permitted in TOON; use block form",
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
