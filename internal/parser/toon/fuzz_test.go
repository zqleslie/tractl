// fuzz_test.go defines fuzz tests for the TOON parser.
// Run with: go test -fuzz=FuzzParseTOON ./internal/parser/toon/
package toon_test

import (
	"testing"

	parsertoon "github.com/tractl/tractl/internal/parser/toon"
)

func FuzzParseTOON(f *testing.F) {
	// Seed: valid minimal TOON spec (double-quoted strings, no anchors/aliases)
	f.Add([]byte(`schemaVersion: 1
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
`))

	// Seed: valid spec with assertions, extracts and dependsOn
	f.Add([]byte(`schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wf-001
    steps:
      - id: step-1
        kind: request
        request:
          protocol: http
          target: "https://example.com"
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
      - id: step-2
        kind: request
        dependsOn:
          - step-1
        request:
          protocol: http
          target: "https://example.com/b"
          operation: GET
`))

	// Seed: empty input — must return error, not panic
	f.Add([]byte(``))

	// Seed: whitespace only
	f.Add([]byte(`   `))

	// Seed: document separator (rejected by TOON validator)
	f.Add([]byte(`schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wf-001
    steps: []
---
extra: data
`))

	// Seed: anchor usage (rejected by TOON validator)
	f.Add([]byte(`schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: &wf wf-001
    steps: []
`))

	// Seed: tab indentation (rejected by TOON validator)
	f.Add([]byte("schemaVersion: 1\ncapabilities:\n\t- protocol.http\nworkflows: []\n"))

	// Seed: malformed structure — must return error, not panic
	f.Add([]byte(`{invalid: [toon`))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Parse must never panic regardless of input.
		// Returning an error for malformed input is correct behaviour.
		result, err := parsertoon.Parse(data, "fuzz")
		if err != nil {
			return
		}
		if result == nil {
			t.Error("Parse returned nil result with nil error")
		}
	})
}
