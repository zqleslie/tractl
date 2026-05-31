// spec_test.go tests the exported methods of the spec package,
// covering DiagnosticsConfig.EffectiveFor across all meaningful
// nil/non-nil and field-level combinations.
package spec_test

import (
	"reflect"
	"testing"

	"github.com/tractl/tractl/internal/spec"
)

func TestDiagnosticsConfig_EffectiveFor(t *testing.T) {
	// workflowCfg and stepCfg are reusable configs with all fields populated
	// and distinct values so accidental field sourcing is immediately visible.
	workflowCfg := &spec.DiagnosticsConfig{
		Enabled:     true,
		Kinds:       []spec.DiagnosticsKind{spec.DiagnosticsKindTLS, spec.DiagnosticsKindTCP},
		CaptureBody: true,
		Retention:   spec.RetentionWorkflow,
	}
	stepCfg := &spec.DiagnosticsConfig{
		Enabled:     false,
		Kinds:       []spec.DiagnosticsKind{spec.DiagnosticsKindDNS},
		CaptureBody: false,
		Retention:   spec.RetentionStep,
	}

	tests := []struct {
		name     string
		receiver *spec.DiagnosticsConfig // workflow-level config (nil receiver is valid)
		arg      *spec.DiagnosticsConfig // step-level config
		want     spec.DiagnosticsConfig
	}{
		// ── Three-branch coverage ────────────────────────────────────────────────
		{
			// Branch 3: both nil → disabled zero-value (spec §15.4)
			name:     "nil receiver nil arg returns zero value disabled config",
			receiver: nil,
			arg:      nil,
			want:     spec.DiagnosticsConfig{},
		},
		{
			// Branch 1: stepConfig non-nil → step wins unconditionally
			name:     "nil receiver non-nil arg returns arg",
			receiver: nil,
			arg:      stepCfg,
			want:     *stepCfg,
		},
		{
			// Branch 2: receiver non-nil, stepConfig nil → workflow config used
			name:     "non-nil receiver nil arg returns receiver",
			receiver: workflowCfg,
			arg:      nil,
			want:     *workflowCfg,
		},
		{
			// Branch 1 again: step beats workflow when both non-nil
			name:     "non-nil receiver non-nil arg returns arg step wins unconditionally",
			receiver: workflowCfg,
			arg:      stepCfg,
			want:     *stepCfg,
		},
		{
			// A zero-value but non-nil stepConfig is still non-nil, so it wins.
			// This confirms the "unconditionally" part of §15.4: presence of the
			// step config pointer, not its content, determines the winner.
			name:     "non-nil zero-value arg beats non-nil receiver",
			receiver: workflowCfg,
			arg:      &spec.DiagnosticsConfig{},
			want:     spec.DiagnosticsConfig{},
		},

		// ── Field-level coverage: each field comes from the correct source ────────
		{
			// Enabled: false on step overrides true on workflow
			name:     "Enabled false on step overrides true on workflow",
			receiver: &spec.DiagnosticsConfig{Enabled: true},
			arg:      &spec.DiagnosticsConfig{Enabled: false},
			want:     spec.DiagnosticsConfig{Enabled: false},
		},
		{
			// Enabled: true on step overrides false on workflow
			name:     "Enabled true on step overrides false on workflow",
			receiver: &spec.DiagnosticsConfig{Enabled: false},
			arg:      &spec.DiagnosticsConfig{Enabled: true},
			want:     spec.DiagnosticsConfig{Enabled: true},
		},
		{
			// Kinds: the entire step Kinds slice replaces the workflow Kinds slice
			name: "Kinds from step replace workflow Kinds entirely",
			receiver: &spec.DiagnosticsConfig{
				Kinds: []spec.DiagnosticsKind{spec.DiagnosticsKindTLS, spec.DiagnosticsKindTCP},
			},
			arg: &spec.DiagnosticsConfig{
				Kinds: []spec.DiagnosticsKind{spec.DiagnosticsKindDNS, spec.DiagnosticsKindConnection},
			},
			want: spec.DiagnosticsConfig{
				Kinds: []spec.DiagnosticsKind{spec.DiagnosticsKindDNS, spec.DiagnosticsKindConnection},
			},
		},
		{
			// CaptureBody: false on step overrides true on workflow
			name:     "CaptureBody false on step overrides true on workflow",
			receiver: &spec.DiagnosticsConfig{CaptureBody: true},
			arg:      &spec.DiagnosticsConfig{CaptureBody: false},
			want:     spec.DiagnosticsConfig{CaptureBody: false},
		},
		{
			// Retention: step-scope on step overrides workflow-scope on workflow
			name:     "Retention step-scope on step overrides workflow-scope on workflow",
			receiver: &spec.DiagnosticsConfig{Retention: spec.RetentionWorkflow},
			arg:      &spec.DiagnosticsConfig{Retention: spec.RetentionStep},
			want:     spec.DiagnosticsConfig{Retention: spec.RetentionStep},
		},
		{
			// Retention: spec-scope on step overrides step-scope on workflow
			name:     "Retention spec-scope on step overrides step-scope on workflow",
			receiver: &spec.DiagnosticsConfig{Retention: spec.RetentionStep},
			arg:      &spec.DiagnosticsConfig{Retention: spec.RetentionSpec},
			want:     spec.DiagnosticsConfig{Retention: spec.RetentionSpec},
		},
		{
			// All fields: step config with all fields set wins over workflow config
			name:     "all fields set on step win over all fields set on workflow",
			receiver: workflowCfg,
			arg: &spec.DiagnosticsConfig{
				Enabled:     true,
				Kinds:       []spec.DiagnosticsKind{spec.DiagnosticsKindLifecycle, spec.DiagnosticsKindTransport},
				CaptureBody: true,
				Retention:   spec.RetentionSpec,
			},
			want: spec.DiagnosticsConfig{
				Enabled:     true,
				Kinds:       []spec.DiagnosticsKind{spec.DiagnosticsKindLifecycle, spec.DiagnosticsKindTransport},
				CaptureBody: true,
				Retention:   spec.RetentionSpec,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.receiver.EffectiveFor(tc.arg)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("EffectiveFor():\n  got  %+v\n  want %+v", got, tc.want)
			}
		})
	}
}
