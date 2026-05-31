# traCtl UI Screen Inventory

Authority: tractl_ui_spec.md
Phase: 15 (Multi-Surface Delivery)
Shell: Wails (desktop) + React/TypeScript (web)
Last updated: 2026-05-31

| ID  | Screen                  | Milestone | Status          | Notes                                             |
|-----|-------------------------|-----------|-----------------|---------------------------------------------------|
| S1  | App Shell + Fast Start  | 15.5      | Mockup approved | 48px activity bar, tab bar, command palette       |
| S2  | Request Editor          | 15.5      | Mockup approved | Body encoding switcher, bottom toolbar, shortcuts |
| S3  | Workflow Canvas         | 15.5      | Mockup approved | DAG, filter bar, code view, result popup          |
| S4  | Step Detail panel       | 15.5      | Covered in S3   | Popup inside canvas, reuses S2 form panels        |
| S5  | Workflow Run Results    | 15.5      | Covered in S3   | Full results popup inside canvas                  |
| S6  | Environment Manager     | 15.5      | Not started     | TBD — model decision pending                      |
| S7  | Overlay Editor          | 15.5      | Mockup approved | Patch list, action picker, diff/result preview    |
| S8  | Script Editor           | 15.5      | Not started     | Full-screen Monaco for workflow-level hooks       |
| S9  | Run History             | 15.5      | Mockup approved | Chronological list, filters, detail popup         |
| S10 | Settings                | 15.5      | Mockup approved | 6 sections, all prefs, about, reset               |
| S11 | Credential Manager      | later     | Not started     | Single phase, keychain-backed, masked values      |

## Shell changes (v4 redesign — 2026-05-31)
All screens inherit the following shell updates based on developer feedback:
- Activity bar collapsed to 48px icon-only (was 210px tabbed sidebar)
- Side panel slides in/out on activity bar icon click
- Browser-style tab bar for open requests and workflows
- Command palette (⌘K) as primary navigation
- Environment selector pill in topbar — blue (staging) / red (production)
- Bottom toolbar: layout toggle (stacked/side-by-side) + word wrap + shortcuts button
- Keyboard shortcuts reference popup (⌘/)
- Body encoding: full switcher (JSON/Form data/Multipart/Raw/Binary/None)
- "Send" replaces "Run" for single requests (HTTP mental model alignment)
- Results panel stacked by default, side-by-side via toolbar toggle

## Remaining for mockup phase
- S6: Environment Manager (pending model decision)
- S8: Script Editor (workflow-level beforeAll/afterAll hooks, full Monaco view)
