// Package runtime manages execution state for a tractl workflow run.
//
// It provides two complementary responsibilities:
//
//   - ExecutionContext tracks step lifecycle, eligibility, results, variables,
//     and run history. It is the authoritative record of what has happened
//     during a run.
//
//   - FrozenContext is an immutable snapshot taken for safe script execution.
//     Scripts receive copied scope maps and cannot mutate the live run state.
//
// ExpressionEvaluator resolves ${...} template strings against an ExecutionContext.
//
// This package owns the runtime state formerly split with internal/context,
// eliminating the dual-import requirement for pipeline consumers.
//
// Import constraints: internal/runtime may be imported by pipeline packages.
// It must not import internal/engine, internal/scheduler, or internal/executor.
package runtime
