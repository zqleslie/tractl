---
name: UI Task
about: A React/TypeScript frontend task — screen, component, or store change
title: '[UI] '
labels: ['track:ui', 'priority:p2']
type: Task
---

## Context

_Describe the current screen state from the screen inventory, store names, and what is stubbed._

## What to Build

_Describe the UI component, interaction, or state change needed._

## Files to Create/Modify

_List exact `frontend/src/` paths._

| File | Action | Purpose |
|------|--------|---------|
| `frontend/src/...` | Create/Modify | ... |

## Execution Paths

- [ ] Web WASM
- [ ] Desktop Wails
- [ ] Both (verify on each surface)

## Design Constraints

| ADR | Constraint |
|-----|------------|
| ADR-016 | Tailwind core only |
| Icons | `@tabler/icons-react` outline only |
| Editor | Monaco only |

## Agent Notes

- React/TS pitfalls: ...
- Store patterns: ...

## Acceptance Criteria

- [ ] `npx tsc --noEmit` exits 0
- [ ] `npm --prefix frontend run test` passes (no regressions)
- [ ] `npm --prefix frontend run lint` exits 0
- [ ] Component renders correctly on Web WASM surface
- [ ] Component renders correctly on Desktop Wails surface (if applicable)
