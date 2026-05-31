//go:build e2e

// Package e2e_test runs the full traCtl canonical pipeline end-to-end with
// real HTTP via httptest.NewServer. No mocks, no stubs.
//
// Run with: go test ./test/e2e/... -tags e2e -timeout 120s
package e2e_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tractl/tractl/internal/compiler"
	"github.com/tractl/tractl/internal/executor"
	"github.com/tractl/tractl/internal/normalizer"
	jsonparser "github.com/tractl/tractl/internal/parser/json"
	yamlparser "github.com/tractl/tractl/internal/parser/yaml"
	"github.com/tractl/tractl/internal/planner"
	"github.com/tractl/tractl/internal/runtime"
	"github.com/tractl/tractl/internal/scheduler"
	"github.com/tractl/tractl/internal/spec"
	"github.com/tractl/tractl/internal/validation"
)

// runWorkflow runs the complete canonical pipeline for a YAML input string
// and returns the resulting WorkflowResult.
func runWorkflow(t *testing.T, yamlInput, workflowID string) *scheduler.WorkflowResult {
	t.Helper()
	s, err := yamlparser.Parse([]byte(yamlInput), "e2e.yaml")
	if err != nil {
		t.Fatalf("yaml parse: %v", err)
	}
	return runSpec(t, s, workflowID)
}

func runWorkflowJSON(t *testing.T, jsonInput, workflowID string) *scheduler.WorkflowResult {
	t.Helper()
	s, err := jsonparser.Parse([]byte(jsonInput), "e2e.json")
	if err != nil {
		t.Fatalf("json parse: %v", err)
	}
	return runSpec(t, s, workflowID)
}

func runSpec(t *testing.T, s *spec.TraCtlSpec, workflowID string) *scheduler.WorkflowResult {
	t.Helper()
	norm, err := normalizer.Normalize(s)
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if errs := validation.NewSpecValidator().Validate(norm); len(errs) > 0 {
		t.Fatalf("validate: %v", errs[0])
	}
	plan, err := planner.NewPlanner(planner.DefaultRegistry()).Plan(norm)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	cp, err := compiler.NewCompiler().Compile(plan, norm)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	exec := executor.NewHTTPExecutor(nil)
	sched := scheduler.NewScheduler(cp, exec)
	res, err := sched.Run(context.Background(), workflowID)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	return res
}

func findOutcome(r *scheduler.WorkflowResult, stepID string) *scheduler.StepOutcome {
	for i := range r.Outcomes {
		if r.Outcomes[i].StepID == stepID {
			return &r.Outcomes[i]
		}
	}
	return nil
}

// TestE2E_LoginThenAuthenticatedFetch — login extracts a token; fetch-profile
// uses the token via implicit dep on login.
func TestE2E_LoginThenAuthenticatedFetch(t *testing.T) {
	var profileAuth string
	loginServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Auth-Token", "test-token-abc")
		w.WriteHeader(http.StatusOK)
	}))
	defer loginServer.Close()

	profileServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		profileAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer profileServer.Close()

	yamlInput := fmt.Sprintf(`
schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: auth-workflow
    steps:
      - id: login
        kind: request
        request:
          protocol: http
          target: %s
          operation: POST
          headers:
            Content-Type: application/json
          body:
            encoding: json
            content:
              username: testuser
              password: testpass
        extracts:
          - id: token-extract
            source: header
            path: X-Auth-Token
            as: authToken
            scope: workflow
      - id: fetch-profile
        kind: request
        request:
          protocol: http
          target: %s
          operation: GET
          headers:
            Authorization: "Bearer ${steps.login.extracts.authToken}"
        assertions:
          - id: status-ok
            kind: status
            op: equals
            expected: "200"
            severity: error
`, loginServer.URL, profileServer.URL)
	res := runWorkflow(t, yamlInput, "auth-workflow")

	login := findOutcome(res, "login")
	fetch := findOutcome(res, "fetch-profile")
	if login == nil || login.State != runtime.StateSucceeded {
		t.Fatalf("login failed: %+v", login)
	}
	if fetch == nil || fetch.State != runtime.StateSucceeded {
		t.Fatalf("fetch-profile failed: %+v", fetch)
	}
	if profileAuth != "Bearer test-token-abc" {
		t.Fatalf("profile server saw Authorization=%q", profileAuth)
	}
	// Order: login finished before fetch started.
	if !login.Result.FinishedAt.Before(fetch.Result.StartedAt.Add(time.Microsecond)) {
		t.Fatalf("login should finish before fetch-profile starts")
	}
	if res.OverallState != runtime.StateSucceeded {
		t.Fatalf("overall: expected succeeded, got %s", res.OverallState)
	}
}

