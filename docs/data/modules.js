window.MODULES_DATA = {
  "pipeline": {
    "id": "pipeline",
    "title": "Execution Pipeline",
    "path": "internal/engine",
    "category": "critical",
    "tagline": "The master 6-stage orchestrator \u2014 parse \u2192 overlay \u2192 validate \u2192 plan \u2192 compile \u2192 execute",
    "description": "The engine package is the only entry point callers should use. <code>Engine.Run(cfg)</code> and <code>Engine.RunDocument()</code> wire all six pipeline stages together, catch every error into a structured <code>RunResult</code> (never panics), and return exit codes 0/1/2. Everything else in the codebase is a stage implementation that the engine orchestrates. Think of it as the Express.js request pipeline \u2014 each stage transforms the data and hands it to the next.",
    "keyTypes": [
      {
        "name": "Engine",
        "kind": "struct",
        "purpose": "Top-level orchestrator. Holds injected dependencies (Planner, Compiler, Evaluator, ScriptRunner)."
      },
      {
        "name": "Config",
        "kind": "struct",
        "purpose": "Input to Engine.Run(). Carries file path, env name, overlay files, trace ID, verbosity flags, and optional mock overrides for testing."
      },
      {
        "name": "RunResult",
        "kind": "struct",
        "purpose": "Output of every run. Contains Passed bool, per-workflow outcomes, and error strings for each stage that can fail."
      },
      {
        "name": "WorkflowOutcome",
        "kind": "struct",
        "purpose": "Per-workflow result inside RunResult. Steps, assertions, diagnostics."
      }
    ],
    "interfaces": [
      {
        "name": "Planner",
        "desc": "Accepts a parsed spec, returns an ExecutionPlan. The real implementation does DAG topological sort. Tests inject stubs.",
        "methods": "type Planner interface {\n    Plan(s *spec.TraCtlSpec) (*planner.ExecutionPlan, error)\n}"
      },
      {
        "name": "Compiler",
        "desc": "Converts an ExecutionPlan + spec into a CompiledPlan (regex pre-compiled, runtimes selected).",
        "methods": "type Compiler interface {\n    Compile(plan *planner.ExecutionPlan, s *spec.TraCtlSpec) (*compiler.CompiledPlan, error)\n}"
      },
      {
        "name": "Evaluator",
        "desc": "Evaluates a single compiled assertion against a StepResult. Returns AssertionResult.",
        "methods": "type Evaluator interface {\n    EvaluateCompiled(a compiler.CompiledAssertion, result *runtime.StepResult, ...) (assertion.AssertionResult, error)\n}"
      },
      {
        "name": "ScriptRunner",
        "desc": "Runs a JavaScript string in a sandbox, returns a MutationSet of variable changes.",
        "methods": "type ScriptRunner interface {\n    Execute(source string, frozen runtime.FrozenContext, seed int64) (sandbox.MutationSet, error)\n}"
      }
    ],
    "goConceptsUsed": [
      "Interfaces (DI)",
      "Error wrapping",
      "Struct embedding",
      "context.Context"
    ],
    "nodeAnalogy": "Like an Express.js middleware chain \u2014 each stage is a middleware, <code>RunResult</code> is the final response object. Interfaces are like TypeScript interfaces for dependency injection in NestJS.",
    "codeExample": "// Calling the engine from any surface\ncfg := engine.Config{\n    WorkflowFile: \"workflow.yaml\",\n    Verbose:      true,\n}\nresult := engine.Run(cfg)\nif !result.Passed {\n    fmt.Println(result.ValidationError) // or ParseError, PlanError\n    os.Exit(2)\n}",
    "relatedModules": [
      {
        "id": "parser",
        "name": "Parser",
        "rel": "Stage 1 \u2014 parse file to *TraCtlSpec"
      },
      {
        "id": "planner",
        "name": "Planner",
        "rel": "Stage 4 \u2014 build ExecutionPlan DAG"
      },
      {
        "id": "compiler",
        "name": "Compiler",
        "rel": "Stage 5 \u2014 compile plan to executable"
      },
      {
        "id": "scheduler",
        "name": "Scheduler",
        "rel": "Stage 6 \u2014 run workflows in parallel"
      }
    ]
  },
  "parser": {
    "id": "parser",
    "title": "Parser",
    "path": "internal/parser",
    "category": "critical",
    "tagline": "Format-agnostic parser registry \u2014 YAML, JSON, TOON \u2192 TraCtlSpec",
    "description": "Detects format from file extension and delegates to the matching sub-parser (<code>internal/parser/yaml</code>, <code>/json</code>, <code>/toon</code>). Each sub-parser: validates format syntax \u2192 unmarshals \u2192 assigns ULIDs to entities \u2192 returns <code>*spec.TraCtlSpec</code>. Parsers do NOT call the canonical validator \u2014 that is the caller's (engine's) responsibility. Adding a new format means implementing the <code>Parser</code> interface in a new sub-package and registering it.",
    "keyTypes": [
      {
        "name": "Parser",
        "kind": "interface",
        "purpose": "One method: Parse([]byte, sourceRef) (*spec.TraCtlSpec, error). Each format implements this."
      },
      {
        "name": "Registry",
        "kind": "struct",
        "purpose": "Maps file extensions to Parser implementations. Detects format automatically."
      },
      {
        "name": "TraCtlSpec",
        "kind": "struct",
        "purpose": "The canonical Go representation of a workflow document. Lives in internal/spec."
      }
    ],
    "interfaces": [
      {
        "name": "Parser",
        "desc": "Implemented once per format. The registry delegates to the correct one based on file extension.",
        "methods": "type Parser interface {\n    Parse(src []byte, sourceRef string) (*spec.TraCtlSpec, error)\n}"
      }
    ],
    "goConceptsUsed": [
      "Interfaces",
      "Struct tags (YAML/JSON)",
      "ULID assignment"
    ],
    "nodeAnalogy": "Like a multer/busboy parser that detects content-type and delegates to the right handler. Each sub-parser is like a content-type handler that normalises input into the same output shape.",
    "codeExample": "// Three parsers, same interface\nvar (\n    _ Parser = (*yamlparser.Parser)(nil)  // compile-time check\n    _ Parser = (*jsonparser.Parser)(nil)\n    _ Parser = (*toonparser.Parser)(nil)\n)",
    "relatedModules": [
      {
        "id": "validation",
        "name": "Validation",
        "rel": "Runs after parsing to validate the spec"
      },
      {
        "id": "overlay",
        "name": "Overlay",
        "rel": "Applied to parsed spec before validation"
      },
      {
        "id": "pipeline",
        "name": "Pipeline",
        "rel": "Engine calls parser as Stage 1"
      }
    ]
  },
  "validation": {
    "id": "validation",
    "title": "Validation",
    "path": "internal/validation",
    "category": "critical",
    "tagline": "Schema and structural validation \u2014 runs before any execution",
    "description": "Validates the parsed spec for semantic correctness: required fields present, valid step kinds, <code>dependsOn</code> references exist, DAG has no cycles, assertion/extract IDs unique. Returns a list of <code>ValidationError</code> \u2014 never mutates the spec. Five validators are built independently so their import graphs never cross (architectural constraint). Validation always runs post-overlay so the final merged document is what gets checked.",
    "keyTypes": [
      {
        "name": "ValidationError",
        "kind": "struct",
        "purpose": "One error per violation. Contains path, rule ID, and human-readable message."
      },
      {
        "name": "SpecValidator",
        "kind": "struct",
        "purpose": "Main validator. Runs 12+ rules against a TraCtlSpec."
      }
    ],
    "goConceptsUsed": [
      "Slice of errors",
      "Recursive DFS for cycle detection",
      "Visitor pattern"
    ],
    "nodeAnalogy": "Like Joi or Zod schema validation, but operating on the already-parsed Go struct (not raw JSON). Returns a list of all errors, not just the first one.",
    "codeExample": "// Validation result\nerrs := validator.Validate(spec)\nif len(errs) > 0 {\n    for _, e := range errs {\n        fmt.Printf(\"[%s] %s: %s\\n\", e.Code, e.Path, e.Message)\n    }\n    os.Exit(2)\n}",
    "relatedModules": [
      {
        "id": "parser",
        "name": "Parser",
        "rel": "Validation runs after parsing"
      },
      {
        "id": "overlay",
        "name": "Overlay",
        "rel": "Validation runs after overlay application"
      },
      {
        "id": "planner",
        "name": "Planner",
        "rel": "Planning runs after validation passes"
      }
    ]
  },
  "planner": {
    "id": "planner",
    "title": "Planner",
    "path": "internal/planner",
    "category": "critical",
    "tagline": "Builds the execution DAG \u2014 topological sort, capability resolution, wave ordering",
    "description": "Takes the validated spec and produces an <code>ExecutionPlan</code> \u2014 an ordered list of step 'waves' within each workflow where steps in the same wave have no dependencies between them and can run in parallel. Uses DFS-based topological sort on the <code>dependsOn</code> edges. Also resolves <code>capabilities</code> (checks the engine supports the declared protocols). Cycle detection happens here as a secondary check (primary is validation).",
    "keyTypes": [
      {
        "name": "ExecutionPlan",
        "kind": "struct",
        "purpose": "Output of planning. Contains ordered WorkflowPlans, each with steps in execution wave order."
      },
      {
        "name": "WorkflowPlan",
        "kind": "struct",
        "purpose": "Per-workflow plan: concurrency limit, failure policy, ordered steps."
      },
      {
        "name": "StepPlan",
        "kind": "struct",
        "purpose": "Per-step plan: resolved dependencies, determined wave number, selected capability."
      }
    ],
    "goConceptsUsed": [
      "DFS topological sort",
      "Graph cycle detection",
      "Map for adjacency list"
    ],
    "nodeAnalogy": "Like how npm resolves package dependency order before installing \u2014 topological sort ensures dependencies install before the packages that need them. The 'waves' are like Webpack chunks that can be loaded in parallel.",
    "codeExample": "// Topological sort concept\n// Steps A, B (no deps) \u2192 C (depends on A) \u2192 D (depends on B, C)\n// Wave 1: A, B (parallel)\n// Wave 2: C (waits for A)\n// Wave 3: D (waits for B and C)",
    "relatedModules": [
      {
        "id": "compiler",
        "name": "Compiler",
        "rel": "Compiler takes ExecutionPlan as input"
      },
      {
        "id": "validation",
        "name": "Validation",
        "rel": "Validation must pass before planning runs"
      },
      {
        "id": "pipeline",
        "name": "Pipeline",
        "rel": "Stage 4 in the engine pipeline"
      }
    ]
  },
  "compiler": {
    "id": "compiler",
    "title": "Compiler",
    "path": "internal/compiler",
    "category": "critical",
    "tagline": "Converts ExecutionPlan \u2192 CompiledPlan: pre-compiles regex, selects runtimes, binds payloads",
    "description": "The compiler makes the plan truly executable. It selects the correct runtime for each step (http, script, etc.), pre-compiles all regex assertion patterns (<code>regexp.Compile()</code>) so they're not re-created per request, binds payload schemas, and produces a <code>CompiledPlan</code>. After this stage, no more spec lookups are needed \u2014 the compiled plan is self-contained. This separation (plan vs compile) follows ADR-003.",
    "keyTypes": [
      {
        "name": "CompiledPlan",
        "kind": "struct",
        "purpose": "Fully executable plan. No more spec lookups needed after compilation."
      },
      {
        "name": "CompiledWorkflow",
        "kind": "struct",
        "purpose": "Per-workflow compiled data including all compiled steps."
      },
      {
        "name": "CompiledStep",
        "kind": "struct",
        "purpose": "Per-step: selected runtime, pre-compiled regex, resolved request descriptor."
      },
      {
        "name": "CompiledAssertion",
        "kind": "struct",
        "purpose": "Pre-compiled assertion including compiled regexp.Regexp pattern if kind=matches."
      }
    ],
    "goConceptsUsed": [
      "regexp.MustCompile",
      "Struct copying",
      "Type switching for runtime selection"
    ],
    "nodeAnalogy": "Like webpack compilation \u2014 takes a module graph (ExecutionPlan) and produces an optimised bundle (CompiledPlan) where all paths are resolved and heavy processing (regex compilation) is done once upfront.",
    "codeExample": "// Assertions with kind=matches have their pattern pre-compiled\ntype CompiledAssertion struct {\n    AssertionID     string\n    Kind            string\n    Op              string\n    Expected        interface{}\n    CompiledPattern *regexp.Regexp  // set if Op == \"matches\"\n    Severity        string\n}",
    "relatedModules": [
      {
        "id": "planner",
        "name": "Planner",
        "rel": "Compiler takes ExecutionPlan from planner"
      },
      {
        "id": "scheduler",
        "name": "Scheduler",
        "rel": "Scheduler runs the CompiledPlan"
      },
      {
        "id": "assertion",
        "name": "Assertion",
        "rel": "Uses CompiledAssertion at evaluation time"
      }
    ]
  },
  "scheduler": {
    "id": "scheduler",
    "title": "Scheduler",
    "path": "internal/scheduler",
    "category": "critical",
    "tagline": "Runs workflows in parallel, executes steps in DAG wave order with concurrency control",
    "description": "Runs multiple workflows concurrently (up to <code>WorkflowConcurrency</code> limit). Within each workflow, fires steps as their dependencies complete \u2014 using goroutines and a semaphore to respect the <code>concurrency</code> limit. Collects <code>WorkflowResult</code> for each. Applies the workflow's <code>failurePolicy</code>: <code>resilient</code> (default) continues running unblocked steps even if some fail; <code>failFast</code> cancels all remaining steps on first failure.",
    "keyTypes": [
      {
        "name": "Scheduler",
        "kind": "struct",
        "purpose": "Top-level parallel workflow runner."
      },
      {
        "name": "WorkflowResult",
        "kind": "struct",
        "purpose": "Per-workflow outcome: step results, timing, pass/fail."
      },
      {
        "name": "StepOutcome",
        "kind": "struct",
        "purpose": "Per-step final state and any assertion failures."
      }
    ],
    "goConceptsUsed": [
      "Goroutines",
      "sync.WaitGroup",
      "Buffered channels (semaphore)",
      "context.Context cancellation"
    ],
    "nodeAnalogy": "Like <code>p-limit</code> + <code>Promise.all()</code> where each 'promise' is a Go goroutine. The semaphore (buffered channel) is exactly <code>p-limit(N)</code>. Dependency tracking is like <code>Promise.all([dep1, dep2])</code> before firing the next step.",
    "codeExample": "// Semaphore pattern (buffered channel as p-limit)\nsem := make(chan struct{}, maxConcurrency)\n\ngo func(step CompiledStep) {\n    sem <- struct{}{}        // acquire slot (blocks if full)\n    defer func() { <-sem }() // release on done\n    executeStep(step)\n}(step)",
    "relatedModules": [
      {
        "id": "executor",
        "name": "Executor",
        "rel": "Scheduler calls Executor for each step"
      },
      {
        "id": "runtime",
        "name": "Runtime",
        "rel": "Scheduler creates and manages ExecutionContext"
      },
      {
        "id": "compiler",
        "name": "Compiler",
        "rel": "Scheduler runs a CompiledPlan"
      }
    ]
  },
  "executor": {
    "id": "executor",
    "title": "Executor",
    "path": "internal/executor",
    "category": "critical",
    "tagline": "HTTP request execution with retry, timing waterfall, and response capture",
    "description": "Builds the actual <code>http.Request</code> from a <code>CompiledStep</code> (resolving variable expressions in URLs, headers, body), executes with retry+backoff policy, captures response body/headers/status code, and attaches an HTTP trace for precise millisecond timing (DNS lookup, TCP connect, TLS handshake, time-to-first-byte). Authorization headers are masked in all diagnostic output. Returns a <code>*StepResult</code>.",
    "keyTypes": [
      {
        "name": "HTTPExecutor",
        "kind": "struct",
        "purpose": "Concrete implementation of Executor interface. Manages HTTP client lifecycle."
      },
      {
        "name": "RequestTimeline",
        "kind": "struct",
        "purpose": "Per-request timing breakdown: DNS, TCP, TLS, TTFB, total duration."
      },
      {
        "name": "RetryState",
        "kind": "struct",
        "purpose": "Tracks attempt count and accumulated wait time across retries."
      }
    ],
    "interfaces": [
      {
        "name": "Executor",
        "desc": "The interface the scheduler calls. Makes testing easy \u2014 inject a mock that returns pre-canned StepResults.",
        "methods": "type Executor interface {\n    Execute(ctx context.Context, step *compiler.CompiledStep, execCtx *runtime.ExecutionContext) (*runtime.StepResult, error)\n}"
      }
    ],
    "goConceptsUsed": [
      "net/http/httptrace",
      "context.WithTimeout",
      "exponential backoff",
      "HTTP client pooling"
    ],
    "nodeAnalogy": "Like <code>axios</code> with a retry interceptor + request timing instrumentation. The HTTP trace attachment is like axios interceptors that measure DNS/TCP/TLS timing \u2014 same concept, Go stdlib version.",
    "codeExample": "// HTTP trace captures precise timing\ntrace := &httptrace.ClientTrace{\n    DNSStart:          func(info httptrace.DNSStartInfo) { ... },\n    ConnectDone:       func(net, addr string, err error) { ... },\n    GotFirstResponseByte: func() { ... },\n}\nctx = httptrace.WithClientTrace(ctx, trace)",
    "relatedModules": [
      {
        "id": "runtime",
        "name": "Runtime",
        "rel": "Writes StepResult into ExecutionContext"
      },
      {
        "id": "assertion",
        "name": "Assertion",
        "rel": "Evaluates assertions after execution"
      },
      {
        "id": "extract",
        "name": "Extract",
        "rel": "Extracts variables from response"
      }
    ]
  },
  "sandbox": {
    "id": "sandbox",
    "title": "JS Sandbox",
    "path": "internal/sandbox",
    "category": "support",
    "tagline": "Safe JavaScript execution via goja \u2014 no I/O, 5-second timeout, mutation-set pattern",
    "description": "Runs user-supplied JavaScript hooks (beforeStep, afterStep, transform, when conditions) inside a pure-Go JavaScript engine (goja \u2014 no V8, no Node.js, no native bindings). Scripts receive an immutable <code>FrozenContext</code> snapshot. The only injected APIs are <code>tractl.setVar()</code>, <code>tractl.cancel()</code>, <code>tractl.random()</code>. Scripts cannot fetch URLs, read files, or require modules. A goroutine interrupts the VM after 5 seconds. Script mutations are returned as a <code>MutationSet</code> and applied back to <code>ExecutionContext</code> under a mutex.",
    "keyTypes": [
      {
        "name": "Sandbox",
        "kind": "struct",
        "purpose": "Wraps a goja runtime. Created per script execution."
      },
      {
        "name": "MutationSet",
        "kind": "struct",
        "purpose": "Collected changes from a script run: variable writes by scope."
      },
      {
        "name": "SafeAPI",
        "kind": "struct",
        "purpose": "The tractl.* object injected into each script VM."
      }
    ],
    "goConceptsUsed": [
      "goja (pure-Go JS engine)",
      "Goroutine timeout via interrupt",
      "Deep copy (FrozenContext)",
      "sync.Mutex"
    ],
    "nodeAnalogy": "Like running user code in a vm.Script (Node.js vm module) with a strict sandbox \u2014 no global access, custom context only. The 5s timeout is like <code>AbortSignal.timeout(5000)</code> applied to the VM.",
    "codeExample": "// Script sees this API only \u2014 nothing else\nconst tractl = {\n    setVar: (name, value) => { /* queues mutation */ },\n    cancel: (reason) => { /* marks step cancelled */ },\n    random: () => 0.42,  // deterministic per traceID seed\n    now:    () => '2024-01-01T00:00:00Z'  // deterministic\n}",
    "relatedModules": [
      {
        "id": "runtime",
        "name": "Runtime",
        "rel": "FrozenContext comes from ExecutionContext.Freeze()"
      },
      {
        "id": "pipeline",
        "name": "Pipeline",
        "rel": "Engine injects ScriptRunner (sandbox) into Config"
      }
    ]
  },
  "assertion": {
    "id": "assertion",
    "title": "Assertion",
    "path": "internal/assertion",
    "category": "support",
    "tagline": "Evaluates compiled assertions against HTTP responses \u2014 status, body, header, regex",
    "description": "Takes a <code>CompiledAssertion</code> and a <code>StepResult</code> and returns <code>AssertionResult</code> with passed/failed and a human-readable message. Supports kinds: <code>status</code> (HTTP status code), <code>body</code> (JSONPath via gjson), <code>header</code> (normalised header value), <code>timeout</code> (duration check). Operators: <code>equals</code>, <code>contains</code>, <code>matches</code> (pre-compiled regex), <code>exists</code>. Severity <code>warning</code> is recorded but doesn't fail the step.",
    "keyTypes": [
      {
        "name": "AssertionEvaluator",
        "kind": "struct",
        "purpose": "Main evaluator. Takes compiled assertion + step result, returns AssertionResult."
      },
      {
        "name": "AssertionResult",
        "kind": "struct",
        "purpose": "Outcome: passed bool, message, severity, assertion ID."
      },
      {
        "name": "EvalEvent",
        "kind": "struct",
        "purpose": "Diagnostic event emitted per assertion \u2014 feeds the diagnostics collector."
      }
    ],
    "goConceptsUsed": [
      "gjson (JSONPath)",
      "regexp pre-compiled from compiler stage",
      "switch on kind/op"
    ],
    "nodeAnalogy": "Like Chai.js assertions (<code>expect(status).to.equal(200)</code>) but operating on a Go struct instead of live response. The JSONPath evaluation is like using lodash.get() on the parsed response body.",
    "codeExample": "// Three assertion kinds\n// Status: res.status == 200\n// Body:   gjson.Get(body, \"user.name\") == \"Alice\"\n// Header: res.headers[\"content-type\"] contains \"json\"\n// Regex:  gjson.Get(body, \"id\") matches \"^[0-9a-f-]{36}$\"",
    "relatedModules": [
      {
        "id": "compiler",
        "name": "Compiler",
        "rel": "CompiledAssertion comes from compiler (regex pre-compiled)"
      },
      {
        "id": "executor",
        "name": "Executor",
        "rel": "Executor calls assertion after HTTP response"
      },
      {
        "id": "extract",
        "name": "Extract",
        "rel": "Extraction and assertion both operate on StepResult"
      }
    ]
  },
  "extract": {
    "id": "extract",
    "title": "Extract",
    "path": "internal/extract",
    "category": "support",
    "tagline": "Pulls values from HTTP responses into variable scopes \u2014 the step-chaining mechanism",
    "description": "After each successful HTTP execution, the extract engine reads configured extracts and writes their values into <code>ExecutionContext</code> at the declared scope (step/workflow/spec). Source can be <code>body</code> (JSONPath via gjson), <code>header</code>, or <code>status</code>. The extracted value is available to later steps via <code>${steps.stepId.extracts.varName}</code>. This is the primary mechanism for chaining steps \u2014 e.g. extract a JWT from a login response, use it in the Authorization header of the next request.",
    "keyTypes": [
      {
        "name": "ExtractEngine",
        "kind": "struct",
        "purpose": "Evaluates all extracts for a step and writes to ExecutionContext."
      },
      {
        "name": "ExtractResult",
        "kind": "struct",
        "purpose": "Per-extract outcome: resolved bool, scope, as-name. Value is redacted in logs."
      }
    ],
    "goConceptsUsed": [
      "gjson for JSONPath",
      "scope-aware writes to ExecutionContext",
      "sync.RWMutex (via ExecutionContext)"
    ],
    "nodeAnalogy": "Like express-session or cookie-parser \u2014 extracts specific values from a response and makes them available to subsequent middleware/handlers. The 'scope' is like variable scoping: local (step) vs module (workflow) vs global (spec).",
    "codeExample": "# YAML \u2014 extract JWT, use in next step\nextracts:\n  - id: get-token\n    source: body\n    path: access_token   # JSONPath\n    as: jwt\n    scope: workflow      # available to all later steps\n\n# Later step uses it:\nrequest:\n  headers:\n    Authorization: \"Bearer ${steps.login.extracts.jwt}\"",
    "relatedModules": [
      {
        "id": "runtime",
        "name": "Runtime",
        "rel": "Writes extracted vars into ExecutionContext"
      },
      {
        "id": "assertion",
        "name": "Assertion",
        "rel": "Both operate on StepResult post-execution"
      },
      {
        "id": "executor",
        "name": "Executor",
        "rel": "Executor triggers extraction after HTTP response"
      }
    ]
  },
  "overlay": {
    "id": "overlay",
    "title": "Overlay",
    "path": "internal/overlay",
    "category": "support",
    "tagline": "Merges patch files onto the parsed spec before validation \u2014 environment-specific overrides",
    "description": "Applies overlay documents to the canonical spec using path, match, and source targeting. Supports merge actions: <code>replace</code>, <code>deepMerge</code>, <code>append</code>, <code>appendUnique</code>, <code>remove</code>. The engine operates on a JSON-clone (<code>map[string]any</code>) so it never imports <code>internal/spec</code> \u2014 architectural constraint preserved. Validation always runs <em>after</em> overlay so the final merged document is validated. Used for environment-specific overrides (different base URLs, different auth tokens) without forking the spec file.",
    "keyTypes": [
      {
        "name": "OverlayDocument",
        "kind": "struct",
        "purpose": "A parsed overlay file: metadata + list of Patches."
      },
      {
        "name": "Patch",
        "kind": "struct",
        "purpose": "One change operation: Target (what to select) + Action (how to merge) + Data (what to set)."
      },
      {
        "name": "Target",
        "kind": "struct",
        "purpose": "Selects what to patch: JSON pointer path, semantic match (by ID), or source type."
      }
    ],
    "goConceptsUsed": [
      "JSON pointer (RFC 6901)",
      "Deep merge on map[string]any",
      "Provenance tracking"
    ],
    "nodeAnalogy": "Like JSON Patch (RFC 6902) but with semantic selectors. Similar to how Kustomize overlays work in Kubernetes \u2014 base file stays untouched, overlay files describe what to change for each environment.",
    "codeExample": "# overlay-staging.yaml \u2014 change base URL for staging\npatches:\n  - target:\n      match:\n        type: variables\n    action: deepMerge\n    data:\n      baseUrl: \"https://staging.api.example.com\"",
    "relatedModules": [
      {
        "id": "parser",
        "name": "Parser",
        "rel": "Overlay runs on the parsed spec"
      },
      {
        "id": "validation",
        "name": "Validation",
        "rel": "Validation runs after overlay"
      }
    ]
  },
  "runtime": {
    "id": "runtime",
    "title": "Runtime / ExecutionContext",
    "path": "internal/runtime",
    "category": "support",
    "tagline": "Mutable shared state during a workflow run \u2014 variable scopes, step states, mutex-guarded",
    "description": "<code>ExecutionContext</code> is the single source of truth during execution. It holds all variable scopes (spec/workflow/env/step), step states (pending/running/succeeded/failed), step results (HTTP responses, assertion outcomes), and the active step ID during hook execution. All maps are protected by a <code>sync.RWMutex</code> because multiple goroutines (concurrent steps) read/write simultaneously. <code>Freeze()</code> produces an immutable deep-copy snapshot for script execution.",
    "keyTypes": [
      {
        "name": "ExecutionContext",
        "kind": "struct",
        "purpose": "The live mutable state bag for one workflow run. Thread-safe via sync.RWMutex."
      },
      {
        "name": "FrozenContext",
        "kind": "struct",
        "purpose": "Immutable deep copy of ExecutionContext. Passed to JS sandbox scripts."
      },
      {
        "name": "StepResult",
        "kind": "struct",
        "purpose": "Per-step outcome: status, headers, body, extracts, timeline, assertions."
      },
      {
        "name": "StepState",
        "kind": "string enum",
        "purpose": "pending | running | succeeded | failed | skipped | cancelled"
      }
    ],
    "goConceptsUsed": [
      "sync.RWMutex",
      "Deep copy (map cloning)",
      "Enum via const iota",
      "Getter/setter methods"
    ],
    "nodeAnalogy": "Like a Redux store that's shared across concurrent workers \u2014 the mutex is the critical difference from Node.js where a plain object would work (single thread). The FrozenContext is like an Immer.js draft snapshot \u2014 immutable copy that scripts work on.",
    "codeExample": "// Thread-safe write (Step result stored)\nfunc (ctx *ExecutionContext) SetStepResult(id string, result *StepResult) {\n    ctx.mu.Lock()\n    defer ctx.mu.Unlock()\n    ctx.stepResults[id] = result\n}\n\n// Thread-safe read\nfunc (ctx *ExecutionContext) GetStepResult(id string) (*StepResult, bool) {\n    ctx.mu.RLock()  // multiple readers OK simultaneously\n    defer ctx.mu.RUnlock()\n    result, ok := ctx.stepResults[id]\n    return result, ok\n}",
    "relatedModules": [
      {
        "id": "sandbox",
        "name": "Sandbox",
        "rel": "Scripts get FrozenContext from ExecutionContext.Freeze()"
      },
      {
        "id": "extract",
        "name": "Extract",
        "rel": "Writes extracted vars into ExecutionContext"
      },
      {
        "id": "scheduler",
        "name": "Scheduler",
        "rel": "Creates and owns ExecutionContext per workflow run"
      }
    ]
  },
  "diagnostics": {
    "id": "diagnostics",
    "title": "Diagnostics",
    "path": "internal/diagnostics",
    "category": "support",
    "tagline": "Execution tracing \u2014 waterfall timings, per-step records, deterministic with fixed TraceID",
    "description": "Collects an <code>ExecutionRecord</code> during a run: per-workflow timing, per-step timing (including wait-for-dependency duration), per-assertion outcomes, HTTP waterfall. All diagnostic collection is driven by event hooks emitted from each pipeline stage (ADR-014). With a fixed <code>TraceID</code>, the seed is deterministic \u2014 identical executions produce identical diagnostic output, useful for test assertions on timing behaviour.",
    "keyTypes": [
      {
        "name": "ExecutionRecord",
        "kind": "struct",
        "purpose": "Top-level trace for one run. Contains WorkflowRecords + DiagnosticsEvents."
      },
      {
        "name": "WorkflowRecord",
        "kind": "struct",
        "purpose": "Per-workflow: start time, duration, StepRecords, outcome."
      },
      {
        "name": "StepRecord",
        "kind": "struct",
        "purpose": "Per-step: timing, wait duration (dependency wait), assertion records, request records."
      }
    ],
    "goConceptsUsed": [
      "Event-driven collection (ADR-014 hooks)",
      "Deterministic seeding via traceID hash",
      "Time.Duration arithmetic"
    ],
    "nodeAnalogy": "Like OpenTelemetry trace spans but built-in. Each stage emits events; diagnostics aggregates them into a waterfall. The deterministic seed is like using a fixed random seed in Jest tests for reproducible outputs.",
    "relatedModules": [
      {
        "id": "pipeline",
        "name": "Pipeline",
        "rel": "Engine attaches diagnostics collector to the run"
      },
      {
        "id": "scheduler",
        "name": "Scheduler",
        "rel": "Scheduler emits step-level timing events"
      },
      {
        "id": "executor",
        "name": "Executor",
        "rel": "Executor emits HTTP waterfall events"
      }
    ]
  }
};
