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
      statusLabel: '200',
      contentType: 'application/json',
      body: '{}',
      headers: [],
      assertionResults: [],
      extractResults: [],
      passedCount: 0,
      totalCount: 0,
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
