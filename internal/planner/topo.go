package planner

import (
	"sort"

	"github.com/tractl/tractl/internal/spec"
)

// topoSortWorkflows topologically sorts workflows by their DependsOn edges using
// Kahn's algorithm with lexicographic tie-breaking. Returns ErrCyclicDependency
// if a cycle is detected among workflow dependencies.
func topoSortWorkflows(workflows []spec.Workflow) ([]spec.Workflow, error) {
	if len(workflows) == 0 {
		return []spec.Workflow{}, nil
	}

	wfByID := make(map[string]*spec.Workflow, len(workflows))
	for i := range workflows {
		wfByID[workflows[i].ID] = &workflows[i]
	}

	inDegree := make(map[string]int, len(workflows))
	adjList := make(map[string][]string)

	for _, wf := range workflows {
		if _, exists := inDegree[wf.ID]; !exists {
			inDegree[wf.ID] = 0
		}
		for _, dep := range wf.DependsOn {
			inDegree[wf.ID]++
			adjList[dep] = append(adjList[dep], wf.ID)
		}
	}

	for key := range adjList {
		sort.Strings(adjList[key])
	}

	queue := []string{}
	for _, wf := range workflows {
		if inDegree[wf.ID] == 0 {
			queue = append(queue, wf.ID)
		}
	}
	sort.Strings(queue)

	result := []spec.Workflow{}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		result = append(result, *wfByID[cur])
		for _, dep := range adjList[cur] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
				sort.Strings(queue)
			}
		}
	}

	if len(result) != len(workflows) {
		return nil, plannerErr(ErrCyclicDependency, "", "", "cyclic dependency detected among workflows")
	}

	return result, nil
}

// topoSort implements Kahn's algorithm (iterative, queue-based) to topologically
// sort steps based on their DependsOn edges. It ensures deterministic ordering
// by processing steps in lexicographic ID order when multiple are eligible.
//
// If a cycle is detected, returns an error with ErrCyclicDependency.
// Otherwise returns the sorted []spec.Step.
func topoSort(workflowID string, steps []spec.Step) ([]spec.Step, error) {
	if len(steps) == 0 {
		return []spec.Step{}, nil
	}

	// Build maps for efficient lookup.
	stepByID := make(map[string]*spec.Step)
	for i := range steps {
		stepByID[steps[i].ID] = &steps[i]
	}

	// Build in-degree map and adjacency list (edges A→S for each S with DependsOn [A, ...]).
	inDegree := make(map[string]int)
	adjList := make(map[string][]string)

	for _, step := range steps {
		if _, exists := inDegree[step.ID]; !exists {
			inDegree[step.ID] = 0
		}
		for _, dep := range step.DependsOn {
			inDegree[step.ID]++
			adjList[dep] = append(adjList[dep], step.ID)
		}
	}

	// Sort adjacency lists for determinism.
	for _, deps := range adjList {
		sort.Strings(deps)
	}

	// Initialize queue with all steps having in-degree 0, sorted by ID.
	queue := []string{}
	for _, step := range steps {
		if inDegree[step.ID] == 0 {
			queue = append(queue, step.ID)
		}
	}
	sort.Strings(queue)

	result := []spec.Step{}

	for len(queue) > 0 {
		// Pop first item from queue.
		current := queue[0]
		queue = queue[1:]
		result = append(result, *stepByID[current])

		// Process dependents (steps that depend on current).
		for _, dependent := range adjList[current] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
				// Re-sort queue to maintain lexicographic order.
				sort.Strings(queue)
			}
		}
	}

	// If we didn't process all steps, a cycle exists.
	if len(result) != len(steps) {
		return nil, plannerErr(ErrCyclicDependency, workflowID, "", "cyclic dependency detected")
	}

	return result, nil
}
