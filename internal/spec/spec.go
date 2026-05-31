// Package spec defines the canonical traCtlSpec Go structs.
// These types are the authoritative Go representation of the traCtlSpec
// canonical semantic model defined in docs/spec/tractl_spec.md.
//
// TraCtlSpec is the document root (§5). Workflow is a child entity inside
// TraCtlSpec.Workflows (§6). Do not conflate these levels.
package spec

// TraCtlSpec is the top-level document root. Reference: tractl_spec.md §5
type TraCtlSpec struct {
	// SchemaVersion is the integer version of the traCtl schema this document
	// conforms to. Must be 1 in schema v1.
	SchemaVersion int `json:"schemaVersion" yaml:"schemaVersion"`
	// Capabilities declares which protocol adapters are required to execute
	// this spec. Each entry is a dotted platform identifier (e.g. "protocol.http").
	Capabilities []string `json:"capabilities" yaml:"capabilities"`
	// Metadata holds optional human-readable descriptors and system-assigned
	// tracking fields (source format, overlay provenance).
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	// Variables declares document-level variable bindings available to all
	// workflows and steps via ${vars.<name>} expressions.
	Variables map[string]string `json:"variables,omitempty" yaml:"variables,omitempty"`
	// Auth declares a document-level authentication profile applied to all
	// steps unless overridden at the workflow or step level.
	Auth *AuthProfile `json:"auth,omitempty" yaml:"auth,omitempty"`
	// Environments holds named environment definitions. TODO: Phase 4+ per §5.4.
	Environments *Environments `json:"environments,omitempty" yaml:"environments,omitempty"`
	// WorkflowConcurrency caps how many workflows run in parallel. Default 10.
	WorkflowConcurrency int `json:"workflowConcurrency,omitempty" yaml:"workflowConcurrency,omitempty"`
	// Workflows is the ordered list of workflow definitions. At least one
	// workflow is required.
	Workflows []Workflow `json:"workflows" yaml:"workflows"`
	// Extensions declares extension dependencies used by extensionCall steps.
	// TODO: Phase 4+ per §5.6.
	Extensions []ExtensionRef `json:"extensions,omitempty" yaml:"extensions,omitempty"`
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
	// ULID is the system-assigned unique identifier for this workflow.
	// It is populated by the parser and must never be authored.
	ULID string `yaml:"-" json:"_ulid,omitempty"`
	// ID is the workflow identifier, unique within the spec. Reference: §6.1.
	ID string `json:"id" yaml:"id"`
	// Name is an optional human-readable display name for the workflow.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Description is an optional human-readable description of the workflow's purpose.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// DependsOn lists workflow IDs that must complete before this workflow starts.
	DependsOn []string `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`
	// Variables declares workflow-scoped variable overrides that shadow document-level
	// variables for all steps in this workflow.
	Variables map[string]string `json:"variables,omitempty" yaml:"variables,omitempty"`
	// Auth overrides the document-level authentication profile for all steps in
	// this workflow.
	Auth *AuthProfile `json:"auth,omitempty" yaml:"auth,omitempty"`
	// Concurrency is the maximum number of steps that may execute in parallel
	// within this workflow. Zero means no limit beyond the scheduler default.
	Concurrency int `json:"concurrency,omitempty" yaml:"concurrency,omitempty"`
	// FailurePolicy controls what happens when a step in this workflow fails.
	// Defaults to DefaultFailurePolicy when empty.
	FailurePolicy FailurePolicy `json:"failurePolicy,omitempty" yaml:"failurePolicy,omitempty"`
	// Steps is the ordered list of steps to execute in this workflow.
	// At least one step is required.
	Steps []Step `json:"steps" yaml:"steps"`
	// Hooks declares workflow-scope lifecycle scripts that run before/after
	// all steps or before/after each step.
	Hooks *WorkflowHooks `json:"hooks,omitempty" yaml:"hooks,omitempty"`
	// Diagnostics declares workflow-level diagnostics collection settings.
	// Overridden per-step by Step.Diagnostics.
	Diagnostics *DiagnosticsConfig `json:"diagnostics,omitempty" yaml:"diagnostics,omitempty"`
}

