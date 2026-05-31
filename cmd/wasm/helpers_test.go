// helpers_test.go tests the pure (non-syscall/js) bridge helper functions.
// These functions live in helpers_pure.go and have no WASM build constraint,
// so they can be exercised with a standard `go test` run.

package main

import (
	"errors"
	"testing"

	"github.com/tractl/tractl/internal/engine"
	"github.com/tractl/tractl/internal/validation"
)

func TestBridgeError(t *testing.T) {
	result := bridgeError("SOME_CODE", "something went wrong")
	errMap, ok := result["error"].(map[string]any)
	if !ok {
		t.Fatal("bridgeError: missing 'error' key")
	}
	if errMap["code"] != "SOME_CODE" {
		t.Errorf("bridgeError: code = %q, want %q", errMap["code"], "SOME_CODE")
	}
	if errMap["message"] != "something went wrong" {
		t.Errorf("bridgeError: message = %q, want %q", errMap["message"], "something went wrong")
	}
}

func TestExecutionError(t *testing.T) {
	result := executionError("timeout")
	errMap, ok := result["error"].(map[string]any)
	if !ok {
		t.Fatal("executionError: missing 'error' key")
	}
	if errMap["code"] != "TRACTL_EXECUTION_ERROR" {
		t.Errorf("executionError: code = %q, want TRACTL_EXECUTION_ERROR", errMap["code"])
	}
	if errMap["message"] != "timeout" {
		t.Errorf("executionError: message = %q, want %q", errMap["message"], "timeout")
	}
}

func TestRunResultPayload_AddsSurfaceMetadata(t *testing.T) {
	result := &engine.RunResult{
		Passed: true,
		Workflows: []engine.WorkflowOutcome{
			{
				WorkflowID: "wf-1",
				Passed:     true,
				Steps: []engine.StepOutcome{
					{
						StepID:          "step-1",
						ResponseStatus:  200,
						ResponseHeaders: map[string]string{"content-type": "application/json"},
						ResponseBody:    `{"ok":true}`,
					},
				},
			},
		},
	}
	payload, err := runResultPayload(result)
	if err != nil {
		t.Fatalf("runResultPayload: unexpected error: %v", err)
	}
	if payload["surface"] != surfaceName {
		t.Errorf("runResultPayload: surface = %v, want %q", payload["surface"], surfaceName)
	}
	if payload["executionMode"] != executionMode {
		t.Errorf("runResultPayload: executionMode = %v, want %q", payload["executionMode"], executionMode)
	}
	if payload["networkProvider"] != networkProvider {
		t.Errorf("runResultPayload: networkProvider = %v, want %q", payload["networkProvider"], networkProvider)
	}
	// Verify output uses lowercase keys (not raw engine capitalized names).
	if _, hasWorkflows := payload["Workflows"]; hasWorkflows {
		t.Error("runResultPayload: output contains raw 'Workflows' key — should use localapi RunResult shape")
	}
	if _, hasStatusCode := payload["statusCode"]; !hasStatusCode {
		t.Error("runResultPayload: output missing 'statusCode' key from localapi.RunResult")
	}
}

