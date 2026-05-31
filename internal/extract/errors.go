package extract

import "fmt"

// ErrorCode is a typed sentinel for extract engine failures.
type ErrorCode string

// Extract error code constants.
const (
	ErrUnknownSource   ErrorCode = "EXTRACT_UNKNOWN_SOURCE"
	ErrBodyParseFailed ErrorCode = "EXTRACT_BODY_PARSE_FAILED"
	ErrPathNotFound    ErrorCode = "EXTRACT_PATH_NOT_FOUND"
	ErrTimingField     ErrorCode = "EXTRACT_TIMING_FIELD_UNKNOWN"
	ErrMetadataField   ErrorCode = "EXTRACT_METADATA_FIELD_UNKNOWN"
	ErrExtensionStub   ErrorCode = "EXTRACT_EXTENSION_STUB"
	ErrInvalidScope    ErrorCode = "EXTRACT_INVALID_SCOPE"
)

// ExtractError is returned by Engine.Extract for domain-level failures.
type ExtractError struct { //nolint:revive
	Code    ErrorCode
	Message string
}

func (e *ExtractError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}
