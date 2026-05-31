// Package engine is the top-level orchestrator of the tractl pipeline.
//
// It accepts a Config for a workflow file or in-memory document and runs the
// full pipeline:
//
//  1. Parse — format-detect and parse the spec into *spec.TraCtlSpec.
//  2. Overlay — apply overlay documents in declaration order.
//  3. Validate — validate the merged spec against canonical rules.
//  4. Plan — topological sort and concurrency planning via Planner.
//  5. Compile — resolve executable step configuration via Compiler.
//  6. Execute — schedule and run steps via Scheduler, Executor, and Evaluator.
//
// Entry points: Engine.Run for file paths and Engine.RunDocument for raw
// document content.
//
// Dependency injection: Planner, Compiler, Evaluator, and ScriptRunner are
// declared as interfaces in interfaces.go. Pass concrete implementations via
// Config fields for testing; leave nil to use the production defaults.
//
// Output formatting: use FormatterFor(format) from format.go to obtain an
// OutputFormatter.
package engine
