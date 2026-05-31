import { describe, expect, it } from 'vitest'
import { buildRecentItems } from '@/lib/fastStart/buildRecentItems'
import type { RunHistoryEntry } from '@/stores/runHistoryStore'

function historyEntry(
  partial: Partial<RunHistoryEntry> & Pick<RunHistoryEntry, 'id' | 'timestamp'>,
): RunHistoryEntry {
  return {
    requestName: 'Test',
    method: 'GET',
    url: 'https://example.com',
    statusCode: 200,
    durationMs: 10,
    outcome: 'success',
    result: {
      passed: true,
      durationMs: 10,
      statusCode: 200,
      statusText: '200 OK',
      body: '{}',
      headers: {},
      timing: { dns: 0, tcp: 0, tls: 0, ttfb: 0, transfer: 0, total: 10, unit: 'ms' },
      assertionResults: [],
      extractResults: [],
      assertionsPassed: 0,
      assertionsTotal: 0,
    },
    ...partial,
  }
}

describe('buildRecentItems', () => {
  it('lists run history newest first', () => {
    const items = buildRecentItems([
      historyEntry({
        id: 'newer',
        timestamp: '2026-05-29T12:00:00.000Z',
        requestName: 'Newer run',
        sourceType: 'workflow',
        sourceName: '02-parallel-with-dependency.yaml',
      }),
      historyEntry({
        id: 'older',
        timestamp: '2026-05-29T10:00:00.000Z',
        requestName: 'Older run',
        sourceType: 'request',
      }),
    ])

    expect(items).toHaveLength(2)
    expect(items[0]?.name).toBe('02-parallel-with-dependency.yaml')
    expect(items[0]?.type).toBe('workflow')
    expect(items[1]?.name).toBe('Older run')
    expect(items[1]?.type).toBe('request')
  })
})
