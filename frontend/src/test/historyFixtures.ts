import type { RunHistoryEntry } from '@/stores/runHistoryStore'

export function makeHistoryEntry(overrides: Partial<RunHistoryEntry> & Pick<RunHistoryEntry, 'id' | 'timestamp'>): RunHistoryEntry {
  return {
    requestName: 'Test',
    method: 'GET',
    url: 'https://example.com',
    statusCode: 200,
    durationMs: 1,
    outcome: 'success',
    result: {
      passed: true,
      durationMs: 1,
      statusCode: 200,
      statusText: '200',
      contentType: 'application/json',
      body: '{}',
      headers: {},
      timing: { dns: 0, tcp: 0, tls: 0, ttfb: 0, transfer: 0, total: 1, unit: 'ms' },
      assertionResults: [],
      extractResults: [],
      assertionsPassed: 0,
      assertionsTotal: 0,
      timeline: [],
    },
    ...overrides,
  }
}
