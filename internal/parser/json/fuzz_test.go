// fuzz_test.go defines fuzz tests for the JSON parser.
// Run with: go test -fuzz=FuzzParseJSON ./internal/parser/json/
package json_test

import (
	"testing"

	parserjson "github.com/tractl/tractl/internal/parser/json"
)

func FuzzParseJSON(f *testing.F) {
	// Seed: valid minimal spec in JSON
	f.Add([]byte(`{
  "schemaVersion": 1,
  "capabilities": ["protocol.http"],
  "workflows": [
    {
      "id": "wf-001",
      "steps": [
        {
          "id": "step-001",
          "kind": "request",
          "request": {
            "protocol": "http",
            "target": "https://example.com",
            "operation": "GET"
          }
        }
      ]
    }
  ]
}`))

	// Seed: valid spec with assertions and extracts
	f.Add([]byte(`{
  "schemaVersion": 1,
  "capabilities": ["protocol.http"],
  "workflows": [
    {
      "id": "wf-001",
      "steps": [
        {
          "id": "step-001",
          "kind": "request",
          "request": {
            "protocol": "http",
            "target": "https://example.com",
            "operation": "GET"
          },
          "assertions": [
            {"id": "assert-001", "kind": "status", "op": "eq", "expected": 200}
          ],
          "extracts": [
            {"id": "extract-001", "source": "response", "path": "$.body.token"}
          ]
        }
      ]
    }
  ]
}`))

	// Seed: valid spec with dependsOn
	f.Add([]byte(`{
  "schemaVersion": 1,
  "capabilities": ["protocol.http"],
  "workflows": [
    {
      "id": "wf-001",
      "steps": [
        {
          "id": "step-1",
          "kind": "request",
          "request": {"protocol": "http", "target": "https://example.com", "operation": "GET"}
        },
        {
          "id": "step-2",
          "kind": "request",
          "dependsOn": ["step-1"],
          "request": {"protocol": "http", "target": "https://example.com/b", "operation": "GET"}
        }
      ]
    }
  ]
}`))

	// Seed: empty input — must return error, not panic
	f.Add([]byte(``))

	// Seed: empty object (missing required fields — validator concern, not parser)
	f.Add([]byte(`{}`))

	// Seed: malformed JSON — must return error, not panic
	f.Add([]byte(`{"schemaVersion": 1, "workflows": [`))

	// Seed: null values at top level
	f.Add([]byte(`{"schemaVersion": null, "workflows": null}`))

	// Seed: wrong types
	f.Add([]byte(`{"schemaVersion": "1", "workflows": "not-an-array"}`))

	// Seed: JSON5-style comment (rejected by validator)
	f.Add([]byte(`// comment
{"schemaVersion": 1}`))

	// Seed: trailing comma (rejected by validator)
	f.Add([]byte(`{"schemaVersion": 1, "workflows": [],}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Parse must never panic regardless of input.
		// Returning an error for malformed input is correct behaviour.
		result, err := parserjson.Parse(data, "fuzz")
		if err != nil {
			return
		}
		if result == nil {
			t.Error("Parse returned nil result with nil error")
		}
	})
}
