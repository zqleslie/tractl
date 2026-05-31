// Package assertion implements evaluation of spec.Assertion rules against step results.
package assertion

import "fmt"

// ErrorCode identifies the category of an assertion evaluation error.
type ErrorCode string

// Assertion error code constants.
const (
	ErrUnknownKind      ErrorCode = "ASSERTION_UNKNOWN_KIND"
	ErrUnknownOperator  ErrorCode = "ASSERTION_UNKNOWN_OPERATOR"
	ErrOperatorMismatch ErrorCode = "ASSERTION_OPERATOR_MISMATCH"
	ErrBodyParseFailed  ErrorCode = "ASSERTION_BODY_PARSE_FAILED"
	ErrRegexInvalid     ErrorCode = "ASSERTION_REGEX_INVALID"
	ErrSchemaStub       ErrorCode = "ASSERTION_SCHEMA_STUB"
	ErrScriptStub       ErrorCode = "ASSERTION_SCRIPT_STUB"
	ErrExtensionStub    ErrorCode = "ASSERTION_EXTENSION_STUB"
)

// AssertionError is a structured evaluation error carrying a machine-readable code.
type AssertionError struct { //nolint:revive
	Code    ErrorCode
	Message string
}

func (e *AssertionError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}
