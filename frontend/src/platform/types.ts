// Mirror of internal/localapi/types.go.
// This is the single source of truth for all engine I/O types.

export type KVRow = { enabled: boolean; key: string; value: string }

export type BodyDef = {
  encoding: string
  content?: string
  formRows?: FormRow[]
  rawContentType?: string
}

export type FormRow = {
  enabled: boolean; key: string; value?: string; type: string; filename?: string
}

export type AuthDef = {
  type: string; token?: string; username?: string; password?: string
  keyName?: string; keyValue?: string; placement?: string
}

export type AssertionDef = {
  id: string; kind: string; op: string; expected: string; severity: string
}

export type ExtractDef = {
  id: string; source: string; path: string; variableName: string; scope: string
}

export type RetryDef = {
  strategy: string; maxAttempts: number; delayMs: number; backoffFactor?: number
}

export type RequestSettings = {
  timeoutMs?: number; failurePolicy?: string; retry?: RetryDef
}

export type RequestDef = {
  id?: string
  name?: string
  method: string
  url: string
  params?: KVRow[]
  headers?: KVRow[]
  body?: BodyDef
  auth?: AuthDef
  preScript?: string
  postScript?: string
  assertions?: AssertionDef[]
  extracts?: ExtractDef[]
  settings?: RequestSettings
  env?: Record<string, string>
}

export type TimingResult = {
  dns: number; tcp: number; tls: number
  ttfb: number; transfer: number; total: number
  unit: 'ms'
}

export type AssertionResult = {
  id: string; kind: string; op: string
  expected: string; received: string
  passed: boolean; severity: string
}

export type ExtractResult = {
  id: string; variableName: string; scope: string
  resolvedValue: string; error?: string
}

export type RunResult = {
  statusCode: number
  statusText: string
  durationMs: number
  body: string
  headers: Record<string, string>
  timing: TimingResult
  assertionResults: AssertionResult[]
  extractResults: ExtractResult[]
  assertionsPassed: number
  assertionsTotal: number
  error?: string
  passed: boolean
}

export type StepRef = { id: string; dependsOn: string[] }

export type StepScanDef = StepRef & {
  url?: string
  headers?: Record<string, string>
  body?: string
  preScript?: string
  postScript?: string
}

export type LayoutRow = { rowIndex: number; stepIds: string[] }

export type LayoutEdge = {
  from: string; to: string; kind: string; implicit: boolean
}

export type LayoutGroup = {
  groupIndex: number; rootStepId: string; stepIds: string[]
}

export type WorkflowLayoutResponse = {
  rows: LayoutRow[]
  edges: LayoutEdge[]
  groups: LayoutGroup[]
  topologySummary: string
  error?: string
}

export type StepWithImplicitDeps = {
  id: string; dependsOn: string[]; implicitDependsOn: string[]
}

export type InferDepsResponse = { steps: StepWithImplicitDeps[] }

export type WorkflowExportRequest = {
  id: string
  name?: string
  variables?: Record<string, string>
  concurrency?: number
  failurePolicy?: string
  steps: Array<{
    id: string
    dependsOn?: string[]
    url: string
    method: string
    headers?: Record<string, string>
    body?: BodyDef
    assertions?: AssertionDef[]
    extracts?: ExtractDef[]
    preScript?: string
    postScript?: string
  }>
}

export type WorkflowExportResponse = { yaml: string; error?: string }

export type EngineDefaults = {
  failurePolicy: string
  timeoutMs: number
  concurrency: number
  httpSuccessMin: number
  httpSuccessMax: number
  retry: { maxAttempts: number; backoffFactor: number; strategies: string[] }
}

export type WorkspaceStatus = {
  rootDir: string; isGitRepo: boolean; branch?: string; dirtyFiles: number
}

export type ApiStatusResponse = { ok: boolean; version: string }

export type FileWriteResponse = { path: string; modifiedAt: string }
