BINARY_NAME  := tractl
CLI_DIR      := ./cmd/tractl
DESKTOP_DIR  := ./cmd/desktop
FRONTEND_DIR := ./frontend
BIN_DIR      := ./bin
BINARY_OUT   := $(BIN_DIR)/$(BINARY_NAME)

# ── Go workspace mode ─────────────────────────────────────────────────────────
# cmd/desktop is a SEPARATE Go module (Wails generates its own go.mod).
# GOWORK=off ensures ./... in CLI / test / lint targets does not accidentally
# pull in the desktop module. The workspace target re-enables it for IDE use.
export GOWORK := off

# ── Module + version ──────────────────────────────────────────────────────────
MODULE      := github.com/tractl/tractl
VERSION     := $(shell git describe --tags --always --dirty 2>/dev/null \
                 || echo "0.1.0-alpha")
COMMIT      := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE        := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
VERSION_PKG := $(MODULE)/internal/version

# -s -w strips symbol table and DWARF debug info (smaller binary).
# -X bakes values into package-level variables at link time.
LDFLAGS := -s -w \
	-X '$(VERSION_PKG).Version=$(VERSION)' \
	-X '$(VERSION_PKG).Commit=$(COMMIT)'   \
	-X '$(VERSION_PKG).Date=$(DATE)'

# CGO_ENABLED=0 → fully static binary, no C runtime dependency.
# -trimpath     → removes local $GOPATH from binary (reproducible builds).
CGO_ENABLED := 0
BUILD_FLAGS := -trimpath -ldflags "$(LDFLAGS)"

# ── Docker ────────────────────────────────────────────────────────────────────
DOCKER_REPO := tractl/tractl
DOCKER_TAG  ?= $(VERSION)

.PHONY: all help deps deps-frontend deps-all install install-cli \
        build build-cli build-desktop build-frontend-web build-frontend-desktop build-wasm build-server build-web \
        dev dev-web dev-desktop dev-cli dev-localapi dev-frontend-web \
        test test-verbose test-cover test-race test-phase1 test-integration test-e2e \
        test-frontend test-frontend-e2e \
        embed-stub vet fmt fmt-check lint vuln check check-full \
        setup workspace \
        release release-snapshot \
        docker docker-web docker-ci \
        sync tidy clean

# ── Default ───────────────────────────────────────────

all: help

