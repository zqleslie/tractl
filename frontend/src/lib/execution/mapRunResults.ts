import type { RequestRunResult } from '@/platform/localApi/types'
import type { RunResult } from '@/stores/executionStore'

export function timelineToTiming(
  result: RequestRunResult,
): RunResult['timing'] {
  if (result.timing) return result.timing
  const find = (label: string) =>
    result.timeline?.find(
      (segment) => segment.label.toLowerCase() === label.toLowerCase(),
    )?.ms ?? 0
  return {
    dns: find('DNS'),
    tcp: find('TCP'),
    tls: find('TLS'),
    ttfb: find('TTFB'),
    transfer: find('Transfer'),
    total: result.durationMs,
  }
}

export function requestRunResultToExecutionResult(
  result: RequestRunResult,
): RunResult {
  return {
    statusCode: result.statusCode,
    statusText: result.statusText || result.statusLabel || String(result.statusCode),
    durationMs: result.durationMs,
    body: result.body,
    headers: Array.isArray(result.headers)
      ? Object.fromEntries(
          result.headers
            .filter((row) => row.key.trim())
            .map((row) => [row.key, row.value]),
        )
      : result.headers,
    timing: timelineToTiming(result),
    assertionResults: result.assertionResults.map((row) => ({
      id: row.id,
      kind: row.kind ?? 'assertion',
      op: row.op ?? 'equals',
      expected: row.expected ?? row.label ?? '',
      received: row.received ?? row.detail ?? '',
      passed: row.passed,
      severity: row.severity ?? 'error',
    })),
    extractResults: result.extractResults.map((row) => ({
      id: row.id ?? row.variableName ?? row.variable ?? 'extract',
      variableName: row.variableName ?? row.variable ?? 'extract',
      scope:
        row.scope === 'spec' || row.scope === 'step' ? row.scope : 'workflow',
      resolvedValue: row.resolvedValue ?? row.value ?? '',
    })),
    assertionsPassed: result.assertionsPassed ?? result.passedCount ?? 0,
    assertionsTotal: result.assertionsTotal ?? result.totalCount ?? 0,
  }
}

export function executionResultToRequestRunResult(
  result: RunResult,
): RequestRunResult {
  return {
    durationMs: result.durationMs,
    statusCode: result.statusCode,
    statusText: result.statusText,
    body: result.body,
    headers: result.headers,
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
      variableName: row.variableName,
      scope: row.scope,
      resolvedValue: row.resolvedValue ?? '',
    })),
    assertionsPassed: result.assertionsPassed,
    assertionsTotal: result.assertionsTotal,
    timing: { ...result.timing, unit: 'ms' },
  }
}
