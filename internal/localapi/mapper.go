package localapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/tractl/tractl/internal/assertion"
	"github.com/tractl/tractl/internal/engine"
)

// DefaultEngineSettings is the single source of truth for engine defaults.
// The same values are used when fields are omitted from a run request.
var DefaultEngineSettings = EngineDefaults{
	FailurePolicy:  "resilient",
	TimeoutMs:      30_000,
	Concurrency:    4,
	HTTPSuccessMin: 200,
	HTTPSuccessMax: 399,
	Retry: EngineRetryDefaults{
		MaxAttempts:   3,
		BackoffFactor: 2.0,
		Strategies:    []string{"none", "fixed", "linear", "exponential"},
	},
}

// MapRunResult maps an engine.RunResult to a localapi.RunResult.
// All JSON field names use lowercase tags to match the HTTP API contract.
func MapRunResult(result *engine.RunResult) (*RunResult, error) {
	if result.ParseError != "" {
		return nil, fmt.Errorf("parse error: %s", result.ParseError)
	}
	if result.ValidationError != "" {
		return nil, fmt.Errorf("validation error: %s", result.ValidationError)
	}
	if result.PlanError != "" {
		return nil, fmt.Errorf("plan error: %s", result.PlanError)
	}
	if len(result.Workflows) == 0 {
		return nil, fmt.Errorf("run produced no workflow outcomes")
	}

	wf := result.Workflows[0]
	if wf.Skipped {
		return nil, fmt.Errorf("workflow skipped due to dependency failure")
	}
	if len(wf.Steps) == 0 {
		return nil, fmt.Errorf("workflow produced no step outcomes")
	}

	step := wf.Steps[0]
	passedCount := 0
	assertionRows := make([]AssertionResult, 0, len(step.AssertionResults))
	for _, ar := range step.AssertionResults {
		passed := ar.Outcome == assertion.OutcomePass
		if passed {
			passedCount++
		}
		expected, received := parseAssertionMessage(ar.Message)
		severity := string(ar.Severity)
		if severity == "" {
			severity = "error"
		}
		assertionRows = append(assertionRows, AssertionResult{
			ID:       ar.AssertionID,
			Kind:     ar.Kind,
			Op:       "",
			Expected: expected,
			Received: received,
			Passed:   passed,
			Severity: severity,
		})
	}

	timing := timingFromDiagnostics(result, step)

	statusLabel := http.StatusText(step.ResponseStatus)
	if statusLabel == "" && step.ResponseStatus > 0 {
		statusLabel = fmt.Sprintf("%d", step.ResponseStatus)
	}
	if step.ResponseStatus > 0 {
		statusLabel = fmt.Sprintf("%d %s", step.ResponseStatus, statusLabel)
	}

	contentType := step.ResponseHeaders["content-type"]
	if contentType == "" {
		contentType = "application/json"
	}

	extractRows := extractResultsFromDiagnostics(result)
	passed := step.Error == "" && !assertion.StepFailed(step.AssertionResults)

	return &RunResult{
		StatusCode:       step.ResponseStatus,
		StatusText:       strings.TrimSpace(statusLabel),
		DurationMs:       timing.Total,
		Body:             step.ResponseBody,
		Headers:          responseHeaders(step.ResponseHeaders, contentType),
		Timing:           timing,
		AssertionResults: assertionRows,
		ExtractResults:   extractRows,
		AssertionsPassed: passedCount,
		AssertionsTotal:  len(step.AssertionResults),
		Passed:           passed,
		Error:            step.Error,
	}, nil
}

// RunRequestDef is the single implementation for request execution.
// Called by both the HTTP /api/v1/run handler and the WASM handleRunRequest.
// Expression resolution happens here — frontend always sends raw unresolved values.
func RunRequestDef(def RequestDef) (RunResult, error) {
	resolved := resolveRequestEnv(def)
	document, err := requestDefToDocument(resolved)
	if err != nil {
		return RunResult{}, err
	}
	raw, err := json.Marshal(document)
	if err != nil {
		return RunResult{}, err
	}
	result := engine.New().RunDocument(engine.DocumentConfig{
		Document:  string(raw),
		Format:    "json",
		SourceRef: "local-api-request",
		Quiet:     true,
		Verbose:   true,
	})
	mapped, err := MapRunResult(result)
	if err != nil {
		return RunResult{}, err
	}
	if len(mapped.ExtractResults) == 0 && len(def.Extracts) > 0 {
		mapped.ExtractResults = requestExtractResults(def.Extracts, mapped.Body)
	}
	return *mapped, nil
}

