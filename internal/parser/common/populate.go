package common

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tractl/tractl/internal/spec"
	validcommon "github.com/tractl/tractl/internal/validation/common"

	goyaml "gopkg.in/yaml.v3"
)

// AssignULIDs walks the spec tree depth-first and assigns a new ULID to every
// entity whose ULID field is currently empty. Document order is preserved.
// Entities: Workflow, Step, Assertion, Extract.
func AssignULIDs(s *spec.TraCtlSpec) {
	for i := range s.Workflows {
		if s.Workflows[i].ULID == "" {
			s.Workflows[i].ULID = NewULID()
		}
		for j := range s.Workflows[i].Steps {
			if s.Workflows[i].Steps[j].ULID == "" {
				s.Workflows[i].Steps[j].ULID = NewULID()
			}
			for k := range s.Workflows[i].Steps[j].Assertions {
				if s.Workflows[i].Steps[j].Assertions[k].ULID == "" {
					s.Workflows[i].Steps[j].Assertions[k].ULID = NewULID()
				}
			}
			for k := range s.Workflows[i].Steps[j].Extracts {
				if s.Workflows[i].Steps[j].Extracts[k].ULID == "" {
					s.Workflows[i].Steps[j].Extracts[k].ULID = NewULID()
				}
			}
		}
	}
}

// SetMetadata ensures s.Metadata is non-nil and sets SourceFormat and SourceRef.
// Name and Description from the authored document are preserved.
func SetMetadata(s *spec.TraCtlSpec, sourceFormat, sourceRef string) {
	if s.Metadata == nil {
		s.Metadata = &spec.Metadata{}
	}
	s.Metadata.SourceFormat = sourceFormat
	s.Metadata.SourceRef = sourceRef
}

// UnmarshalYAMLBytes deserialises YAML bytes into a *spec.TraCtlSpec using
// struct tags. ULID assignment and metadata population are the caller's
// responsibility.
func UnmarshalYAMLBytes(src []byte) (*spec.TraCtlSpec, error) {
	var s spec.TraCtlSpec
	if err := goyaml.Unmarshal(src, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// UnmarshalJSONBytes deserialises JSON bytes into a *spec.TraCtlSpec using
// struct tags. ULID assignment and metadata population are the caller's
// responsibility.
func UnmarshalJSONBytes(src []byte) (*spec.TraCtlSpec, error) {
	var s spec.TraCtlSpec
	if err := json.Unmarshal(src, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// FormatValidationError converts a slice of ValidationError into a single
// error value. Each error is formatted as "[CODE] message (field: path)".
func FormatValidationError(errs []validcommon.ValidationError) error {
	if len(errs) == 0 {
		return nil
	}
	parts := make([]string, len(errs))
	for i, e := range errs {
		parts[i] = e.Error()
	}
	return fmt.Errorf("format validation failed:\n%s", strings.Join(parts, "\n"))
}
