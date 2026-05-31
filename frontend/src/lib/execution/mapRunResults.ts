import type { RequestRunResult } from '@/platform/localApi/types'
import type { RunResult } from '@/stores/executionStore'

export function timelineToTiming(
  timeline: RequestRunResult['timeline'],
  durationMs: number,
): RunResult['timing'] {
  const find = (label: string) =>
    timeline.find((segment) => segment.label.toLowerCase() === label.toLowerCase())
      ?.ms ?? 0

  const dns = find('DNS')
  const tcp = find('TCP')
  const tls = find('TLS')
  const ttfb = find('TTFB')
  const transfer = find('Transfer')
  const total =
    timeline.length > 0
      ? dns + tcp + tls + ttfb + transfer
      : durationMs

  return { dns, tcp, tls, ttfb, transfer, total: total || durationMs }
}

export function requestRunResultToExecutionResult(
  result: RequestRunResult,
): RunResult {
  const headers: Record<string, string> = {}
  for (const row of result.headers) {
    if (row.key.trim()) headers[row.key] = row.value
  }

  return {
    statusCode: result.statusCode,
    statusText: result.statusLabel || String(result.statusCode),
    durationMs: result.durationMs,
    body: result.body,
    headers,
    timing: timelineToTiming(result.timeline, result.durationMs),
    assertionResults: result.assertionResults.map((row) => ({
      id: row.id,
      kind: 'assertion',
      op: 'equals',
      expected: row.label,
      received: row.detail,
      passed: row.passed,
      severity: 'error' as const,
    })),
    extractResults: result.extractResults.map((row, index) => ({
      id: `extract-${index}`,
      variableName: row.variable,
      scope: (row.scope as 'workflow' | 'spec' | 'step') ?? 'workflow',
      resolvedValue: row.value,
    })),
    assertionsPassed: result.passedCount,
    assertionsTotal: result.totalCount,
  }
}

export function executionResultToRequestRunResult(
  result: RunResult,
): RequestRunResult {
  const headers = Object.entries(result.headers).map(([key, value], index) => ({
    id: `header-${index}`,
    enabled: true,
    key,
    value,
  }))

  return {
    passed:
      result.assertionsTotal === 0 ||
      result.assertionsPassed === result.assertionsTotal,
    durationMs: result.durationMs,
    statusCode: result.statusCode,
    statusLabel: result.statusText,
    contentType: result.headers['content-type'] ?? 'application/json',
    body: result.body,
    headers,
    assertionResults: result.assertionResults.map((row) => ({
      id: row.id,
      passed: row.passed,
      label: `${row.kind} ${row.op} ${row.expected}`,
      detail: row.received,
    })),
    extractResults: result.extractResults.map((row) => ({
      variable: row.variableName,
      value: row.resolvedValue ?? '',
      scope: row.scope,
    })),
    passedCount: result.assertionsPassed,
    totalCount: result.assertionsTotal,
    timeline: [
      { label: 'DNS', ms: result.timing.dns },
      { label: 'TCP', ms: result.timing.tcp },
      { label: 'TLS', ms: result.timing.tls },
      { label: 'TTFB', ms: result.timing.ttfb },
      { label: 'Transfer', ms: result.timing.transfer },
    ],
  }
}
