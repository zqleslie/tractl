window.SECURITY_DATA = {
  "good": [
    {
      "title": "JS Sandbox \u2014 Zero I/O Access",
      "desc": "Scripts run in goja (pure-Go engine) with no filesystem, network, or require access. Only tractl.setVar(), tractl.cancel(), tractl.random() are injected."
    },
    {
      "title": "No Panics in Critical Path",
      "desc": "engine.Run() and engine.RunDocument() never panic. All errors are captured in RunResult fields and surfaced as exit code 0/1/2."
    },
    {
      "title": "Authorization Header Masking",
      "desc": "Authorization, X-API-Key and similar headers are automatically redacted in all diagnostic output and trace logs."
    },
    {
      "title": "Thread-Safe ExecutionContext",
      "desc": "sync.RWMutex protects all shared maps. Scripts work on FrozenContext deep copies \u2014 no goroutine can corrupt another's state mid-execution."
    },
    {
      "title": "Validation Before Execution",
      "desc": "Full spec validation (including cycle detection) runs before any HTTP requests are made. Malformed specs fail loudly at Stage 3, not mid-execution."
    },
    {
      "title": "Deterministic Sandbox Seeding",
      "desc": "tractl.random() and tractl.now() in scripts use a seed derived from TraceID \u2014 reproducible across runs, not truly random (prevents timing-based attacks in scripts)."
    }
  ],
  "warn": [
    {
      "title": "Credential Masking Incomplete \u2014 Response Bodies",
      "desc": "Headers are masked in traces, but response body contents are not. If an endpoint echoes a token in its response body, it appears in plaintext in diagnostic output."
    },
    {
      "title": "LocalAPI Has No Authentication",
      "desc": "The server on :7428 assumes a trusted local environment. Permissive CORS means any web page open in the browser can call it. Acceptable for local dev; needs auth if ever exposed to a network."
    },
    {
      "title": "No Rate Limiting on LocalAPI",
      "desc": "A runaway script or malicious page could hammer :7428 with thousands of concurrent workflow runs. Add a token bucket before any production exposure."
    },
    {
      "title": "Retry Backoff Has No Jitter",
      "desc": "Deterministic exponential backoff means many clients retrying the same endpoint simultaneously create synchronised load spikes. Add \u00b110% random jitter \u2014 standard practice (AWS SDK does this)."
    },
    {
      "title": "HTTP Client TLS \u2014 Verify Default",
      "desc": "Confirm the default HTTP client in internal/executor does not set InsecureSkipVerify anywhere. Go's http.DefaultTransport is strict by default but custom clients sometimes disable it."
    }
  ],
  "bad": [
    {
      "title": "Future: Env Var Storage Must Encrypt at Rest",
      "desc": "Phase 4+ plans a workspace Store for environment variables (API keys, tokens). These MUST be encrypted at rest (OS keychain or AES-GCM). Storing as plaintext JSON on disk is a common mistake that leads to credential leaks."
    },
    {
      "title": "No Audit Logging",
      "desc": "No trail of who ran which workflow when. Fine for a single-user CLI. A multi-user server mode MUST add structured audit logging before going to production."
    }
  ],
  "antipatterns": [
    {
      "title": "Step Hooks Run After Execution \u2014 Not Before",
      "desc": "beforeStep and afterStep hooks currently run after the scheduler has already executed the step (architectural limitation). A beforeStep mutation won't affect the current step. Pending scheduler refactor. Developers writing hooks will debug for hours if this isn't documented clearly."
    },
    {
      "title": "Variable Scope Shadowing Is Silent",
      "desc": "A workflow-level variable with the same name as a spec-level variable silently wins. No warning is emitted. Best practice: prefix workflow variables (e.g. wf_myVar) to avoid accidental shadowing."
    },
    {
      "title": "Unresolved Variables Become Empty Strings",
      "desc": "If you reference ${vars.typo} and 'typo' doesn't exist, it resolves to an empty string. The request is sent with a blank URL segment or header giving a confusing HTTP error instead of 'variable not found'."
    }
  ]
};
