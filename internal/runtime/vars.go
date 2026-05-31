// vars.go defines code for the runtime package.

package runtime

// VariableResolver resolves variables across multiple scopes using an
// ExecutionContext as the backing store.
type VariableResolver struct {
	ctx *ExecutionContext
}

// NewVariableResolver creates a resolver bound to an ExecutionContext.
func NewVariableResolver(ctx *ExecutionContext) *VariableResolver {
	return &VariableResolver{ctx: ctx}
}

// Resolve resolves a variable by scope and key.
// Supported scopes:
//   - "step": key must be "<stepId>.<extractId>"; reads from step extracts
//   - "workflow", "env", "spec": reads from the named variable scope via ExecutionContext.GetVar
func (vr *VariableResolver) Resolve(scope, key string) (string, bool) {
	if vr == nil || vr.ctx == nil {
		return "", false
	}
	switch scope {
	case "step":
		// key format: "stepId.extractId"
		var stepID, extractID string
		for i, ch := range key {
			if ch == '.' {
				stepID = key[:i]
				extractID = key[i+1:]
				break
			}
		}
		if stepID == "" || extractID == "" {
			return "", false
		}
		return vr.ctx.GetExtract(stepID, extractID)
	default:
		return vr.ctx.GetVar(scope, key)
	}
}
