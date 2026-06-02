window.SURFACES_DATA = {
  "overview": {
    "id": "overview",
    "icon": "\ud83d\udce1",
    "title": "Surfaces Overview",
    "tagline": "Four ways to deliver the same engine \u2014 same spec, same semantics, different transport",
    "description": "traCtl is designed so the core engine (<code>internal/engine</code>) is completely surface-agnostic. Any surface just constructs an <code>engine.Config</code> and calls <code>engine.Run()</code>. The four surfaces differ only in how input arrives (CLI flags, WASM JS call, Wails IPC, HTTP request) and how output is delivered (terminal text, JSON to JS, Wails response, HTTP response).",
    "tags": [
      {
        "label": "Same Engine",
        "cls": "blue"
      },
      {
        "label": "Same Spec Semantics",
        "cls": "green"
      },
      {
        "label": "Different Transport",
        "cls": "purple"
      }
    ],
    "keyFiles": [
      {
        "path": "cmd/tractl/",
        "purpose": "CLI surface \u2014 reads flags, calls engine, formats terminal output"
      },
      {
        "path": "cmd/wasm/",
        "purpose": "WASM surface \u2014 registers Go functions as JS globals in the browser"
      },
      {
        "path": "cmd/desktop/",
        "purpose": "Desktop surface \u2014 Wails app shell, Go methods exposed via IPC"
      },
      {
        "path": "cmd/localapi/ + internal/localapi/",
        "purpose": "Local API surface \u2014 REST server on :7428"
      }
    ],
    "whenToUse": "CLI for CI/CD and scripting. Web WASM for hosted web app (engine in browser). Desktop for native app. Local API for IDE plugins and external tool integration."
  },
  "cli": {
    "id": "cli",
    "icon": "\ud83d\udcbb",
    "title": "CLI Surface",
    "tagline": "tractl run workflow.yaml \u2014 the terminal-first surface for CI/CD and scripting",
    "description": "The CLI is the simplest surface. It reads command-line flags, constructs an <code>engine.Config</code>, calls <code>engine.Run()</code>, formats the result (text or JSON), and exits with code 0, 1, or 2. No HTTP server, no UI, no persistent state. Perfect for CI pipelines, shell scripts, and developer terminal workflows.",
    "tags": [
      {
        "label": "cmd/tractl",
        "cls": "blue"
      },
      {
        "label": "Exit 0/1/2",
        "cls": "green"
      },
      {
        "label": "No frontend",
        "cls": ""
      }
    ],
    "entryPoint": {
      "path": "cmd/tractl/main.go",
      "desc": "Parses flags, dispatches to RunCommand, formats output, calls os.Exit()"
    },
    "howFrontendTalks": "No frontend. The CLI is a pure terminal tool.",
    "keyFiles": [
      {
        "path": "cmd/tractl/main.go",
        "purpose": "Entry point \u2014 flag parsing, command dispatch"
      },
      {
        "path": "cmd/tractl/run_command.go",
        "purpose": "RunCommand implementation \u2014 wires engine.Config from flags"
      },
      {
        "path": "cmd/tractl/output.go",
        "purpose": "Text/JSON output formatting for terminal"
      }
    ],
    "codeExample": "# Run a workflow\ntractl run workflow.yaml\n\n# With options\ntractl run workflow.yaml --verbose --output json\ntractl run workflow.yaml --overlay staging.overlay.yaml\n\n# Exit codes\n# 0 = all assertions passed\n# 1 = assertions failed (workflow ran, checks failed)\n# 2 = pipeline error (parse/validate/plan failed)",
    "whenToUse": "CI/CD pipelines (GitHub Actions, Jenkins), shell scripts, cron jobs, developer terminal workflows. The only surface that doesn't require Node.js or a running server."
  },
  "web-wasm": {
    "id": "web-wasm",
    "icon": "\ud83c\udf10",
    "title": "Web + WASM Surface",
    "tagline": "Two parts: cmd/server serves the SPA; cmd/wasm is the engine compiled to WebAssembly",
    "description": "This surface is actually <strong>two separate binaries</strong> that work together \u2014 a common source of confusion. <code>cmd/server</code> is a plain HTTP server that serves the React SPA (HTML/CSS/JS) as static files. It has no engine logic. <code>cmd/wasm</code> is the Go engine compiled to a <code>.wasm</code> binary that the browser downloads and executes locally. The engine runs <em>inside the user's browser</em> \u2014 there's no backend API call when you run a workflow in the web app.",
    "tags": [
      {
        "label": "cmd/server (host)",
        "cls": "blue"
      },
      {
        "label": "cmd/wasm (engine)",
        "cls": "green"
      },
      {
        "label": "Engine in browser",
        "cls": "purple"
      }
    ],
    "entryPoint": {
      "path": "cmd/server/server.go + cmd/wasm/bridge.go",
      "desc": "server.go hosts static files; bridge.go registers tractl.* JS globals via syscall/js"
    },
    "howFrontendTalks": "The React app calls <code>window.tractl.run(document, format)</code> \u2014 a Go function registered as a JavaScript global. No HTTP round-trip. The engine runs synchronously in the browser's WASM thread.",
    "keyFiles": [
      {
        "path": "cmd/server/server.go",
        "purpose": "Static file server \u2014 serves dist/ + /api/v1/status health check"
      },
      {
        "path": "cmd/wasm/bridge.go",
        "purpose": "Registers Go engine functions as JS globals using syscall/js"
      },
      {
        "path": "frontend/public/wasm_exec.js",
        "purpose": "Go WASM runtime loader (official Go stdlib file)"
      },
      {
        "path": "frontend/src/platform/web/",
        "purpose": "Frontend adapter that calls window.tractl.* functions"
      }
    ],
    "routes": [
      {
        "method": "GET",
        "path": "/",
        "desc": "Serves React SPA (index.html fallback for SPA routing)"
      },
      {
        "method": "GET",
        "path": "/api/v1/status",
        "desc": "Health check \u2014 returns version and surface name"
      }
    ],
    "codeExample": "// cmd/wasm/bridge.go \u2014 register engine as JS global\nfunc main() {\n    js.Global().Set(\"tractl\", js.ValueOf(map[string]interface{}{\n        \"run\":      js.FuncOf(runHandler),\n        \"parse\":    js.FuncOf(parseHandler),\n        \"validate\": js.FuncOf(validateHandler),\n        \"version\":  js.FuncOf(versionHandler),\n    }))\n    select {} // keep WASM alive\n}",
    "whenToUse": "Hosted web application. Users get the full UI without installing anything. The engine runs in their browser \u2014 no server-side execution, no data leaves the client."
  },
  "desktop": {
    "id": "desktop",
    "icon": "\ud83d\udda5\ufe0f",
    "title": "Desktop Surface (Wails)",
    "tagline": "Native app \u2014 React UI in a native OS window, engine runs in the same Go process",
    "description": "Built with Wails v2, which wraps a WebView (native browser component) in an OS window and exposes Go methods to JavaScript via a custom IPC bridge (not WASM, not HTTP). The React frontend is the same codebase as the web surface, built with <code>vite build --mode desktop</code>. The engine runs in the same Go process as the window, making it faster than the local API approach and offline-capable.",
    "tags": [
      {
        "label": "cmd/desktop",
        "cls": "blue"
      },
      {
        "label": "Wails v2",
        "cls": "green"
      },
      {
        "label": "Native Window",
        "cls": "purple"
      },
      {
        "label": "Offline",
        "cls": "orange"
      }
    ],
    "entryPoint": {
      "path": "cmd/desktop/app.go",
      "desc": "Wails App struct \u2014 defines Go methods exposed to frontend via window.wails IPC"
    },
    "howFrontendTalks": "JavaScript calls <code>window.wails.MethodName(args)</code>. Wails serialises args to JSON, calls the Go method, serialises the result back to JSON, and resolves the JS Promise. No HTTP server involved.",
    "keyFiles": [
      {
        "path": "cmd/desktop/app.go",
        "purpose": "App struct with bound Go methods (Run, Parse, Validate, etc.)"
      },
      {
        "path": "cmd/desktop/wails.json",
        "purpose": "Wails config: frontend paths, window dimensions, app metadata"
      },
      {
        "path": "frontend/src/platform/desktop/",
        "purpose": "Frontend adapter that calls window.wails.* methods"
      }
    ],
    "codeExample": "// cmd/desktop/app.go \u2014 bound Go methods\ntype App struct {\n    ctx context.Context\n    eng *engine.Engine\n}\n\n// This becomes window.wails.Run(...) in JavaScript\nfunc (a *App) Run(document, format string) string {\n    result := a.eng.RunDocument(document, format)\n    b, _ := json.Marshal(result)\n    return string(b)\n}",
    "whenToUse": "When users want a native app experience (system tray, native file dialogs, offline). Faster than the local API surface because there's no HTTP serialisation layer."
  },
  "localapi": {
    "id": "localapi",
    "icon": "\ud83d\udd0c",
    "title": "Local API Surface",
    "tagline": "REST daemon on :7428 \u2014 for external tool integration, IDE plugins, and advanced frontend scenarios",
    "description": "A full REST API server running on localhost port 7428. The engine runs server-side in the Go process. The React frontend can connect to this instead of WASM when running in <em>Tier 2 mode</em>. Unlike the WASM surface, the local API persists workspace state across page refreshes and can access the local filesystem directly. Primary use cases: IDE extensions, external scripts calling the engine via HTTP, and the web frontend needing file system access.",
    "tags": [
      {
        "label": "internal/localapi",
        "cls": "blue"
      },
      {
        "label": "localhost:7428",
        "cls": "green"
      },
      {
        "label": "No auth (trusted local)",
        "cls": "yellow"
      }
    ],
    "entryPoint": {
      "path": "internal/localapi/server.go",
      "desc": "HTTP server setup \u2014 routes, middleware (CORS, logging), handler registration"
    },
    "howFrontendTalks": "Standard <code>fetch('http://localhost:7428/api/v1/...')</code> HTTP calls. The frontend detects if the local API is running and switches from WASM mode to local API mode automatically.",
    "keyFiles": [
      {
        "path": "internal/localapi/server.go",
        "purpose": "Server setup, route registration, CORS middleware"
      },
      {
        "path": "internal/localapi/handlers.go",
        "purpose": "HTTP handlers for each route"
      },
      {
        "path": "internal/localapi/store.go",
        "purpose": "Workspace state persistence (file list, recent runs)"
      },
      {
        "path": "frontend/src/platform/localApi/",
        "purpose": "Frontend adapter: fetch() calls to :7428"
      }
    ],
    "routes": [
      {
        "method": "GET",
        "path": "/api/v1/status",
        "desc": "Health check \u2014 version, surface, runtime info"
      },
      {
        "method": "GET",
        "path": "/api/v1/files",
        "desc": "List files in the workspace directory"
      },
      {
        "method": "GET",
        "path": "/api/v1/files/{path...}",
        "desc": "Read a file's content from the workspace"
      },
      {
        "method": "POST",
        "path": "/api/v1/run",
        "desc": "Execute a workflow from a file path on disk"
      },
      {
        "method": "POST",
        "path": "/api/v1/workflows/run",
        "desc": "Execute a workflow from inline document (request body)"
      },
      {
        "method": "GET",
        "path": "/api/v1/workflow/layout",
        "desc": "Compute DAG layout coordinates for the canvas UI"
      },
      {
        "method": "GET",
        "path": "/api/v1/workflow/infer-deps",
        "desc": "Auto-detect step dependencies from variable references"
      },
      {
        "method": "GET",
        "path": "/api/v1/engine/defaults",
        "desc": "Engine default values (concurrency limits, timeout, etc.)"
      },
      {
        "method": "GET",
        "path": "/api/v1/workspace/status",
        "desc": "Workspace metadata: path, name, recent files"
      }
    ],
    "whenToUse": "When you need filesystem access from the web frontend (Tier 2 mode). For IDE plugin integration. For external scripts/tools that want to call the engine via HTTP. For advanced scenarios where WASM's in-browser execution is not sufficient."
  }
};
