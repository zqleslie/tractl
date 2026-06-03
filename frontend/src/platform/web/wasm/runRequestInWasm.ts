import { looksLikeViteDevShell } from '@/components/request-editor/normalizeRequestUrl'
import type { AssertionResultRow, ExtractResultRow, RequestRunResult } from '@/platform/localApi/types'
import type { EngineRunResult } from '@/platform/web/workflowRunAdapter'
import {
  isTractlWasmRunSuccess,
  loadTractlWasmRuntime,
  type TractlWasmRunResult,
} from '@/platform/web/wasm/loadTractlWasmRuntime'
import type { RequestDef } from '@/types/requestDef'

export async function runRequestDefInWasm(def: RequestDef): Promise<RequestRunResult> {
  await loadTractlWasmRuntime()
  if (!window.tractl?.runRequest) {
    return mapWasmFailure('TRACTL_WASM_UNAVAILABLE', 'WASM runtime is not ready')
  }
  const result = await window.tractl.runRequest(JSON.stringify(def))
  if (!isTractlWasmRunSuccess(result)) {
    return mapWasmFailure(result.error.code, result.error.message)
  }
  return mapLocalapiRunResult(result as unknown as LocalapiRunResult)
}

// localapi.RunResult shape — camelCase JSON tags from internal/localapi/types.go.
type LocalapiRunResult = {
  statusCode: number
  statusText: string
  durationMs: number
  body: string
  headers: Record<string, string>
  timing: { dns: number; tcp: number; tls: number; ttfb: number; transfer: number; total: number; unit: string }
  assertionResults: Array<{ id: string; kind: string; op: string; expected: string; received: string; passed: boolean; severity: string }>
  extractResults: Array<{ id: string; variableName: string; scope: string; resolvedValue: string }>
  assertionsPassed: number
  assertionsTotal: number
  passed: boolean
  error?: string
}

function mapLocalapiRunResult(r: LocalapiRunResult): RequestRunResult {
  const contentType = r.headers?.['content-type'] ?? 'application/json'
  return {
    passed: r.passed,
    statusCode: r.statusCode ?? 0,
    statusText: r.statusText ?? '',
    durationMs: r.timing?.total ?? r.durationMs ?? 0,
    body: r.body ?? '',
    contentType,
    headers: r.headers ?? {},
    timing: {
      dns: r.timing?.dns ?? 0,
      tcp: r.timing?.tcp ?? 0,
      tls: r.timing?.tls ?? 0,
      ttfb: r.timing?.ttfb ?? 0,
      transfer: r.timing?.transfer ?? 0,
      total: r.timing?.total ?? 0,
      unit: 'ms',
    },
    assertionResults: (r.assertionResults ?? []).map((ar) => ({
      id: ar.id,
      kind: ar.kind,
      op: ar.op,
      expected: ar.expected,
      received: ar.received,
      passed: ar.passed,
      severity: (ar.severity === 'warning' ? 'warning' : 'error') as 'error' | 'warning',
    })),
    extractResults: (r.extractResults ?? []).map((ex) => ({
      id: ex.id,
      variable: ex.variableName,
      value: ex.resolvedValue,
      scope: ex.scope,
    })),
    assertionsPassed: r.assertionsPassed ?? 0,
    assertionsTotal: r.assertionsTotal ?? 0,
    timeline: [],
    error: r.error || undefined,
  }
}

export function mapWasmRunResultToRequestRunResult(
  result: TractlWasmRunResult,
): RequestRunResult {
  if (!isTractlWasmRunSuccess(result)) {
    return mapWasmFailure(result.error.code, result.error.message)
  }

  // handleRun returns the raw engine.RunResult in PascalCase (Go default encoding).
  // Extract the first workflow's first step result to produce a flat RequestRunResult.
  const mapped = mapEngineRunResultToRequestRunResult(result as unknown as EngineRunResult)

  // WASM-only: detect when browser fetch accidentally hit the Vite dev server instead
  // of an external API. The engine sees a 200 OK with HTML body and marks it as passed.
  if (looksLikeViteDevShell(mapped.body)) {
    return {
      ...mapped,
      passed: false,
      body: 'The response is the traCtl dev app page, not an API. Use an absolute external URL such as https://httpbin.org/get.',
      error: 'Request hit the traCtl dev server instead of an external API. Check the URL.',
    }
  }

  return mapped
}

