package compiler

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/tractl/tractl/internal/planner"
	"github.com/tractl/tractl/internal/spec"
)

// Compiler converts a planner.ExecutionPlan into a CompiledPlan.
type Compiler struct{}

// NewCompiler returns a new Compiler.
func NewCompiler() *Compiler {
	return &Compiler{}
}

// Compile validates the plan structurally and builds the CompiledPlan.
// It collects all errors before returning; it does not fail fast.
// The input plan is never mutated.
// NOTE: spec is used only to copy step payloads into the compiled plan.
func (c *Compiler) Compile(plan *planner.ExecutionPlan, s *spec.TraCtlSpec) (*CompiledPlan, error) {
	if plan == nil {
		return nil, compilerErr(ErrNilPlan, "", "", "nil ExecutionPlan received")
	}
	if len(plan.Workflows) == 0 {
		return nil, compilerErr(ErrEmptyPlan, "", "", "ExecutionPlan has no workflows")
	}

	var errs []error
	for _, wp := range plan.Workflows {
		if wp.WorkflowID == "" {
			errs = append(errs, compilerErr(ErrEmptyWorkflowID, "", "", "WorkflowPlan has empty WorkflowID"))
			continue
		}
		for _, sp := range wp.Steps {
			if sp.StepID == "" {
				errs = append(errs, compilerErr(ErrEmptyStepID, wp.WorkflowID, "", "StepPlan has empty StepID"))
			}
			if sp.Runtime == "" {
				errs = append(errs, compilerErr(ErrEmptyRuntime, wp.WorkflowID, sp.StepID, "StepPlan has empty Runtime"))
			}
		}
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	compiled := &CompiledPlan{
		PlanID:              plan.PlanID,
		CompiledAt:          time.Now().UTC().Format(time.RFC3339),
		WorkflowConcurrency: plan.WorkflowConcurrency,
		Workflows:           make([]CompiledWorkflow, 0, len(plan.Workflows)),
	}

	// helper to find spec step
	findSpecStep := func(workflowID, stepID string) *spec.Step {
		if s == nil {
			return nil
		}
		for _, wf := range s.Workflows {
			if wf.ID != workflowID {
				continue
			}
			for _, st := range wf.Steps {
				if st.ID == stepID {
					return &st
				}
			}
		}
		return nil
	}

	for _, wp := range plan.Workflows {
		cw := CompiledWorkflow{
			WorkflowID:       wp.WorkflowID,
			DependsOn:        append([]string{}, wp.DependsOn...),
			ConcurrencyLimit: wp.ConcurrencyLimit,
			FailurePolicy:    wp.FailurePolicy,
			Steps:            make([]CompiledStep, 0, len(wp.Steps)),
		}
		for _, sp := range wp.Steps {
			cs := CompiledStep{
				StepID:          sp.StepID,
				WorkflowID:      sp.WorkflowID,
				Runtime:         sp.Runtime,
				FailurePolicy:   sp.FailurePolicy,
				DependsOn:       append([]string{}, sp.DependsOn...),
				ConcurrencySlot: sp.ConcurrencySlot,
				CanSkip:         sp.CanSkip,
			}

			// populate payload fields from spec.Step
			sStep := findSpecStep(sp.WorkflowID, sp.StepID)
			if sStep == nil {
				// Spec step missing: this is a compilation error per Phase 5 requirements
				return nil, compilerErr(ErrMissingStepPayload, sp.WorkflowID, sp.StepID, fmt.Sprintf("missing payload for step %s in workflow %s", sp.StepID, sp.WorkflowID))
			}

			cs.When = sStep.When
			cs.Timeout = sStep.Timeout

			if sStep.Request != nil {
				cr := &CompiledRequest{
					Protocol: sStep.Request.Protocol,
					Target:   sStep.Request.Target,
					Method:   sStep.Request.Operation,
					Headers:  nil,
					Body:     nil,
				}
				if len(sStep.Request.Headers) > 0 {
					cr.Headers = make(map[string]string, len(sStep.Request.Headers))
					for k, v := range sStep.Request.Headers {
						cr.Headers[k] = v
					}
				}
				if sStep.Request.Body != nil {
					cr.Body = &CompiledBody{Encoding: sStep.Request.Body.Encoding, Content: sStep.Request.Body.Content}
				}
				cs.Request = cr
			}

			if sStep.Retry != nil {
				cs.Retry = &CompiledRetry{
					MaxAttempts: sStep.Retry.MaxAttempts,
					Backoff:     sStep.Retry.Backoff,
					Delay:       sStep.Retry.Delay,
					RetryOn:     append([]string{}, sStep.Retry.RetryOn...),
				}
			}

			// assertions
			if len(sStep.Assertions) > 0 {
				cs.Assertions = make([]CompiledAssertion, 0, len(sStep.Assertions))
				for _, a := range sStep.Assertions {
					exp := ""
					if a.Expected != nil {
						exp = fmt.Sprintf("%v", a.Expected)
					}
					ca := CompiledAssertion{
						ID:       a.ID,
						Kind:     a.Kind,
						Target:   a.Target,
						Op:       a.Op,
						Expected: exp,
						Severity: string(a.Severity),
					}
					if a.Op == "matches" && exp != "" {
						if rx, err := regexp.Compile(exp); err == nil {
							ca.CompiledPattern = rx
						}
					}
					cs.Assertions = append(cs.Assertions, ca)
				}
			}

			// extracts
			if len(sStep.Extracts) > 0 {
				cs.Extracts = make([]CompiledExtract, 0, len(sStep.Extracts))
				for _, ex := range sStep.Extracts {
					as := ex.As
					if as == "" {
						as = ex.ID
					}
					scopeStr := string(ex.Scope)
					if scopeStr == "" {
						scopeStr = string(spec.ExtractScopeWorkflow)
					}
					cs.Extracts = append(cs.Extracts, CompiledExtract{
						ID:     ex.ID,
						Source: ex.Source,
						Path:   ex.Path,
						As:     as,
						Scope:  scopeStr,
					})
				}
			}

			cw.Steps = append(cw.Steps, cs)
		}
		compiled.Workflows = append(compiled.Workflows, cw)
	}

	return compiled, nil
}
