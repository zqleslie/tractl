package localapi

// RequestDef is the single representation of a request.
// Protocol defaults to "http" when absent — ADR-017.
type RequestDef struct {
	Protocol   string            `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	ID         string            `json:"id" yaml:"id"`
	Name       string            `json:"name" yaml:"name"`
	Method     string            `json:"method" yaml:"method"`
	URL        string            `json:"url" yaml:"url"`
	Params     []KVRow           `json:"params" yaml:"params,omitempty"`
	Headers    []KVRow           `json:"headers" yaml:"headers,omitempty"`
	Body       *BodyDef          `json:"body" yaml:"body,omitempty"`
	Auth       *AuthDef          `json:"auth" yaml:"auth,omitempty"`
	PreScript  string            `json:"preScript" yaml:"preScript,omitempty"`
	PostScript string            `json:"postScript" yaml:"postScript,omitempty"`
	Assertions []AssertionDef    `json:"assertions" yaml:"assertions,omitempty"`
	Extracts   []ExtractDef      `json:"extracts" yaml:"extracts,omitempty"`
	Settings   RequestSettings   `json:"settings" yaml:"settings,omitempty"`
	Env        map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
}

// KVRow is one enabled key-value row in params or headers.
type KVRow struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Key     string `json:"key" yaml:"key"`
	Value   string `json:"value" yaml:"value"`
}

// BodyDef describes the request body encoding and content.
type BodyDef struct {
	// Encoding identifies the content encoding. Accepted values:
	// "json", "form", "multipart", "raw", "binary", "none", "graphql"
	// "xml" is reserved for v1.1 (SOAP support).
	Encoding       string    `json:"encoding" yaml:"encoding"`
	Content        string    `json:"content" yaml:"content,omitempty"`
	FormRows       []FormRow `json:"formRows" yaml:"formRows,omitempty"`
	RawContentType string    `json:"rawContentType" yaml:"rawContentType,omitempty"`
}

// FormRow is one row in a form or multipart body.
type FormRow struct {
	Enabled  bool   `json:"enabled" yaml:"enabled"`
	Key      string `json:"key" yaml:"key"`
	Value    string `json:"value" yaml:"value,omitempty"`
	Type     string `json:"type" yaml:"type"`
	Filename string `json:"filename" yaml:"filename,omitempty"`
}

// AuthDef holds request authentication settings.
type AuthDef struct {
	Type      string `json:"type" yaml:"type"`
	Token     string `json:"token" yaml:"token,omitempty"`
	Username  string `json:"username" yaml:"username,omitempty"`
	Password  string `json:"password" yaml:"password,omitempty"`
	KeyName   string `json:"keyName" yaml:"keyName,omitempty"`
	KeyValue  string `json:"keyValue" yaml:"keyValue,omitempty"`
	Placement string `json:"placement" yaml:"placement,omitempty"`
}

// AssertionDef is one assertion on the request response.
type AssertionDef struct {
	ID       string `json:"id" yaml:"id"`
	Kind     string `json:"kind" yaml:"kind"`
	Op       string `json:"op" yaml:"op"`
	Expected string `json:"expected" yaml:"expected"`
	Severity string `json:"severity" yaml:"severity"`
}

// ExtractDef is one variable extract from the response.
type ExtractDef struct {
	ID           string `json:"id" yaml:"id"`
	Source       string `json:"source" yaml:"source"`
	Path         string `json:"path" yaml:"path"`
	VariableName string `json:"variableName" yaml:"variableName"`
	Scope        string `json:"scope" yaml:"scope"`
}

// RequestSettings holds timeout, retry, and failure policy.
type RequestSettings struct {
	TimeoutMs     int       `json:"timeoutMs" yaml:"timeoutMs,omitempty"`
	FailurePolicy string    `json:"failurePolicy" yaml:"failurePolicy,omitempty"`
	Retry         *RetryDef `json:"retry" yaml:"retry,omitempty"`
}

// RetryDef configures step retry behavior.
type RetryDef struct {
	Strategy      string  `json:"strategy" yaml:"strategy"`
	MaxAttempts   int     `json:"maxAttempts" yaml:"maxAttempts"`
	DelayMs       int     `json:"delayMs" yaml:"delayMs"`
	BackoffFactor float64 `json:"backoffFactor" yaml:"backoffFactor,omitempty"`
}

// StatusResponse is returned by GET /api/v1/status.
type StatusResponse struct {
	OK      bool   `json:"ok"`
	Version string `json:"version"`
}

// RunResult is the ephemeral response from POST /api/v1/run.
type RunResult struct {
	StatusCode       int               `json:"statusCode"`
	StatusText       string            `json:"statusText"`
	DurationMs       int64             `json:"durationMs"`
	Body             string            `json:"body"`
	Headers          map[string]string `json:"headers"`
	Timing           TimingResult      `json:"timing"`
	AssertionResults []AssertionResult `json:"assertionResults"`
	ExtractResults   []ExtractResult   `json:"extractResults"`
	AssertionsPassed int               `json:"assertionsPassed"`
	AssertionsTotal  int               `json:"assertionsTotal"`
	Passed           bool              `json:"passed"`
	Error            string            `json:"error,omitempty"`
}

// TimingResult breaks down request timing in milliseconds.
type TimingResult struct {
	DNS      int64  `json:"dns"`
	TCP      int64  `json:"tcp"`
	TLS      int64  `json:"tls"`
	TTFB     int64  `json:"ttfb"`
	Transfer int64  `json:"transfer"`
	Total    int64  `json:"total"`
	Unit     string `json:"unit"`
}

// AssertionResult is one evaluated assertion outcome.
type AssertionResult struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Op       string `json:"op"`
	Expected string `json:"expected"`
	Received string `json:"received"`
	Passed   bool   `json:"passed"`
	Severity string `json:"severity"`
}

// ExtractResult is one resolved extract outcome.
type ExtractResult struct {
	ID            string `json:"id"`
	VariableName  string `json:"variableName"`
	Scope         string `json:"scope"`
	ResolvedValue string `json:"resolvedValue"`
	Error         string `json:"error,omitempty"`
}

// WorkflowRunRequest is the POST /api/v1/workflows/run body.
type WorkflowRunRequest struct {
	Document string `json:"document"`
	Format   string `json:"format,omitempty"`
	Env      string `json:"env,omitempty"`
}

// FileWriteRequest is the POST /api/v1/files/{path} body.
type FileWriteRequest struct {
	Content string `json:"content"`
}

// FileEntry summarizes one workspace YAML file.
type FileEntry struct {
	Path       string `json:"path"`
	Name       string `json:"name"`
	ModifiedAt string `json:"modifiedAt"`
	Size       int64  `json:"size"`
}

// FileReadResponse is returned by GET /api/v1/files/{path}.
type FileReadResponse struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// FileWriteResponse is returned after writing a workspace file.
type FileWriteResponse struct {
	Path       string `json:"path"`
	ModifiedAt string `json:"modifiedAt"`
}

// ErrorResponse is a generic API error payload.
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// StepRef is a lightweight step reference for DAG computation.
type StepRef struct {
	ID        string   `json:"id"`
	DependsOn []string `json:"dependsOn"`
}

// LayoutRow is one execution tier in a workflow DAG layout.
type LayoutRow struct {
	RowIndex int      `json:"rowIndex"`
	StepIDs  []string `json:"stepIds"`
}

// LayoutEdge is one directed edge in the workflow DAG.
type LayoutEdge struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Kind     string `json:"kind"` // "sequential" | "fanout" | "merge"
	Implicit bool   `json:"implicit"`
}

// LayoutGroup is one independent sub-DAG (connected component).
type LayoutGroup struct {
	GroupIndex int      `json:"groupIndex"`
	RootStepID string   `json:"rootStepId"`
	StepIDs    []string `json:"stepIds"`
}

// WorkflowLayoutResponse is returned by POST /api/v1/workflow/layout.
type WorkflowLayoutResponse struct {
	Rows            []LayoutRow   `json:"rows"`
	Edges           []LayoutEdge  `json:"edges"`
	Groups          []LayoutGroup `json:"groups"`
	TopologySummary string        `json:"topologySummary"`
	Error           string        `json:"error,omitempty"`
}

// StepScanDef carries step field content to scan for implicit dependencies.
type StepScanDef struct {
	ID         string            `json:"id"`
	DependsOn  []string          `json:"dependsOn"`
	URL        string            `json:"url,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
	PreScript  string            `json:"preScript,omitempty"`
	PostScript string            `json:"postScript,omitempty"`
}