function buildStatusText(code: number | undefined): string {
  if (!code) return ''
  const labels: Record<number, string> = {
    200: 'OK', 201: 'Created', 204: 'No Content',
    400: 'Bad Request', 401: 'Unauthorized', 403: 'Forbidden',
    404: 'Not Found', 405: 'Method Not Allowed', 422: 'Unprocessable Entity',
    429: 'Too Many Requests', 500: 'Internal Server Error',
    502: 'Bad Gateway', 503: 'Service Unavailable', 504: 'Gateway Timeout',
  }
  const text = labels[code]
  return text ? `${code} ${text}` : String(code)
}

function mapEngineRunResultToRequestRunResult(run: EngineRunResult): RequestRunResult {
  const wf = run.Workflows?.[0]
  if (!wf || wf.Skipped) {
    return mapWasmFailure('TRACTL_EXECUTION_ERROR', 'Run produced no workflow result')
  }
  const step = wf.Steps?.[0]
  if (!step) {
    return mapWasmFailure('TRACTL_EXECUTION_ERROR', 'Run produced no step result')
  }

  const headers = step.ResponseHeaders ?? {}
  const contentType = headers['content-type'] ?? 'application/json'

  const assertionResults: AssertionResultRow[] = (step.AssertionResults ?? []).map((ar) => ({
    id: ar.AssertionID ?? '',
    kind: ar.Kind ?? '',
    op: ar.Op ?? '',
    expected: ar.Expected ?? '',
    received: ar.Received ?? '',
    passed: ar.Outcome === 'pass',
    severity: (ar.Severity === 'warning' ? 'warning' : 'error') as 'error' | 'warning',
  }))

  const diagWf = run.diagnostics?.Workflows?.[0]
  const diagStep = diagWf?.Steps?.[0]
  const tl = diagStep?.Requests?.[0]?.Timeline

  const timing = {
    dns: tl?.DNSMs ?? 0,
    tcp: tl?.TCPMs ?? 0,
    tls: tl?.TLSMs ?? 0,
    ttfb: tl?.TTFBMs ?? 0,
    transfer: tl?.TransferMs ?? 0,
    total: tl?.TotalMs ?? diagWf?.Duration ?? 0,
    unit: 'ms' as const,
  }

  const extractResults: ExtractResultRow[] = (diagStep?.Extracts ?? []).map((ex) => ({
    id: ex.ID ?? '',
    variable: ex.As ?? ex.ID ?? '',
    value: ex.Source ?? '',
    scope: 'workflow' as const,
  }))

  return {
    passed: !step.Error && !step.CausesFailure,
    statusCode: step.ResponseStatus ?? 0,
    statusText: buildStatusText(step.ResponseStatus),
    durationMs: timing.total,
    body: step.ResponseBody ?? '',
    contentType,
    headers,
    timing,
    assertionResults,
    extractResults,
    assertionsPassed: assertionResults.filter((r) => r.passed).length,
    assertionsTotal: assertionResults.length,
    timeline: [],
    error: step.Error || undefined,
  }
}

function mapWasmFailure(code: string, message: string): RequestRunResult {
  const detail = `${code}: ${message}`
  return {
    passed: false,
    durationMs: 0,
    statusCode: 0,
    statusText: 'Error',
    contentType: 'text/plain',
    body: detail,
    headers: {},
    timing: { dns: 0, tcp: 0, tls: 0, ttfb: 0, transfer: 0, total: 0, unit: 'ms' },
    assertionResults: [],
    extractResults: [],
    assertionsPassed: 0,
    assertionsTotal: 0,
    timeline: [],
    error: detail,
  }
}
