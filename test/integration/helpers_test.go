//go:build integration

package integration_test

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tractl/tractl/internal/spec"
	"github.com/tractl/tractl/internal/validation"
)

// validationErr wraps a slice of validation errors as a single error value.
type validationErr struct{ errs []validation.ValidationError }

func (e *validationErr) Error() string {
	parts := make([]string, len(e.errs))
	for i, v := range e.errs {
		parts[i] = fmt.Sprintf("[%s] %s: %s", v.Code, v.Field, v.Message)
	}
	return strings.Join(parts, "; ")
}

// overlayMarshalUnmarshal round-trips the overlay engine's map[string]any
// back into a typed *spec.TraCtlSpec.
func overlayMarshalUnmarshal(v any) (*spec.TraCtlSpec, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var out spec.TraCtlSpec
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
