import type {
  TraCtlAssertion,
  TraCtlExtract,
} from '@/components/request-editor/tractlSpecDocument'

export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
export type StepOutcome = 'idle' | 'running' | 'passed' | 'failed' | 'skipped'
export type FailurePolicy = 'resilient' | 'failFast'

export interface StepAssertionResult {
  description: string
  passed: boolean
  received?: string
  expected?: string
  severity: 'error' | 'warn' | 'info'
}

export interface StepExtractResult {
  variable: string
  expression: string
  value?: string
}

export interface StepResult {
  outcome: StepOutcome
  statusCode?: number
  durationMs?: number
  /** Milliseconds from workflow start when this step began executing. */
  startMs?: number
  assertions: StepAssertionResult[]
  extracts: StepExtractResult[]
  responseBody?: string
  responseHeaders?: Record<string, string>
  /** Actual URL used by the engine after variable resolution. */
  requestUrl?: string
}

export interface WorkflowStep {
  id: string
  method: HttpMethod
  url: string
  dependsOn: string[]
  implicitDependsOn?: string[]
  hasAuth: boolean
  assertionCount: number
  hasPreScript: boolean
  extractCount: number
  assertions?: TraCtlAssertion[]
  extracts?: TraCtlExtract[]
  result?: StepResult
}

export interface WorkflowConfig {
  /** Omitted = engine platform default (4 for step scheduling). */
  concurrency?: number
  timeoutMs: number
  failurePolicy: FailurePolicy
  hasHooks: boolean
  hasOverlay: boolean
  overlayName?: string
}

export interface Workflow {
  id: string
  name: string
  steps: WorkflowStep[]
  config: WorkflowConfig
  variables?: Record<string, string>
  yaml?: string
  lastModified?: string
}

export type RunState = 'idle' | 'running' | 'complete' | 'error'

export interface WorkflowRunSummary {
  outcome: 'passed' | 'failed' | 'error'
  totalDurationMs: number
  stepsRan: number
  totalSteps: number
  assertionsPassed: number
  assertionsTotal: number
  waterfallText: string
  errorMessage?: string
}
