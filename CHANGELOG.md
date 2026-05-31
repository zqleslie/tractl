# Changelog

All notable changes to traCtl will be documented in this file.

Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added

- Monorepo scaffolding: CLI (`cmd/tractl/`), desktop app (`cmd/desktop/` — Wails v2 + React TypeScript), and core engine package stubs (`internal/`)
- `AGENTS.md` — universal agent context file with architecture baseline, active phase guardrail, and build commands
- `Makefile` — GNU Make build runner covering build, test, quality, workspace, and clean targets
- Architecture documentation: ADRs (`docs/adr/`), canonical specs (`docs/spec/`), HLD, and terminology registry
- Open-source community files: `README.md`, `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `SECURITY.md`, `CHANGELOG.md`, `LICENSE`, `NOTICE`
- GitHub issue templates (bug report, feature request) and pull request template