// ComputeWorkflowLayout computes topological rows, edges, groups, and a summary
// for a slice of step references using Kahn's algorithm.
func ComputeWorkflowLayout(steps []StepRef) WorkflowLayoutResponse {
	if len(steps) == 0 {
		return WorkflowLayoutResponse{
			Rows:            []LayoutRow{},
			Edges:           []LayoutEdge{},
			Groups:          []LayoutGroup{},
			TopologySummary: "0 steps",
		}
	}

	inDegree := make(map[string]int, len(steps))
	children := make(map[string][]string, len(steps))
	for _, s := range steps {
		if _, ok := inDegree[s.ID]; !ok {
			inDegree[s.ID] = 0
		}
		for _, dep := range s.DependsOn {
			inDegree[s.ID]++
			children[dep] = append(children[dep], s.ID)
		}
	}
	for k := range children {
		sort.Strings(children[k])
	}

	queue := make([]string, 0, len(steps))
	for _, s := range steps {
		if inDegree[s.ID] == 0 {
			queue = append(queue, s.ID)
		}
	}
	sort.Strings(queue)

	rows := []LayoutRow{}
	processed := make(map[string]bool, len(steps))
	for rowIndex := 0; len(queue) > 0; rowIndex++ {
		level := make([]string, len(queue))
		copy(level, queue)
		queue = queue[:0]
		rows = append(rows, LayoutRow{RowIndex: rowIndex, StepIDs: level})
		for _, id := range level {
			processed[id] = true
			for _, child := range children[id] {
				inDegree[child]--
				if inDegree[child] == 0 {
					queue = append(queue, child)
				}
			}
		}
		sort.Strings(queue)
	}

	if len(processed) != len(steps) {
		return WorkflowLayoutResponse{
			Rows:   []LayoutRow{},
			Edges:  []LayoutEdge{},
			Groups: []LayoutGroup{},
			Error:  "cyclic dependency detected",
		}
	}

	parentCount := make(map[string]int, len(steps))
	childCount := make(map[string]int, len(steps))
	for _, s := range steps {
		parentCount[s.ID] = len(s.DependsOn)
		for _, dep := range s.DependsOn {
			childCount[dep]++
		}
	}

	edges := []LayoutEdge{}
	for _, s := range steps {
		for _, dep := range s.DependsOn {
			kind := "sequential"
			if childCount[dep] > 1 {
				kind = "fanout"
			}
			if parentCount[s.ID] > 1 {
				kind = "merge"
			}
			edges = append(edges, LayoutEdge{From: dep, To: s.ID, Kind: kind})
		}
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}
		return edges[i].To < edges[j].To
	})

	// Connected components via union-find.
	ufParent := make(map[string]string, len(steps))
	for _, s := range steps {
		ufParent[s.ID] = s.ID
	}
	var ufFind func(string) string
	ufFind = func(x string) string {
		if ufParent[x] != x {
			ufParent[x] = ufFind(ufParent[x])
		}
		return ufParent[x]
	}
	for _, s := range steps {
		for _, dep := range s.DependsOn {
			ra, rb := ufFind(s.ID), ufFind(dep)
			if ra != rb {
				ufParent[ra] = rb
			}
		}
	}
	compMembers := make(map[string][]string)
	for _, s := range steps {
		root := ufFind(s.ID)
		compMembers[root] = append(compMembers[root], s.ID)
	}
	groupRoots := make([]string, 0, len(compMembers))
	for root := range compMembers {
		groupRoots = append(groupRoots, root)
	}
	sort.Strings(groupRoots)
	groups := make([]LayoutGroup, 0, len(compMembers))
	for gi, root := range groupRoots {
		members := compMembers[root]
		sort.Strings(members)
		rootStep := members[0]
		for _, m := range members {
			if parentCount[m] == 0 {
				rootStep = m
				break
			}
		}
		groups = append(groups, LayoutGroup{GroupIndex: gi, RootStepID: rootStep, StepIDs: members})
	}

	rootCount, fanoutCount, mergeCount := 0, 0, 0
	for _, s := range steps {
		if len(s.DependsOn) == 0 {
			rootCount++
		}
	}
	for id := range childCount {
		if childCount[id] > 1 {
			fanoutCount++
		}
	}
	for id := range parentCount {
		if parentCount[id] > 1 {
			mergeCount++
		}
	}
	summary := fmt.Sprintf("%d steps", len(steps))
	switch rootCount {
	case 1:
		summary += " · 1 root"
	default:
		if rootCount > 1 {
			summary += fmt.Sprintf(" · %d roots", rootCount)
		}
	}
	switch fanoutCount {
	case 1:
		summary += " · 1 fan-out"
	default:
		if fanoutCount > 1 {
			summary += fmt.Sprintf(" · %d fan-outs", fanoutCount)
		}
	}
	switch mergeCount {
	case 1:
		summary += " · 1 merge"
	default:
		if mergeCount > 1 {
			summary += fmt.Sprintf(" · %d merges", mergeCount)
		}
	}

	return WorkflowLayoutResponse{
		Rows:            rows,
		Edges:           edges,
		Groups:          groups,
		TopologySummary: summary,
	}
}

