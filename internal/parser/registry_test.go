// registry_test.go tests the parser registry and ForFormat factory.
package parser

import (
	"strings"
	"testing"
)

func TestForFormat_KnownFormats(t *testing.T) {
	for _, format := range []string{"yaml", "json", "toon"} {
		t.Run(format, func(t *testing.T) {
			p, err := ForFormat(format)
			if err != nil {
				t.Fatalf("ForFormat(%q) returned unexpected error: %v", format, err)
			}
			if p == nil {
				t.Fatalf("ForFormat(%q) returned nil parser", format)
			}
		})
	}
}

func TestForFormat_UnknownFormat(t *testing.T) {
	_, err := ForFormat("toml")
	if err == nil {
		t.Fatal("ForFormat(\"toml\") expected error for unknown format, got nil")
	}
}

func TestForFormat_EmptyFormat(t *testing.T) {
	_, err := ForFormat("")
	if err == nil {
		t.Fatal("ForFormat(\"\") expected error for empty format, got nil")
	}
}

func TestSupportedFormats_ContainsExpected(t *testing.T) {
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

func TestSupportedFormats_IsSorted(t *testing.T) {
	formats := SupportedFormats()
	for i := 1; i < len(formats); i++ {
		if formats[i] < formats[i-1] {
			t.Errorf("SupportedFormats() not sorted: %q before %q", formats[i-1], formats[i])
		}
	}
}

func TestForFormat_ParseCallable(t *testing.T) {
	// Verify the returned parser can be called without panicking.
	// Empty input will return a validation error — that is acceptable.
	p, err := ForFormat("yaml")
	if err != nil {
		t.Fatalf("unexpected error getting yaml parser: %v", err)
	}
	// Must not panic — an error for empty/invalid input is correct behaviour.
	result, _ := p.Parse([]byte(""), "test")
	_ = result
}

func TestForFormat_ParseValidInput(t *testing.T) {
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
	p, err := ForFormat("yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s, err := p.Parse(validYAML, "test-source")
	if err != nil {
		t.Fatalf("Parse returned unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("Parse returned nil spec with nil error")
	}
	if s.SchemaVersion != 1 {
		t.Errorf("SchemaVersion: got %d, want 1", s.SchemaVersion)
	}
}

func TestForFormat_ErrorMessageContainsFormat(t *testing.T) {
	_, err := ForFormat("badformat")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "badformat") {
		t.Errorf("error message should contain the bad format name, got: %v", err)
	}
}

func TestForFormat_ErrorMessageContainsSupportedFormats(t *testing.T) {
	_, err := ForFormat("badformat")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// The error should hint at what formats are supported.
	for _, expected := range []string{"yaml", "json", "toon"} {
		if !strings.Contains(err.Error(), expected) {
			t.Errorf("error message should list supported format %q; got: %v", expected, err)
		}
	}
}
