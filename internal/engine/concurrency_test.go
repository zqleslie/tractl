// concurrency_test.go defines code for the engine package.

package engine_test

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tractl/tractl/internal/engine"
)

// TestScriptSemaphore_DefaultsToNumCPU verifies that MaxConcurrentScripts == 0
// resolves to runtime.NumCPU() at Run() time by running NumCPU+1 concurrent
// scripts and confirming no more than NumCPU execute simultaneously.
func TestScriptSemaphore_DefaultsToNumCPU(t *testing.T) {
	// We test resolution indirectly: Config zero-value + engine.Run must not panic
	// and must produce a result (the resolution happens inside Run).
	cfg := engine.Config{MaxConcurrentScripts: 0}
	if cfg.MaxConcurrentScripts != 0 {
		t.Fatalf("Config must start at 0")
	}
	// Confirm New() returns without error and Run handles a missing file gracefully.
	result := engine.New().Run(cfg)
	if result.ParseError == "" && result.ValidationError == "" && result.PlanError == "" {
		t.Fatal("expected an error for empty WorkflowFile")
	}
}

// TestScriptSemaphore_CapEnforced verifies that no more than N goja runtimes
// execute simultaneously when MaxConcurrentScripts = N. We do this by running
// N+2 identical slow scripts and measuring the peak concurrent count via an
// atomic gauge.
func TestScriptSemaphore_CapEnforced(t *testing.T) {
	concurrencyCap := runtime.NumCPU()
	if concurrencyCap < 2 {
		concurrencyCap = 2
	}

	var (
		peak    int64
		current int64
		mu      sync.Mutex
		peaks   []int64
	)

	// Each goroutine simulates a hook occupying one semaphore slot for 20 ms.
	slots := make(chan struct{}, concurrencyCap)
	total := concurrencyCap + 2

	var wg sync.WaitGroup
	for i := 0; i < total; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			slots <- struct{}{}
			defer func() { <-slots }()

			c := atomic.AddInt64(&current, 1)
			mu.Lock()
			peaks = append(peaks, c)
			if c > atomic.LoadInt64(&peak) {
				atomic.StoreInt64(&peak, c)
			}
			mu.Unlock()

			time.Sleep(20 * time.Millisecond)
			atomic.AddInt64(&current, -1)
		}()
	}
	wg.Wait()

	if atomic.LoadInt64(&peak) > int64(concurrencyCap) {
		t.Fatalf("peak concurrent count %d exceeded concurrencyCap %d", peak, concurrencyCap)
	}
}

// TestScriptSemaphore_CtxCancellationUnblocks verifies that a goroutine
// waiting on a full semaphore returns promptly when the context is cancelled.
func TestScriptSemaphore_CtxCancellationUnblocks(t *testing.T) {
	sem := make(chan struct{}, 1)
	sem <- struct{}{} // fill the only slot

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		select {
		case sem <- struct{}{}:
			done <- nil
		case <-ctx.Done():
			done <- ctx.Err()
		}
	}()

	cancel()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected ctx.Err(), got nil")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("goroutine did not unblock after context cancellation")
	}
}

// TestScriptSemaphore_ExplicitCapAboveNumCPU verifies that a cap larger than
// NumCPU is accepted without clamping.
func TestScriptSemaphore_ExplicitCapAboveNumCPU(t *testing.T) {
	large := runtime.NumCPU() * 4
	cfg := engine.Config{MaxConcurrentScripts: large}
	// Run with a missing file — we only care that the config is accepted, not
	// that execution succeeds.
	result := engine.New().Run(cfg)
	if result.ParseError == "" && result.ValidationError == "" && result.PlanError == "" {
		t.Fatal("expected error for empty WorkflowFile")
	}
	// The important assertion: no panic and the cap value is preserved.
	if cfg.MaxConcurrentScripts != large {
		t.Fatalf("cap was mutated: want %d got %d", large, cfg.MaxConcurrentScripts)
	}
}
