// Package json provides the traCtl JSON format parser.
package json

import (
	"fmt"

	parscommon "github.com/tractl/tractl/internal/parser/common"
	"github.com/tractl/tractl/internal/spec"
	jsonvalidator "github.com/tractl/tractl/internal/validation/json"
)

// Parse reads JSON source bytes and returns a canonical *spec.TraCtlSpec.
//
// Parse is a three-phase operation:
//  1. Format validation  — runs internal/validation/json.Validate(src).
//     Returns error if any format rule is violated.
//  2. Unmarshal          — deserialises JSON bytes into *spec.TraCtlSpec
//     using struct tags.
//  3. Enrichment         — assigns _ulid to all identifiable entities,
//     sets metadata.sourceFormat = "json",
//     sets metadata.sourceRef = sourceRef argument.
//
// The canonical validator (internal/validation.SpecValidator) is NOT called
// here. Callers run it separately on the returned *spec.TraCtlSpec.
func Parse(src []byte, sourceRef string) (*spec.TraCtlSpec, error) {
	result := jsonvalidator.Validate(src)
	if !result.Valid {
		return nil, parscommon.FormatValidationError(result.Errors)
	}

	s, err := parscommon.UnmarshalJSONBytes(src)
	if err != nil {
		return nil, fmt.Errorf("json: unmarshal: %w", err)
	}

	parscommon.AssignULIDs(s)
	parscommon.SetMetadata(s, "json", sourceRef)
	return s, nil
}
