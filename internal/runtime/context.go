// context.go defines ExecutionContext, which tracks workflow run state and variables.

package runtime

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	ulid "github.com/oklog/ulid/v2"
)

// ulidEntropy is a process-level monotonic entropy source for ULID generation.
// Shared and reused across calls to preserve monotonicity within the same millisecond.
var (
	ulidEntropyMu sync.Mutex
	ulidEntropy   = ulid.Monotonic(rand.Reader, 0)
)

// ExecutionContext holds runtime state for a single execution.
type ExecutionContext struct {
	executionID  string
	workflowID   string
	startedAt    time.Time
	mu           sync.RWMutex
	stepStates   map[string]StepState
	stepResults  map[string]*StepResult
	vars         map[string]string // spec-level variables
	workflowVars map[string]string
	envVars      map[string]string
	activeStepID string // step ID while applying step-scoped hook mutations (SetVar step scope)
}

// Scope identifies one of the five variable scopes defined in traCtlSpec §13.1.
type Scope string

// Variable scope constants per traCtlSpec §13.1.
const (
	ScopeSpec        Scope = "spec"
	ScopeEnvironment Scope = "env"
	ScopeWorkflow    Scope = "workflow"
	ScopeStep        Scope = "step"
	ScopeRuntime     Scope = "runtime"
)

// FrozenContext is an immutable snapshot of an ExecutionContext, delivered to scripts
// and assertion evaluators (traCtlSpec §13.6). All scope maps are deep-copied at
// snapshot time so sandbox mutations cannot affect the live ExecutionContext.
type FrozenContext struct {
	Spec        map[string]string
	Environment map[string]string
	Workflow    map[string]string
	Step        map[string]string
	Runtime     map[string]string
}

// NewExecutionContext creates a new execution context with a ULID executionID.
func NewExecutionContext(workflowID string, specVars, workflowVars, envVars map[string]string) (*ExecutionContext, error) {
	ulidEntropyMu.Lock()
	id, err := ulid.New(ulid.Timestamp(time.Now()), ulidEntropy)
	ulidEntropyMu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("runtime: generate execution ID: %w", err)
	}
	execID := id.String()

	if specVars == nil {
		specVars = make(map[string]string)
	}
	if workflowVars == nil {
		workflowVars = make(map[string]string)
	}
	if envVars == nil {
		envVars = make(map[string]string)
	}

	return &ExecutionContext{
		executionID:  execID,
		workflowID:   workflowID,
		startedAt:    time.Now(),
		stepStates:   make(map[string]StepState),
		stepResults:  make(map[string]*StepResult),
		vars:         specVars,
		workflowVars: workflowVars,
		envVars:      envVars,
	}, nil
}

// SetStepState sets the state for a step (thread-safe).
func (e *ExecutionContext) SetStepState(stepID string, state StepState) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.stepStates[stepID] = state
}

// GetStepState returns the step state, defaulting to StatePending if absent.
func (e *ExecutionContext) GetStepState(stepID string) StepState {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if s, ok := e.stepStates[stepID]; ok {
		return s
	}
	return StatePending
}

// SetStepResult stores a StepResult (thread-safe).
func (e *ExecutionContext) SetStepResult(result *StepResult) {
	if result == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	// ensure extracts map exists
	if result.Extracts == nil {
		result.Extracts = make(map[string]string)
	}
	// Preserve hook- or script-preseeded extracts when the executor stores its result.
	if existing, ok := e.stepResults[result.StepID]; ok && existing != nil && len(existing.Extracts) > 0 {
		for k, v := range existing.Extracts {
			if _, present := result.Extracts[k]; !present {
				result.Extracts[k] = v
			}
		}
	}
	e.stepResults[result.StepID] = result
	// also mirror state
	e.stepStates[result.StepID] = result.State
}

// GetStepResult returns the StepResult and a bool indicating presence.
func (e *ExecutionContext) GetStepResult(stepID string) (*StepResult, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	r, ok := e.stepResults[stepID]
	return r, ok
}

// GetExtract reads an extract value from a step result.
func (e *ExecutionContext) GetExtract(stepID, extractAs string) (string, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if r, ok := e.stepResults[stepID]; ok && r != nil && r.Extracts != nil {
		v, found := r.Extracts[extractAs]
		return v, found
	}
	return "", false
}

// SetStepExtract writes an extract value into a step's StepResult (creates result if necessary).
func (e *ExecutionContext) SetStepExtract(stepID, extractAs, value string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.setStepExtractLocked(stepID, extractAs, value)
}

// setStepExtractLocked writes an extract; caller must hold e.mu.
func (e *ExecutionContext) setStepExtractLocked(stepID, extractAs, value string) {
	r, ok := e.stepResults[stepID]
	if !ok || r == nil {
		r = &StepResult{StepID: stepID, WorkflowID: e.workflowID, Extracts: make(map[string]string)}
		e.stepResults[stepID] = r
	}
	if r.Extracts == nil {
		r.Extracts = make(map[string]string)
	}
	r.Extracts[extractAs] = value
}

