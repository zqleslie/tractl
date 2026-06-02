import { describe, expect, it } from 'vitest'
import { groupHistoryByDate } from '@/lib/history/groupHistoryByDate'
import type { RunHistoryEntry } from '@/stores/runHistoryStore'

function entry(id: string, timestamp: string): RunHistoryEntry {
  return {
    id,
    timestamp,
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
  }
}

describe('groupHistoryByDate', () => {
  it('groups entries under Today and Yesterday labels', () => {
    const now = new Date()
    const today = now.toISOString()
    const yesterday = new Date(now)
    yesterday.setDate(yesterday.getDate() - 1)

    const groups = groupHistoryByDate([
      entry('1', today),
      entry('2', yesterday.toISOString()),
    ])

    expect(groups[0]?.label).toBe('Today')
    expect(groups[1]?.label).toBe('Yesterday')
  })
})
