// fuzz_test.go defines fuzz tests for the spec validator.
// Run with: go test -fuzz=FuzzValidate ./internal/validation/
//
// Strategy: unmarshal fuzzer bytes as JSON into *spec.TraCtlSpec, then run
// SpecValidator.Validate. This exercises the validator against the full range
// of structurally-diverse spec shapes that JSON can express, including
// impossible combinations that correctly-authored documents would never
// produce. The JSON unmarshal step is a structural filter only — validation
// errors from Validate are expected and correct for malformed specs.
package validation

import (
	"encoding/json"
	"testing"

	"github.com/tractl/tractl/internal/spec"
)

func FuzzValidate(f *testing.F) {
	// Seed: valid minimal spec as JSON (canonical representation of *spec.TraCtlSpec)
	f.Add([]byte(`{
  "schemaVersion": 1,
  "capabilities": ["protocol.http"],
  "workflows": [
    {
      "id": "wf1",
      "steps": [
        {
          "id": "step1",
          "kind": "request",
          "request": {"protocol": "http", "target": "https://example.com"}
        }
      ]
    }
  ]
}`))

	// Seed: spec with multiple workflows and steps
	f.Add([]byte(`{
  "schemaVersion": 1,
  "capabilities": ["protocol.http"],
  "workflows": [
    {
      "id": "wf1",
      "steps": [
        {"id": "step1", "kind": "request", "request": {"protocol": "http", "target": "https://a.com"}},
        {"id": "step2", "kind": "request", "dependsOn": ["step1"], "request": {"protocol": "http", "target": "https://b.com"}}
      ]
    },
    {
      "id": "wf2",
      "steps": [
        {"id": "step1", "kind": "script", "script": {}}
      ]
    }
  ]
}`))

	// Seed: spec with cyclic dependsOn (validator must detect and return error, not loop)
	f.Add([]byte(`{
  "schemaVersion": 1,
  "capabilities": ["protocol.http"],
  "workflows": [
    {
      "id": "wf1",
      "steps": [
        {"id": "a", "kind": "request", "dependsOn": ["b"], "request": {"protocol": "http", "target": "https://example.com"}},
        {"id": "b", "kind": "request", "dependsOn": ["a"], "request": {"protocol": "http", "target": "https://example.com"}}
      ]
    }
  ]
}`))

	// Seed: empty object (zero-value spec — all rules should fire)
	f.Add([]byte(`{}`))

	// Seed: missing capabilities
	f.Add([]byte(`{"schemaVersion": 1, "workflows": []}`))

	// Seed: duplicate workflow IDs
	f.Add([]byte(`{
  "schemaVersion": 1,
  "capabilities": ["protocol.http"],
  "workflows": [
    {"id": "wf1", "steps": []},
    {"id": "wf1", "steps": []}
  ]
}`))

	// Seed: duplicate step IDs within a workflow
	f.Add([]byte(`{
  "schemaVersion": 1,
  "capabilities": ["protocol.http"],
  "workflows": [
    {
      "id": "wf1",
      "steps": [
        {"id": "step1", "kind": "request", "request": {"protocol": "http", "target": "https://example.com"}},
        {"id": "step1", "kind": "request", "request": {"protocol": "http", "target": "https://example.com/b"}}
      ]
    }
  ]
}`))

	// Seed: empty input — JSON unmarshal will fail, fuzz body returns early
	f.Add([]byte(``))

	// Seed: null
	f.Add([]byte(`null`))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Attempt to construct a *spec.TraCtlSpec from the raw bytes.
		// If the bytes are not valid JSON or cannot be unmarshalled into the struct,
		// skip — we are fuzzing the validator, not the JSON decoder.
		var s spec.TraCtlSpec
		if err := json.Unmarshal(data, &s); err != nil {
			return
		}

		// Validate must never panic on any spec shape.
		// Returning validation errors is expected and correct for malformed specs.
		v := NewSpecValidator()
		_ = v.Validate(&s)
	})
}
