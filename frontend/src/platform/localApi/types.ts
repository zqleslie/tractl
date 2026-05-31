import type { KeyValueRow } from '@/components/request-editor/types'
import type { RequestDef, RunResult } from '@/types/requestDef'

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

type LegacyAssertionResult = {
  id: string
  passed: boolean
  label?: string
  detail?: string
  kind?: string
  op?: string
  expected?: string
  received?: string
  severity?: 'error' | 'warning'
}

type LegacyExtractResult = {
  id?: string
  variable?: string
  value?: string
  variableName?: string
  resolvedValue?: string
  scope: 'workflow' | 'spec' | 'step' | string
}

export type RequestRunResult = Partial<Omit<RunResult, 'headers' | 'assertionResults' | 'extractResults'>> & {
  durationMs: number
  statusCode: number
  body: string
  headers: Record<string, string> | KeyValueRow[]
  assertionResults: LegacyAssertionResult[]
  extractResults: LegacyExtractResult[]
  passed?: boolean
  statusLabel?: string
  contentType?: string
  passedCount?: number
  totalCount?: number
  timeline?: RequestTimelineSegment[]
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
