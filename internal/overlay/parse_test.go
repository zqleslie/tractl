// parse_test.go tests overlay document parsing, cross-format parity, and
// the ApplyToJSON convenience wrapper.
// Split from engine_test.go (1012 lines) per standard 12.1.
package overlay

import (
	"reflect"
	"testing"
)

// ─── Parse ────────────────────────────────────────────────────────────────────

func TestParseYAML(t *testing.T) {
	src := []byte(`
metadata:
  name: test-overlay
patches:
  - target:
      path: variables
    action: deepMerge
    data:
      env: staging
`)
	doc, err := Parse(src, "yaml")
	if err != nil {
		t.Fatalf("Parse yaml: %v", err)
	}
	if doc.Metadata == nil || doc.Metadata.Name != "test-overlay" {
		t.Errorf("metadata.name: got %v", doc.Metadata)
	}
	if len(doc.Patches) != 1 {
		t.Fatalf("patches count: want 1, got %d", len(doc.Patches))
	}
	if doc.Patches[0].Target.Path != "variables" {
		t.Errorf("patch target.path: got %q", doc.Patches[0].Target.Path)
	}
}

func TestParseJSON(t *testing.T) {
	src := []byte(`{
  "metadata": {"name": "json-overlay"},
  "patches": [
    {
      "target": {"path": "variables"},
      "action": "deepMerge",
      "data": {"env": "production"}
    }
  ]
}`)
	doc, err := Parse(src, "json")
	if err != nil {
		t.Fatalf("Parse json: %v", err)
	}
	if doc.Metadata == nil || doc.Metadata.Name != "json-overlay" {
		t.Errorf("metadata.name: got %v", doc.Metadata)
	}
}

func TestParseUnsupportedFormat(t *testing.T) {
	_, err := Parse([]byte("x"), "toml")
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
}

// ─── Cross-format parity ──────────────────────────────────────────────────────

// TestCrossFormatParity verifies that semantically equivalent overlays in different
// formats produce identical results when applied to the same source.
// Reference: overlay spec §6, §20, ADR-002 §3a
func TestCrossFormatParity(t *testing.T) {
	src := specDoc()
	src["variables"] = map[string]any{"env": "dev"}

	yamlOverlay := []byte(`
metadata:
  name: parity-overlay
patches:
  - target:
      path: variables
    action: deepMerge
    data:
      region: us-east-1
`)
	jsonOverlay := []byte(`{
  "metadata": {"name": "parity-overlay"},
  "patches": [
    {
      "target": {"path": "variables"},
      "action": "deepMerge",
      "data": {"region": "us-east-1"}
    }
  ]
}`)

	docYAML, err := Parse(yamlOverlay, "yaml")
	if err != nil {
		t.Fatalf("Parse yaml: %v", err)
	}
	docJSON, err := Parse(jsonOverlay, "json")
	if err != nil {
		t.Fatalf("Parse json: %v", err)
	}

	resYAML, err := eng().Apply(src, []*OverlayDocument{docYAML})
	if err != nil {
		t.Fatalf("Apply yaml overlay: %v", err)
	}
	resJSON, err := eng().Apply(src, []*OverlayDocument{docJSON})
	if err != nil {
		t.Fatalf("Apply json overlay: %v", err)
	}

	// Compare variables (execution fields only; provenance may differ in timestamps).
	mY, _ := resYAML.(map[string]any)
	mJ, _ := resJSON.(map[string]any)

	varsY, _ := mY["variables"].(map[string]any)
	varsJ, _ := mJ["variables"].(map[string]any)

	if !reflect.DeepEqual(varsY, varsJ) {
		t.Errorf("cross-format parity violated:\nyaml result vars: %v\njson result vars: %v", varsY, varsJ)
	}
}

// ─── ApplyToJSON ─────────────────────────────────────────────────────────────

func TestApplyToJSON_RoundTrip(t *testing.T) {
	src := specDoc()
	src["variables"] = map[string]any{"x": "1"}
	srcJSON := mustMarshal(t, src)

	doc := &OverlayDocument{
		Patches: []Patch{
			{Target: Target{Path: "variables"}, Action: ActionDeepMerge, Data: map[string]any{"y": "2"}},
		},
	}

	out, err := eng().ApplyToJSON(srcJSON, []*OverlayDocument{doc})
	if err != nil {
		t.Fatalf("ApplyToJSON: %v", err)
	}

	result := mustUnmarshal(t, out)
	vars, _ := result["variables"].(map[string]any)
	if vars["x"] != "1" || vars["y"] != "2" {
		t.Errorf("ApplyToJSON round-trip vars: %v", vars)
	}
}