// TestE2E_DiamondDAG — 4-step diamond.
func TestE2E_DiamondDAG(t *testing.T) {
	hits := make(map[string]*int32, 4)
	for _, n := range []string{"auth", "user", "orders", "merge"} {
		var v int32
		hits[n] = &v
	}
	mk := func(name string) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(hits[name], 1)
			w.WriteHeader(http.StatusOK)
		}))
	}
	authS := mk("auth")
	defer authS.Close()
	userS := mk("user")
	defer userS.Close()
	ordersS := mk("orders")
	defer ordersS.Close()
	mergeS := mk("merge")
	defer mergeS.Close()

	yamlInput := fmt.Sprintf(`
schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: diamond
    concurrency: 3
    steps:
      - id: auth
        kind: request
        request:
          protocol: http
          target: %s
      - id: fetch-user
        kind: request
        dependsOn:
          - auth
        request:
          protocol: http
          target: %s
      - id: fetch-orders
        kind: request
        dependsOn:
          - auth
        request:
          protocol: http
          target: %s
      - id: merge-results
        kind: request
        dependsOn:
          - fetch-user
          - fetch-orders
        request:
          protocol: http
          target: %s
`, authS.URL, userS.URL, ordersS.URL, mergeS.URL)

	res := runWorkflow(t, yamlInput, "diamond")
	if res.OverallState != runtime.StateSucceeded {
		t.Fatalf("expected succeeded, got %s", res.OverallState)
	}
	for n, v := range hits {
		if atomic.LoadInt32(v) != 1 {
			t.Fatalf("%s: expected 1 hit, got %d", n, atomic.LoadInt32(v))
		}
	}
	// merge-results must execute last.
	merge := findOutcome(res, "merge-results")
	user := findOutcome(res, "fetch-user")
	orders := findOutcome(res, "fetch-orders")
	if !merge.Result.StartedAt.After(user.Result.FinishedAt) || !merge.Result.StartedAt.After(orders.Result.FinishedAt) {
		t.Fatalf("merge should start after both user and orders finish")
	}
}

// TestE2E_ResilientFailure — auth fails; fetch is dependency-skipped; health
// (independent) succeeds; overall failed.
func TestE2E_ResilientFailure(t *testing.T) {
	auth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer auth.Close()
	fetch := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("fetch should not be called")
		w.WriteHeader(http.StatusOK)
	}))
	defer fetch.Close()
	health := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer health.Close()

	yamlInput := fmt.Sprintf(`
schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wf
    failurePolicy: resilient
    concurrency: 5
    steps:
      - id: auth
        kind: request
        request:
          protocol: http
          target: %s
        assertions:
          - id: a1
            kind: status
            op: equals
            expected: "200"
            severity: error
      - id: fetch
        kind: request
        dependsOn:
          - auth
        request:
          protocol: http
          target: %s
      - id: health
        kind: request
        request:
          protocol: http
          target: %s
`, auth.URL, fetch.URL, health.URL)
	res := runWorkflow(t, yamlInput, "wf")

	au := findOutcome(res, "auth")
	fe := findOutcome(res, "fetch")
	he := findOutcome(res, "health")
	if au == nil || au.State != runtime.StateFailed {
		t.Fatalf("auth: expected failed, got %v", au)
	}
	if fe == nil || fe.State != runtime.StateDependencySkipped {
		t.Fatalf("fetch: expected dependency-skipped, got %v", fe)
	}
	if he == nil || he.State != runtime.StateSucceeded {
		t.Fatalf("health: expected succeeded (independent branch), got %v", he)
	}
	if res.OverallState != runtime.StateFailed {
		t.Fatalf("overall: expected failed, got %s", res.OverallState)
	}
}