// SetActiveStepID pins the step for step-scoped SetVar during hook MutationSet.Apply.
func (e *ExecutionContext) SetActiveStepID(stepID string) {
	e.mu.Lock()
	e.activeStepID = stepID
	e.mu.Unlock()
}

// ClearActiveStepID releases the active step pin set by SetActiveStepID.
func (e *ExecutionContext) ClearActiveStepID() {
	e.SetActiveStepID("")
}

// ExecutionID returns the execution ULID.
func (e *ExecutionContext) ExecutionID() string {
	return e.executionID
}

// WorkflowID returns the workflow id for this context.
func (e *ExecutionContext) WorkflowID() string {
	return e.workflowID
}

// ElapsedSince returns monotonic duration since context start.
func (e *ExecutionContext) ElapsedSince() time.Duration {
	return time.Since(e.startedAt)
}

// StartedAt returns the wall-clock time when this context was created.
func (e *ExecutionContext) StartedAt() time.Time {
	return e.startedAt
}

// ResolveAll evaluates all ${...} placeholders in s against this context.
// It is the bridge used by assertion and extract layers.
func (e *ExecutionContext) ResolveAll(s string) (string, error) {
	resolver := NewVariableResolver(e)
	eval := NewExpressionEvaluator(resolver, e)
	// context.Background() used: this bridge has no request context.
	// TODO: thread context through when assertion/extract callers gain one.
	return eval.Evaluate(context.Background(), s)
}

// GetVar reads a named variable from the given scope.
// Supported scopes: "spec", "workflow", "env".
func (e *ExecutionContext) GetVar(scope, key string) (string, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	switch scope {
	case "spec":
		v, ok := e.vars[key]
		return v, ok
	case "workflow":
		v, ok := e.workflowVars[key]
		return v, ok
	case string(ScopeEnvironment):
		v, ok := e.envVars[key]
		return v, ok
	case string(ScopeRuntime):
		return "", false
	default:
		return "", false
	}
}

// SetVar writes a named variable into the given scope.
// Supported scopes: "spec", "workflow", "env", "step".
// Step scope writes the extract on the active step (SetActiveStepID) via SetStepExtract semantics.
func (e *ExecutionContext) SetVar(scope, key, value string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	switch scope {
	case "spec":
		e.vars[key] = value
	case "workflow":
		e.workflowVars[key] = value
	case string(ScopeEnvironment):
		e.envVars[key] = value
	case string(ScopeStep):
		if e.activeStepID == "" {
			return fmt.Errorf("runtime: step-scoped SetVar requires active step ID")
		}
		e.setStepExtractLocked(e.activeStepID, key, value)
	case string(ScopeRuntime):
		return fmt.Errorf("runtime: ScopeRuntime is read-only")
	default:
		return fmt.Errorf("runtime: unknown variable scope %q", scope)
	}
	return nil
}

// Freeze returns a read-only snapshot of the execution context's variable scopes
// for delivery to sandbox scripts via the frozen ctx object (spec §13.6).
// The returned FrozenContext is a deep copy; mutations inside the sandbox do not
// affect the live ExecutionContext.
func (e *ExecutionContext) Freeze() FrozenContext {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return FrozenContext{
		Spec:        copyMap(e.vars),
		Environment: copyMap(e.envVars),
		Workflow:    copyMap(e.workflowVars),
		Step:        map[string]string{},
		Runtime:     map[string]string{},
	}
}

func copyMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return map[string]string{}
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// IsEligible answers whether ALL dependencies are terminal (done).
func (e *ExecutionContext) IsEligible(_ string, dependsOn []string) bool {
	// no deps -> eligible
	if len(dependsOn) == 0 {
		return true
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, d := range dependsOn {
		st, ok := e.stepStates[d]
		if !ok {
			// not present -> considered pending
			return false
		}
		if !IsTerminal(st) {
			return false
		}
	}
	return true
}

// ShouldExecute answers whether the step should execute (all deps succeeded or conditional-skip).
// Precondition: IsEligible must have returned true for this step; behaviour is undefined otherwise.
func (e *ExecutionContext) ShouldExecute(_ string, dependsOn []string) bool {
	// if any dep is Failed, DependencySkipped, or Cancelled -> do not execute
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, d := range dependsOn {
		st := e.stepStates[d]
		switch st {
		case StateFailed, StateDependencySkipped, StateCancelled:
			return false
		case StatePending, StateRunning, StateSucceeded, StateConditionalSkip:
			// acceptable states — continue checking remaining deps
		}
	}
	return true
}
