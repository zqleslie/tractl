// Protocol identifies the request protocol. Defaults to 'http' when absent — ADR-017.
// 'soap' and 'odata' are reserved; SOAP requires v1.1, OData is plain HTTP.
export type Protocol = 'http' | 'graphql' | 'soap' | 'odata'

export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
export type BodyEncoding = 'json' | 'form' | 'multipart' | 'raw' | 'binary' | 'none' | 'graphql'
export type AuthType = 'none' | 'bearer' | 'basic' | 'apikey'
export type AssertionKind = 'status' | 'header' | 'body' | 'schema' | 'script'
export type AssertionSeverity = 'error' | 'warning'
export type ExtractSource = 'body' | 'header' | 'status' | 'timing' | 'metadata' | 'extension'
export type ExtractScope = 'workflow' | 'spec' | 'step'
export type FailurePolicy = 'resilient' | 'failFast'
export type RetryStrategy = 'fixed' | 'linear' | 'exponential'

export interface KVRow {
  enabled: boolean
  key: string
  value: string
}

export interface FormRow {
  enabled: boolean
  key: string
  value?: string
  type: 'text' | 'file'
  filename?: string
}

export interface BodyDef {
  encoding: BodyEncoding
  content?: string
  formRows?: FormRow[]
  rawContentType?: string
}

export interface AuthDef {
  type: AuthType
  token?: string
  username?: string
  password?: string
  keyName?: string
  keyValue?: string
  placement?: 'header' | 'query'
}

export interface AssertionDef {
  id: string
  kind: AssertionKind
  op: string
  expected: string
  severity: AssertionSeverity
}

export interface ExtractDef {
  id: string
  source: ExtractSource
  path: string
  variableName: string
  scope: ExtractScope
}

export interface RetryDef {
  strategy: RetryStrategy
  maxAttempts: number
  delayMs: number
  backoffFactor?: number
}

export interface RequestSettings {
  timeoutMs?: number
  failurePolicy?: FailurePolicy
  retry?: RetryDef | null
}

export interface RequestDef {
  protocol?: Protocol  // defaults to 'http' when absent — ADR-017
  id: string
  name: string
  method: HttpMethod
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
  /** Runtime environment variable overrides — matches Go RequestDef.Env (types.go). */
  env?: Record<string, string>
}

export interface AssertionResult {
  id: string
  kind: string
  op: string
  expected: string
  received: string
  passed: boolean
  severity: AssertionSeverity
}

export interface ExtractResult {
  id: string
  variableName: string
  scope: ExtractScope
  resolvedValue: string
  error?: string
}

export interface TimingResult {
  dns: number
  tcp: number
  tls: number
  ttfb: number
  transfer: number
  total: number
  unit: 'ms'
}

export interface RunResult {
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
}
