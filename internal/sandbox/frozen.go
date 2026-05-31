package sandbox

import (
	"github.com/dop251/goja"

	"github.com/tractl/tractl/internal/runtime"
)

// injectFrozenContext serialises a runtime.FrozenContext into the goja runtime
// as a global "ctx" variable.
//
// The ctx object shape:
//
//	ctx.spec     — map of spec-scope variables
//	ctx.env      — map of environment-scope variables
//	ctx.workflow  — map of workflow-scope variables
//	ctx.step     — map of step-scope variables
//	ctx.runtime  — map of runtime-scope variables
//
// The Go-side FrozenContext is never mutated. Any JS assignment to ctx.*
// affects only the goja runtime's copy of the object.
//
// Implementation: convert each scope's map[string]any to a goja object
// using vm.ToValue(). Nest them under a parent "ctx" object set via vm.Set("ctx", ...).
func injectFrozenContext(vm *goja.Runtime, frozen runtime.FrozenContext) error {
	// make copies to ensure JS mutations don't affect Go-side frozen maps
	ctxObj := map[string]interface{}{
		"spec":     copyStringMap(frozen.Spec),
		"env":      copyStringMap(frozen.Environment),
		"workflow": copyStringMap(frozen.Workflow),
		"step":     copyStringMap(frozen.Step),
		"runtime":  copyStringMap(frozen.Runtime),
	}
	_ = vm.Set("ctx", vm.ToValue(ctxObj))
	return nil
}

func copyStringMap(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}
	out := make(map[string]string, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}
