import type { ExtractResultRow, RequestRunResult } from '@/platform/localApi/types'
import type { RunResult } from '@/stores/executionStore'

/**
 * GoExtractResult matches the JSON shape emitted by internal/localapi types.go ExtractResult.
 * Used for both the Go HTTP API response and the WASM handleRun response (both go through
 * localapi.MapRunResult, producing the same flat wire format).
 */
export type GoExtractResult = {
  id: string
  variableName: string
  scope: string
  resolvedValue: string
}

/**
 * GoRunResult is the raw flat JSON shape from POST /api/v1/run and window.tractl.run().
 * Both paths go through localapi.MapRunResult on the Go side, so the wire format is identical.
 * Omits fields that the Go side does not emit and must be derived client-side.
 */
export type GoRunResult = Omit<RequestRunResult, 'extractResults' | 'contentType' | 'timeline'> & {
  extractResults: GoExtractResult[]
}

/**
 * Maps the flat localapi.RunResult wire format to the canonical RequestRunResult.
 * Shared by the HTTP API client (client.ts) and the WASM run mapper (runRequestInWasm.ts).
 */
export function mapFlatRunResult(raw: GoRunResult): RequestRunResult {
  const headers = raw.headers ?? {}
  return {
    ...raw,
    headers,
    contentType: headers['content-type'] ?? 'application/json',
    timeline: [],
    extractResults: (raw.extractResults ?? []).map(
      (row): ExtractResultRow => ({
        id: row.id,
        variable: row.variableName,
        value: row.resolvedValue,
        scope: row.scope,
      }),
    ),
  }
}

export function timelineToTiming(result: RequestRunResult): RunResult['timing'] {
  return result.timing
}

export function requestRunResultToExecutionResult(
  result: RequestRunResult,
): RunResult {
  return {
    statusCode: result.statusCode,
    statusText: result.statusText || String(result.statusCode),
    durationMs: result.durationMs,
    body: result.body,
    headers: result.headers ?? {},
    timing: result.timing,
    assertionResults: (result.assertionResults ?? []).map((row) => ({
      id: row.id,
      kind: row.kind,
      op: row.op,
      expected: row.expected,
      received: row.received,
      passed: row.passed,
      severity: row.severity,
    })),
    extractResults: (result.extractResults ?? []).map((row) => ({
      id: row.id,
      variableName: row.variable,
      scope: row.scope === 'spec' || row.scope === 'step' ? row.scope : 'workflow',
      resolvedValue: row.value,
    })),
    assertionsPassed: result.assertionsPassed,
    assertionsTotal: result.assertionsTotal,
  }
}

export function executionResultToRequestRunResult(
  result: RunResult,
): RequestRunResult {
  return {
    passed: result.assertionsPassed >= result.assertionsTotal,
    durationMs: result.durationMs,
    statusCode: result.statusCode,
    statusText: result.statusText,
    body: result.body,
    contentType: result.headers?.['content-type'] ?? 'application/json',
    headers: result.headers,
    timing: { ...result.timing, unit: 'ms' as const },
    assertionResults: result.assertionResults.map((row) => ({
      id: row.id,
      kind: row.kind,
      op: row.op,
      expected: row.expected,
      received: row.received,
      passed: row.passed,
      severity: row.severity,
    })),
    extractResults: result.extractResults.map((row) => ({
      id: row.id,
      variable: row.variableName,
      value: row.resolvedValue ?? '',
      scope: row.scope,
    })),
    assertionsPassed: result.assertionsPassed,
    assertionsTotal: result.assertionsTotal,
    timeline: [],
  }
}