// stepRefRE matches ${steps.<id>} expressions — same pattern as internal/normalize/scan.go.
var stepRefRE = regexp.MustCompile(`\$\{steps\.([a-zA-Z][a-zA-Z0-9_-]*)\b`)

// InferImplicitDeps scans step fields for ${steps.x} references and returns each
// step with explicit dependsOn plus inferred implicitDependsOn.
func InferImplicitDeps(steps []StepScanDef) InferDepsResponse {
	knownIDs := make(map[string]struct{}, len(steps))
	for _, s := range steps {
		knownIDs[s.ID] = struct{}{}
	}
	result := make([]StepWithImplicitDeps, 0, len(steps))
	for _, s := range steps {
		refs := scanStepRefs(s.URL, s.Body, s.PreScript, s.PostScript)
		for _, v := range s.Headers {
			for ref := range scanStepRefs(v) {
				refs[ref] = struct{}{}
			}
		}
		explicitSet := make(map[string]struct{}, len(s.DependsOn))
		for _, dep := range s.DependsOn {
			explicitSet[dep] = struct{}{}
		}
		implicit := []string{}
		for ref := range refs {
			if ref == s.ID {
				continue
			}
			if _, ok := knownIDs[ref]; !ok {
				continue
			}
			if _, ok := explicitSet[ref]; !ok {
				implicit = append(implicit, ref)
			}
		}
		sort.Strings(implicit)
		dependsOn := s.DependsOn
		if dependsOn == nil {
			dependsOn = []string{}
		}
		result = append(result, StepWithImplicitDeps{
			ID:                s.ID,
			DependsOn:         dependsOn,
			ImplicitDependsOn: implicit,
		})
	}
	return InferDepsResponse{Steps: result}
}

// ExportWorkflow builds the canonical traCtlSpec from canvas workflow state
// and serializes it to YAML.
func ExportWorkflow(req WorkflowExportRequest) (WorkflowExportResponse, error) {
	steps := make([]map[string]any, 0, len(req.Steps))
	for _, s := range req.Steps {
		step := map[string]any{"id": s.ID, "kind": "request"}
		if len(s.DependsOn) > 0 {
			step["dependsOn"] = s.DependsOn
		}
		method := strings.ToUpper(strings.TrimSpace(s.Method))
		if method == "" {
			method = "GET"
		}
		request := map[string]any{
			"protocol":  "http",
			"target":    s.URL,
			"operation": method,
		}
		if len(s.Headers) > 0 {
			request["headers"] = s.Headers
		}
		if body := requestBody(s.Body); body != nil {
			request["body"] = body
		}
		step["request"] = request
		if len(s.Assertions) > 0 {
			step["assertions"] = assertionDocs(s.Assertions)
		}
		if len(s.Extracts) > 0 {
			step["extracts"] = extractDocs(s.Extracts)
		}
		if s.PreScript != "" || s.PostScript != "" {
			hooks := map[string]any{}
			if s.PreScript != "" {
				hooks["beforeStep"] = map[string]string{"language": "js", "source": s.PreScript}
			}
			if s.PostScript != "" {
				hooks["afterStep"] = map[string]string{"language": "js", "source": s.PostScript}
			}
			step["hooks"] = hooks
		}
		steps = append(steps, step)
	}

	wfName := defaultString(req.Name, req.ID)
	workflow := map[string]any{
		"id":    req.ID,
		"name":  wfName,
		"steps": steps,
	}
	if req.Concurrency > 0 {
		workflow["concurrency"] = req.Concurrency
	}
	if req.FailurePolicy != "" {
		workflow["failurePolicy"] = req.FailurePolicy
	}
	doc := map[string]any{
		"schemaVersion": 1,
		"capabilities":  []string{"protocol.http"},
		"metadata":      map[string]any{"name": wfName},
		"workflows":     []any{workflow},
	}
	if len(req.Variables) > 0 {
		doc["variables"] = req.Variables
	}
	yamlBytes, err := yaml.Marshal(doc)
	if err != nil {
		return WorkflowExportResponse{Error: err.Error()}, err
	}
	return WorkflowExportResponse{YAML: string(yamlBytes)}, nil
}

