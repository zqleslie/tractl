package localapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/tractl/tractl/internal/engine"
)

func (s *Server) handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req RequestDef
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	result, err := RunRequestDef(req)
	if err != nil {
		httpError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

func (s *Server) handleWorkflowRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req WorkflowRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if strings.TrimSpace(req.Document) == "" {
		httpError(w, http.StatusBadRequest, "document required")
		return
	}
	format := strings.ToLower(strings.TrimSpace(req.Format))
	if format == "" {
		format = "yaml"
	}

	result := engine.New().RunDocument(engine.DocumentConfig{
		Document:  req.Document,
		Format:    format,
		SourceRef: "local-api-workflow",
		EnvName:   req.Env,
		Quiet:     true,
		Verbose:   true,
	})
	jsonResponse(w, http.StatusOK, result)
}

func (s *Server) handleWorkflowLayout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var steps []StepRef
	if err := json.NewDecoder(r.Body).Decode(&steps); err != nil {
		httpError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, ComputeWorkflowLayout(steps))
}

func (s *Server) handleWorkflowInferDeps(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var steps []StepScanDef
	if err := json.NewDecoder(r.Body).Decode(&steps); err != nil {
		httpError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, InferImplicitDeps(steps))
}

func (s *Server) handleWorkflowExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req WorkflowExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	result, err := ExportWorkflow(req)
	if err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

func (s *Server) handleEngineDefaults(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	jsonResponse(w, http.StatusOK, DefaultEngineSettings)
}

func requestExtractResults(extracts []ExtractDef, body string) []ExtractResult {
	rows := make([]ExtractResult, 0, len(extracts))
	var decoded map[string]any
	_ = json.Unmarshal([]byte(body), &decoded)
	for _, extract := range extracts {
		value, errText := resolveBodyPath(decoded, extract.Path)
		rows = append(rows, ExtractResult{
			ID:            extract.ID,
			VariableName:  extract.VariableName,
			Scope:         defaultString(extract.Scope, "workflow"),
			ResolvedValue: value,
			Error:         errText,
		})
	}
	return rows
}

func resolveBodyPath(decoded map[string]any, path string) (string, string) {
	if decoded == nil {
		return "", "response body is not a JSON object"
	}
	key := strings.TrimPrefix(strings.TrimSpace(path), "$.")
	if key == "" || key == path {
		return "", "unsupported extract path"
	}
	value, ok := decoded[key]
	if !ok {
		return "", "path not found"
	}
	return strings.Trim(strings.ReplaceAll(strings.TrimSpace(toJSON(value)), "\n", ""), `"`), ""
}

func toJSON(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(raw)
}
