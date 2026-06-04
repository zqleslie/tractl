---
name: Engine Milestone
about: A specific piece of work — Go engine development task
title: '[Engine] '
labels: ['track:engine', 'priority:p2']
type: Task
---

## Context

_Describe the current file state, stubs, and TODOs that motivate this task._

## What to Build

_Describe the algorithm, data flow, or architectural change needed._

## Files to Create/Modify

_List exact file paths within the repo._

| File | Action | Purpose |
|------|--------|---------|
| `engine/...` | Create/Modify | ... |

## Import Boundary

| Constraint | Detail |
|------------|--------|
| MUST NOT import | ... |
| MAY import | ... |

## Agent Notes

- Go-specific pitfalls to avoid: ...
- Test expectations: `go vet ./...` must pass

## Acceptance Criteria

- [ ] `go vet ./...` exits 0
- [ ] `go test ./...` passes (no regressions)
- [ ] `npx tsc --noEmit` passes (if frontend touched)
- [ ] `npm --prefix frontend run test` passes
- [ ] Handoff log entry added in PR diff

## Handoff Log Entry

> Add a note in the PR diff documenting any architectural decisions or deviations from this spec.
