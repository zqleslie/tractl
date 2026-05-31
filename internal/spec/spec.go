// Package spec defines the canonical traCtlSpec Go structs.
// These types are the authoritative Go representation of the traCtlSpec
// canonical semantic model defined in docs/spec/tractl_spec.md.
//
// TraCtlSpec is the document root (§5). Workflow is a child entity inside
// TraCtlSpec.Workflows (§6). Do not conflate these levels.
package spec

// TraCtlSpec is the top-level document root. Reference: tractl_spec.md §5
type TraCtlSpec struct {
	SchemaVersion int               `json:"schemaVersion" yaml:"schemaVersion"`
	Capabilities  []string          `json:"capabilities" yaml:"capabilities"`
	Metadata      *Metadata         `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Variables     map[string]string `json:"variables,omitempty" yaml:"variables,omitempty"`
	Auth          *AuthProfile      `json:"auth,omitempty" yaml:"auth,omitempty"`
	Environments  *Environments     `json:"environments,omitempty" yaml:"environments,omitempty"`
	// WorkflowConcurrency caps how many workflows run in parallel. Default 10.
	WorkflowConcurrency int            `json:"workflowConcurrency,omitempty" yaml:"workflowConcurrency,omitempty"`
	Workflows           []Workflow     `json:"workflows" yaml:"workflows"`
	Extensions          []ExtensionRef `json:"extensions,omitempty" yaml:"extensions,omitempty"`
}

// FailurePolicy is a typed enum for Workflow.FailurePolicy.
// Reference: tractl_spec.md §6.3
type FailurePolicy string

const (
	// FailurePolicyResilient records a step failure and passively skips dependent steps.
	// Independent branches always continue. Spec default per Revision 1.5.
	FailurePolicyResilient FailurePolicy = "resilient"

	// FailurePolicyFailFast actively cancels the transitive dependency subgraph on failure.
	FailurePolicyFailFast FailurePolicy = "failFast"

	// DefaultFailurePolicy is the spec-mandated default. tractl_spec.md §6.3.
	DefaultFailurePolicy = FailurePolicyResilient
)

// Workflow is an ordered, dependency-aware collection of steps sharing context.
// Reference: tractl_spec.md §6.1
type Workflow struct {
	ULID        string `yaml:"-" json:"_ulid,omitempty"`
	ID          string `json:"id" yaml:"id"`
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// DependsOn lists workflow IDs that must complete before this workflow starts.
	DependsOn     []string           `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`
	Variables     map[string]string  `json:"variables,omitempty" yaml:"variables,omitempty"`
	Auth          *AuthProfile       `json:"auth,omitempty" yaml:"auth,omitempty"`
	Concurrency   int                `json:"concurrency,omitempty" yaml:"concurrency,omitempty"`
	FailurePolicy FailurePolicy      `json:"failurePolicy,omitempty" yaml:"failurePolicy,omitempty"`
	Steps         []Step             `json:"steps" yaml:"steps"`
	Hooks         *WorkflowHooks     `json:"hooks,omitempty" yaml:"hooks,omitempty"`
	Diagnostics   *DiagnosticsConfig `json:"diagnostics,omitempty" yaml:"diagnostics,omitempty"`
}

// Step is the atomic unit of execution within a workflow.
// Reference: tractl_spec.md §7.1
type Step struct {
	ULID          string                   `yaml:"-" json:"_ulid,omitempty"`
	ID            string                   `json:"id" yaml:"id"`
	Description   string                   `json:"description,omitempty" yaml:"description,omitempty"`
	Kind          string                   `json:"kind" yaml:"kind"`
	DependsOn     []string                 `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`
	When          string                   `json:"when,omitempty" yaml:"when,omitempty"`
	Request       *RequestDescriptor       `json:"request,omitempty" yaml:"request,omitempty"`
	Script        *ScriptDescriptor        `json:"script,omitempty" yaml:"script,omitempty"`
	ExtensionCall *ExtensionCallDescriptor `json:"extensionCall,omitempty" yaml:"extensionCall,omitempty"`
	Composite     *CompositeDescriptor     `json:"composite,omitempty" yaml:"composite,omitempty"`
	Assertions    []Assertion              `json:"assertions,omitempty" yaml:"assertions,omitempty"`
	Extracts      []Extract                `json:"extracts,omitempty" yaml:"extracts,omitempty"`
	Auth          *AuthProfile             `json:"auth,omitempty" yaml:"auth,omitempty"`
	Retry         *RetryPolicy             `json:"retry,omitempty" yaml:"retry,omitempty"`
	Timeout       string                   `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Diagnostics   *DiagnosticsConfig       `json:"diagnostics,omitempty" yaml:"diagnostics,omitempty"`
	Hooks         *StepHooks               `json:"hooks,omitempty" yaml:"hooks,omitempty"`
}

// AssertionSeverity controls whether an assertion failure fails the step.
// Reference: tractl_spec.md §11.1, §11.4
type AssertionSeverity string

