import type { TraCtlSpecDocument } from '@/components/request-editor/tractlSpecDocument'
import type { RunResult } from '@/platform/types'
import {
  useRunHistoryStore,
  type RunHistorySourceFormat,
  type RunHistorySourceType,
} from '@/stores/runHistoryStore'
import type { HttpMethod } from '@/components/primitives'

export type RecordRunHistoryInput = {
  requestName: string
  method: HttpMethod
  url: string
  result: RunResult
  sourceType?: RunHistorySourceType
  sourceName?: string
  sourceFormat?: RunHistorySourceFormat
  document?: TraCtlSpecDocument
}

export function recordRunHistoryEntry(input: RecordRunHistoryInput): string {
  const id = crypto.randomUUID()
  useRunHistoryStore.getState().addEntry({
    id,
    timestamp: new Date().toISOString(),
    sourceType: input.sourceType ?? 'request',
    sourceName: input.sourceName,
    sourceFormat: input.sourceFormat,
    document: input.document,
    requestName: input.requestName,
    method: input.method,
    url: input.url,
    statusCode: input.result.statusCode,
    durationMs: input.result.durationMs,
    outcome: input.result.error ? 'error' : 'success',
    result: input.result,
  })
  return id
}
