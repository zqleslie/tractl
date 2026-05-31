// parser_test.go tests the YAML format parser. Covers: valid minimal documents,
// ULID assignment to all spec entities, rejection of authored _ulid keys,
// format validation errors, multi-step dependsOn, metadata preservation, and
// failurePolicy round-trip.
package yaml_test

import (
	"strings"
	"testing"

	parseryaml "github.com/tractl/tractl/internal/parser/yaml"
)

func TestParse_MinimalValidDocument(t *testing.T) {
	src := []byte(`schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wf-001
    steps:
      - id: step-001
        kind: request
        request:
          protocol: http
          target: https://example.com
          operation: GET
`)
	s, err := parseryaml.Parse(src, "test-source")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.SchemaVersion != 1 {
		t.Errorf("SchemaVersion: got %d, want 1", s.SchemaVersion)
	}
	if len(s.Capabilities) != 1 || s.Capabilities[0] != "protocol.http" {
		t.Errorf("Capabilities: got %v, want [protocol.http]", s.Capabilities)
	}
	if len(s.Workflows) == 0 || s.Workflows[0].ID != "wf-001" {
		t.Errorf("Workflows[0].ID: got %q, want wf-001", s.Workflows[0].ID)
	}
	if s.Metadata == nil {
		t.Fatal("Metadata is nil")
	}
	if s.Metadata.SourceFormat != "yaml" {
		t.Errorf("SourceFormat: got %q, want yaml", s.Metadata.SourceFormat)
	}
	if s.Metadata.SourceRef != "test-source" {
		t.Errorf("SourceRef: got %q, want test-source", s.Metadata.SourceRef)
	}
}

func TestParse_ULIDsAssigned(t *testing.T) {
	src := []byte(`schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wf-001
    steps:
      - id: step-001
        kind: request
        request:
          protocol: http
          target: https://example.com
          operation: GET
        assertions:
          - id: assert-001
            kind: status
            op: eq
            expected: 200
        extracts:
          - id: extract-001
            source: response
            path: "$.body.token"
      - id: step-002
        kind: request
        request:
          protocol: http
          target: https://example.com/b
          operation: GET
`)
	s, err := parseryaml.Parse(src, "test-source")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	seen := make(map[string]bool)
	checkUnique := func(label, ulid string) {
		t.Helper()
		if ulid == "" {
			t.Errorf("%s: ULID is empty", label)
			return
		}
		if seen[ulid] {
			t.Errorf("%s: duplicate ULID %q", label, ulid)
		}
		seen[ulid] = true
	}

	wf := s.Workflows[0]
	checkUnique("Workflow[0]", wf.ULID)
	checkUnique("Step[0]", wf.Steps[0].ULID)
	checkUnique("Step[1]", wf.Steps[1].ULID)
	checkUnique("Assertion[0]", wf.Steps[0].Assertions[0].ULID)
	checkUnique("Extract[0]", wf.Steps[0].Extracts[0].ULID)
}

func TestParse_ULIDsNotInAuthored(t *testing.T) {
	// _ulid in authored YAML must be rejected by the format validator.
	src := []byte(`schemaVersion: 1
capabilities:
  - protocol.http
_ulid: foo
workflows:
  - id: wf-001
    steps:
      - id: step-001
        kind: request
`)
	s, err := parseryaml.Parse(src, "test-source")
	if err == nil {
		t.Fatal("expected error for authored _ulid key, got nil")
	}
	if s != nil {
		t.Errorf("expected nil spec on error, got non-nil")
	}
}

func TestParse_InvalidFormat_TabIndentation(t *testing.T) {
	src := []byte("schemaVersion: 1\ncapabilities:\n\t- protocol.http\nworkflows: []\n")
	_, err := parseryaml.Parse(src, "test-source")
	if err == nil {
		t.Fatal("expected error for tab indentation, got nil")
	}
	if !strings.Contains(err.Error(), "YAML_TAB_INDENTATION") {
		t.Errorf("expected YAML_TAB_INDENTATION in error, got: %v", err)
	}
}

func TestParse_InvalidFormat_AnchorRejected(t *testing.T) {
	src := []byte(`schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: &wf wf-001
    steps: []
`)
	_, err := parseryaml.Parse(src, "test-source")
	if err == nil {
		t.Fatal("expected error for anchor, got nil")
	}
}

func TestParse_MultiStepWithDependsOn(t *testing.T) {
	src := []byte(`schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wf-001
    steps:
      - id: step-1
        kind: request
        request:
          protocol: http
          target: https://example.com
          operation: GET
      - id: step-2
        kind: request
        dependsOn:
          - step-1
        request:
          protocol: http
          target: https://example.com/b
          operation: GET
`)
	s, err := parseryaml.Parse(src, "test-source")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	steps := s.Workflows[0].Steps
	if len(steps[1].DependsOn) != 1 || steps[1].DependsOn[0] != "step-1" {
		t.Errorf("DependsOn: got %v, want [step-1]", steps[1].DependsOn)
	}
	if steps[0].ULID == "" || steps[1].ULID == "" {
		t.Error("expected non-empty ULIDs on both steps")
	}
	if steps[0].ULID == steps[1].ULID {
		t.Error("expected distinct ULIDs on steps")
	}
}

func TestParse_MetadataPreserved(t *testing.T) {
	src := []byte(`schemaVersion: 1
capabilities:
  - protocol.http
metadata:
  name: "My Workflow"
  description: "Integration suite"
workflows:
  - id: wf-001
    steps:
      - id: step-001
        kind: request
        request:
          protocol: http
          target: https://example.com
          operation: GET
`)
	s, err := parseryaml.Parse(src, "ref-path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Metadata.Name != "My Workflow" {
		t.Errorf("Name: got %q, want 'My Workflow'", s.Metadata.Name)
	}
	if s.Metadata.Description != "Integration suite" {
		t.Errorf("Description: got %q", s.Metadata.Description)
	}
	if s.Metadata.SourceFormat != "yaml" {
		t.Errorf("SourceFormat: got %q, want yaml", s.Metadata.SourceFormat)
	}
}

func TestParse_FailurePolicyPreserved(t *testing.T) {
	src := []byte(`schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wf-001
    failurePolicy: failFast
    steps:
      - id: step-001
        kind: request
        request:
          protocol: http
          target: https://example.com
          operation: GET
`)
	s, err := parseryaml.Parse(src, "test-source")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Workflows[0].FailurePolicy != "failFast" {
		t.Errorf("FailurePolicy: got %q, want failFast", s.Workflows[0].FailurePolicy)
	}
}

func TestParse_EmptyFailurePolicy_DefaultsToEmpty(t *testing.T) {
	src := []byte(`schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wf-001
    steps:
      - id: step-001
        kind: request
        request:
          protocol: http
          target: https://example.com
          operation: GET
`)
	s, err := parseryaml.Parse(src, "test-source")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Workflows[0].FailurePolicy != "" {
		t.Errorf("FailurePolicy: got %q, want empty", s.Workflows[0].FailurePolicy)
	}
}
