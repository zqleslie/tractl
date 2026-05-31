package extract

import (
	"time"

	"github.com/tractl/tractl/internal/spec"
)

// ExtractResult is the output of running one spec.Extract.
type ExtractResult struct { //nolint:revive
	ExtractID   string
	ExtractULID string
	Source      string
	// As is the variable name bound in context (defaults to ExtractID).
	As    string
	Scope spec.ExtractScope
	// ValuePlaceholder is the credential-safe reference used in observability.
	// It is NOT the actual extracted value.
	// Format: "[extract:<extractID>]"
	ValuePlaceholder string
	// Resolved is true when extraction succeeded.
	Resolved bool
	Duration time.Duration
}
