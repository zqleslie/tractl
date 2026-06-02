package localapi

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestRequestDefToDocument_DefaultProtocol verifies that an empty Protocol field
// produces request.protocol == "http" in the output document (ADR-017 §2).
func TestRequestDefToDocument_DefaultProtocol(t *testing.T) {
	req := RequestDef{
		Protocol: "",
		ID:       "step-1",
		Method:   "GET",
		URL:      "https://api.example.com/users",
	}
	doc, err := requestDefToDocument(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wfs, ok := doc["workflows"].([]map[string]any)
	if !ok || len(wfs) == 0 {
		t.Fatal("missing workflows in output")
	}
	steps, ok := wfs[0]["steps"].([]map[string]any)
	if !ok || len(steps) == 0 {
		t.Fatal("missing steps in output")
	}
	request, ok := steps[0]["request"].(map[string]any)
	if !ok {
		t.Fatal("missing request in step")
	}
	if got := request["protocol"]; got != "http" {
		t.Errorf("expected protocol=%q, got %q", "http", got)
	}
}

// TestRequestDefToDocument_ExplicitProtocol verifies that an explicit Protocol
// value is passed through unchanged (ADR-017 §2).
func TestRequestDefToDocument_ExplicitProtocol(t *testing.T) {
	tests := []struct {
		protocol string
	}{
		{"graphql"},
		{"http"},
		{"soap"},
	}
	for _, tt := range tests {
		t.Run(tt.protocol, func(t *testing.T) {
			req := RequestDef{
				Protocol: tt.protocol,
				ID:       "step-1",
				Method:   "POST",
				URL:      "https://api.example.com/graphql",
				Body: &BodyDef{
					Encoding: "json",
					Content:  `{"key":"value"}`,
				},
			}
			doc, err := requestDefToDocument(req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			wfs := doc["workflows"].([]map[string]any)
			steps := wfs[0]["steps"].([]map[string]any)
			request := steps[0]["request"].(map[string]any)
			if got := request["protocol"]; got != tt.protocol {
				t.Errorf("expected protocol=%q, got %q", tt.protocol, got)
			}
		})
	}
}

// TestGraphQLBodyEncoding_ValidContent verifies that a graphql body with query
// and variables produces a JSON body, Content-Type: application/json, and
// operation: POST (ADR-017 §4).
func TestGraphQLBodyEncoding_ValidContent(t *testing.T) {
	content := `{"query":"{ users { id name } }","variables":{"limit":10}}`
	req := RequestDef{
		ID:     "gql-step",
		Method: "GET", // must be overridden to POST
		URL:    "https://api.example.com/graphql",
		Body: &BodyDef{
			Encoding: "graphql",
			Content:  content,
		},
	}
	doc, err := requestDefToDocument(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wfs := doc["workflows"].([]map[string]any)
	steps := wfs[0]["steps"].([]map[string]any)
	request := steps[0]["request"].(map[string]any)

	// operation must be POST
	if got := request["operation"]; got != "POST" {
		t.Errorf("expected operation=POST, got %q", got)
	}

	// Content-Type must be application/json
	headers, ok := request["headers"].(map[string]string)
	if !ok {
		t.Fatal("expected headers map in request")
	}
	if ct := headers["Content-Type"]; ct != "application/json" {
		t.Errorf("expected Content-Type=application/json, got %q", ct)
	}

	// Body encoding must be "graphql" and content must be valid JSON
	body, ok := request["body"].(map[string]any)
	if !ok {
		t.Fatal("expected body map in request")
	}
	if enc := body["encoding"]; enc != "graphql" {
		t.Errorf("expected encoding=graphql, got %q", enc)
	}
	bodyContent, ok := body["content"].(string)
	if !ok {
		t.Fatal("expected body content to be a string")
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(bodyContent), &parsed); err != nil {
		t.Errorf("body content is not valid JSON: %v", err)
	}
	if _, ok := parsed["query"]; !ok {
		t.Error("body content missing query key")
	}
	if _, ok := parsed["variables"]; !ok {
		t.Error("body content missing variables key")
	}
}

// TestGraphQLBodyEncoding_MissingQuery verifies that a graphql body without
// a "query" key returns a non-nil error (ADR-017 §4).
func TestGraphQLBodyEncoding_MissingQuery(t *testing.T) {
	req := RequestDef{
		ID:  "gql-step",
		URL: "https://api.example.com/graphql",
		Body: &BodyDef{
			Encoding: "graphql",
			Content:  `{"variables":{"limit":10}}`,
		},
	}
	_, err := requestDefToDocument(req)
	if err == nil {
		t.Fatal("expected error for missing query key, got nil")
	}
}

// TestProtocolODataNormalizedToHttp verifies that "odata" is normalized to
// "http" at the mapper boundary (ADR-017 §4).
func TestProtocolODataNormalizedToHttp(t *testing.T) {
	req := RequestDef{
		Protocol: "odata",
		ID:       "step-1",
		Method:   "GET",
		URL:      "https://api.example.com/odata/Products",
	}
	doc, err := requestDefToDocument(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wfs := doc["workflows"].([]map[string]any)
	steps := wfs[0]["steps"].([]map[string]any)
	request := steps[0]["request"].(map[string]any)
	if got := request["protocol"]; got != "http" {
		t.Errorf("expected protocol=%q after odata normalization, got %q", "http", got)
	}
}

// TestCapabilitiesHTTP verifies that a plain HTTP request produces only
// ["protocol.http"] in capabilities (ADR-017 §4).
func TestCapabilitiesHTTP(t *testing.T) {
	req := RequestDef{Protocol: "http", ID: "s", Method: "GET", URL: "https://example.com"}
	doc, err := requestDefToDocument(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	caps, ok := doc["capabilities"].([]string)
	if !ok {
		t.Fatalf("capabilities not []string: %T", doc["capabilities"])
	}
	if len(caps) != 1 || caps[0] != "protocol.http" {
		t.Errorf("expected [\"protocol.http\"], got %v", caps)
	}
}

// TestCapabilitiesGraphQL verifies that a GraphQL request produces both
// "protocol.http" and "protocol.graphql" in capabilities (ADR-017 §4).
func TestCapabilitiesGraphQL(t *testing.T) {
	req := RequestDef{
		Protocol: "graphql",
		ID:       "s",
		Method:   "POST",
		URL:      "https://api.example.com/graphql",
		Body:     &BodyDef{Encoding: "graphql", Content: `{"query":"{ me { id } }"}`},
	}
	doc, err := requestDefToDocument(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	caps, ok := doc["capabilities"].([]string)
	if !ok {
		t.Fatalf("capabilities not []string: %T", doc["capabilities"])
	}
	if len(caps) != 2 || caps[0] != "protocol.http" || caps[1] != "protocol.graphql" {
		t.Errorf("expected [\"protocol.http\",\"protocol.graphql\"], got %v", caps)
	}
}

// TestUnknownEncodingReturnsError verifies that an unrecognized body encoding
// returns an explicit error rather than silently passing through (ADR-017 §5).
func TestUnknownEncodingReturnsError(t *testing.T) {
	req := RequestDef{
		ID:  "s",
		URL: "https://example.com",
		Body: &BodyDef{
			Encoding: "xml",
			Content:  "<root/>",
		},
	}
	_, err := requestDefToDocument(req)
	if err == nil {
		t.Fatal("expected error for unknown encoding xml, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported body encoding") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestKnownEncodingsPassThrough verifies that form, raw, binary, and multipart
// encodings do not return an error (ADR-017 §5).
func TestKnownEncodingsPassThrough(t *testing.T) {
	for _, enc := range []string{"form", "raw", "binary", "multipart"} {
		t.Run(enc, func(t *testing.T) {
			req := RequestDef{
				ID:   "s",
				URL:  "https://example.com",
				Body: &BodyDef{Encoding: enc, Content: "data"},
			}
			_, err := requestDefToDocument(req)
			if err != nil {
				t.Errorf("encoding %q should not error, got: %v", enc, err)
			}
		})
	}
}

// TestGraphQLBodyEncoding_StringContent verifies that a graphql body with a
// JSON string as content is parsed and serialized correctly (ADR-017 §4).
func TestGraphQLBodyEncoding_StringContent(t *testing.T) {
	req := RequestDef{
		ID:  "gql-step",
		URL: "https://api.example.com/graphql",
		Body: &BodyDef{
			Encoding: "graphql",
			Content:  `{"query":"{ me { id } }"}`,
		},
	}
	doc, err := requestDefToDocument(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wfs := doc["workflows"].([]map[string]any)
	steps := wfs[0]["steps"].([]map[string]any)
	request := steps[0]["request"].(map[string]any)
	body := request["body"].(map[string]any)
	bodyContent, ok := body["content"].(string)
	if !ok {
		t.Fatal("expected body content to be a string")
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(bodyContent), &parsed); err != nil {
		t.Errorf("body content is not valid JSON: %v", err)
	}
	if q, ok := parsed["query"].(string); !ok || q == "" {
		t.Error("expected non-empty query in parsed body content")
	}
}
