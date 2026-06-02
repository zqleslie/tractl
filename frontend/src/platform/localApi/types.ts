import type { RequestDef, TimingResult } from '@/types/requestDef'

export type ApiStatusResponse = {
  ok: boolean
  version: string
}

export type SaveRequestFileInput = {
  id?: string | null
  name: string
  request: RequestDef
  path?: string
}

export type SaveRequestFileResponse = {
  path: string
  updatedAt: string
}

export type RunRequestInput = {
  request: RequestDef
  env?: string
}

export type RequestTimelineSegment = {
  label: string
  ms: number
}

export type AssertionResultRow = {
  id: string
  kind: string
  op: string
  expected: string
  received: string
  passed: boolean
  severity: 'error' | 'warning'
}

export type ExtractResultRow = {
  id: string
  variable: string
  value: string
  scope: string
}

export type RequestRunResult = {
  passed: boolean
  statusCode: number
  statusText: string
  durationMs: number
  body: string
  contentType: string
  headers: Record<string, string>
  timing: TimingResult
  assertionResults: AssertionResultRow[]
  extractResults: ExtractResultRow[]
  assertionsPassed: number
  assertionsTotal: number
  timeline: RequestTimelineSegment[]
  error?: string
}

export type ApiErrorBody = {
  error: string
  details?: string
}

export type WorkflowRunDocumentInput = {
  document: string
  format?: 'yaml' | 'yml' | 'json' | 'toon'
  env?: string
}

/** engine.RunResult JSON from POST /api/v1/workflows/run (same shape as WASM run success). */
export type EngineWorkflowRunResult = Record<string, unknown>

export type FileWriteInput = {
  path: string
  content: string
}

export type FileWriteResponse = {
  path: string
  modifiedAt: string
}
