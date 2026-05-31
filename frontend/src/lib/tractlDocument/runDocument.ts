import {
  formatMissingEnvironmentVariables,
  resolveTraCtlSpecEnvironmentVariables,
} from '@/lib/resolveEnvironmentVariables'
import { extractRunSummaryFromDocument } from '@/lib/tractlDocument/extractRunSummary'
import { parseAndValidateDocument } from '@/lib/tractlDocument/parseValidateDocument'
import type { TraCtlSpecDocument } from '@/components/request-editor/tractlSpecDocument'
import { getRequestExecutionRunner } from '@/platform/requestExecution/getRequestExecutionRunner'
import type { RunResult } from '@/platform/types'
import { getEngineDefaults } from '@/stores/engineDefaultsStore'
import type { RequestDef } from '@/types/requestDef'
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
  stage: 'summary' | 'environment' | 'execution'
  message: string
  errors: Array<{ message: string; code?: string }>
}

export type RunDocumentResult = RunDocumentSuccess | RunDocumentFailure

function executionFailure(
  message: string,
  code?: string,
): RunDocumentFailure {
  return {
    ok: false,
    stage: 'execution',
    message,
    errors: [{ message, code }],
  }
}

export async function runTraCtlDocument(
  input: RunDocumentInput,
): Promise<RunDocumentResult> {
  const validated = await parseAndValidateDocument(input.raw, input.format)
  if (!validated.ok) {
    return validated
  }

  const summary = extractRunSummaryFromDocument(
    validated.spec,
    input.sourceName,
  )
  if (!summary) {
    return {
      ok: false,
      stage: 'summary',
      message: 'No HTTP request step found in document',
      errors: [{ message: 'No HTTP request step found in document' }],
    }
  }

  const resolved = resolveTraCtlSpecEnvironmentVariables(
    validated.spec,
    input.environmentVariables ?? {},
  )
  if (resolved.missing.length > 0) {
    const message = formatMissingEnvironmentVariables(resolved.missing)
    return {
      ok: false,
      stage: 'environment',
      message,
      errors: resolved.missing.map((name) => ({
        message: `Missing environment variable: ${name}`,
        code: 'TRACTL_ENV_VARIABLE_MISSING',
      })),
    }
  }

  const runner = getRequestExecutionRunner()

  try {
    await runner.checkAvailable()
    const request: RequestDef = {
      id: input.sourceName,
      name: summary.requestName,
      method: summary.method as RequestDef['method'],
      url: summary.url,
      headers: [],
      params: [],
      assertions: [],
      extracts: [],
      settings: { failurePolicy: getEngineDefaults().failurePolicy },
    }
    const result = await runner.runRequest({
      request,
      name: summary.requestName,
    })

    const sourceFormat =
      input.format === 'yml' ? 'yaml' : input.format

    return {
      ok: true,
      spec: validated.spec,
      result,
      summary,
      sourceFormat,
    }
  } catch (error) {
    const message =
      error instanceof Error ? error.message : 'Execution failed'
    return executionFailure(message)
  }
}
