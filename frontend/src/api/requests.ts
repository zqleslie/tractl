import type {
  AssertionResult,
  ExtractResult,
  RunResult,
} from '@/stores/executionStore'

export type AssertionDef = {
  id: string
  kind: string
  op: string
  expected: string
  severity: 'error' | 'warning'
}

export type ExtractDef = {
  id: string
  source: string
  path: string
  variableName: string
  scope: string
}

export type RunRequestPayload = {
  method: string
  url: string
  headers: Record<string, string>
  body?: { encoding: string; content: string }
  auth?: { type: string; token?: string }
  assertions: AssertionDef[]
  extracts: ExtractDef[]
  settings: { timeoutMs: number; failurePolicy: string }
  environmentId: string | null
}

export async function runRequest(_payload: RunRequestPayload): Promise<RunResult> {
  await new Promise((resolve) => setTimeout(resolve, 1200))
  return {
    statusCode: 201,
    statusText: 'Created',
    durationMs: 234,
    body: JSON.stringify(
      { userId: 'usr_01HXYZ9ABC', email: 'test@example.com' },
      null,
      2,
    ),
    headers: {
      'content-type': 'application/json',
      'x-request-id': 'req_01HXYZ',
    },
    timing: {
      dns: 4,
      tcp: 12,
      tls: 23,
      ttfb: 187,
      transfer: 8,
      total: 234,
    },
    assertionResults: [] as AssertionResult[],
    extractResults: [] as ExtractResult[],
    assertionsPassed: 0,
    assertionsTotal: 0,
  }
}

export async function saveRequestFile(
  _payload: { id?: string | null; name: string; yaml: string },
): Promise<{ id: string; updatedAt: string }> {
  await new Promise((resolve) => setTimeout(resolve, 80))
  return { id: _payload.id ?? `file-${Date.now()}`, updatedAt: new Date().toISOString() }
}
