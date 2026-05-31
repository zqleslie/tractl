// Package validation implements the canonical contract enforcement layer.
// All rules run independently; the complete error list is always returned.
// Reference: tractl_spec.md §18.3
package validation

import "github.com/tractl/tractl/internal/spec"

// ErrorCode is a typed validation error code.
type ErrorCode string

// Validation error code constants.
const (
	MissingRequiredField      ErrorCode = "MISSING_REQUIRED_FIELD"
	InvalidSchemaVersion      ErrorCode = "INVALID_SCHEMA_VERSION"
	InvalidIdentifier         ErrorCode = "INVALID_IDENTIFIER"
	ReservedIdentifier        ErrorCode = "RESERVED_IDENTIFIER"
	DuplicateWorkflowID       ErrorCode = "DUPLICATE_WORKFLOW_ID"
	DuplicateStepID           ErrorCode = "DUPLICATE_STEP_ID"
	UnknownDependencyRef      ErrorCode = "UNKNOWN_DEPENDENCY_REF"
	CyclicDependency          ErrorCode = "CYCLIC_DEPENDENCY"
	InvalidStepKind           ErrorCode = "INVALID_STEP_KIND"
	StepBodyKindMismatch      ErrorCode = "STEP_BODY_KIND_MISMATCH"
	InvalidCapabilityContract ErrorCode = "INVALID_CAPABILITY_CONTRACT"
)

// ValidationError describes a single validation failure.
// Field uses dot-bracket notation matching the traCtlSpec document structure.
//
// Examples:
//
//	schemaVersion
//	capabilities[0]
//	workflows[0].id
//	workflows[0].steps[1].dependsOn[0]
type ValidationError struct { //nolint:revive
	Field   string
	Code    ErrorCode
	Message string
}

// Validator is the contract enforcement interface for canonical TraCtlSpec documents.
// Implementations must run all rules independently and return the complete error list.
// Reference: tractl_spec.md §18.3
type Validator interface {
	Validate(s *spec.TraCtlSpec) []ValidationError
}
