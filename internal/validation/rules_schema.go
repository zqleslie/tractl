// rules_schema.go defines validation rules for top-level document fields:
// schemaVersion (Rule 1), capabilities (Rule 2), and the shared identifier
// and capability-format validators used by other rule files.
package validation

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tractl/tractl/internal/spec"
)

// validateSchemaVersion enforces Rule 1. Reference: tractl_spec.md §3.1
func validateSchemaVersion(s *spec.TraCtlSpec) []ValidationError {
	if s.SchemaVersion < 1 {
		return []ValidationError{{
			Field:   "schemaVersion",
			Code:    InvalidSchemaVersion,
			Message: "schemaVersion must be a positive integer >= 1",
		}}
	}
	return nil
}

// validateCapabilities enforces Rule 2. Reference: tractl_spec.md §3.2
func validateCapabilities(s *spec.TraCtlSpec) []ValidationError {
	if len(s.Capabilities) == 0 {
		return []ValidationError{{
			Field:   "capabilities",
			Code:    MissingRequiredField,
			Message: "capabilities is required and must not be empty",
		}}
	}
	var errs []ValidationError
	for i, cap := range s.Capabilities {
		field := fmt.Sprintf("capabilities[%d]", i)
		if err := validateCapabilityContract(cap, field); err != nil {
			errs = append(errs, *err)
		}
	}
	return errs
}

// validateCapabilityContract validates one capability string.
// Form A (platform): dotted identifier, no @.
// Form B (extension): dotted identifier @ positive integer >= 1.
// Reference: tractl_spec.md §3.2
func validateCapabilityContract(capStr, field string) *ValidationError {
	if capStr == "" {
		return &ValidationError{
			Field:   field,
			Code:    InvalidCapabilityContract,
			Message: "capability must not be empty",
		}
	}

	atIdx := strings.Index(capStr, "@")
	if atIdx < 0 {
		if !isValidPlatformCapability(capStr) {
			return &ValidationError{
				Field:   field,
				Code:    InvalidCapabilityContract,
				Message: fmt.Sprintf("invalid platform capability %q: must be a dotted identifier (e.g. protocol.http)", capStr),
			}
		}
		return nil
	}

	// Form B: <identifier>@<version>
	identifier, versionStr := capStr[:atIdx], capStr[atIdx+1:]
	if !isValidPlatformCapability(identifier) {
		return &ValidationError{
			Field:   field,
			Code:    InvalidCapabilityContract,
			Message: fmt.Sprintf("invalid extension capability %q: identifier part must be a dotted identifier", capStr),
		}
	}
	if versionStr == "" {
		return &ValidationError{
			Field:   field,
			Code:    InvalidCapabilityContract,
			Message: fmt.Sprintf("invalid extension capability %q: version after @ must be a positive integer >= 1", capStr),
		}
	}
	version, err := strconv.Atoi(versionStr)
	if err != nil || version < 1 {
		return &ValidationError{
			Field:   field,
			Code:    InvalidCapabilityContract,
			Message: fmt.Sprintf("invalid extension capability %q: version after @ must be a positive integer >= 1", capStr),
		}
	}
	return nil
}

// isValidPlatformCapability returns true when s is a dotted multi-segment identifier.
func isValidPlatformCapability(s string) bool {
	return strings.Contains(s, ".") && capPlatformRe.MatchString(s)
}
