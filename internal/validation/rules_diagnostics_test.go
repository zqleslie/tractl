package validation

import (
	"strings"
	"testing"

	"github.com/tractl/tractl/internal/spec"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func specWithStepDiag(kinds []spec.DiagnosticsKind, retention spec.DiagnosticsRetention) *spec.TraCtlSpec {
	s := minimalValidSpec()
	s.Workflows[0].Steps[0].Diagnostics = &spec.DiagnosticsConfig{
		Kinds:     kinds,
		Retention: retention,
	}
	return s
}

func specWithWFDiag(kinds []spec.DiagnosticsKind, retention spec.DiagnosticsRetention) *spec.TraCtlSpec {
	s := minimalValidSpec()
	s.Workflows[0].Diagnostics = &spec.DiagnosticsConfig{
		Kinds:     kinds,
		Retention: retention,
	}
	return s
}

// ── validateDiagnosticsKinds ──────────────────────────────────────────────────

func TestValidateDiagnosticsKinds_AllSixAccepted(t *testing.T) {
	s := specWithStepDiag([]spec.DiagnosticsKind{
		spec.DiagnosticsKindTLS,
		spec.DiagnosticsKindTCP,
		spec.DiagnosticsKindTransport,
		spec.DiagnosticsKindLifecycle,
		spec.DiagnosticsKindDNS,
		spec.DiagnosticsKindConnection,
	}, "")
	errs := validateDiagnosticsKinds(s)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errs), errs)
	}
}

func TestValidateDiagnosticsKinds_UnknownKindRejected(t *testing.T) {
	s := specWithStepDiag([]spec.DiagnosticsKind{"tls", "wireshark"}, "")
	errs := validateDiagnosticsKinds(s)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if errs[0].Code != DiagnosticsUnknownKind {
		t.Errorf("expected code %q, got %q", DiagnosticsUnknownKind, errs[0].Code)
	}
	if !strings.Contains(errs[0].Message, "wireshark") {
		t.Errorf("expected message to contain 'wireshark', got: %s", errs[0].Message)
	}
}

func TestValidateDiagnosticsKinds_MultipleUnknownKinds(t *testing.T) {
	s := specWithStepDiag([]spec.DiagnosticsKind{"bad1", "bad2"}, "")
	errs := validateDiagnosticsKinds(s)
	if len(errs) != 2 {
		t.Errorf("expected 2 errors, got %d: %v", len(errs), errs)
	}
}

func TestValidateDiagnosticsKinds_WorkflowLevelUnknownKind(t *testing.T) {
	s := specWithWFDiag([]spec.DiagnosticsKind{"nmap"}, "")
	errs := validateDiagnosticsKinds(s)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if errs[0].Code != DiagnosticsUnknownKind {
		t.Errorf("expected code %q, got %q", DiagnosticsUnknownKind, errs[0].Code)
	}
	// Workflow-level error message must not reference a step ID.
	if strings.Contains(errs[0].Message, "step") && strings.Contains(errs[0].Message, "step1") {
		t.Errorf("workflow-level error message must not contain step ID, got: %s", errs[0].Message)
	}
}

func TestValidateDiagnosticsKinds_DisabledBlockStillValidated(t *testing.T) {
	s := minimalValidSpec()
	s.Workflows[0].Steps[0].Diagnostics = &spec.DiagnosticsConfig{
		Enabled: false,
		Kinds:   []spec.DiagnosticsKind{"bad"},
	}
	errs := validateDiagnosticsKinds(s)
	if len(errs) != 1 {
		t.Errorf("expected 1 error even when Enabled=false, got %d: %v", len(errs), errs)
	}
}

func TestValidateDiagnosticsKinds_NilDiagnosticsNoError(t *testing.T) {
	s := minimalValidSpec()
	errs := validateDiagnosticsKinds(s)
	if len(errs) != 0 {
		t.Errorf("expected no errors for nil diagnostics, got %d: %v", len(errs), errs)
	}
}

func TestValidateDiagnosticsKinds_EmptyKindsListNoError(t *testing.T) {
	s := specWithStepDiag([]spec.DiagnosticsKind{}, "")
	errs := validateDiagnosticsKinds(s)
	if len(errs) != 0 {
		t.Errorf("expected no errors for empty kinds list, got %d: %v", len(errs), errs)
	}
}

