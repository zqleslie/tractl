package localapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndRunRequestFile(t *testing.T) {
	root := t.TempDir()
	store, err := NewStore(root)
	if err != nil {
		t.Fatal(err)
	}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":1,"title":"ping"}`))
	}))
	t.Cleanup(upstream.Close)

	server := NewServer("127.0.0.1:0", store)
	ts := httptest.NewServer(server.handler)
	t.Cleanup(ts.Close)

	document := map[string]any{
		"schemaVersion": 1,
		"capabilities":  []string{"protocol.http"},
		"metadata":      map[string]string{"name": "ping"},
		"workflows": []map[string]any{
			{
				"id": "request-flow",
				"steps": []map[string]any{
					{
						"id":   "request-step",
						"kind": "request",
						"request": map[string]any{
							"protocol":  "http",
							"target":    upstream.URL,
							"operation": "GET",
						},
						"assertions": []map[string]any{
							{
								"id":       "status-ok",
								"kind":     "status",
								"op":       "equals",
								"expected": 200,
							},
						},
					},
				},
			},
		},
	}
	docBytes, _ := json.Marshal(document)

	saveBody, _ := json.Marshal(map[string]any{
		"kind":     "request",
		"name":     "ping",
		"document": json.RawMessage(docBytes),
	})
	saveReq, err := http.NewRequestWithContext(context.Background(), http.MethodPost, ts.URL+"/api/v1/files", bytes.NewReader(saveBody))
	if err != nil {
		t.Fatal(err)
	}
	saveReq.Header.Set("Content-Type", "application/json")
	saveResp, err := http.DefaultClient.Do(saveReq)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := saveResp.Body.Close(); err != nil {
			t.Errorf("close save response body: %v", err)
		}
	}()
	if saveResp.StatusCode != http.StatusOK {
		t.Fatalf("save status %d", saveResp.StatusCode)
	}

	var saved SaveFileResponse
	if err := json.NewDecoder(saveResp.Body).Decode(&saved); err != nil {
		t.Fatal(err)
	}
	if saved.ID == "" {
		t.Fatal("expected file id")
	}

	if _, err := os.Stat(filepath.Join(root, "requests", saved.ID+".yaml")); err != nil {
		t.Fatalf("yaml file missing: %v", err)
	}

	runBody, _ := json.Marshal(map[string]string{"fileId": saved.ID})
	runReq, err := http.NewRequestWithContext(context.Background(), http.MethodPost, ts.URL+"/api/v1/run", bytes.NewReader(runBody))
	if err != nil {
		t.Fatal(err)
	}
	runReq.Header.Set("Content-Type", "application/json")
	runResp, err := http.DefaultClient.Do(runReq)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := runResp.Body.Close(); err != nil {
			t.Errorf("close run response body: %v", err)
		}
	}()
	if runResp.StatusCode != http.StatusOK {
		t.Fatalf("run status %d", runResp.StatusCode)
	}

	var run RunResponse
	if err := json.NewDecoder(runResp.Body).Decode(&run); err != nil {
		t.Fatal(err)
	}
	if run.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", run.StatusCode)
	}
	if run.TotalCount != 1 {
		t.Fatalf("expected 1 assertion, got %d", run.TotalCount)
	}
}

func TestWorkflowRunDocument(t *testing.T) {
	root := t.TempDir()
	store, err := NewStore(root)
	if err != nil {
		t.Fatal(err)
	}

	server := NewServer("127.0.0.1:0", store)
	ts := httptest.NewServer(server.handler)
	t.Cleanup(ts.Close)

	runBody, _ := json.Marshal(WorkflowRunRequest{
		Document: "not: valid: yaml: [[[",
		Format:   "yaml",
	})
	runReq, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		ts.URL+"/api/v1/workflows/run",
		bytes.NewReader(runBody),
	)
	if err != nil {
		t.Fatal(err)
	}
	runReq.Header.Set("Content-Type", "application/json")
	runResp, err := http.DefaultClient.Do(runReq)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := runResp.Body.Close(); err != nil {
			t.Errorf("close workflow run response body: %v", err)
		}
	}()
	if runResp.StatusCode != http.StatusOK {
		t.Fatalf("workflow run status %d", runResp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(runResp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result["ParseError"] == "" && result["parseError"] == nil {
		t.Fatalf("expected parse error in engine result, got %#v", result)
	}
}
