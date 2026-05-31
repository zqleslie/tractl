// normalize.go defines the Normalizer type and its Normalize method, responsible
// for resolving implicit step dependencies by scanning expression-bearing fields
// for ${steps.<id>...} references and merging them into each step's DependsOn list.
package normalize

import (
	"sort"

	"github.com/tractl/tractl/internal/spec"
)

// Normalizer scans steps for expression-bearing references and merges
// discovered step IDs into each step's DependsOn list.
//
// Cross-workflow references (a step referencing another workflow's step ID)
// are NOT merged into intra-workflow DependsOn — they are resolved at
// runtime via shared execution context.
type Normalizer struct{}

// NewNormalizer returns a ready-to-use Normalizer.
func NewNormalizer() *Normalizer { return &Normalizer{} }

// Normalize returns a new *spec.TraCtlSpec with implicit dependencies merged.
// The source spec is never mutated.
//
// Errors:
//   - ErrSelfReference: a step references itself in an expression
//   - ErrDeadReference: a step references a non-existent step in the same workflow
//     (an identifier that does not appear in any workflow in the spec)
func (n *Normalizer) Normalize(src *spec.TraCtlSpec) (*spec.TraCtlSpec, error) {
	if src == nil {
		return nil, nil
	}

	// Pre-build: set of step IDs per workflow, and a global set of step IDs.
	perWorkflow := make(map[string]map[string]struct{}, len(src.Workflows))
	allIDs := make(map[string]struct{})
	for _, wf := range src.Workflows {
		ids := make(map[string]struct{}, len(wf.Steps))
		for _, st := range wf.Steps {
			if st.ID != "" {
				ids[st.ID] = struct{}{}
				allIDs[st.ID] = struct{}{}
			}
		}
		perWorkflow[wf.ID] = ids
	}

	// Build the output spec as a shallow copy of the source, then deep-copy
	// the Workflows slice so we can mutate steps independently.
	out := *src
	out.Workflows = make([]spec.Workflow, len(src.Workflows))
	for wi, wf := range src.Workflows {
		nw := wf
		nw.Steps = make([]spec.Step, len(wf.Steps))
		for si, st := range wf.Steps {
			ns := st

			// Determine the implicit refs from this step's expression-bearing fields.
			refs := collectStepRefs(st)
			localIDs := perWorkflow[wf.ID]

			// Merge with explicit DependsOn for dedup.
			explicit := make(map[string]struct{}, len(st.DependsOn))
			for _, d := range st.DependsOn {
				explicit[d] = struct{}{}
			}

			// Sort the refs for deterministic error ordering.
			refIDs := make([]string, 0, len(refs))
			for r := range refs {
				refIDs = append(refIDs, r)
			}
			sort.Strings(refIDs)

			for _, ref := range refIDs {
				// Self reference?
				if ref == st.ID {
					return nil, normalizerErr(ErrSelfReference, wf.ID, st.ID, "step references itself in an expression")
				}
				// Reference unknown anywhere in the spec → dead.
				if _, exists := allIDs[ref]; !exists {
					return nil, normalizerErr(ErrDeadReference, wf.ID, st.ID, "expression references unknown step "+ref)
				}
				// Cross-workflow reference (exists somewhere but not in this workflow):
				// do NOT add to DependsOn — resolved at runtime via shared context.
				if _, inLocal := localIDs[ref]; !inLocal {
					continue
				}
				explicit[ref] = struct{}{}
			}

			// Materialise a sorted, deduped DependsOn slice.
			merged := make([]string, 0, len(explicit))
			for d := range explicit {
				merged = append(merged, d)
			}
			sort.Strings(merged)
			if len(merged) == 0 {
				ns.DependsOn = nil
			} else {
				ns.DependsOn = merged
			}

			nw.Steps[si] = ns
		}
		out.Workflows[wi] = nw
	}

	return &out, nil
}

// collectStepRefs gathers the set of step IDs referenced by all expression-bearing
// fields of a step: When, Request.Target, Request.Headers (values),
// Request.Body content (recursively).
func collectStepRefs(st spec.Step) map[string]struct{} {
	refs := make(map[string]struct{})

	if st.When != "" {
		for k := range scanStepRefs(st.When) {
			refs[k] = struct{}{}
		}
	}

	if st.Request != nil {
		if st.Request.Target != "" {
			for k := range scanStepRefs(st.Request.Target) {
				refs[k] = struct{}{}
			}
		}
		for _, hv := range st.Request.Headers {
			for k := range scanStepRefs(hv) {
				refs[k] = struct{}{}
			}
		}
		if st.Request.Body != nil && st.Request.Body.Content != nil {
			scanRefsInValue(st.Request.Body.Content, refs)
		}
	}

	return refs
}

// Normalize is a package-level convenience wrapper.
func Normalize(src *spec.TraCtlSpec) (*spec.TraCtlSpec, error) {
	return NewNormalizer().Normalize(src)
}
