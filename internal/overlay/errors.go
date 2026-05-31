// Package overlay implements the overlay engine per overlay spec §3.
package overlay

import "fmt"

// ErrorCode is a typed string for overlay engine error codes.
type ErrorCode string

const (
	// ErrPathNotFound is returned when a path selector resolves to zero nodes.
	// Reference: overlay spec §9.1
	ErrPathNotFound ErrorCode = "OVERLAY_PATH_NOT_FOUND"

	// ErrModeOneNoMatch is returned when a mode:one selector resolves to zero nodes.
	// Reference: overlay spec §9.2, §9.3
	ErrModeOneNoMatch ErrorCode = "OVERLAY_MODE_ONE_NO_MATCH"

	// ErrModeOneMultipleMatches is returned when a mode:one selector resolves to more than one node.
	// Reference: overlay spec §9.2, §9.3
	ErrModeOneMultipleMatches ErrorCode = "OVERLAY_MODE_ONE_MULTIPLE_MATCHES"

	// ErrRemoveWithPatch is returned when a removal patch has a non-nil patch body.
	// Reference: overlay spec §11
	ErrRemoveWithPatch ErrorCode = "OVERLAY_REMOVE_WITH_PATCH"

	// ErrNoIdentityContract is returned when appendUnique is applied to an array
	// with no declared identity contract.
	// Reference: overlay spec §10.5
	ErrNoIdentityContract ErrorCode = "OVERLAY_NO_IDENTITY_CONTRACT"

	// ErrInvalidAction is returned when an action is incompatible with target field
	// semantics or is an unrecognised string.
	// Reference: overlay spec §10.4
	ErrInvalidAction ErrorCode = "OVERLAY_INVALID_ACTION"

	// ErrSourceSelectorUnknown is returned when a source selector type is not recognised.
	// Reference: overlay spec §9.3
	ErrSourceSelectorUnknown ErrorCode = "OVERLAY_SOURCE_SELECTOR_UNKNOWN"

	// ErrCloneFailed is returned when the JSON round-trip deep clone fails.
	ErrCloneFailed ErrorCode = "OVERLAY_CLONE_FAILED"
)

// EngineError is a structured error from the overlay engine.
// It carries the error code, the patch index that caused the failure, the target
// selector description, and the action attempted.
type EngineError struct {
	Code        ErrorCode
	PatchIndex  int
	OverlayName string
	Target      string
	Action      string
	Detail      string
}

func (e *EngineError) Error() string {
	name := e.OverlayName
	if name == "" {
		name = "<unnamed>"
	}
	return fmt.Sprintf("[%s] overlay %q patch[%d] target=%q action=%q: %s",
		e.Code, name, e.PatchIndex, e.Target, e.Action, e.Detail)
}

// engineErr is a convenience constructor for EngineError.
func engineErr(code ErrorCode, overlayName string, patchIndex int, target, action, detail string) *EngineError {
	return &EngineError{
		Code:        code,
		PatchIndex:  patchIndex,
		OverlayName: overlayName,
		Target:      target,
		Action:      action,
		Detail:      detail,
	}
}
