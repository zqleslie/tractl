package localapi

import "encoding/json"

// SaveFileRequest is the POST /api/v1/files body.
type SaveFileRequest struct {
	ID       string          `json:"id,omitempty"`
	Kind     string          `json:"kind"`
	Name     string          `json:"name"`
	Document json.RawMessage `json:"document"`
}

// SaveFileResponse is returned after persisting a request/workflow file.
type SaveFileResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	UpdatedAt string `json:"updatedAt"`
}

// RunRequest is the POST /api/v1/run body.
type RunRequest struct {
	FileID string `json:"fileId"`
	Env    string `json:"env,omitempty"`
}

// WorkflowRunRequest is the POST /api/v1/workflows/run body.
// Returns the same engine.RunResult JSON shape as the browser WASM bridge.
type WorkflowRunRequest struct {
	Document string `json:"document"`
	Format   string `json:"format,omitempty"`
	Env      string `json:"env,omitempty"`
}

// StatusResponse is returned by GET /api/v1/status.
type StatusResponse struct {
	OK      bool   `json:"ok"`
	Version string `json:"version"`
}

// TimelineSegment is one segment in the request timeline UI.
type TimelineSegment struct {
	Label string `json:"label"`
	Ms    int64  `json:"ms"`
}

// AssertionResult is a UI-friendly assertion outcome row.
type AssertionResult struct {
	ID     string `json:"id"`
	Passed bool   `json:"passed"`
	Label  string `json:"label"`
	Detail string `json:"detail"`
}

// ExtractResult is a UI-friendly extract outcome row.
type ExtractResult struct {
	Variable string `json:"variable"`
	Value    string `json:"value"`
	Scope    string `json:"scope"`
}

// HeaderRow mirrors the frontend key-value table row shape.
type HeaderRow struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
	Key     string `json:"key"`
	Value   string `json:"value"`
}

// RunResponse is returned by POST /api/v1/run for the request editor UI.
type RunResponse struct {
	Passed           bool              `json:"passed"`
	DurationMs       int64             `json:"durationMs"`
	StatusCode       int               `json:"statusCode"`
	StatusLabel      string            `json:"statusLabel"`
	ContentType      string            `json:"contentType"`
	Body             string            `json:"body"`
	Headers          []HeaderRow       `json:"headers"`
	AssertionResults []AssertionResult `json:"assertionResults"`
	ExtractResults   []ExtractResult   `json:"extractResults"`
	PassedCount      int               `json:"passedCount"`
	TotalCount       int               `json:"totalCount"`
	Timeline         []TimelineSegment `json:"timeline"`
	Error            string            `json:"error,omitempty"`
}

// ErrorResponse is a generic API error payload.
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}