help:
	@echo ""
	@echo "  tractl build targets  ($(VERSION)  commit=$(COMMIT))"
	@echo ""
	@echo "  Setup"
	@echo "    make setup                 initialize workspace and install git hooks"
	@echo ""
	@echo "  Dependencies"
	@echo "    make deps                  download Go module dependencies"
	@echo "    make deps-frontend         install shared frontend npm dependencies"
	@echo "    make deps-all             install all dependencies"
	@echo "    make install              alias for deps-all"
	@echo "    make workspace             initialize/sync local Go workspace"
	@echo ""
	@echo "  Build"
	@echo "    make build-cli             build CLI binary to ./bin/"
	@echo "    make build-frontend-web    build WASM runtime + web UI bundle"
	@echo "    make build-frontend-desktop  build desktop UI bundle for Wails embed"
	@echo "    make build-wasm            build browser WASM runtime to frontend/public/"
	@echo "    make build-server          build web server binary (cmd/server — needs build-web first)"
	@echo "    make embed-stub            placeholder dist/web for go:embed (vet/test without build-web)"
	@echo "    make build-desktop         build Wails desktop app (production)"
	@echo "    make install-cli           install current build to ~/.local/bin/tractl"
	@echo "    make build                 build everything"
	@echo ""
	@echo "  Development"
	@echo "    make dev-web               build WASM and run web dev server"
	@echo "    make dev-desktop           run desktop app + localhost API"
	@echo "    make dev-cli               run CLI without building"
	@echo "    make dev-localapi          run localhost API on :7428 (web UI dev)"
	@echo ""
	@echo "  Test"
	@echo "    make test                  run all tests"
	@echo "    make test-verbose          run all tests verbose"
	@echo "    make test-cover            run tests with HTML coverage report"
	@echo "    make test-race             run tests with -race detector"
	@echo "    make test-phase1           run Phase 1 validation tests only"
	@echo "    make test-integration      run integration tests"
	@echo "    make test-e2e              run e2e tests (uses httptest, no external services)"
	@echo "    make check-full            vet + fmt-check + lint + all tests + race"
	@echo "    make test-frontend         run frontend unit/component tests"
	@echo "    make test-frontend-e2e     run frontend Playwright tests"
	@echo ""
	@echo "  Quality"
	@echo "    make vet                   go vet"
	@echo "    make fmt                   format all Go code"
	@echo "    make fmt-check             check formatting (CI use)"
	@echo "    make lint                  run golangci-lint"
	@echo "    make vuln                  run govulncheck (stdlib + module vulnerabilities)"
	@echo "    make check                 vet + fmt-check + lint + vuln + test (CI gate)"
	@echo "    make check-full            check + test-race + integration + e2e"
	@echo ""
	@echo "  Release"
	@echo "    make release               goreleaser release (needs GITHUB_TOKEN + git tag)"
	@echo "    make release-snapshot      local snapshot build, nothing published"
	@echo "    make docker                multi-arch OCI image (linux/amd64 + linux/arm64)"
	@echo "    make docker-web            build web server Docker image"
	@echo "    make docker-ci             CI runner Docker image"
	@echo ""
	@echo "  Workspace"
	@echo "    make sync                  go work sync"
	@echo "    make tidy                  tidy all go modules"
	@echo "    make clean                 remove build artifacts"
	@echo ""

# ── Dependencies ──────────────────────────────────────

workspace:
	@root_go_version="$$(awk '/^go / {print $$2; exit}' go.mod)" && \
		if [ -f go.work ]; then \
			env -u GOWORK go work use . $(DESKTOP_DIR); \
		else \
			env -u GOWORK go work init . $(DESKTOP_DIR); \
		fi && \
		env -u GOWORK go work edit -go="$$root_go_version"
	@env -u GOWORK go work sync

deps: workspace
	go mod download

deps-frontend:
	cd $(FRONTEND_DIR) && npm ci

deps-all: deps deps-frontend

install: deps-all

# install-cli puts the current local build into ~/.local/bin — equivalent to
# npm link in Node.js. Useful for testing the CLI as if installed.
install-cli: build-cli
	@mkdir -p ~/.local/bin
	@install -m 755 $(BINARY_OUT) ~/.local/bin/$(BINARY_NAME)
	@echo "✓ tractl installed to ~/.local/bin/$(BINARY_NAME)"
	@~/.local/bin/$(BINARY_NAME) version

# ── Build ─────────────────────────────────────────────

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

# build-cli: produces a fully static binary with version info embedded.
# CGO_ENABLED=0  → no C runtime dependency (runs on any Linux/macOS/Windows)
# -trimpath      → removes your local $GOPATH from debug info (reproducible)
# -ldflags       → injects VERSION/COMMIT/DATE into internal/version variables
build-cli: $(BIN_DIR)
	CGO_ENABLED=$(CGO_ENABLED) go build $(BUILD_FLAGS) -o $(BINARY_OUT) $(CLI_DIR)

# Two separate frontend builds because Wails and the web surface need
# different Vite configs / environment variables.
build-frontend-web: build-wasm
	cd $(FRONTEND_DIR) && npm run build:web

# build-web: convenience alias used by BUILD-005 verification and docker builds.
# Outputs to cmd/server/dist/web/ so cmd/server can embed via //go:embed all:dist/web.
build-web: build-frontend-web

build-frontend-desktop:
	cd $(FRONTEND_DIR) && npm run build:desktop

