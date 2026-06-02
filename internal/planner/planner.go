package planner

import (
	"crypto/rand"
	"errors"
	"slices"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/tractl/tractl/internal/spec"
)

// Planner orchestrates execution planning for a canonical traCtlSpec.
// It is responsible for:
//   - topological sorting of steps using DependsOn edges (all edges pre-merged by Phase 3.5)
//   - composite cycle detection via recursive DAG flattening (spec §16.3)
//   - capability resolution
//   - runtime selection
//   - concurrency planning
//   - failure policy planning
//   - planning trace event emission (ADR-003 §8)
//   - surface-specific transport capability guards (ADR-017 §7)
type Planner struct {
	registry CapabilityRegistry
	surface  string
}

// NewPlanner constructs a new Planner with the given capability registry.
// The surface defaults to "cli". Panics if registry is nil.
func NewPlanner(registry CapabilityRegistry) *Planner {
	if registry == nil {
		panic("planner: capability registry cannot be nil")
	}
	return &Planner{registry: registry, surface: "cli"}
}

// NewPlannerWithConfig constructs a Planner with an explicit Config.
// Use this to set the delivery surface for capability guards (ADR-017 §7).
// Panics if registry is nil.
func NewPlannerWithConfig(registry CapabilityRegistry, cfg Config) *Planner {
	if registry == nil {
		panic("planner: capability registry cannot be nil")
	}
	surface := cfg.Surface
	if surface == "" {
		surface = "cli"
	}
	return &Planner{registry: registry, surface: surface}
}