// StepWithImplicitDeps extends StepRef with inferred implicit dependencies.
type StepWithImplicitDeps struct {
	ID                string   `json:"id"`
	DependsOn         []string `json:"dependsOn"`
	ImplicitDependsOn []string `json:"implicitDependsOn"`
}

// InferDepsResponse is returned by POST /api/v1/workflow/infer-deps.
type InferDepsResponse struct {
	Steps []StepWithImplicitDeps `json:"steps"`
}

// WorkflowExportStep is one step in a workflow export request.
type WorkflowExportStep struct {
	ID         string            `json:"id"`
	DependsOn  []string          `json:"dependsOn,omitempty"`
	URL        string            `json:"url"`
	Method     string            `json:"method"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       *BodyDef          `json:"body,omitempty"`
	Assertions []AssertionDef    `json:"assertions,omitempty"`
	Extracts   []ExtractDef      `json:"extracts,omitempty"`
	PreScript  string            `json:"preScript,omitempty"`
	PostScript string            `json:"postScript,omitempty"`
}

// WorkflowExportRequest is the POST /api/v1/workflow/export body.
type WorkflowExportRequest struct {
	ID            string               `json:"id"`
	Name          string               `json:"name,omitempty"`
	Variables     map[string]string    `json:"variables,omitempty"`
	Concurrency   int                  `json:"concurrency,omitempty"`
	FailurePolicy string               `json:"failurePolicy,omitempty"`
	Steps         []WorkflowExportStep `json:"steps"`
}

// WorkflowExportResponse is returned by POST /api/v1/workflow/export.
type WorkflowExportResponse struct {
	YAML  string `json:"yaml"`
	Error string `json:"error,omitempty"`
}

// EngineRetryDefaults holds the default retry configuration values.
type EngineRetryDefaults struct {
	MaxAttempts   int      `json:"maxAttempts"`
	BackoffFactor float64  `json:"backoffFactor"`
	Strategies    []string `json:"strategies"`
}

// EngineDefaults holds the canonical engine default values.
type EngineDefaults struct {
	FailurePolicy  string              `json:"failurePolicy"`
	TimeoutMs      int                 `json:"timeoutMs"`
	Concurrency    int                 `json:"concurrency"`
	HTTPSuccessMin int                 `json:"httpSuccessMin"`
	HTTPSuccessMax int                 `json:"httpSuccessMax"`
	Retry          EngineRetryDefaults `json:"retry"`
}

// WorkspaceStatus is returned by GET /api/v1/workspace/status.
type WorkspaceStatus struct {
	RootDir    string `json:"rootDir"`
	IsGitRepo  bool   `json:"isGitRepo"`
	Branch     string `json:"branch,omitempty"`
	DirtyFiles int    `json:"dirtyFiles"`
}
