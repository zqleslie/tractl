// spec_validator.go defines SpecValidator and its Validate method, which
// orchestrate all Phase 1.1 validation rules against a parsed *spec.TraCtlSpec.
package validation

import "github.com/tractl/tractl/internal/spec"

// SpecValidator validates TraCtlSpec documents against Phase 1.1 rules.
// All rules run independently; the complete error list is returned.
type SpecValidator struct{}

// NewSpecValidator returns a ready-to-use SpecValidator.
func NewSpecValidator() *SpecValidator { return &SpecValidator{} }

// Validate runs all Phase 1.1 rules and returns every error found.
// Returns nil when the document is valid.
func (v *SpecValidator) Validate(s *spec.TraCtlSpec) []ValidationError {
	var errs []ValidationError
	errs = append(errs, validateSchemaVersion(s)...)
	errs = append(errs, validateCapabilities(s)...)
	errs = append(errs, validateWorkflowsPresent(s)...)
	errs = append(errs, validateWorkflows(s)...)
	errs = append(errs, validateDiagnosticsKinds(s)...)
	errs = append(errs, validateDiagnosticsRetention(s)...)
	return errs
}