// AssertionSeverity constants.
const (
	SeverityError   AssertionSeverity = "error"   // step enters "failed" state — default
	SeverityWarning AssertionSeverity = "warning" // records diagnostic; step still "succeeded"
)

// Assertion is a validation rule evaluated against a step result.
// Reference: tractl_spec.md §11.1
type Assertion struct {
	ULID        string                 `yaml:"-" json:"_ulid,omitempty"`
	ID          string                 `json:"id" yaml:"id"`
	Description string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Kind        string                 `json:"kind" yaml:"kind"`
	Op          string                 `json:"op,omitempty" yaml:"op,omitempty"`
	Target      string                 `json:"target,omitempty" yaml:"target,omitempty"`
	Expected    interface{}            `json:"expected,omitempty" yaml:"expected,omitempty"`
	Severity    AssertionSeverity      `json:"severity,omitempty" yaml:"severity,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty" yaml:"config,omitempty"`
}

// ExtractScope declares where an extracted value is written in context.
// Reference: tractl_spec.md §11.3
type ExtractScope string

// ExtractScope constants.
const (
	ExtractScopeStep     ExtractScope = "step"
	ExtractScopeWorkflow ExtractScope = "workflow" // default per §11.3
	ExtractScopeSpec     ExtractScope = "spec"
)

// Extract is a named value pulled from a step result into context.
// Reference: tractl_spec.md §11.3
type Extract struct {
	ULID   string `yaml:"-" json:"_ulid,omitempty"`
	ID     string `json:"id" yaml:"id"`
	Source string `json:"source" yaml:"source"`
	Path   string `json:"path,omitempty" yaml:"path,omitempty"`
	// As is the variable binding name; defaults to ID when empty.
	As string `json:"as,omitempty" yaml:"as,omitempty"`
	// Scope defaults to ExtractScopeWorkflow when empty.
	Scope ExtractScope `json:"scope,omitempty" yaml:"scope,omitempty"`
}

