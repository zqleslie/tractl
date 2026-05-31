package sandbox

import "fmt"

// ErrorCode is a string enumerating sandbox error kinds.
type ErrorCode string

// Sandbox error code constants.
const (
	ErrSyntax               ErrorCode = "sandbox.syntax_error"
	ErrRuntime              ErrorCode = "sandbox.runtime_error"
	ErrTimeout              ErrorCode = "sandbox.timeout"
	ErrInvalidResult        ErrorCode = "sandbox.invalid_result"
	ErrSourceRefUnsupported ErrorCode = "sandbox.sourceref_not_supported"
	ErrLanguageUnsupported  ErrorCode = "sandbox.language_not_supported"
)

// SandboxError wraps a categorized sandbox error.
type SandboxError struct { //nolint:revive
	Code    ErrorCode
	Message string
	Cause   error
}

func (e *SandboxError) Error() string { return fmt.Sprintf("[%s] %s", e.Code, e.Message) }
func (e *SandboxError) Unwrap() error { return e.Cause }
