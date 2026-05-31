// Package sandbox executes JavaScript scripts in a bounded, I/O-free environment.
package sandbox

import (
	"fmt"
	"math/rand"
	"sync/atomic"

	"github.com/dop251/goja"
)

// sandboxState holds deterministic API state for a single VM execution.
type sandboxState struct {
	baseNow  int64
	nowCount int64 // increment per call
	randSrc  *rand.Rand
	logs     []string
}

// injectSafeAPIs sets the global "tractl" object in the goja runtime.
// Only deterministic, I/O-free APIs are exposed.
//
// Exposed as tractl.*:
//
//	tractl.now()      — returns a deterministic integer (seeded epoch ms)
//	tractl.random()   — seeded pseudo-random float64 in [0, 1).
//	tractl.log(msg)   — appends msg (coerced to string) to an internal log slice.
//
// injectSafeAPIs returns a function that drains the accumulated logs.
func injectSafeAPIs(vm *goja.Runtime, seed int64) func() []string {
	st := &sandboxState{
		baseNow:  1_700_000_000_000, // fixed base epoch ms as per spec
		nowCount: 0,
		randSrc:  rand.New(rand.NewSource(seed)),
		logs:     make([]string, 0),
	}

	tractl := vm.NewObject()

	_ = tractl.Set("now", func(_ goja.FunctionCall) goja.Value {
		next := atomic.AddInt64(&st.nowCount, 1)
		v := st.baseNow + next - 1
		return vm.ToValue(v)
	})

	_ = tractl.Set("random", func(_ goja.FunctionCall) goja.Value {
		v := st.randSrc.Float64()
		return vm.ToValue(v)
	})

	_ = tractl.Set("log", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) > 0 {
			st.logs = append(st.logs, fmt.Sprint(call.Arguments[0].Export()))
		}
		return goja.Undefined()
	})

	_ = vm.Set("tractl", tractl)

	return func() []string {
		logs := st.logs
		st.logs = nil
		return logs
	}
}
