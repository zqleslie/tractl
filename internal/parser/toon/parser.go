// Package toon provides the traCtl TOON format parser.
//
// TOON is a strict YAML subset (tractl_toon_spec.md §3). Unmarshaling reuses
// gopkg.in/yaml.v3 via parscommon.UnmarshalYAMLBytes; only the format
// validator differs.
package toon

import (
	"fmt"

	parscommon "github.com/tractl/tractl/internal/parser/common"
	"github.com/tractl/tractl/internal/spec"
	toonvalidator "github.com/tractl/tractl/internal/validation/toon"
)

// Parse reads TOON source bytes and returns a canonical *spec.TraCtlSpec.
//
// Parse is a three-phase operation:
//  1. Format validation  — runs internal/validation/toon.Validate(src).
//     Returns error if any TOON-specific rule is violated.
//  2. Unmarshal          — reuses UnmarshalYAMLBytes because TOON is a YAML
//     subset; struct tags are identical.
//  3. Enrichment         — assigns _ulid to all identifiable entities,
//     sets metadata.sourceFormat = "toon",
//     sets metadata.sourceRef = sourceRef argument.
//
// The canonical validator (internal/validation.SpecValidator) is NOT called
// here. Callers run it separately on the returned *spec.TraCtlSpec.
func Parse(src []byte, sourceRef string) (*spec.TraCtlSpec, error) {
	result := toonvalidator.Validate(src)
	if !result.Valid {
		return nil, parscommon.FormatValidationError(result.Errors)
	}

	s, err := parscommon.UnmarshalYAMLBytes(src)
	if err != nil {
		return nil, fmt.Errorf("toon: unmarshal: %w", err)
	}

	if err := parscommon.AssignULIDs(s); err != nil {
		return nil, fmt.Errorf("toon: assign ULIDs: %w", err)
	}
	parscommon.SetMetadata(s, "toon", sourceRef)
	return s, nil
}
