package sandbox

import (
	"fmt"
	"strings"
	"time"

	"github.com/dop251/goja"

	"github.com/tractl/tractl/internal/runtime"
)

// DefaultTimeout is the maximum wall-clock time a single script may run.
// Scripts exceeding this limit are interrupted with ErrTimeout.
const DefaultTimeout = 5 * time.Second

// Sandbox executes JavaScript scripts in a bounded, I/O-free environment.
// Each Execute call creates a fresh goja.Runtime; Sandbox itself is safe for
// concurrent use.
type Sandbox struct {
	timeout time.Duration
}

// NewSandbox returns a Sandbox with the specified timeout.
// If timeout ≤ 0, DefaultTimeout is used.
func NewSandbox(timeout time.Duration) *Sandbox {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	return &Sandbox{timeout: timeout}
}

// Execute runs the JavaScript source string inside a fresh sandbox environment.
//
// The script receives:
//
//	ctx     — the frozen context object (read from frozen; Go-side is unchanged after return)
//	tractl  — injected safe API object (see apis.go)
//
// The script MAY return a mutation set descriptor object (see Mutation Set Contract below).
// If the script returns undefined/null or nothing, an empty MutationSet is returned.
//
// seed is used to initialise the deterministic RNG and clock APIs.
// Same seed + same script → same tractl.random() / tractl.now() sequence.
//
// Errors:
//
//	*SandboxError with Code ErrSyntax     — JavaScript syntax error
//	*SandboxError with Code ErrRuntime    — uncaught JavaScript exception
//	*SandboxError with Code ErrTimeout    — script exceeded s.timeout
//	*SandboxError with Code ErrInvalidResult — return value has wrong shape
func (s *Sandbox) Execute(source string, frozen runtime.FrozenContext, seed int64) (MutationSet, error) {
	vm := goja.New()

	// Prohibit all I/O: do NOT set console, fetch, require, process, or fs.
	// The default goja runtime has none of these; just do not inject them.

	// Inject frozen context
	if err := injectFrozenContext(vm, frozen); err != nil {
		return MutationSet{}, err
	}

	// Inject safe APIs
	drainLogs := injectSafeAPIs(vm, seed)

	// Enforce timeout via interrupt from a separate goroutine
	timer := time.AfterFunc(s.timeout, func() {
		vm.Interrupt("execution timeout exceeded")
	})
	defer timer.Stop()

	// Wrap source in an IIFE so top-level `return` is allowed.
	wrapped := "(function(){\n" + source + "\n})()"
	val, err := vm.RunString(wrapped)
	if err != nil {
		return MutationSet{}, classifyError(err)
	}

	ms, perr := parseMutationSet(val)
	if perr != nil {
		return MutationSet{}, perr
	}
	// Drain logs from API closure into the MutationSet
	ms.Logs = append(ms.Logs, drainLogs()...)
	return ms, nil
}

// classifyError converts a goja error into a *SandboxError with the appropriate code.
// Distinguish between syntax errors, runtime exceptions, and interrupt (timeout).
func classifyError(err error) *SandboxError {
	if err == nil {
		return nil
	}
	errStr := err.Error()
	if strings.Contains(errStr, "SyntaxError") || strings.Contains(errStr, "syntax error") {
		return &SandboxError{Code: ErrSyntax, Message: errStr, Cause: err}
	}
	if strings.Contains(errStr, "execution timeout") || strings.Contains(errStr, "execution timeout exceeded") {
		return &SandboxError{Code: ErrTimeout, Message: "script timed out", Cause: err}
	}
	return &SandboxError{Code: ErrRuntime, Message: errStr, Cause: err}
}

// parseMutationSet converts the goja.Value returned by the script into a MutationSet.
// Accepts undefined/null (→ empty MutationSet) or a plain object with optional
// fields: variables, extracts, assertions, logs, cancel.
// Returns ErrInvalidResult if the shape is wrong.
func parseMutationSet(val goja.Value) (MutationSet, error) {
	if goja.IsUndefined(val) || goja.IsNull(val) {
		return MutationSet{}, nil
	}
	exported := val.Export()
	obj, ok := exported.(map[string]interface{})
	if !ok {
		return MutationSet{}, &SandboxError{Code: ErrInvalidResult, Message: "script must return an object"}
	}
	ms := MutationSet{
		Variables: make(map[string]string),
		Extracts:  make(map[string]string),
		Logs:      []string{},
	}
	if v, ok := obj["variables"]; ok {
		if m, ok := v.(map[string]interface{}); ok {
			for kk, vv := range m {
				ms.Variables[kk] = fmt.Sprint(vv)
			}
		} else {
			return MutationSet{}, &SandboxError{Code: ErrInvalidResult, Message: "variables must be an object"}
		}
	}
	if v, ok := obj["extracts"]; ok {
		if m, ok := v.(map[string]interface{}); ok {
			for kk, vv := range m {
				ms.Extracts[kk] = fmt.Sprint(vv)
			}
		} else {
			return MutationSet{}, &SandboxError{Code: ErrInvalidResult, Message: "extracts must be an object"}
		}
	}
	if v, ok := obj["assertions"]; ok {
		if arr, ok := v.([]interface{}); ok {
			for _, item := range arr {
				if ai, ok := item.(map[string]interface{}); ok {
					add := AssertionAddition{}
					if vv, ok := ai["id"].(string); ok {
						add.ID = vv
					}
					if vv, ok := ai["kind"].(string); ok {
						add.Kind = vv
					}
					if vv, ok := ai["operator"].(string); ok {
						add.Operator = vv
					}
					if vv, ok := ai["target"].(string); ok {
						add.Target = vv
					}
					if vv, ok := ai["expected"].(string); ok {
						add.Expected = vv
					} else if vv, ok := ai["expected"]; ok {
						add.Expected = fmt.Sprint(vv)
					}
					if vv, ok := ai["severity"].(string); ok {
						if vv == "error" || vv == "warning" {
							add.Severity = vv
						}
					}
					ms.Assertions = append(ms.Assertions, add)
				}
			}
		} else {
			return MutationSet{}, &SandboxError{Code: ErrInvalidResult, Message: "assertions must be an array"}
		}
	}
	if v, ok := obj["logs"]; ok {
		if arr, ok := v.([]interface{}); ok {
			for _, item := range arr {
				ms.Logs = append(ms.Logs, fmt.Sprint(item))
			}
		} else {
			return MutationSet{}, &SandboxError{Code: ErrInvalidResult, Message: "logs must be an array"}
		}
	}
	if v, ok := obj["cancel"]; ok {
		if b, ok := v.(bool); ok {
			ms.Cancel = b
		} else {
			return MutationSet{}, &SandboxError{Code: ErrInvalidResult, Message: "cancel must be a boolean"}
		}
	}
	return ms, nil
}