// ── validateDiagnosticsRetention ─────────────────────────────────────────────

func TestValidateDiagnosticsRetention_ValidValuesAccepted(t *testing.T) {
	for _, r := range []spec.DiagnosticsRetention{
		spec.RetentionStep,
		spec.RetentionWorkflow,
		spec.RetentionSpec,
	} {
		r := r
		t.Run(string(r), func(t *testing.T) {
			errs := validateDiagnosticsRetention(specWithStepDiag(nil, r))
			if len(errs) != 0 {
				t.Errorf("expected no errors for retention=%q, got %d: %v", r, len(errs), errs)
			}
		})
	}
}

func TestValidateDiagnosticsRetention_EmptyRetentionAccepted(t *testing.T) {
	s := specWithStepDiag(nil, "")
	errs := validateDiagnosticsRetention(s)
	if len(errs) != 0 {
		t.Errorf("expected no errors for empty retention, got %d: %v", len(errs), errs)
	}
}

func TestValidateDiagnosticsRetention_UnknownRetentionRejected(t *testing.T) {
	s := specWithStepDiag(nil, "session")
	errs := validateDiagnosticsRetention(s)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if errs[0].Code != DiagnosticsUnknownRetention {
		t.Errorf("expected code %q, got %q", DiagnosticsUnknownRetention, errs[0].Code)
	}
	if !strings.Contains(errs[0].Message, "session") {
		t.Errorf("expected message to contain 'session', got: %s", errs[0].Message)
	}
}

func TestValidateDiagnosticsRetention_WorkflowLevelRejected(t *testing.T) {
	s := specWithWFDiag(nil, "cache")
	errs := validateDiagnosticsRetention(s)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if errs[0].Code != DiagnosticsUnknownRetention {
		t.Errorf("expected code %q, got %q", DiagnosticsUnknownRetention, errs[0].Code)
	}
	// Workflow-level: message must not reference a step ID.
	if strings.Contains(errs[0].Message, "step1") {
		t.Errorf("workflow-level error must not mention step ID, got: %s", errs[0].Message)
	}
}

func TestValidateDiagnosticsRetention_NilDiagnosticsNoError(t *testing.T) {
	s := minimalValidSpec()
	errs := validateDiagnosticsRetention(s)
	if len(errs) != 0 {
		t.Errorf("expected no errors for nil diagnostics, got %d: %v", len(errs), errs)
	}
}

// ── SpecValidator integration ─────────────────────────────────────────────────

func TestSpecValidator_DiagnosticsIntegrated_ValidSpec(t *testing.T) {
	s := minimalValidSpec()
	s.Workflows[0].Diagnostics = &spec.DiagnosticsConfig{
		Enabled:   true,
		Kinds:     []spec.DiagnosticsKind{spec.DiagnosticsKindTLS, spec.DiagnosticsKindDNS},
		Retention: spec.RetentionWorkflow,
	}
	s.Workflows[0].Steps[0].Diagnostics = &spec.DiagnosticsConfig{
		Enabled:   true,
		Kinds:     []spec.DiagnosticsKind{spec.DiagnosticsKindTCP},
		Retention: spec.RetentionStep,
	}
	v := NewSpecValidator()
	errs := v.Validate(s)
	for _, e := range errs {
		if e.Code == DiagnosticsUnknownKind || e.Code == DiagnosticsUnknownRetention {
			t.Errorf("unexpected diagnostics error: [%s] %s", e.Code, e.Message)
		}
	}
}

func TestSpecValidator_DiagnosticsIntegrated_InvalidKindFails(t *testing.T) {
	s := minimalValidSpec()
	s.Workflows[0].Steps[0].Diagnostics = &spec.DiagnosticsConfig{
		Kinds: []spec.DiagnosticsKind{"wiretap"},
	}
	v := NewSpecValidator()
	errs := v.Validate(s)
	if !hasError(errs, DiagnosticsUnknownKind) {
		t.Errorf("expected %q error in full validator output", DiagnosticsUnknownKind)
	}
}