// --- private helpers ---

func responseHeaders(headers map[string]string, contentType string) map[string]string {
	out := make(map[string]string, len(headers)+1)
	for key, value := range headers {
		out[key] = value
	}
	if _, ok := out["content-type"]; !ok && contentType != "" {
		out["content-type"] = contentType
	}
	return out
}

func timingFromDiagnostics(result *engine.RunResult, step engine.StepOutcome) TimingResult {
	if result.Diagnostics != nil && len(result.Diagnostics.Workflows) > 0 {
		wf := result.Diagnostics.Workflows[0]
		if len(wf.Steps) > 0 && len(wf.Steps[0].Requests) > 0 {
			t := wf.Steps[0].Requests[0].Timeline
			return TimingResult{
				DNS:      t.DNSMs,
				TCP:      t.TCPMs,
				TLS:      t.TLSMs,
				TTFB:     t.TTFBMs,
				Transfer: t.TransferMs,
				Total:    t.TotalMs,
				Unit:     "ms",
			}
		}
	}
	totalMs := step.Duration.Milliseconds()
	if totalMs < 0 {
		totalMs = 0
	}
	return TimingResult{Total: totalMs, Unit: "ms"}
}

func extractResultsFromDiagnostics(result *engine.RunResult) []ExtractResult {
	if result.Diagnostics == nil || len(result.Diagnostics.Workflows) == 0 {
		return []ExtractResult{}
	}
	wf := result.Diagnostics.Workflows[0]
	if len(wf.Steps) == 0 {
		return []ExtractResult{}
	}
	rows := make([]ExtractResult, 0, len(wf.Steps[0].Extracts))
	for _, ex := range wf.Steps[0].Extracts {
		variable := ex.As
		if variable == "" {
			variable = ex.ID
		}
		rows = append(rows, ExtractResult{
			ID:            ex.ID,
			VariableName:  variable,
			Scope:         "workflow",
			ResolvedValue: ex.Source,
		})
	}
	return rows
}

// parseAssertionMessage splits a human-readable message like "expected 200, got 404"
// into separate expected and received strings.
func parseAssertionMessage(msg string) (expected, received string) {
	msg = strings.TrimSpace(msg)
	const expectPrefix = "expected "
	const gotInfix = ", got "
	if strings.HasPrefix(msg, expectPrefix) {
		rest := msg[len(expectPrefix):]
		if idx := strings.Index(rest, gotInfix); idx >= 0 {
			return rest[:idx], rest[idx+len(gotInfix):]
		}
	}
	return "", msg
}

func scanStepRefs(strs ...string) map[string]struct{} {
	out := make(map[string]struct{})
	for _, s := range strs {
		if s == "" {
			continue
		}
		for _, m := range stepRefRE.FindAllStringSubmatch(s, -1) {
			if len(m) >= 2 && m[1] != "" {
				out[m[1]] = struct{}{}
			}
		}
	}
	return out
}

func resolveRequestEnv(def RequestDef) RequestDef {
	if len(def.Env) == 0 {
		return def
	}
	def.URL = resolveEnv(def.URL, def.Env)
	resolved := make([]KVRow, len(def.Headers))
	for i, h := range def.Headers {
		h.Value = resolveEnv(h.Value, def.Env)
		resolved[i] = h
	}
	def.Headers = resolved
	if def.Body != nil {
		cp := *def.Body
		cp.Content = resolveEnv(cp.Content, def.Env)
		def.Body = &cp
	}
	return def
}

func resolveEnv(s string, env map[string]string) string {
	for k, v := range env {
		s = strings.ReplaceAll(s, "{{"+k+"}}", v)
		s = strings.ReplaceAll(s, "${vars."+k+"}", v)
	}
	return s
}
