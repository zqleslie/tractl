import { extractRunSummaryFromDocument } from '@/lib/tractlDocument/extractRunSummary'
import { parseAndValidateDocument } from '@/lib/tractlDocument/parseValidateDocument'
import type { TraCtlSpecDocument } from '@/components/request-editor/tractlSpecDocument'
import { engine } from '@/platform/engine'
import type { RunResult } from '@/platform/types'
import { getEngineDefaults } from '@/stores/engineDefaultsStore'
import type {
  RunHistorySourceFormat,
  RunHistorySourceType,
} from '@/stores/runHistoryStore'
import type { TractlWasmParseFormat } from '@/platform/web/wasm/loadTractlWasmRuntime'
import type { ParseValidateFailure } from '@/lib/tractlDocument/parseValidateDocument'

export type RunDocumentInput = {
  raw: string
  format: TractlWasmParseFormat
  sourceName: string
  sourceType: RunHistorySourceType
  environmentVariables?: Record<string, string>
}

export type RunDocumentSuccess = {
  ok: true
  spec: TraCtlSpecDocument
  result: RunResult
  summary: NonNullable<ReturnType<typeof extractRunSummaryFromDocument>>
  sourceFormat: RunHistorySourceFormat
}

export type RunDocumentFailure = ParseValidateFailure | {
  ok: false
  stage: 'summary' | 'execution'
  message: string
  errors: Array<{ message: string; code?: string }>
}

export type RunDocumentResult = RunDocumentSuccess | RunDocumentFailure

function executionFailure(message: string, code?: string): RunDocumentFailure {
  return { ok: false, stage: 'execution', message, errors: [{ message, code }] }
}

export async function runTraCtlDocument(
  input: RunDocumentInput,
): Promise<RunDocumentResult> {
  const validated = await parseAndValidateDocument(input.raw, input.format)
  if (!validated.ok) return validated

  const summary = extractRunSummaryFromDocument(validated.spec, input.sourceName)
  if (!summary) {
    return {
      ok: false,
      stage: 'summary',
      message: 'No HTTP request step found in document',
      errors: [{ message: 'No HTTP request step found in document' }],
    }
  }

  try {
    const result = await engine.runRequest({
      method: summary.method,
      url: summary.url,
      settings: { failurePolicy: getEngineDefaults().failurePolicy },
      env: input.environmentVariables,
    })

    const sourceFormat = input.format === 'yml' ? 'yaml' : input.format
    return { ok: true, spec: validated.spec, result, summary, sourceFormat }
  } catch (error) {
    return executionFailure(error instanceof Error ? error.message : 'Execution failed')
  }
}