# build-wasm: cross-compiles the Go engine to WebAssembly.
# GOOS=js GOARCH=wasm tells Go to target the browser WASM runtime instead of
# a native OS. wasm_exec.js is the Go runtime shim that bridges Go ↔ JavaScript.
# Go 1.26 moved the shim from misc/wasm/ to lib/wasm/ — we handle both.
build-wasm:
	@mkdir -p $(FRONTEND_DIR)/public
	GOOS=js GOARCH=wasm CGO_ENABLED=0 go build $(BUILD_FLAGS) \
		-o $(FRONTEND_DIR)/public/tractl.wasm ./cmd/wasm
	@wasm_exec="$$(go env GOROOT)/lib/wasm/wasm_exec.js"; \
		if [ ! -f "$$wasm_exec" ]; then wasm_exec="$$(go env GOROOT)/misc/wasm/wasm_exec.js"; fi; \
		cp "$$wasm_exec" $(FRONTEND_DIR)/public/wasm_exec.js

build-desktop:
	@command -v wails >/dev/null 2>&1 || { \
		echo ""; \
		echo "  wails not found. Install it first:"; \
		echo "    go install github.com/wailsapp/wails/v2/cmd/wails@latest"; \
		echo ""; \
		exit 1; \
	}
	cd $(DESKTOP_DIR) && wails build -ldflags "$(LDFLAGS)" -trimpath

build: build-cli build-frontend-web build-desktop

# ── Development ───────────────────────────────────────

dev: dev-web

dev-web: build-wasm
	cd $(FRONTEND_DIR) && npm run dev:web

dev-frontend-web: dev-web

dev-desktop:
	@command -v wails >/dev/null 2>&1 || { \
		echo ""; \
		echo "  wails not found. Install it first:"; \
		echo "    go install github.com/wailsapp/wails/v2/cmd/wails@latest"; \
		echo ""; \
		exit 1; \
	}
	cd $(DESKTOP_DIR) && wails dev

dev-cli:
	go run $(CLI_DIR)

dev-localapi:
	go run ./cmd/localapi

# ── Test ──────────────────────────────────────────────

test: embed-stub
	go test ./...

test-verbose: embed-stub
	go test -v ./...

test-cover: embed-stub
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report written to coverage.html"

# test-race: runs tests with the Go race detector.
# The race detector instruments memory accesses at runtime and reports
# concurrent reads/writes that could cause data corruption — critical for
# a codebase with goroutines (scheduler, executor, context engine, etc.).
test-race: embed-stub
	go test -race -timeout 120s -count=1 ./...

test-phase1:
	go test -v ./internal/validation/...
	go test -v ./internal/spec/...

test-integration:
	go test -tags integration ./test/integration/...

test-e2e:
	go test -tags e2e -timeout 120s ./test/e2e/...

# ── Server ────────────────────────────────────────────────────────────────────

SERVER_EMBED_DIR    := cmd/server/dist/web
SERVER_EMBED_INDEX  := $(SERVER_EMBED_DIR)/index.html

# embed-stub: minimal dist/web tree so go vet / go test / golangci-lint succeed on a
# fresh clone without running the full frontend build. Overwritten by make build-web.
$(SERVER_EMBED_INDEX):
	@mkdir -p $(SERVER_EMBED_DIR)
	@echo '<!DOCTYPE html><html><head><meta charset="utf-8"><title>tractl</title></head><body><p>Run <code>make build-web</code> for the production UI bundle.</p></body></html>' > $@

.PHONY: embed-stub
embed-stub: $(SERVER_EMBED_INDEX)

# build-server builds the web-serving binary that embeds the frontend.
# Requires: make build-web must have been run first (produces cmd/server/dist/web/).
# If cmd/server/dist/web/ does not exist, go build will fail — this is correct.
build-server:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=$(CGO_ENABLED) go build $(BUILD_FLAGS) \
		-o $(BIN_DIR)/tractl-server ./cmd/server

test-frontend:
	cd $(FRONTEND_DIR) && npm test

test-frontend-e2e:
	cd $(FRONTEND_DIR) && npm run test:e2e

# ── Quality ───────────────────────────────────────────

vet: embed-stub
	go vet ./...