// RequestDescriptor describes a protocol-level invocation in a protocol-neutral manner.
// Reference: tractl_spec.md §8.1
type RequestDescriptor struct {
	Protocol  string                 `json:"protocol" yaml:"protocol"`
	Target    string                 `json:"target" yaml:"target"`
	Operation string                 `json:"operation,omitempty" yaml:"operation,omitempty"`
	Headers   map[string]string      `json:"headers,omitempty" yaml:"headers,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Body      *BodyDescriptor        `json:"body,omitempty" yaml:"body,omitempty"`
	// TLS and Transport are Phase 1 structural placeholders. Reference: tractl_spec.md §8.4
	TLS       map[string]interface{} `json:"tls,omitempty" yaml:"tls,omitempty"`
	Transport map[string]interface{} `json:"transport,omitempty" yaml:"transport,omitempty"`
}

// BodyDescriptor describes a request or response payload.
// Reference: tractl_spec.md §8.3
type BodyDescriptor struct {
	Encoding  string      `json:"encoding,omitempty" yaml:"encoding,omitempty"`
	Content   interface{} `json:"content,omitempty" yaml:"content,omitempty"`
	SchemaRef string      `json:"schemaRef,omitempty" yaml:"schemaRef,omitempty"`
}

// RetryPolicy declares retry behavior for a step.
// Reference: tractl_spec.md §7.4
type RetryPolicy struct {
	MaxAttempts int      `json:"maxAttempts,omitempty" yaml:"maxAttempts,omitempty"`
	Backoff     string   `json:"backoff,omitempty" yaml:"backoff,omitempty"`
	Delay       string   `json:"delay,omitempty" yaml:"delay,omitempty"`
	RetryOn     []string `json:"retryOn,omitempty" yaml:"retryOn,omitempty"`
}

// ─── Phase 4+ stubs ──────────────────────────────────────────────────────────

// AuthProfile is a named authentication profile. TODO: Phase 4+
type AuthProfile struct{}

// ScriptDescriptor describes a script step body. TODO: Phase 4+ per tractl_spec.md §9.2
type ScriptDescriptor struct{}

// ExtensionCallDescriptor describes an extensionCall step body. TODO: Phase 4+ per §14.1
type ExtensionCallDescriptor struct{}

// CompositeDescriptor describes a composite step body. TODO: Phase 4+ per §16.1
type CompositeDescriptor struct{}

// DiagnosticsKind is a typed string for the six permitted diagnostics kinds.
// No additional kinds are admitted without explicit architecture extension (§15.2).
type DiagnosticsKind string

// DiagnosticsKind constants.
const (
	DiagnosticsKindTLS        DiagnosticsKind = "tls"
	DiagnosticsKindTCP        DiagnosticsKind = "tcp"
	DiagnosticsKindTransport  DiagnosticsKind = "transport"
	DiagnosticsKindLifecycle  DiagnosticsKind = "lifecycle"
	DiagnosticsKindDNS        DiagnosticsKind = "dns"
	DiagnosticsKindConnection DiagnosticsKind = "connection"
)

// DiagnosticsRetention is a typed string for the three permitted retention scopes.
type DiagnosticsRetention string

// DiagnosticsRetention constants.
const (
	RetentionStep     DiagnosticsRetention = "step"
	RetentionWorkflow DiagnosticsRetention = "workflow" // default
	RetentionSpec     DiagnosticsRetention = "spec"
)

// DiagnosticsConfig is the opt-in diagnostics directive (traCtlSpec §15.1).
// It may be declared at workflow scope (applies to all steps) or at step scope
// (overrides workflow-level config). Diagnostics MUST NOT change execution
// semantics (§15.3).
type DiagnosticsConfig struct {
	Enabled     bool                 `yaml:"enabled"     json:"enabled"`
	Kinds       []DiagnosticsKind    `yaml:"kinds"       json:"kinds"`
	CaptureBody bool                 `yaml:"captureBody" json:"captureBody"`
	Retention   DiagnosticsRetention `yaml:"retention"   json:"retention"`
}

// EffectiveFor returns the diagnostics config that governs a step,
// applying the workflow-scope override rule from §15.4:
//   - If stepConfig is non-nil, it wins unconditionally.
//   - Otherwise the workflow-level config (receiver) is used.
//   - If both are nil, returns a disabled zero-value config (Enabled: false).
//
// The receiver may be nil (a nil *DiagnosticsConfig is valid).
func (wf *DiagnosticsConfig) EffectiveFor(stepConfig *DiagnosticsConfig) DiagnosticsConfig {
	if stepConfig != nil {
		return *stepConfig
	}
	if wf != nil {
		return *wf
	}
	return DiagnosticsConfig{}
}

// Script describes a JavaScript script inline or by reference (spec §9.2).
// language MUST be "js" in schema v1; other values are rejected at validation.
// Exactly one of Source or SourceRef must be non-empty.
// SourceRef (file asset reference) is accepted in the type but returns
// ErrSourceRefUnsupported at runtime — deferred to a future phase.
type Script struct {
	Language  string         `yaml:"language"            json:"language"`
	Source    string         `yaml:"source,omitempty"    json:"source,omitempty"`
	SourceRef string         `yaml:"sourceRef,omitempty" json:"sourceRef,omitempty"`
	Inputs    map[string]any `yaml:"inputs,omitempty"    json:"inputs,omitempty"`
}

// WorkflowHooks declares the four workflow-scope lifecycle hook scripts (spec §9.1).
type WorkflowHooks struct {
	BeforeAll  *Script `yaml:"beforeAll,omitempty"  json:"beforeAll,omitempty"`
	AfterAll   *Script `yaml:"afterAll,omitempty"   json:"afterAll,omitempty"`
	BeforeEach *Script `yaml:"beforeEach,omitempty" json:"beforeEach,omitempty"`
	AfterEach  *Script `yaml:"afterEach,omitempty"  json:"afterEach,omitempty"`
}

// StepHooks declares the three step-scope lifecycle hook scripts (spec §9.1).
type StepHooks struct {
	BeforeStep *Script `yaml:"beforeStep,omitempty" json:"beforeStep,omitempty"`
	AfterStep  *Script `yaml:"afterStep,omitempty"  json:"afterStep,omitempty"`
	Transform  *Script `yaml:"transform,omitempty"  json:"transform,omitempty"`
}

// ProvenanceEntry records which overlay patch contributed a canonical field.
// The map key in Metadata.Provenance is the canonical dotted field path,
// e.g. "workflows.main.steps.login.request.headers.Authorization".
// Reference: tractl_spec.md §5.1.1
type ProvenanceEntry struct {
	// Source must reference an entry in OverlayRefs.
	Source     string `json:"source,omitempty" yaml:"source,omitempty"`
	PatchIndex int    `json:"patchIndex,omitempty" yaml:"patchIndex,omitempty"`
	Agent      string `json:"agent,omitempty" yaml:"agent,omitempty"`
	Timestamp  string `json:"timestamp,omitempty" yaml:"timestamp,omitempty"`
	SourceRef  string `json:"sourceRef,omitempty" yaml:"sourceRef,omitempty"`
}

// Metadata is free-form descriptive metadata attached to a TraCtlSpec document.
// Reference: tractl_spec.md §5.1
type Metadata struct {
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// SourceFormat is set by the parser ("yaml", "json", "toon"); never authored.
	SourceFormat string `json:"sourceFormat,omitempty" yaml:"sourceFormat,omitempty"`
	// SourceRef is an opaque origin identifier set by the parser (file path, URL, etc.).
	SourceRef   string   `json:"sourceRef,omitempty" yaml:"sourceRef,omitempty"`
	OverlayRefs []string `json:"overlayRefs,omitempty" yaml:"overlayRefs,omitempty"`
	// Provenance maps canonical dotted field paths to overlay contribution records.
	// Populated by the overlay engine; omitted when no overlay was applied.
	Provenance map[string]ProvenanceEntry `json:"provenance,omitempty" yaml:"provenance,omitempty"`
}

// ExtensionRef declares an extension dependency. TODO: Phase 4+ per §5.6
type ExtensionRef struct{}

// Environments holds named environment definitions. TODO: Phase 4+ per §5.4
type Environments struct{}
