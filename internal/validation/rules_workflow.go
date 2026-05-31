// rules_workflow.go defines validation rules for the workflows array: presence
// (Rule 3), workflow identifier format (Rule 4), duplicate workflow IDs
// (Rule 5), and dispatch to per-step and per-DAG validation.
package validation

import (
	"fmt"

	"github.com/tractl/tractl/internal/spec"
)

// validateWorkflowsPresent enforces Rule 3. Reference: tractl_spec.md §5.5
func validateWorkflowsPresent(s *spec.TraCtlSpec) []ValidationError {
	if len(s.Workflows) == 0 {
		return []ValidationError{{
			Field:   "workflows",
			Code:    MissingRequiredField,
			Message: "workflows is required and must contain at least one workflow",
		}}
	}
	return nil
}

// validateWorkflows enforces Rules 4–12 across all workflows.
func validateWorkflows(s *spec.TraCtlSpec) []ValidationError {
	var errs []ValidationError
	seenIDs := make(map[string]int) // id → first index; Rule 5

	for wi, wf := range s.Workflows {
		wField := fmt.Sprintf("workflows[%d]", wi)

		// Rule 4 — workflow id format.
		errs = append(errs, validateIdentifier(wf.ID, wField+".id")...)

		// Rule 5 — workflow id uniqueness (spec-global).
		if wf.ID != "" {
			if firstIdx, seen := seenIDs[wf.ID]; seen {
				errs = append(errs, ValidationError{
					Field:   wField + ".id",
					Code:    DuplicateWorkflowID,
					Message: fmt.Sprintf("duplicate workflow id %q (first seen at workflows[%d])", wf.ID, firstIdx),
				})
			} else {
				seenIDs[wf.ID] = wi
			}
		}

		// Rule — failurePolicy enum. Empty string → DefaultFailurePolicy at normalisation time.
		if fp := wf.FailurePolicy; fp != "" &&
			fp != spec.FailurePolicyResilient &&
			fp != spec.FailurePolicyFailFast {
			errs = append(errs, ValidationError{
				Field:   wField + ".failurePolicy",
				Code:    MissingRequiredField,
				Message: fmt.Sprintf("%s.failurePolicy %q is not valid; allowed: %q, %q", wField, fp, spec.FailurePolicyResilient, spec.FailurePolicyFailFast),
			})
		}

		// Rule 6 — steps required.
		if len(wf.Steps) == 0 {
			errs = append(errs, ValidationError{
				Field:   wField + ".steps",
				Code:    MissingRequiredField,
				Message: fmt.Sprintf("workflow %q must contain at least one step", wf.ID),
			})
			continue
		}

		errs = append(errs, validateSteps(wf, wi)...)
	}

	// Validate workflow-level dependsOn: unknown refs and cycles.
	errs = append(errs, validateWorkflowDependsOn(s)...)
	return errs
}

// validateWorkflowDependsOn checks that workflow dependsOn IDs all exist and
// that there are no cycles in the workflow dependency graph.
func validateWorkflowDependsOn(s *spec.TraCtlSpec) []ValidationError {
	wfIDs := make(map[string]int, len(s.Workflows))
	for i, wf := range s.Workflows {
		if wf.ID != "" {
			wfIDs[wf.ID] = i
		}
	}

	var errs []ValidationError

	// Unknown ref check.
	for wi, wf := range s.Workflows {
		for ki, dep := range wf.DependsOn {
			field := fmt.Sprintf("workflows[%d].dependsOn[%d]", wi, ki)
			if dep == wf.ID {
				errs = append(errs, ValidationError{
					Field:   field,
					Code:    UnknownDependencyRef,
					Message: fmt.Sprintf("workflow %q depends on itself", wf.ID),
				})
				continue
			}
			if _, ok := wfIDs[dep]; !ok {
				errs = append(errs, ValidationError{
					Field:   field,
					Code:    UnknownDependencyRef,
					Message: fmt.Sprintf("workflow %q dependsOn references unknown workflow %q", wf.ID, dep),
				})
			}
		}
	}

	// Cycle detection: build adjacency (dep → dependents) and run DFS.
	adj := make(map[string][]string, len(s.Workflows))
	for _, wf := range s.Workflows {
		for _, dep := range wf.DependsOn {
			if dep != wf.ID {
				adj[dep] = append(adj[dep], wf.ID)
			}
		}
	}

	const (
		unvisited = 0
		inStack   = 1
		done      = 2
	)
	state := make(map[string]int, len(s.Workflows))
	parent := make(map[string]string, len(s.Workflows))

	var dfs func(node string) bool
	dfs = func(node string) bool {
		state[node] = inStack
		for _, next := range adj[node] {
			if state[next] == inStack {
				path := reconstructCyclePath(node, next, parent)
				errs = append(errs, ValidationError{
					Field:   "workflows",
					Code:    CyclicDependency,
					Message: fmt.Sprintf("cyclic workflow dependency detected: %s", path),
				})
				return true
			}
			if state[next] == unvisited {
				parent[next] = node
				if dfs(next) {
					return true
				}
			}
		}
		state[node] = done
		return false
	}

	for _, wf := range s.Workflows {
		if wf.ID != "" && state[wf.ID] == unvisited {
			dfs(wf.ID)
		}
	}

	return errs
}
