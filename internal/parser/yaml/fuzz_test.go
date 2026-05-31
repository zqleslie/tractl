// fuzz_test.go defines fuzz tests for the YAML parser.
// Run with: go test -fuzz=FuzzParseYAML ./internal/parser/yaml/
package yaml_test

import (
	"testing"
	"unicode/utf8"

	parseryaml "github.com/tractl/tractl/internal/parser/yaml"
)

func FuzzParseYAML(f *testing.F) {
	// Seed: valid minimal spec
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

	// Seed: valid spec with assertions and extracts
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
        assertions:
          - id: assert-001
            kind: status
            op: eq
            expected: 200
        extracts:
          - id: extract-001
            source: response
            path: "$.body.token"
`))

	// Seed: valid spec with dependsOn
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

	// Seed: malformed YAML — must return error, not panic
	f.Add([]byte(`{invalid: [yaml`))

	// Seed: tab indentation (rejected by validator)
	f.Add([]byte("schemaVersion: 1\ncapabilities:\n\t- protocol.http\nworkflows: []\n"))

	// Seed: YAML with anchor (rejected by validator)
	f.Add([]byte(`schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: &wf wf-001
    steps: []
`))

	// Seed: deeply nested key
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
          headers:
            Authorization: "Bearer token"
            Content-Type: "application/json"
`))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Informational only — the parser must handle non-UTF-8 gracefully.
		_ = utf8.Valid(data)

		// Parse must never panic regardless of input.
		// Returning an error for malformed input is correct behaviour.
		result, err := parseryaml.Parse(data, "fuzz")
		if err != nil {
			return
		}
		if result == nil {
			t.Error("Parse returned nil result with nil error")
		}
	})
}