// Step is the atomic unit of execution within a workflow.
// Reference: tractl_spec.md §7.1
type Step struct {
	// ULID is the system-assigned unique identifier for this step.
	// It is populated by the parser and must never be authored.
	ULID string `yaml:"-" json:"_ulid,omitempty"`
	// ID is the step identifier, unique within its parent workflow.
	ID string `json:"id" yaml:"id"`
	// Description is an optional human-readable description of the step's purpose.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Kind identifies the step type discriminant. Permitted values: "request",
	// "script", "extensionCall", "composite". Reference: tractl_spec.md §7.2.
	Kind string `json:"kind" yaml:"kind"`
	// DependsOn lists step IDs within the same workflow that must complete
	// successfully before this step may start.
	DependsOn []string `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`
	// When is an optional boolean expression that gates step execution.
	// The step is skipped when the expression evaluates to false.
	When string `json:"when,omitempty" yaml:"when,omitempty"`
	// Request describes the protocol-level invocation for kind="request" steps.
	Request *RequestDescriptor `json:"request,omitempty" yaml:"request,omitempty"`
	// Script describes the JavaScript script for kind="script" steps.
	// TODO: Phase 4+ per tractl_spec.md §9.2.
	Script *ScriptDescriptor `json:"script,omitempty" yaml:"script,omitempty"`
	// ExtensionCall describes the extension invocation for kind="extensionCall" steps.
	// TODO: Phase 4+ per §14.1.
	ExtensionCall *ExtensionCallDescriptor `json:"extensionCall,omitempty" yaml:"extensionCall,omitempty"`
	// Composite describes the nested workflow for kind="composite" steps.
	// TODO: Phase 4+ per §16.1.
	Composite *CompositeDescriptor `json:"composite,omitempty" yaml:"composite,omitempty"`
	// Assertions is the ordered list of validation rules evaluated against the
	// step result after execution.
	Assertions []Assertion `json:"assertions,omitempty" yaml:"assertions,omitempty"`
	// Extracts is the ordered list of value extraction rules that bind data from
	// the step result into the execution context.
	Extracts []Extract `json:"extracts,omitempty" yaml:"extracts,omitempty"`
	// Auth overrides the workflow-level authentication profile for this step only.
	Auth *AuthProfile `json:"auth,omitempty" yaml:"auth,omitempty"`
	// Retry declares retry behavior if the step fails.
	Retry *RetryPolicy `json:"retry,omitempty" yaml:"retry,omitempty"`
	// Timeout is the maximum wall-clock duration allowed for this step,
	// expressed as an ISO 8601 duration string (e.g. "PT30S").
	Timeout string `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	// Diagnostics overrides the workflow-level diagnostics settings for this
	// step. When non-nil it wins unconditionally per §15.4.
	Diagnostics *DiagnosticsConfig `json:"diagnostics,omitempty" yaml:"diagnostics,omitempty"`
	// Hooks declares step-scope lifecycle scripts that run before and after
	// this step's main action.
	Hooks *StepHooks `json:"hooks,omitempty" yaml:"hooks,omitempty"`
}

// AssertionSeverity controls whether an assertion failure fails the step.
// Reference: tractl_spec.md §11.1, §11.4
type AssertionSeverity string

// AssertionSeverity constants.
const (
	// SeverityError causes the step to enter the "failed" state when the
	// assertion condition is not met. This is the default severity.
	SeverityError AssertionSeverity = "error"
	// SeverityWarning records a diagnostic event but allows the step to remain
	// in the "succeeded" state even when the assertion is not met.
	SeverityWarning AssertionSeverity = "warning"
)

// Assertion is a validation rule evaluated against a step result.
// Reference: tractl_spec.md §11.1
type Assertion struct {
	// ULID is the system-assigned unique identifier for this assertion.
	// It is populated by the parser and must never be authored.
	ULID string `yaml:"-" json:"_ulid,omitempty"`
	// ID is the assertion identifier, unique within its parent step.
	ID string `json:"id" yaml:"id"`
	// Description is an optional human-readable description of the assertion's intent.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Kind identifies the assertion type (e.g. "status", "jsonSchema", "custom").
	Kind string `json:"kind" yaml:"kind"`
	// Op is the comparison operator used by this assertion (e.g. "eq", "gt").
	Op string `json:"op,omitempty" yaml:"op,omitempty"`
	// Target is the extraction path within the step result to evaluate (e.g. a JSONPath).
	Target string `json:"target,omitempty" yaml:"target,omitempty"`
	// Expected is the value that the extracted target must satisfy for the assertion to pass.
	Expected interface{} `json:"expected,omitempty" yaml:"expected,omitempty"`
	// Severity controls whether a failure blocks the step. Defaults to SeverityError.
	Severity AssertionSeverity `json:"severity,omitempty" yaml:"severity,omitempty"`
	// Config holds assertion-kind-specific configuration that does not fit the
	// Op/Target/Expected model.
	Config map[string]interface{} `json:"config,omitempty" yaml:"config,omitempty"`
}

// ExtractScope declares where an extracted value is written in context.
// Reference: tractl_spec.md §11.3
type ExtractScope string

// ExtractScope constants.
const (
	// ExtractScopeStep binds the extracted value to the step-local context only.
	// The value is discarded after the step completes.
	ExtractScopeStep ExtractScope = "step"
	// ExtractScopeWorkflow binds the extracted value to the workflow context,
	// making it available to subsequent steps in the same workflow. This is the
	// default scope per §11.3.
	ExtractScopeWorkflow ExtractScope = "workflow"
	// ExtractScopeSpec binds the extracted value to the spec-wide context,
	// making it available across all workflows.
	ExtractScopeSpec ExtractScope = "spec"
)

// Extract is a named value pulled from a step result into context.
// Reference: tractl_spec.md §11.3
type Extract struct {
	// ULID is the system-assigned unique identifier for this extract.
	// It is populated by the parser and must never be authored.
	ULID string `yaml:"-" json:"_ulid,omitempty"`
	// ID is the extract identifier, unique within its parent step.
	ID string `json:"id" yaml:"id"`
	// Source identifies the data origin to extract from (e.g. "response", "header").
	Source string `json:"source" yaml:"source"`
	// Path is the extraction path within the source (e.g. a JSONPath expression).
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
	// As is the variable binding name; defaults to ID when empty.
	As string `json:"as,omitempty" yaml:"as,omitempty"`
	// Scope defaults to ExtractScopeWorkflow when empty.
	Scope ExtractScope `json:"scope,omitempty" yaml:"scope,omitempty"`
}

// RequestDescriptor describes a protocol-level invocation in a protocol-neutral manner.
// Reference: tractl_spec.md §8.1
type RequestDescriptor struct {
	// Protocol identifies the transport protocol (e.g. "http", "grpc").
	Protocol string `json:"protocol" yaml:"protocol"`
	// Target is the fully-qualified endpoint address (URL or host:port).
	Target string `json:"target" yaml:"target"`
	// Operation is the optional protocol-level operation name (e.g. HTTP method,
	// gRPC method path).
	Operation string `json:"operation,omitempty" yaml:"operation,omitempty"`
	// Headers is the map of request headers to send.
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	// Metadata holds protocol-specific key/value pairs that do not map to
	// standard request fields.
	Metadata map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	// Body describes the request payload, including encoding and content.
	Body *BodyDescriptor `json:"body,omitempty" yaml:"body,omitempty"`
	// TLS and Transport are Phase 1 structural placeholders. Reference: tractl_spec.md §8.4
	TLS       map[string]interface{} `json:"tls,omitempty" yaml:"tls,omitempty"`
	Transport map[string]interface{} `json:"transport,omitempty" yaml:"transport,omitempty"`
}

// BodyDescriptor describes a request or response payload.
// Reference: tractl_spec.md §8.3
type BodyDescriptor struct {
	// Encoding identifies the content encoding (e.g. "json", "form", "binary").
	Encoding string `json:"encoding,omitempty" yaml:"encoding,omitempty"`
	// Content is the body payload. Its type depends on Encoding.
	Content interface{} `json:"content,omitempty" yaml:"content,omitempty"`
	// SchemaRef is an optional reference to a schema definition that describes
	// the structure of Content.
	SchemaRef string `json:"schemaRef,omitempty" yaml:"schemaRef,omitempty"`
}

// RetryPolicy declares retry behavior for a step.
// Reference: tractl_spec.md §7.4
type RetryPolicy struct {
	// MaxAttempts is the total number of execution attempts, including the
	// initial attempt. A value of 1 means no retries.
	MaxAttempts int `json:"maxAttempts,omitempty" yaml:"maxAttempts,omitempty"`
	// Backoff is the retry backoff strategy (e.g. "exponential", "linear").
	Backoff string `json:"backoff,omitempty" yaml:"backoff,omitempty"`
	// Delay is the initial wait duration between retry attempts, expressed as
	// an ISO 8601 duration string (e.g. "PT1S").
	Delay string `json:"delay,omitempty" yaml:"delay,omitempty"`
	// RetryOn lists the error conditions or status codes that trigger a retry.
	RetryOn []string `json:"retryOn,omitempty" yaml:"retryOn,omitempty"`
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
	// DiagnosticsKindTLS captures TLS handshake events, including certificate
	// negotiation and cipher suite selection.
	DiagnosticsKindTLS DiagnosticsKind = "tls"
	// DiagnosticsKindTCP captures TCP connection lifecycle events including
	// SYN/ACK timing and connection establishment.
	DiagnosticsKindTCP DiagnosticsKind = "tcp"
	// DiagnosticsKindTransport captures transport-layer events such as bytes
	// sent/received and connection reuse.
	DiagnosticsKindTransport DiagnosticsKind = "transport"
	// DiagnosticsKindLifecycle captures step lifecycle events including start,
	// end, and state transitions.
	DiagnosticsKindLifecycle DiagnosticsKind = "lifecycle"
	// DiagnosticsKindDNS captures DNS resolution events including query times
	// and resolved addresses.
	DiagnosticsKindDNS DiagnosticsKind = "dns"
	// DiagnosticsKindConnection captures connection establishment events,
	// including pooling and keep-alive behaviour.
	DiagnosticsKindConnection DiagnosticsKind = "connection"
)

// DiagnosticsRetention is a typed string for the three permitted retention scopes.
type DiagnosticsRetention string

// DiagnosticsRetention constants.
const (
	// RetentionStep retains diagnostics data only for the duration of the step.
	// Data is discarded once the step completes.
	RetentionStep DiagnosticsRetention = "step"
	// RetentionWorkflow retains diagnostics data for the duration of the workflow.
	// This is the default retention scope.
	RetentionWorkflow DiagnosticsRetention = "workflow"
	// RetentionSpec retains diagnostics data for the lifetime of the spec execution,
	// making it available across all workflows.
	RetentionSpec DiagnosticsRetention = "spec"
)

// DiagnosticsConfig is the opt-in diagnostics directive (traCtlSpec §15.1).
// It may be declared at workflow scope (applies to all steps) or at step scope
// (overrides workflow-level config). Diagnostics MUST NOT change execution
// semantics (§15.3).
type DiagnosticsConfig struct {
	// Enabled controls whether diagnostic event collection is active for the
	// scope governed by this config.
	Enabled bool `yaml:"enabled"     json:"enabled"`
	// Kinds is the list of diagnostic event categories to collect. When empty,
	// all kinds are collected (subject to Enabled).
	Kinds []DiagnosticsKind `yaml:"kinds"       json:"kinds"`
	// CaptureBody controls whether request and response bodies are included in
	// the collected diagnostic data.
	CaptureBody bool `yaml:"captureBody" json:"captureBody"`
	// Retention controls how long the collected diagnostic data is retained.
	// Defaults to RetentionWorkflow when empty.
	Retention DiagnosticsRetention `yaml:"retention"   json:"retention"`
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
	// Language is the script language identifier. Must be "js" in schema v1.
	Language string `yaml:"language"            json:"language"`
	// Source is the inline script source code.
	Source string `yaml:"source,omitempty"    json:"source,omitempty"`
	// SourceRef is a reference to an external script asset file. Accepted by
	// the parser but not yet executed at runtime (Phase 4+).
	SourceRef string `yaml:"sourceRef,omitempty" json:"sourceRef,omitempty"`
	// Inputs is the map of named input bindings passed to the script as variables.
	Inputs map[string]any `yaml:"inputs,omitempty"    json:"inputs,omitempty"`
}

// WorkflowHooks declares the four workflow-scope lifecycle hook scripts (spec §9.1).
type WorkflowHooks struct {
	// BeforeAll runs once before any step in the workflow begins executing.
	BeforeAll *Script `yaml:"beforeAll,omitempty"  json:"beforeAll,omitempty"`
	// AfterAll runs once after all steps in the workflow have completed,
	// regardless of individual step outcomes.
	AfterAll *Script `yaml:"afterAll,omitempty"   json:"afterAll,omitempty"`
	// BeforeEach runs before each individual step in the workflow.
	BeforeEach *Script `yaml:"beforeEach,omitempty" json:"beforeEach,omitempty"`
	// AfterEach runs after each individual step in the workflow completes.
	AfterEach *Script `yaml:"afterEach,omitempty"  json:"afterEach,omitempty"`
}

// StepHooks declares the three step-scope lifecycle hook scripts (spec §9.1).
type StepHooks struct {
	// BeforeStep runs immediately before the step's main action executes.
	BeforeStep *Script `yaml:"beforeStep,omitempty" json:"beforeStep,omitempty"`
	// AfterStep runs immediately after the step's main action completes,
	// regardless of success or failure.
	AfterStep *Script `yaml:"afterStep,omitempty"  json:"afterStep,omitempty"`
	// Transform runs after AfterStep and may reshape the step result before
	// extraction rules and assertions are evaluated.
	Transform *Script `yaml:"transform,omitempty"  json:"transform,omitempty"`
}

// ProvenanceEntry records which overlay patch contributed a canonical field.
// The map key in Metadata.Provenance is the canonical dotted field path,
// e.g. "workflows.main.steps.login.request.headers.Authorization".
// Reference: tractl_spec.md §5.1.1
type ProvenanceEntry struct {
	// Source must reference an entry in OverlayRefs.
	Source string `json:"source,omitempty" yaml:"source,omitempty"`
	// PatchIndex is the zero-based index of the patch within the overlay document
	// that contributed this field.
	PatchIndex int `json:"patchIndex,omitempty" yaml:"patchIndex,omitempty"`
	// Agent identifies the tool or process that applied the overlay patch.
	Agent string `json:"agent,omitempty" yaml:"agent,omitempty"`
	// Timestamp is the RFC 3339 time at which the overlay patch was applied.
	Timestamp string `json:"timestamp,omitempty" yaml:"timestamp,omitempty"`
	// SourceRef is an opaque reference to the overlay document source (file path or URL).
	SourceRef string `json:"sourceRef,omitempty" yaml:"sourceRef,omitempty"`
}

// Metadata is free-form descriptive metadata attached to a TraCtlSpec document.
// Reference: tractl_spec.md §5.1
type Metadata struct {
	// Name is a human-readable display name for the spec document.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Description is a human-readable description of the spec's purpose.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// SourceFormat is set by the parser ("yaml", "json", "toon"); never authored.
	SourceFormat string `json:"sourceFormat,omitempty" yaml:"sourceFormat,omitempty"`
	// SourceRef is an opaque origin identifier set by the parser (file path, URL, etc.).
	SourceRef string `json:"sourceRef,omitempty" yaml:"sourceRef,omitempty"`
	// OverlayRefs is the ordered list of overlay document identifiers that were
	// applied to produce this spec candidate.
	OverlayRefs []string `json:"overlayRefs,omitempty" yaml:"overlayRefs,omitempty"`
	// Provenance maps canonical dotted field paths to overlay contribution records.
	// Populated by the overlay engine; omitted when no overlay was applied.
	Provenance map[string]ProvenanceEntry `json:"provenance,omitempty" yaml:"provenance,omitempty"`
}

// ExtensionRef declares an extension dependency. TODO: Phase 4+ per §5.6
type ExtensionRef struct{}

// Environments holds named environment definitions. TODO: Phase 4+ per §5.4
type Environments struct{}
