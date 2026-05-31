// bench_test.go defines benchmark tests for the overlay engine.
// Run with: go test -bench=. -benchmem ./internal/overlay/
//
// Establish a baseline BEFORE any structural refactoring of this package.
// Compare post-refactoring results against bench_baseline.txt.
package overlay

import (
	"fmt"
	"testing"
)

// buildBenchmarkSpec constructs a representative base spec for benchmarking
// as the map[string]any shape the overlay engine operates on.
// 3 workflows × 10 steps each, with mixed dependsOn edges.
func buildBenchmarkSpec() map[string]any {
	workflows := make([]any, 3)
	for wi := 0; wi < 3; wi++ {
		steps := make([]any, 10)
		for si := 0; si < 10; si++ {
			stepID := fmt.Sprintf("step-%02d", si+1)
			step := map[string]any{
				"id":   stepID,
				"kind": "request",
				"request": map[string]any{
					"protocol": "http",
					"target":   fmt.Sprintf("https://api.example.com/wf%d/%s", wi+1, stepID),
					"headers": map[string]any{
						"Accept":     "application/json",
						"X-Workflow": fmt.Sprintf("wf%d", wi+1),
					},
				},
				"assertions": []any{},
				"extracts":   []any{},
			}
			// Non-trivial dependsOn graph: step-03→step-01, step-05→step-02,
			// step-07→step-04, step-10→step-06.
			switch si {
			case 2:
				step["dependsOn"] = []any{"step-01"}
			case 4:
				step["dependsOn"] = []any{"step-02"}
			case 6:
				step["dependsOn"] = []any{"step-04"}
			case 9:
				step["dependsOn"] = []any{"step-06"}
			}
			steps[si] = step
		}
		workflows[wi] = map[string]any{
			"id":    fmt.Sprintf("wf%d", wi+1),
			"name":  fmt.Sprintf("Workflow %d", wi+1),
			"steps": steps,
		}
	}
	return map[string]any{
		"schemaVersion": float64(1),
		"capabilities":  []any{"protocol.http"},
		"variables": map[string]any{
			"timeout": "PT30S",
			"baseURL": "https://api.example.com",
		},
		"workflows": workflows,
	}
}

// buildBenchmarkOverlays constructs 3 overlay documents for benchmarking.
// Overlays exercise different merge paths: path-based scalar replace,
// match-based tree traversal (mode:all), and combined patches.
func buildBenchmarkOverlays() []*OverlayDocument {
	// Overlay 1: three path-based variable overrides — exercises resolvePath
	// and deepMerge on a scalar map.
	overlay1 := &OverlayDocument{
		Metadata: &OverlayMetadata{Name: "variables-overlay"},
		Patches: []Patch{
			{
				Target: Target{Path: "variables"},
				Action: ActionDeepMerge,
				Data:   map[string]any{"timeout": "PT60S"},
			},
			{
				Target: Target{Path: "variables"},
				Action: ActionDeepMerge,
				Data:   map[string]any{"retryCount": "3"},
			},
			{
				Target: Target{Path: "variables"},
				Action: ActionDeepMerge,
				Data:   map[string]any{"region": "us-east-1"},
			},
		},
	}

	// Overlay 2: single match-based patch targeting all request steps
	// (mode:all, 30 matches across 3 workflows × 10 steps) — exercises
	// semanticMatch tree traversal and repeated applyToNode calls.
	overlay2 := &OverlayDocument{
		Metadata: &OverlayMetadata{Name: "step-config-overlay"},
		Patches: []Patch{
			{
				Target: Target{
					Match: MatchSelector{"kind": "request"},
					Mode:  TargetModeAll,
				},
				Action: ActionDeepMerge,
				Data: map[string]any{
					"config": map[string]any{
						"followRedirects": true,
						"maxRedirects":    float64(5),
					},
				},
			},
		},
	}

	// Overlay 3: mixed patches — path replace plus match-based header
	// injection into all HTTP request sub-maps — exercises both path and
	// match code paths in a single overlay.
	overlay3 := &OverlayDocument{
		Metadata: &OverlayMetadata{Name: "env-overlay"},
		Patches: []Patch{
			{
				Target: Target{Path: "variables"},
				Action: ActionDeepMerge,
				Data:   map[string]any{"env": "benchmark"},
			},
			{
				Target: Target{
					Match: MatchSelector{"protocol": "http"},
					Mode:  TargetModeAll,
				},
				Action: ActionDeepMerge,
				Data: map[string]any{
					"headers": map[string]any{
						"X-Benchmark": "true",
					},
				},
			},
		},
	}

	return []*OverlayDocument{overlay1, overlay2, overlay3}
}

// BenchmarkOverlayApply measures applying all 3 overlays to a 3-workflow ×
// 10-step spec. This is the primary hot-path benchmark.
func BenchmarkOverlayApply(b *testing.B) {
	baseSpec := buildBenchmarkSpec()
	overlays := buildBenchmarkOverlays()

	if len(baseSpec) == 0 {
		b.Fatal("buildBenchmarkSpec returned empty map")
	}
	if len(overlays) != 3 {
		b.Fatalf("expected 3 overlays, got %d", len(overlays))
	}

	engine := NewEngine()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result, err := engine.Apply(baseSpec, overlays)
		if err != nil {
			b.Fatalf("Apply failed: %v", err)
		}
		if result == nil {
			b.Fatal("Apply returned nil result")
		}
	}
}

// BenchmarkOverlayApplySingle benchmarks applying a single overlay containing
// only path-based patches. Isolates the path-resolution and deepMerge cost
// from the match traversal cost measured by BenchmarkOverlayApply.
func BenchmarkOverlayApplySingle(b *testing.B) {
	baseSpec := buildBenchmarkSpec()
	overlays := buildBenchmarkOverlays()[:1] // only the path-based overlay

	engine := NewEngine()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result, err := engine.Apply(baseSpec, overlays)
		if err != nil {
			b.Fatalf("Apply failed: %v", err)
		}
		_ = result
	}
}

// BenchmarkOverlayApplyParallel benchmarks concurrent overlay applications.
// Engine is stateless so all goroutines share one instance. Verifies there
// is no lock contention under parallel load.
func BenchmarkOverlayApplyParallel(b *testing.B) {
	baseSpec := buildBenchmarkSpec()
	overlays := buildBenchmarkOverlays()
	engine := NewEngine()

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			result, err := engine.Apply(baseSpec, overlays)
			if err != nil {
				b.Errorf("Apply failed: %v", err)
				return
			}
			_ = result
		}
	})
}
