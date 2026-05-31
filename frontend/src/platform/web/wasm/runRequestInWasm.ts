import type { TraCtlSpecDocument } from '@/components/request-editor/tractlSpecDocument'
import { looksLikeViteDevShell } from '@/components/request-editor/normalizeRequestUrl'
import type { RunResult } from '@/platform/types'
import {
  isTractlWasmRunSuccess,
  loadTractlWasmRuntime,
  type TractlWasmParseFormat,
  type TractlWasmRunResult,
} from '@/platform/web/wasm/loadTractlWasmRuntime'

export const REQUEST_EDITOR_WASM_FORMAT: TractlWasmParseFormat = 'json'

type WasmRunSuccess = Record<string, unknown> & {
  Passed?: boolean
  Workflows?: WasmWorkflowOutcome[]
  diagnostics?: WasmDiagnostics
}

type WasmWorkflowOutcome = {
  Passed?: boolean
  Skipped?: boolean
  Steps?: WasmStepOutcome[]
}

type WasmStepOutcome = {
  StepID?: string
  AssertionResults?: WasmAssertionRecord[]
  ResponseStatus?: number
  ResponseHeaders?: Record<string, string>
  ResponseBody?: string
  Error?: string
}

type WasmAssertionRecord = {
  AssertionID?: string
  Kind?: string
  Outcome?: string
  Message?: string
}

type WasmDiagnostics = {
  Workflows?: Array<{
    Steps?: Array<{
      Requests?: Array<{
        Timeline?: WasmTimingRecord
      }>
      Extracts?: WasmExtractRecord[]
    }>
  }>
}

type WasmTimingRecord = {
  DNSMs?: number
  TCPMs?: number
  TLSMs?: number
  TTFBMs?: number
  TransferMs?: number
  TotalMs?: number
}

type WasmExtractRecord = {
  ID?: string
  Source?: string
  As?: string
}

export async function runRequestInWasm(
  spec: TraCtlSpecDocument,
): Promise<RunResult> {
  await loadTractlWasmRuntime()

  if (!window.tractl?.run) {
    return mapWasmFailure('TRACTL_WASM_UNAVAILABLE', 'WASM runtime is not ready')
  }

  const result = await window.tractl.run(
    JSON.stringify(spec),
    REQUEST_EDITOR_WASM_FORMAT,
  )
  return mapWasmRunResultToRequestRunResult(result)
}

export function mapWasmRunResultToRequestRunResult(
  result: TractlWasmRunResult,
): RunResult {
  if (!isTractlWasmRunSuccess(result)) {
    return mapWasmFailure(result.error.code, result.error.message)
  }

  const run = result as WasmRunSuccess
  const workflow = run.Workflows?.[0]
  const step = workflow?.Steps?.[0]

  if (!workflow || workflow.Skipped || !step) {
    return mapWasmFailure(
      'TRACTL_WASM_EMPTY_RESULT',
      workflow?.Skipped
        ? 'Workflow skipped due to dependency failure'
        : 'Run produced no request result',
    )
  }

  const assertionResults = (step.AssertionResults ?? []).map((assertion, index) => {
    const id = assertion.AssertionID || `assertion-${index + 1}`
    const kind = assertion.Kind || 'assertion'
    return {
      id,
      kind,
      op: 'equals',
      expected: id,
      received: assertion.Message || '',
      passed: assertion.Outcome === 'pass',
      severity: 'error',
    }
  })

  const passedCount = assertionResults.filter((assertion) => assertion.passed).length
  const timing = timingFromDiagnostics(run.diagnostics)
  const headers = step.ResponseHeaders ?? {}
  const statusCode = step.ResponseStatus ?? 0
  const rawBody = step.ResponseBody ?? step.Error ?? ''
  const bodyLooksLikeAppShell = looksLikeViteDevShell(rawBody)
  const body = bodyLooksLikeAppShell
    ? 'The response is the traCtl dev app page, not an API. Use an absolute external URL such as https://httpbin.org/get.'
    : rawBody

  return {
    passed: run.Passed === true && !bodyLooksLikeAppShell,
    durationMs: timing.total,
    statusCode,
    statusText: responseStatusLabel(statusCode),
    body,
    headers,
    timing,
    assertionResults,
    extractResults: extractResultsFromDiagnostics(run.diagnostics),
    assertionsPassed: passedCount,
    assertionsTotal: assertionResults.length,
    error:
      step.Error ||
      (bodyLooksLikeAppShell
        ? 'Request hit the traCtl dev server instead of an external API. Check the URL.'
        : undefined),
  }
}

function mapWasmFailure(code: string, message: string): RunResult {
  const detail = `${code}: ${message}`
  return {
    passed: false,
    durationMs: 0,
    statusCode: 0,
    statusText: 'Error',
    body: detail,
    headers: {},
    timing: { dns: 0, tcp: 0, tls: 0, ttfb: 0, transfer: 0, total: 0, unit: 'ms' },
    assertionResults: [],
    extractResults: [],
    assertionsPassed: 0,
    assertionsTotal: 0,
    error: detail,
  }
}

function responseStatusLabel(statusCode: number): string {
  if (statusCode <= 0) return 'Error'

  const known: Record<number, string> = {
    200: 'OK',
    201: 'Created',
    202: 'Accepted',
    204: 'No Content',
    301: 'Moved Permanently',
    302: 'Found',
    304: 'Not Modified',
    400: 'Bad Request',
    401: 'Unauthorized',
    403: 'Forbidden',
    404: 'Not Found',
    408: 'Request Timeout',
    409: 'Conflict',
    422: 'Unprocessable Entity',
    429: 'Too Many Requests',
    500: 'Internal Server Error',
    502: 'Bad Gateway',
    503: 'Service Unavailable',
    504: 'Gateway Timeout',
  }

  return `${statusCode} ${known[statusCode] ?? ''}`.trim()
}

function timingFromDiagnostics(diagnostics?: WasmDiagnostics): RunResult['timing'] {
  const timeline =
    diagnostics?.Workflows?.[0]?.Steps?.[0]?.Requests?.[0]?.Timeline ?? undefined
  if (!timeline) {
    return { dns: 0, tcp: 0, tls: 0, ttfb: 0, transfer: 0, total: 0, unit: 'ms' }
  }
  return {
    dns: timeline.DNSMs ?? 0,
    tcp: timeline.TCPMs ?? 0,
    tls: timeline.TLSMs ?? 0,
    ttfb: timeline.TTFBMs ?? 0,
    transfer: timeline.TransferMs ?? 0,
    total: timeline.TotalMs ?? 0,
    unit: 'ms',
  }
}

function extractResultsFromDiagnostics(diagnostics?: WasmDiagnostics): RunResult['extractResults'] {
  const extracts = diagnostics?.Workflows?.[0]?.Steps?.[0]?.Extracts ?? []
  return extracts.map((extract, index) => ({
    id: extract.ID || `extract-${index + 1}`,
    variableName: extract.As || extract.ID || 'extract',
    scope: 'workflow',
    resolvedValue: extract.Source || '',
  }))
}
