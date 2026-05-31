import type { TraCtlSpecDocument } from '@/components/request-editor/tractlSpecDocument'
import type { KeyValueRow } from '@/components/request-editor/types'

export type ApiStatusResponse = {
  ok: boolean
  version: string
}

export type SaveRequestFileInput = {
  id?: string | null
  name: string
  document: TraCtlSpecDocument
}

export type SaveRequestFileResponse = {
  id: string
  name: string
  path: string
  updatedAt: string
}

export type RunRequestInput = {
  fileId: string
  env?: string
}

export type RequestTimelineSegment = {
  label: string
  ms: number
}

export type RequestRunResult = {
  passed: boolean
  durationMs: number
  statusCode: number
  statusLabel: string
  contentType: string
  body: string
  headers: KeyValueRow[]
  assertionResults: Array<{
    id: string
    passed: boolean
    label: string
    detail: string
  }>
  extractResults: Array<{
    variable: string
    value: string
    scope: string
  }>
  passedCount: number
  totalCount: number
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
