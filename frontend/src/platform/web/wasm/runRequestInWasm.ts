import type { TraCtlSpecDocument } from '@/components/request-editor/tractlSpecDocument'
import { looksLikeViteDevShell } from '@/components/request-editor/normalizeRequestUrl'
import type { RequestRunResult } from '@/platform/localApi/types'
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
  AssertionResults?: WasmAssertionOutcome[]
  ResponseStatus?: number
  ResponseHeaders?: Record<string, string>
  ResponseBody?: string
  Error?: string
}

type WasmAssertionOutcome = {
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
): Promise<RequestRunResult> {
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
): RequestRunResult {
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
      passed: assertion.Outcome === 'pass',
      label: `${kind} ${id}`,
      detail: assertion.Message || '',
    }
  })

  const passedCount = assertionResults.filter((assertion) => assertion.passed).length
  const timeline = timelineFromDiagnostics(run.diagnostics)
  const durationMs = timeline.durationMs
  const headers = responseHeaders(step.ResponseHeaders ?? {})
  const statusCode = step.ResponseStatus ?? 0
  const statusLabel = responseStatusLabel(statusCode)
  const contentType =
    headerValue(step.ResponseHeaders ?? {}, 'content-type') || 'application/json'
  const rawBody = step.ResponseBody ?? step.Error ?? ''
  const bodyLooksLikeAppShell = looksLikeViteDevShell(rawBody)
  const body = bodyLooksLikeAppShell
    ? 'The response is the traCtl dev app page, not an API. Use an absolute external URL such as https://httpbin.org/get.'
    : rawBody

  return {
    passed: run.Passed === true && !bodyLooksLikeAppShell,
    durationMs,
    statusCode,
    statusLabel,
    contentType,
    body,
    headers,
    assertionResults,
    extractResults: extractResultsFromDiagnostics(run.diagnostics),
    passedCount,
    totalCount: assertionResults.length,
    timeline: timeline.segments,
    error:
      step.Error ||
      (bodyLooksLikeAppShell
        ? 'Request hit the traCtl dev server instead of an external API. Check the URL.'
        : undefined),
  }
}

function mapWasmFailure(code: string, message: string): RequestRunResult {
  const detail = `${code}: ${message}`
  return {
    passed: false,
    durationMs: 0,
    statusCode: 0,
    statusLabel: 'Error',
    contentType: 'text/plain',
    body: detail,
    headers: [],
    assertionResults: [],
    extractResults: [],
    passedCount: 0,
    totalCount: 0,
    timeline: [],
    error: detail,
  }
}

function responseHeaders(headers: Record<string, string>) {
  return Object.entries(headers).map(([key, value], index) => ({
    id: `response-header-${index + 1}`,
    enabled: true,
    key,
    value,
  }))
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

function headerValue(headers: Record<string, string>, key: string): string {
  const match = Object.entries(headers).find(
    ([headerKey]) => headerKey.toLowerCase() === key,
  )
  return match?.[1] ?? ''
}

function timelineFromDiagnostics(diagnostics?: WasmDiagnostics): {
  segments: RequestRunResult['timeline']
  durationMs: number
} {
  const timeline =
    diagnostics?.Workflows?.[0]?.Steps?.[0]?.Requests?.[0]?.Timeline ?? undefined
  if (!timeline) {
    return { segments: [], durationMs: 0 }
  }

  return {
    segments: [
      { label: 'DNS', ms: timeline.DNSMs ?? 0 },
      { label: 'TCP', ms: timeline.TCPMs ?? 0 },
      { label: 'TLS', ms: timeline.TLSMs ?? 0 },
      { label: 'TTFB', ms: timeline.TTFBMs ?? 0 },
      { label: 'Transfer', ms: timeline.TransferMs ?? 0 },
    ],
    durationMs: timeline.TotalMs ?? 0,
  }
}

function extractResultsFromDiagnostics(diagnostics?: WasmDiagnostics) {
  const extracts = diagnostics?.Workflows?.[0]?.Steps?.[0]?.Extracts ?? []
  return extracts.map((extract) => {
    const variable = extract.As || extract.ID || 'extract'
    return {
      variable: `steps.this.extracts.${variable}`,
      value: extract.Source || '',
      scope: 'workflow',
    }
  })
}