// TestE2E_ConditionalSkip — optional-step has when=false; always-runs continues.
func TestE2E_ConditionalSkip(t *testing.T) {
	login := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer login.Close()
	optional := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("optional should not be called")
		w.WriteHeader(http.StatusOK)
	}))
	defer optional.Close()
	always := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer always.Close()

	yamlInput := fmt.Sprintf(`
schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wf
    concurrency: 5
    steps:
      - id: login
        kind: request
        request:
          protocol: http
          target: %s
      - id: optional-step
        kind: request
        when: "false"
        dependsOn:
          - login
        request:
          protocol: http
          target: %s
      - id: always-runs
        kind: request
        request:
          protocol: http
          target: %s
`, login.URL, optional.URL, always.URL)
	res := runWorkflow(t, yamlInput, "wf")

	if findOutcome(res, "login").State != runtime.StateSucceeded {
		t.Fatalf("login should succeed")
	}
	if findOutcome(res, "optional-step").State != runtime.StateConditionalSkip {
		t.Fatalf("optional-step should be conditional-skip, got %s", findOutcome(res, "optional-step").State)
	}
	if findOutcome(res, "always-runs").State != runtime.StateSucceeded {
		t.Fatalf("always-runs should succeed")
	}
	if res.OverallState != runtime.StateSucceeded {
		t.Fatalf("overall: expected succeeded, got %s", res.OverallState)
	}
}

// TestE2E_RetryUntilSuccess — flaky-step returns 503 twice, then 200.
func TestE2E_RetryUntilSuccess(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	yamlInput := fmt.Sprintf(`
schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wf
    steps:
      - id: flaky-step
        kind: request
        request:
          protocol: http
          target: %s
        retry:
          maxAttempts: 3
          backoff: fixed
          delay: PT0S
          retryOn:
            - "503"
        assertions:
          - id: a1
            kind: status
            op: equals
            expected: "200"
            severity: error
`, srv.URL)
	res := runWorkflow(t, yamlInput, "wf")

	flaky := findOutcome(res, "flaky-step")
	if flaky == nil || flaky.State != runtime.StateSucceeded {
		t.Fatalf("flaky-step: expected succeeded, got %v", flaky)
	}
	if atomic.LoadInt32(&calls) != 3 {
		t.Fatalf("expected 3 server hits, got %d", atomic.LoadInt32(&calls))
	}
	if flaky.Result == nil || flaky.Result.Timeline == nil || flaky.Result.Timeline.TotalDuration <= 0 {
		t.Fatalf("expected non-zero timeline duration")
	}
}