fmt:
	gofmt -w .

# fmt-check: used in CI — reports files that need formatting without modifying them.
# gofmt -l lists files that differ from gofmt output; test -z fails if any listed.
fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "Go files need formatting:"; gofmt -l .; exit 1)

lint: embed-stub
	golangci-lint run ./...

# vuln: scan for known vulnerabilities in dependencies and stdlib (CI + pre-commit).
vuln: embed-stub
	@command -v govulncheck >/dev/null 2>&1 || go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

# check: the minimum CI quality gate.
# Matches the original: vet + fmt-check + lint + vuln + test.
check: vet fmt-check lint vuln test

# check-full: the complete gate before a release or major merge.
check-full: check test-race test-integration test-e2e

# ── Release ───────────────────────────────────────────

# release: full goreleaser release.
# goreleaser cross-compiles for all platforms, packages archives, builds+pushes
# Docker images, creates a GitHub Release, and updates the Homebrew tap.
# Requires: goreleaser installed + GITHUB_TOKEN + a git tag on HEAD.
# Install: brew install goreleaser
release:
	@command -v goreleaser >/dev/null 2>&1 || { \
		echo "goreleaser not found. Install: brew install goreleaser"; exit 1; \
	}
	@[ -n "$(GITHUB_TOKEN)" ] || { echo "GITHUB_TOKEN is not set"; exit 1; }
	goreleaser release --clean

# release-snapshot: builds all release artefacts locally — no git tag, no publish.
# Safe to run anytime to verify the release pipeline before tagging.
release-snapshot:
	@command -v goreleaser >/dev/null 2>&1 || { \
		echo "goreleaser not found. Install: brew install goreleaser"; exit 1; \
	}
	goreleaser release --snapshot --clean

# ── Docker ────────────────────────────────────────────

# docker: builds both web and ci images locally.
docker: docker-web docker-ci

# docker-web builds the web server image that serves the React + WASM app.
# User runs: docker run -p 7428:7428 tractl/tractl:web
# Then opens: http://localhost:7428
docker-web:
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--build-arg VERSION=$(VERSION)      \
		--build-arg COMMIT=$(COMMIT)        \
		-t $(DOCKER_REPO):web-$(DOCKER_TAG) \
		-t $(DOCKER_REPO):web               \
		-f build/docker/Dockerfile.web      \
		.

# docker-ci: CI runner image (tractl binary + bundled OpenAPI provider).
docker-ci:
	docker buildx build \
		--platform linux/amd64,linux/arm64  \
		--build-arg VERSION=$(VERSION)      \
		-t $(DOCKER_REPO):ci-$(DOCKER_TAG)  \
		-t $(DOCKER_REPO):ci                \
		-f build/docker/Dockerfile.ci       \
		.

# ── Setup ─────────────────────────────────────────────

setup: workspace
	@hook_path="$$(git rev-parse --git-path hooks/pre-commit)" && \
		mkdir -p "$$(dirname "$$hook_path")" && \
		cp scripts/pre-commit "$$hook_path" && \
		chmod +x "$$hook_path"
	@echo "✓ git hooks installed"

# ── Workspace ─────────────────────────────────────────

sync: workspace

# tidy: run go mod tidy for BOTH the root module and the desktop module.
# env -u GOWORK temporarily disables workspace mode for each tidy run so each
# module tidies against its own dependency graph, not the workspace graph.
tidy:
	env -u GOWORK go mod tidy
	cd $(DESKTOP_DIR) && env -u GOWORK go mod tidy

# ── Clean ─────────────────────────────────────────────

clean:
	rm -rf $(BIN_DIR)/
	rm -rf $(DESKTOP_DIR)/build/
	rm -rf $(FRONTEND_DIR)/dist/
	rm -rf $(DESKTOP_DIR)/frontend/dist/
	rm -rf cmd/server/dist/
	rm -f $(FRONTEND_DIR)/public/tractl.wasm $(FRONTEND_DIR)/public/wasm_exec.js
	rm -f coverage.out coverage.html
