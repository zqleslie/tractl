package localapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func requestDefToDocument(req RequestDef) (map[string]any, error) {
	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = "GET"
	}
	// GraphQL always uses POST regardless of the Method field (ADR-017 §4).
	if req.Body != nil && strings.EqualFold(req.Body.Encoding, "graphql") {
		method = "POST"
	}

	protocol := req.Protocol
	if protocol == "" {
		protocol = "http"
	}
	// OData is plain HTTP (ADR-017 §4) — normalize at the mapper boundary.
	if protocol == "odata" {
		protocol = "http"
	}

	target, err := targetURL(req)
	if err != nil {
		return nil, err
	}
	request := map[string]any{
		"protocol":  protocol,
		"target":    target,
		"operation": method,
	}

	headers := enabledKVMap(req.Headers)
	// GraphQL requires Content-Type: application/json (ADR-017 §4).
	if req.Body != nil && strings.EqualFold(req.Body.Encoding, "graphql") {
		if headers == nil {
			headers = map[string]string{}
		}
		if _, hasContentType := headers["Content-Type"]; !hasContentType {
			headers["Content-Type"] = "application/json"
		}
	}
	if len(headers) > 0 {
		request["headers"] = headers
	}

	body, bodyErr := requestBody(req.Body)
	if bodyErr != nil {
		return nil, bodyErr
	}
	if body != nil {
		request["body"] = body
	}
	step := map[string]any{"id": stepID(req), "kind": "request", "request": request}
	if len(req.Assertions) > 0 {
		step["assertions"] = assertionDocs(req.Assertions)
	}
	if len(req.Extracts) > 0 {
		step["extracts"] = extractDocs(req.Extracts)
	}
	if req.PreScript != "" || req.PostScript != "" {
		step["hooks"] = hookDoc(req)
	}
	if req.Settings.TimeoutMs > 0 {
		step["timeout"] = fmt.Sprintf("PT%dS", max(1, req.Settings.TimeoutMs/1000))
	}
	if retry := retryDoc(req.Settings.Retry); retry != nil {
		step["retry"] = retry
	}
	failurePolicy := req.Settings.FailurePolicy
	if failurePolicy == "" {
		failurePolicy = "resilient"
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = target
	}
	return map[string]any{
		"schemaVersion": 1,
		"capabilities":  buildCapabilities(protocol),
		"metadata":      map[string]any{"name": name},
		"workflows": []map[string]any{{
			"id":            "request-flow",
			"name":          name,
			"failurePolicy": failurePolicy,
			"steps":         []map[string]any{step},
		}},
	}, nil
}

func targetURL(req RequestDef) (string, error) {
	target := strings.TrimSpace(req.URL)
	if target == "" {
		return "", errors.New("url required")
	}
	params := url.Values{}
	for _, row := range req.Params {
		if row.Enabled && strings.TrimSpace(row.Key) != "" {
			params.Add(strings.TrimSpace(row.Key), row.Value)
		}
	}
	if len(params) == 0 {
		return target, nil
	}
	separator := "?"
	if strings.Contains(target, "?") {
		separator = "&"
	}
	return target + separator + params.Encode(), nil
}

func enabledKVMap(rows []KVRow) map[string]string {
	out := map[string]string{}
	for _, row := range rows {
		key := strings.TrimSpace(row.Key)
		if row.Enabled && key != "" {
			out[key] = row.Value
		}
	}
	return out
}

func buildCapabilities(protocol string) []string {
	if protocol == "graphql" {
		return []string{"protocol.http", "protocol.graphql"}
	}
	return []string{"protocol.http"}
}

func requestBody(body *BodyDef) (map[string]any, error) {
	if body == nil || strings.EqualFold(body.Encoding, "none") {
		return nil, nil
	}
	enc := strings.ToLower(body.Encoding)
	doc := map[string]any{"encoding": enc}
	switch enc {
	case "json":
		if strings.TrimSpace(body.Content) == "" {
			return nil, nil
		}
		var decoded any
		if err := json.Unmarshal([]byte(body.Content), &decoded); err == nil {
			doc["content"] = decoded
			return doc, nil
		}
	case "graphql":
		// Content must be a JSON object with at least a "query" key (ADR-017 §4).
		var parsed map[string]any
		if err := json.Unmarshal([]byte(body.Content), &parsed); err != nil {
			return nil, fmt.Errorf("graphql body: content is not valid JSON: %w", err)
		}
		if _, ok := parsed["query"]; !ok {
			return nil, errors.New(`graphql body: required key "query" is missing`)
		}
		serialized, _ := json.Marshal(parsed)
		doc["content"] = string(serialized)
		return doc, nil
	case "form", "multipart", "raw", "binary":
		// pass content through as-is — engine handles serialization
	default:
		return nil, fmt.Errorf("unsupported body encoding %q; accepted: json, graphql, form, multipart, raw, binary, none", enc)
	}
	doc["content"] = body.Content
	return doc, nil
}

func assertionDocs(assertions []AssertionDef) []map[string]any {
	out := make([]map[string]any, 0, len(assertions))
	for i, assertion := range assertions {
		expected := any(assertion.Expected)
		if assertion.Kind == "status" {
			if status, err := strconv.Atoi(assertion.Expected); err == nil {
				expected = status
			}
		}
		out = append(out, map[string]any{
			"id":       fallbackID(assertion.ID, "assertion", i),
			"kind":     assertion.Kind,
			"op":       assertion.Op,
			"expected": expected,
			"severity": defaultString(assertion.Severity, "error"),
		})
	}
	return out
}

func extractDocs(extracts []ExtractDef) []map[string]any {
	out := make([]map[string]any, 0, len(extracts))
	for i, extract := range extracts {
		out = append(out, map[string]any{
			"id":     fallbackID(extract.ID, "extract", i),
			"source": extract.Source,
			"path":   extract.Path,
			"as":     defaultString(extract.VariableName, extract.ID),
			"scope":  defaultString(extract.Scope, "workflow"),
		})
	}
	return out
}

func hookDoc(req RequestDef) map[string]any {
	hooks := map[string]any{}
	if strings.TrimSpace(req.PreScript) != "" {
		hooks["beforeStep"] = map[string]string{"language": "js", "source": req.PreScript}
	}
	if strings.TrimSpace(req.PostScript) != "" {
		hooks["afterStep"] = map[string]string{"language": "js", "source": req.PostScript}
	}
	return hooks
}

func retryDoc(retry *RetryDef) map[string]any {
	if retry == nil || retry.MaxAttempts <= 0 {
		return nil
	}
	return map[string]any{
		"maxAttempts": retry.MaxAttempts,
		"backoff":     retry.Strategy,
		"delay":       fmt.Sprintf("PT%dS", max(1, retry.DelayMs/1000)),
	}
}

func stepID(req RequestDef) string {
	return defaultString(req.ID, "request-step")
}

func fallbackID(id, prefix string, index int) string {
	return defaultString(id, fmt.Sprintf("%s-%d", prefix, index+1))
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