// TestE2E_ExtractAndChain — create-resource extracts the Location header,
// fetch-resource consumes it.
func TestE2E_ExtractAndChain(t *testing.T) {
	var fetchURL string
	var srvMu sync.Mutex
	created := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srvMu.Lock()
		defer srvMu.Unlock()
		w.Header().Set("Location", fetchURL)
		w.WriteHeader(http.StatusCreated)
	}))
	defer created.Close()
	fetch := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer fetch.Close()
	srvMu.Lock()
	fetchURL = fetch.URL
	srvMu.Unlock()

	yamlInput := fmt.Sprintf(`
schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wf
    steps:
      - id: create-resource
        kind: request
        request:
          protocol: http
          target: %s
          operation: POST
        extracts:
          - id: e1
            source: header
            path: Location
            as: resourceUrl
            scope: workflow
      - id: fetch-resource
        kind: request
        request:
          protocol: http
          target: "${steps.create-resource.extracts.resourceUrl}"
        assertions:
          - id: a1
            kind: status
            op: equals
            expected: "200"
            severity: error
`, created.URL)
	res := runWorkflow(t, yamlInput, "wf")

	c := findOutcome(res, "create-resource")
	f := findOutcome(res, "fetch-resource")
	if c == nil || c.State != runtime.StateSucceeded {
		t.Fatalf("create-resource: %v", c)
	}
	if f == nil || f.State != runtime.StateSucceeded {
		t.Fatalf("fetch-resource: %v", f)
	}
	if res.OverallState != runtime.StateSucceeded {
		t.Fatalf("overall: expected succeeded, got %s", res.OverallState)
	}
}

// TestE2E_StepTimeoutEnforced — slow-step times out via PT0.05S.
func TestE2E_StepTimeoutEnforced(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	yamlInput := fmt.Sprintf(`
schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wf
    steps:
      - id: slow-step
        kind: request
        timeout: PT0.05S
        request:
          protocol: http
          target: %s
`, srv.URL)
	res := runWorkflow(t, yamlInput, "wf")

	slow := findOutcome(res, "slow-step")
	if slow == nil || (slow.State != runtime.StateFailed && slow.State != runtime.StateCancelled) {
		t.Fatalf("slow-step: expected failed/cancelled, got %v", slow)
	}
	if res.OverallState != runtime.StateFailed && res.OverallState != runtime.StateCancelled {
		t.Fatalf("overall: expected failed/cancelled, got %s", res.OverallState)
	}
}

// TestE2E_YAMLAndJSONProduceIdenticalResults — same workflow, different formats.
func TestE2E_YAMLAndJSONProduceIdenticalResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	yamlInput := fmt.Sprintf(`
schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wf
    steps:
      - id: a
        kind: request
        request:
          protocol: http
          target: %s
      - id: b
        kind: request
        dependsOn:
          - a
        request:
          protocol: http
          target: %s
`, srv.URL, srv.URL)
	jsonInput := fmt.Sprintf(`{
  "schemaVersion": 1,
  "capabilities": ["protocol.http"],
  "workflows": [
    {"id": "wf", "steps": [
      {"id": "a", "kind": "request", "request": {"protocol": "http", "target": "%s"}},
      {"id": "b", "kind": "request", "dependsOn": ["a"], "request": {"protocol": "http", "target": "%s"}}
    ]}
  ]
}`, srv.URL, srv.URL)

	resY := runWorkflow(t, yamlInput, "wf")
	resJ := runWorkflowJSON(t, jsonInput, "wf")

	if resY.OverallState != runtime.StateSucceeded || resJ.OverallState != runtime.StateSucceeded {
		t.Fatalf("expected both succeeded; yaml=%s json=%s", resY.OverallState, resJ.OverallState)
	}
	if len(resY.Outcomes) != len(resJ.Outcomes) {
		t.Fatalf("outcome count differs: yaml=%d json=%d", len(resY.Outcomes), len(resJ.Outcomes))
	}
	for i := range resY.Outcomes {
		if resY.Outcomes[i].StepID != resJ.Outcomes[i].StepID {
			t.Fatalf("outcome[%d] step differs: yaml=%s json=%s", i, resY.Outcomes[i].StepID, resJ.Outcomes[i].StepID)
		}
		if resY.Outcomes[i].State != resJ.Outcomes[i].State {
			t.Fatalf("outcome[%d] state differs: yaml=%s json=%s", i, resY.Outcomes[i].State, resJ.Outcomes[i].State)
		}
	}
}

// ensure strings import is used in some helper if needed elsewhere
var _ = strings.Builder{}
