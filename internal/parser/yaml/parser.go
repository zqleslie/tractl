// Package yaml provides the traCtl YAML format parser.
package yaml

import (
	"fmt"

	parscommon "github.com/tractl/tractl/internal/parser/common"
	"github.com/tractl/tractl/internal/spec"
	yamlvalidator "github.com/tractl/tractl/internal/validation/yaml"
)

// Parse reads YAML source bytes and returns a canonical *spec.TraCtlSpec.
//
// Parse is a three-phase operation:
//  1. Format validation  — runs internal/validation/yaml.Validate(src).
//     Returns error if any format rule is violated.
//  2. Unmarshal          — deserialises YAML bytes into *spec.TraCtlSpec
//     using struct tags. _ulid fields are skipped (yaml:"-").
//  3. Enrichment         — assigns _ulid to all identifiable entities,
//     sets metadata.sourceFormat = "yaml",
//     sets metadata.sourceRef = sourceRef argument.
//
// The canonical validator (internal/validation.SpecValidator) is NOT called
// here. Callers run it separately on the returned *spec.TraCtlSpec.
//
// sourceRef is an opaque string identifying the document origin (file path,
// URL, or similar). It is stored in metadata.sourceRef verbatim.
func Parse(src []byte, sourceRef string) (*spec.TraCtlSpec, error) {
	result := yamlvalidator.Validate(src)
	if !result.Valid {
		return nil, parscommon.FormatValidationError(result.Errors)
	}

	s, err := parscommon.UnmarshalYAMLBytes(src)
	if err != nil {
		return nil, fmt.Errorf("yaml: unmarshal: %w", err)
	}

	parscommon.AssignULIDs(s)
	parscommon.SetMetadata(s, "yaml", sourceRef)
	return s, nil
}
