// format_validator_test.go tests the FormatValidator interface and
// ValidatorFor factory function.
package validation

import (
	"strings"
	"testing"
)

func TestValidatorFor_KnownFormats(t *testing.T) {
	for _, format := range []string{"yaml", "json", "toon"} {
		t.Run(format, func(t *testing.T) {
			v, err := ValidatorFor(format)
			if err != nil {
				t.Fatalf("ValidatorFor(%q) returned unexpected error: %v", format, err)
			}
			if v == nil {
				t.Fatalf("ValidatorFor(%q) returned nil validator", format)
			}
		})
	}
}

func TestValidatorFor_UnknownFormat(t *testing.T) {
	_, err := ValidatorFor("toml")
	if err == nil {
		t.Fatal("ValidatorFor(\"toml\") expected error for unknown format, got nil")
	}
	if !strings.Contains(err.Error(), "toml") {
		t.Errorf("error message should mention the unknown format, got: %v", err)
	}
}

func TestValidatorFor_EmptyFormat(t *testing.T) {
	_, err := ValidatorFor("")
	if err == nil {
		t.Fatal("ValidatorFor(\"\") expected error for empty format, got nil")
	}
}

func TestSupportedFormats(t *testing.T) {
	formats := SupportedFormats()
	if len(formats) < 3 {
		t.Fatalf("expected at least 3 supported formats, got %d", len(formats))
	}
	found := make(map[string]bool, len(formats))
	for _, f := range formats {
		found[f] = true
	}
	for _, expected := range []string{"yaml", "json", "toon"} {
		if !found[expected] {
			t.Errorf("SupportedFormats() missing expected format %q; got: %v", expected, formats)
		}
	}
}

func TestSupportedFormats_Sorted(t *testing.T) {
	formats := SupportedFormats()
	for i := 1; i < len(formats); i++ {
		if formats[i] < formats[i-1] {
			t.Errorf("SupportedFormats() not sorted: %v", formats)
		}
	}
}

func TestValidatorFor_ValidateCallable(t *testing.T) {
	for _, format := range []string{"yaml", "json", "toon"} {
		t.Run(format, func(t *testing.T) {
			v, err := ValidatorFor(format)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Must not panic — a ValidationResult (valid or not) is acceptable.
			result := v.Validate([]byte(""))
			if result == nil {
				t.Error("Validate returned nil result")
			}
		})
	}
}

func TestValidatorFor_ValidInputAccepted(t *testing.T) {
	validYAML := []byte(`schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wf-001
    steps:
      - id: step-001
        kind: request
        request:
          protocol: http
          target: "https://example.com"
          operation: GET
`)
	v, err := ValidatorFor("yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := v.Validate(validYAML)
	if result == nil {
		t.Fatal("Validate returned nil result")
	}
	if !result.Valid {
		t.Errorf("expected valid result for correct YAML, got errors: %v", result.Errors)
	}
}
