import type { RunResult as PlatformRunResult } from '@/platform/types'
import type { RunResult as ExecutionRunResult } from '@/stores/executionStore'

export function requestRunResultToExecutionResult(
  result: PlatformRunResult,
): ExecutionRunResult {
  return {
    statusCode: result.statusCode,
    statusText: result.statusText || String(result.statusCode),
    durationMs: result.durationMs,
    body: result.body,
    headers: result.headers,
    timing: result.timing,
    assertionResults: result.assertionResults.map((row) => ({
      id: row.id,
      kind: row.kind,
      op: row.op,
      expected: row.expected,
      received: row.received,
      passed: row.passed,
      severity: (row.severity === 'warning' ? 'warning' : 'error') as 'error' | 'warning',
    })),
    extractResults: result.extractResults.map((row) => ({
      id: row.id,
      variableName: row.variableName,
      scope: (row.scope === 'spec' || row.scope === 'step' ? row.scope : 'workflow') as 'workflow' | 'spec' | 'step',
      resolvedValue: row.resolvedValue,
    })),
    assertionsPassed: result.assertionsPassed,
    assertionsTotal: result.assertionsTotal,
  }
}
