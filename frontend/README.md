# traCtl frontend

Shared React 18 + TypeScript UI for **web** and **desktop** surfaces (ADR-016).

## Layout

- `src/` — shared application source
- `src/platform/web/` — web-only adapters and constraints
- `src/platform/desktop/` — desktop-only capabilities (Wails bindings later)
- `dist/` — web build output (`npm run build:web`)
- `../cmd/desktop/frontend/dist/` — desktop build output for Wails `go:embed`

## Scripts

| Command | Purpose |
|---------|---------|
| `npm run dev:web` | Web surface dev server |
| `npm run dev:desktop` | Desktop surface dev (same UI, desktop build target) |
| `npm run build:web` | Production web bundle → `frontend/dist` |
| `npm run build:desktop` | Production desktop bundle → `cmd/desktop/frontend/dist` |
| `npm test` | Vitest component/unit tests |
| `npm run test:e2e` | Playwright integration tests |

## Platform split

Shared screens and primitives must not import Wails or other desktop-only APIs directly.
Use `getPlatformCapabilities()` from `src/platform` and keep desktop integrations in
`src/platform/desktop/`.