// Plan generates an ExecutionPlan for a traCtlSpec.
// It performs all planning steps: topological sort of workflows and steps,
// capability resolution, concurrency planning, and failure policy planning.
// All errors across all workflows are collected before returning (no fail-fast).
func (p *Planner) Plan(s *spec.TraCtlSpec) (*ExecutionPlan, error) {
	// Guard: nil or empty spec
	if s == nil || len(s.Workflows) == 0 {
		return nil, plannerErr(ErrEmptySpec, "", "", "spec is nil or has no workflows")
	}

	// Detect composite cycles early
	if err := detectCompositeCycles(s); err != nil {
		return nil, err
	}

	// Validate workflow-level dependsOn references before topo-sorting.
	wfIDs := make(map[string]struct{}, len(s.Workflows))
	for _, wf := range s.Workflows {
		wfIDs[wf.ID] = struct{}{}
	}
	var errs []error
	for _, wf := range s.Workflows {
		for _, dep := range wf.DependsOn {
			if _, ok := wfIDs[dep]; !ok {
				errs = append(errs, plannerErr(ErrCyclicDependency, wf.ID, "",
					"workflow dependsOn references unknown workflow: "+dep))
			}
		}
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	// Topo-sort workflows so dependent workflows appear after their prerequisites.
	sortedWorkflows, topoErr := topoSortWorkflows(s.Workflows)
	if topoErr != nil {
		return nil, topoErr
	}

	// Resolve workflow-level concurrency cap.
	workflowConcurrency := s.WorkflowConcurrency
	if workflowConcurrency <= 0 {
		workflowConcurrency = DefaultWorkflowConcurrency
	}

	// Generate plan ID (ULID for uniqueness and sorting)
	planID := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String()

	// Plan each workflow, collecting all errors before returning.
	workflowPlans := []WorkflowPlan{}
	totalSteps := 0

	for _, workflow := range sortedWorkflows {
		// Topologically sort steps
		sortedSteps, err := topoSort(workflow.ID, workflow.Steps)
		if err != nil {
			errs = append(errs, err)
		}

		// Resolve concurrency limit: user overrides via concurrency:, else DefaultStepConcurrency.
		concurrencyLimit := workflow.Concurrency
		if concurrencyLimit <= 0 {
			concurrencyLimit = DefaultStepConcurrency
		}

		// Resolve failure policy (default: resilient per spec §6.3)
		failurePolicy := workflow.FailurePolicy
		if failurePolicy == "" {
			failurePolicy = spec.DefaultFailurePolicy
		}

		// Plan each step
		stepPlans := []StepPlan{}
		for stepIndex, step := range sortedSteps {
			// Map step kind to capability contract
			contract, contractErr := kindToContract(step.Kind)
			if contractErr != nil {
				errs = append(errs, contractErr)
			}

			// Resolve capability
			var stepRuntime string
			if contractErr == nil {
				resolved, resolveErr := p.registry.Resolve(contract)
				if resolveErr != nil {
					errs = append(errs, resolveErr)
				} else if resolved != nil {
					stepRuntime = resolved.Runtime
				}
			}

			// Transport capability guard — must run after capability resolution (ADR-017 §7).
			if transportErr := p.checkTransport(workflow.ID, step.ID, step); transportErr != nil {
				errs = append(errs, transportErr)
			}

			stepPlan := StepPlan{
				StepID:          step.ID,
				WorkflowID:      workflow.ID,
				Runtime:         stepRuntime,
				FailurePolicy:   string(failurePolicy),
				DependsOn:       slices.Clone(step.DependsOn),
				ConcurrencySlot: (stepIndex % concurrencyLimit) + 1,
				CanSkip:         len(step.DependsOn) > 0,
			}
			stepPlans = append(stepPlans, stepPlan)
		}

		totalSteps += len(stepPlans)

		workflowPlan := WorkflowPlan{
			WorkflowID:       workflow.ID,
			DependsOn:        slices.Clone(workflow.DependsOn),
			Steps:            stepPlans,
			ConcurrencyLimit: concurrencyLimit,
			FailurePolicy:    string(failurePolicy),
		}
		workflowPlans = append(workflowPlans, workflowPlan)
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	// Build execution plan with trace
	plan := &ExecutionPlan{
		PlanID:              planID,
		WorkflowConcurrency: workflowConcurrency,
		Workflows:           workflowPlans,
		Trace: PlanTrace{
			PlanID:        planID,
			WorkflowCount: len(workflowPlans),
			TotalSteps:    totalSteps,
			GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		},
	}

	return plan, nil
}

// PlanWorkflow generates an ExecutionPlan for a single workflow by ID.
// It runs composite cycle detection against the full spec (all workflows),
// then plans only the requested workflow. Returns an error if the workflow
// ID is not found in the spec.
func (p *Planner) PlanWorkflow(s *spec.TraCtlSpec, workflowID string) (*ExecutionPlan, error) {
	if s == nil || len(s.Workflows) == 0 {
		return nil, plannerErr(ErrEmptySpec, "", "", "spec is nil or has no workflows")
	}

	// Composite cycle detection must run over the full spec.
	if err := detectCompositeCycles(s); err != nil {
		return nil, err
	}

	var targetWorkflow *spec.Workflow
	for i := range s.Workflows {
		if s.Workflows[i].ID == workflowID {
			targetWorkflow = &s.Workflows[i]
			break
		}
	}
	if targetWorkflow == nil {
		return nil, plannerErr("WORKFLOW_NOT_FOUND", workflowID, "", "workflow not found in spec")
	}

	planID := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String()

	concurrencyLimit := targetWorkflow.Concurrency
	if concurrencyLimit <= 0 {
		concurrencyLimit = DefaultStepConcurrency
	}
	failurePolicy := targetWorkflow.FailurePolicy
	if failurePolicy == "" {
		failurePolicy = spec.DefaultFailurePolicy
	}

	sortedSteps, topoErr := topoSort(targetWorkflow.ID, targetWorkflow.Steps)

	var errs []error
	if topoErr != nil {
		errs = append(errs, topoErr)
	}

	stepPlans := make([]StepPlan, 0, len(sortedSteps))
	for stepIndex, step := range sortedSteps {
		contract, contractErr := kindToContract(step.Kind)
		if contractErr != nil {
			errs = append(errs, contractErr)
		}
		var stepRuntime string
		if contractErr == nil {
			resolved, resolveErr := p.registry.Resolve(contract)
			if resolveErr != nil {
				errs = append(errs, resolveErr)
			} else if resolved != nil {
				stepRuntime = resolved.Runtime
			}
		}

		// Transport capability guard — must run after capability resolution (ADR-017 §7).
		if transportErr := p.checkTransport(targetWorkflow.ID, step.ID, step); transportErr != nil {
			errs = append(errs, transportErr)
		}

		stepPlans = append(stepPlans, StepPlan{
			StepID:          step.ID,
			WorkflowID:      targetWorkflow.ID,
			Runtime:         stepRuntime,
			FailurePolicy:   string(failurePolicy),
			DependsOn:       slices.Clone(step.DependsOn),
			ConcurrencySlot: (stepIndex % concurrencyLimit) + 1,
			CanSkip:         len(step.DependsOn) > 0,
		})
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	wp := WorkflowPlan{
		WorkflowID:       targetWorkflow.ID,
		Steps:            stepPlans,
		ConcurrencyLimit: concurrencyLimit,
		FailurePolicy:    string(failurePolicy),
	}
	plan := &ExecutionPlan{
		PlanID:    planID,
		Workflows: []WorkflowPlan{wp},
		Trace: PlanTrace{
			PlanID:        planID,
			WorkflowCount: 1,
			TotalSteps:    len(stepPlans),
			GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		},
	}
	return plan, nil
}

// checkTransport enforces surface-specific transport capability guards (ADR-017 §7).
// It returns a planning error when transport.version "h3" is declared on any
// surface other than "wasm". HTTP/3 is only supported on the WASM surface via
// the browser's fetch() API.
func (p *Planner) checkTransport(wfID, stepID string, step spec.Step) error {
	if step.Request == nil {
		return nil
	}
	version, ok := step.Request.Transport["version"]
	if !ok {
		return nil
	}
	if version != "h3" {
		return nil
	}
	if p.surface == "wasm" {
		return nil
	}
	return plannerErr(ErrTransportNotSupported, wfID, stepID,
		`transport.version "h3" is not supported on the `+p.surface+` surface. `+
			`HTTP/3 is available on the Web WASM surface via browser fetch only. `+
			`CLI, Desktop, and Local API surfaces support HTTP/1.1 and HTTP/2 only.`)
}

// kindToContract maps a step kind to its capability contract.
// Returns ErrUnknownCapability for unsupported kinds.
func kindToContract(kind string) (string, error) {
	switch kind {
	case "request":
		return "protocol.http", nil
	case "script", "extension", "composite":
		// Phase 4+ these will be supported; for now they're unknown
		return "", plannerErr(ErrUnknownCapability, "", "", "unsupported step kind: "+kind)
	default:
		return "", plannerErr(ErrUnknownCapability, "", "", "unknown step kind: "+kind)
	}
}
