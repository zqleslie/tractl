// rules_dag.go defines validation rules for step dependency graphs:
// dependsOn reference validity (Rule 11), self-reference detection, and cycle
// detection via depth-first search (Rule 12). Reference: tractl_spec.md §12.
package validation

import (
	"fmt"
	"strings"

	"github.com/tractl/tractl/internal/spec"
)

// validateDependsOn enforces Rule 11. Reference: tractl_spec.md §12.1
func validateDependsOn(wf spec.Workflow, wi int, stepIDs map[string]bool) []ValidationError {
	var errs []ValidationError
	wField := fmt.Sprintf("workflows[%d]", wi)

	for si, st := range wf.Steps {
		for ki, dep := range st.DependsOn {
			field := fmt.Sprintf("%s.steps[%d].dependsOn[%d]", wField, si, ki)
			if dep == st.ID {
				errs = append(errs, ValidationError{
					Field:   field,
					Code:    UnknownDependencyRef,
					Message: fmt.Sprintf("step %q depends on itself", st.ID),
				})
				continue
			}
			if !stepIDs[dep] {
				errs = append(errs, ValidationError{
					Field:   field,
					Code:    UnknownDependencyRef,
					Message: fmt.Sprintf("step %q depends on unknown step %q (not found in workflow %q)", st.ID, dep, wf.ID),
				})
			}
		}
	}
	return errs
}

// validateCycles enforces Rule 12 using iterative DFS.
// Only valid edges (both endpoints known, non-self) are considered.
// Reference: tractl_spec.md §12.1, §6.2
func validateCycles(wf spec.Workflow, wi int, stepIDs map[string]bool) []ValidationError {
	// Build adjacency list: dep → dependents.
	adj := make(map[string][]string, len(wf.Steps))
	for _, st := range wf.Steps {
		if !stepIDs[st.ID] {
			continue
		}
		for _, dep := range st.DependsOn {
			if dep == st.ID || !stepIDs[dep] {
				continue
			}
			adj[dep] = append(adj[dep], st.ID)
		}
	}

	const (
		unvisited = 0
		inStack   = 1
		done      = 2
	)
	state := make(map[string]int, len(wf.Steps))
	parent := make(map[string]string, len(wf.Steps))

	var errs []ValidationError

	var dfs func(node string) bool
	dfs = func(node string) bool {
		state[node] = inStack
		for _, next := range adj[node] {
			if state[next] == inStack {
				path := reconstructCyclePath(node, next, parent)
				errs = append(errs, ValidationError{
					Field:   fmt.Sprintf("workflows[%d].steps", wi),
					Code:    CyclicDependency,
					Message: fmt.Sprintf("cycle detected: %s", path),
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

	for _, st := range wf.Steps {
		if stepIDs[st.ID] && state[st.ID] == unvisited {
			dfs(st.ID)
		}
	}
	return errs
}

// reconstructCyclePath builds a human-readable cycle path.
// cycleEnd is the node already in-stack when cycleStart was visited.
func reconstructCyclePath(cycleEnd, cycleStart string, parent map[string]string) string {
	var path []string
	for cur := cycleEnd; cur != cycleStart; {
		path = append(path, cur)
		p, ok := parent[cur]
		if !ok {
			break
		}
		cur = p
	}
	path = append(path, cycleStart)

	// Reverse to read start → … → end, then close the cycle.
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	path = append(path, path[0])
	return strings.Join(path, " → ")
}