func TestPreview(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"short string returned unchanged", "hello", "hello"},
		{"empty string returned unchanged", "", ""},
		{"exactly at limit returned unchanged", string(make([]rune, previewRuneLimit)), string(make([]rune, previewRuneLimit))},
		{"truncated at limit", string(make([]rune, previewRuneLimit+1)), string(make([]rune, previewRuneLimit))},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := preview(tc.input)
			if len([]rune(got)) > previewRuneLimit {
				t.Errorf("preview: returned %d runes, want ≤ %d", len([]rune(got)), previewRuneLimit)
			}
			if tc.input == tc.want && got != tc.want {
				t.Errorf("preview(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestPreview_MultiByte(t *testing.T) {
	// Verify preview counts runes, not bytes.
	emoji := "😀"
	long := ""
	for i := 0; i < previewRuneLimit+5; i++ {
		long += emoji
	}
	got := preview(long)
	if runeLen := len([]rune(got)); runeLen != previewRuneLimit {
		t.Errorf("preview multi-byte: got %d runes, want %d", runeLen, previewRuneLimit)
	}
}

func TestProjectVersion_ReturnsFallback(t *testing.T) {
	// In test binaries debug.ReadBuildInfo returns "(devel)", so we always get the fallback.
	v := projectVersion()
	if v == "" {
		t.Error("projectVersion: returned empty string")
	}
}

func TestValidationFailure(t *testing.T) {
	errs := []any{map[string]any{"message": "bad field"}}
	result := validationFailure(errs)
	if result["valid"] != false {
		t.Errorf("validationFailure: valid = %v, want false", result["valid"])
	}
	if result["surface"] != surfaceName {
		t.Errorf("validationFailure: surface = %v, want %q", result["surface"], surfaceName)
	}
}

func TestFormatValidationError(t *testing.T) {
	err := errors.New("bad syntax at line 3")
	result := formatValidationError(err)
	if len(result) != 1 {
		t.Fatalf("formatValidationError: len = %d, want 1", len(result))
	}
	m, ok := result[0].(map[string]any)
	if !ok {
		t.Fatal("formatValidationError: element is not map[string]any")
	}
	if m["message"] != "bad syntax at line 3" {
		t.Errorf("formatValidationError: message = %q, want %q", m["message"], "bad syntax at line 3")
	}
}

func TestSpecValidationErrors(t *testing.T) {
	errs := []validation.ValidationError{
		{Field: "steps", Code: "REQUIRED", Message: "steps is required"},
	}
	result := specValidationErrors(errs)
	if len(result) != 1 {
		t.Fatalf("specValidationErrors: len = %d, want 1", len(result))
	}
	m, ok := result[0].(map[string]any)
	if !ok {
		t.Fatal("specValidationErrors: element is not map[string]any")
	}
	if m["field"] != "steps" {
		t.Errorf("specValidationErrors: field = %q, want %q", m["field"], "steps")
	}
	if m["code"] != "REQUIRED" {
		t.Errorf("specValidationErrors: code = %q, want %q", m["code"], "REQUIRED")
	}
}

func TestJsonSafeObject(t *testing.T) {
	type inner struct {
		Name string `json:"name"`
	}
	val := inner{Name: "tractl"}
	result, err := jsonSafeObject(val)
	if err != nil {
		t.Fatalf("jsonSafeObject: unexpected error: %v", err)
	}
	if result["name"] != "tractl" {
		t.Errorf("jsonSafeObject: name = %v, want %q", result["name"], "tractl")
	}
}

func TestJsonSafeObject_UnmarshalableInput(t *testing.T) {
	// Channels cannot be marshalled to JSON.
	_, err := jsonSafeObject(make(chan int))
	if err == nil {
		t.Error("jsonSafeObject: expected error for unmarshallable input, got nil")
	}
}

func TestRequireSupportedFormat_Unknown(t *testing.T) {
	bad := requireSupportedFormat("xml")
	if bad == nil {
		t.Fatal("requireSupportedFormat: expected error for unknown format, got nil")
	}
	errMap, ok := bad["error"].(map[string]any)
	if !ok {
		t.Fatal("requireSupportedFormat: missing error map")
	}
	if errMap["code"] != "TRACTL_WASM_INVALID_FORMAT" {
		t.Errorf("requireSupportedFormat: code = %q, want TRACTL_WASM_INVALID_FORMAT", errMap["code"])
	}
}

func TestRequireSupportedFormat_YmlAlias(t *testing.T) {
	if bad := requireSupportedFormat("yml"); bad != nil {
		t.Errorf("requireSupportedFormat: yml should be accepted, got %v", bad)
	}
}

func TestParseDocument_UnknownFormat(t *testing.T) {
	_, err := parseDocument("content: 1", "xml")
	if err == nil {
		t.Error("parseDocument: expected error for unknown format, got nil")
	}
}

func TestParseDocument_ValidFormats(t *testing.T) {
	yamlDoc := "name: test\nsteps: []\n"
	jsonDoc := `{"name":"test","steps":[]}`

	tests := []struct {
		format string
		doc    string
	}{
		{"yaml", yamlDoc},
		{"json", jsonDoc},
	}
	for _, tc := range tests {
		t.Run(tc.format, func(t *testing.T) {
			_, err := parseDocument(tc.doc, tc.format)
			// Parse may fail on spec validation but should not fail with "unsupported format".
			if err != nil && err.Error() == `unsupported format "`+tc.format+`"` {
				t.Errorf("parseDocument(%q): unexpectedly rejected valid format", tc.format)
			}
		})
	}
}

func TestValidateDocument_UnknownFormat(t *testing.T) {
	_, err := validateDocument("content: 1", "toml")
	if err == nil {
		t.Error("validateDocument: expected error for unknown format, got nil")
	}
}

func TestValidFormats_Coverage(t *testing.T) {
	for _, f := range engine.SupportedFormats {
		if !engine.IsSupportedFormat(f) {
			t.Errorf("engine.SupportedFormats contains %q but IsSupportedFormat rejects it", f)
		}
	}
	if len(engine.SupportedFormats) != 3 {
		t.Errorf("engine.SupportedFormats length = %d, want 3", len(engine.SupportedFormats))
	}
}
