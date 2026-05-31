package validation

import (
	"testing"

	"github.com/tractl/tractl/internal/spec"
)

func TestSchemaVersion(t *testing.T) {
	runValidatorTests(t, []validatorTestCase{
		{
			name:     "valid schemaVersion=1",
			spec:     minimalValidSpec(),
			wantPass: true,
		},
		{
			name:      "schemaVersion zero",
			spec:      func() *spec.TraCtlSpec { s := minimalValidSpec(); s.SchemaVersion = 0; return s }(),
			wantCodes: []ErrorCode{InvalidSchemaVersion},
		},
		{
			name:      "schemaVersion negative",
			spec:      func() *spec.TraCtlSpec { s := minimalValidSpec(); s.SchemaVersion = -1; return s }(),
			wantCodes: []ErrorCode{InvalidSchemaVersion},
		},
	})
}

func TestCapabilities(t *testing.T) {
	runValidatorTests(t, []validatorTestCase{
		{
			name:      "empty capabilities list",
			spec:      func() *spec.TraCtlSpec { s := minimalValidSpec(); s.Capabilities = []string{}; return s }(),
			wantCodes: []ErrorCode{MissingRequiredField},
		},
		{
			name:      "capability empty string",
			spec:      func() *spec.TraCtlSpec { s := minimalValidSpec(); s.Capabilities = []string{""}; return s }(),
			wantCodes: []ErrorCode{InvalidCapabilityContract},
		},
		{
			name:      "capability bare word (no dot)",
			spec:      func() *spec.TraCtlSpec { s := minimalValidSpec(); s.Capabilities = []string{"http"}; return s }(),
			wantCodes: []ErrorCode{InvalidCapabilityContract},
		},
		{
			name: "capability trailing @",
			spec: func() *spec.TraCtlSpec {
				s := minimalValidSpec()
				s.Capabilities = []string{"protocol.http@"}
				return s
			}(),
			wantCodes: []ErrorCode{InvalidCapabilityContract},
		},
		{
			name: "extension @0",
			spec: func() *spec.TraCtlSpec {
				s := minimalValidSpec()
				s.Capabilities = []string{"protocol.ext.amqp@0"}
				return s
			}(),
			wantCodes: []ErrorCode{InvalidCapabilityContract},
		},
		{
			name: "extension @non-integer",
			spec: func() *spec.TraCtlSpec {
				s := minimalValidSpec()
				s.Capabilities = []string{"provider.foo@x"}
				return s
			}(),
			wantCodes: []ErrorCode{InvalidCapabilityContract},
		},
		{
			name: "valid platform capabilities",
			spec: func() *spec.TraCtlSpec {
				s := minimalValidSpec()
				s.Capabilities = []string{"protocol.http", "scripting.js", "fuzz.builtin", "diagnostics.tls"}
				return s
			}(),
			wantPass: true,
		},
		{
			name: "valid extension capabilities",
			spec: func() *spec.TraCtlSpec {
				s := minimalValidSpec()
				s.Capabilities = []string{"protocol.http", "protocol.ext.amqp@1", "secret.ext.vault@1"}
				return s
			}(),
			wantPass: true,
		},
		{
			// platform.cap@version is treated as Form B (extension cap) — spec does not forbid it.
			name: "platform cap with @version treated as extension — passes",
			spec: func() *spec.TraCtlSpec {
				s := minimalValidSpec()
				s.Capabilities = []string{"protocol.http@1"}
				return s
			}(),
			wantPass: true,
		},
		{
			// dotted with no @ is Form A — syntactically valid.
			name: "dotted cap without @version is valid Form A",
			spec: func() *spec.TraCtlSpec {
				s := minimalValidSpec()
				s.Capabilities = []string{"protocol.ext.amqp"}
				return s
			}(),
			wantPass: true,
		},
	})
}
