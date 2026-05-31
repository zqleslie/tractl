# traCtl

> Git-first backend execution and validation platform for progressive API and workflow testing across CLI, web, desktop, CI, and containers.

![Go](https://img.shields.io/badge/go-1.26+-00ADD8?logo=go&logoColor=white)
![Node.js](https://img.shields.io/badge/node-24+-339933?logo=node.js&logoColor=white)
![License](https://img.shields.io/badge/license-Apache%202.0-blue)
![Status](https://img.shields.io/badge/status-early%20development-orange)
![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen)

---

## What is traCtl?

traCtl is a Git-first backend execution and validation platform that makes progressive engineering validation part of the normal developer workflow. Workflow assets are plain-text, filesystem-native, and repository-native — versioned, reviewed, and merged alongside the code they validate. There is no opaque workspace and no parallel source of truth.

traCtl is built on a single non-negotiable promise: **same workflow definition, equivalent execution semantics** — across Desktop, Web App, CLI, CI, and Container Runtime surfaces. Execution is deterministic by design, every surface goes through the same canonical pipeline, and progressive validation depth — correctness, auth behavior, contract drift, resilience, performance smoke — is incrementally adoptable without heavyweight workflow friction.

The current implementation already includes a runnable Go engine, CLI harness, browser WebAssembly runtime path, and shared React frontend used by the Web App and Wails desktop shell.

---

## Why traCtl?

- **Workflow fragmentation** — backend engineering workflows are scattered across disconnected tools with inconsistent execution models.
- **Validation friction** — meaningful validation is deferred because existing workflows require too much setup, so developers ship with happy-path checks only.
- **Local/CI drift** — what passes locally fails in CI because execution environments, tools, and semantics differ.
- **Duplicated validation logic** — the same assertions are rewritten in API clients, integration tests, and CI pipelines, then maintained independently.
- **Inconsistent workflow ownership** — validation ownership fragments across teams when assets live outside the repository.
- **No progressive adoption path** — deeper checks (auth, contract drift, resilience, performance) are all-or-nothing rather than incrementally adoptable.

---

## Protocol Support

| Category | Protocols |
|---|---|
| Request-Response | HTTP, GraphQL, gRPC, SOAP |
| Streaming | gRPC streaming, GraphQL subscriptions, WebSocket |
| Event / Messaging | MQTT, AMQP, Kafka |
| Raw Transport | TCP, UDP |

> MVP runtime begins with HTTP. Full protocol coverage is delivered progressively through the provider architecture.

---

## Authoring Formats

| Format | Description | Default |
|---|---|---|
| YAML | Standard structured format. Most familiar to developers. | Yes |
| JSON | Machine-friendly. Preferred for programmatic generation. | No |
| TOON | traCtl's own human-readable format. Purpose-built for workflow authoring. | No |

All formats produce equivalent canonical semantics. Choosing a format is an authoring preference, not an execution difference.

---

## Quick Example

```yaml
schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: health-check
    steps:
      - id: ping
        kind: request
        request:
          protocol: http
          target: https://api.example.com/health
          operation: GET
        assertions:
          - id: status-ok
            kind: status
            op: equals
            expected: 200
```

Run it:

```bash
tractl run health-check.yaml
```

---

## Delivery Surfaces

- **CLI** — `tractl run`. Canonical automation surface with deterministic exit codes and structured output.
- **Web App** — React + Go WebAssembly for browser-native Tier 1 execution; optional server-connected Tier 2 is planned.
- **Desktop** — Wails v2 native app using the shared React frontend and in-process Go engine.
- **CI** — same CLI semantics with pinned versions and machine-consumable output.
- **Container Runtime** — OCI execution surface planned for reproducible CI and isolated environments.
- **Browser Extension** — capture-only surface planned for workflow bootstrapping; it does not execute workflows.

---

## Project Status

> traCtl is in early development. The architecture is frozen and documented; implementation is active across the engine, CLI, Web App, and Desktop shell. **Not production ready.**

Current implementation highlights:

- Canonical validation, parsers, overlay engine, planner/compiler, runtime, assertions/extracts, parallel DAG scheduling, and script sandbox are implemented.
- CLI execution works through `tractl run` with text and JSON output.
- Diagnostics/reporting is in progress.
- Shared React frontend, WebAssembly runtime bridge, request editor, workflow canvas, run history, and environment variable resolution are in progress.

See [docs/spec/tractl_roadmap.md](docs/spec/tractl_roadmap.md) and [CHANGELOG.md](CHANGELOG.md) for progress to date.

---

## Getting Started

### Prerequisites

- **Go 1.26+** — [go.dev/dl](https://go.dev/dl/)
- **Node.js 24+** — [nodejs.org](https://nodejs.org/) (for the shared React frontend)
- **Wails v2 CLI** — `go install github.com/wailsapp/wails/v2/cmd/wails@latest` (for desktop builds)
- **GNU Make** — standard on macOS/Linux; on Windows use WSL or `winget install GnuWin32.Make`
- **golangci-lint** (optional, for `make lint`) — [golangci-lint.run/usage/install](https://golangci-lint.run/usage/install/)

### Clone and Set Up

```bash
git clone https://github.com/tractl/tractl.git
cd tractl
make setup
make install
make help
```

### Build

```bash
make build-cli
make build-frontend-web
make build-desktop
```

### Develop

```bash
make dev-cli
make dev-web
make dev-desktop
```

### Test

```bash
make test
make test-frontend
make check
```

---

## Contributing

Contributions are welcome — see [CONTRIBUTING.md](CONTRIBUTING.md).

---

## License

Apache License 2.0 — see [LICENSE](LICENSE).
