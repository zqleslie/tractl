import { beforeEach, describe, expect, it } from 'vitest'
import {
  RUN_HISTORY_MAX_ENTRIES,
  useRunHistoryStore,
  type RunHistoryEntry,
} from '@/stores/runHistoryStore'

function makeEntry(index: number): RunHistoryEntry {
  return {
    id: `entry-${index}`,
    timestamp: new Date().toISOString(),
    requestName: `Request ${index}`,
    method: 'GET',
    url: `https://example.com/${index}`,
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

describe('runHistoryStore', () => {
  beforeEach(() => {
    useRunHistoryStore.setState({ entries: [], pinnedIds: [] })
  })

  it('caps history at fifty entries', () => {
    for (let index = 0; index < RUN_HISTORY_MAX_ENTRIES + 5; index += 1) {
      useRunHistoryStore.getState().addEntry(makeEntry(index))
    }

    expect(useRunHistoryStore.getState().entries).toHaveLength(
      RUN_HISTORY_MAX_ENTRIES,
    )
  })

  it('does not evict pinned entries when trimming', () => {
    const pinned = makeEntry(0)
    useRunHistoryStore.getState().addEntry(pinned)
    useRunHistoryStore.getState().pinEntry(pinned.id)

    for (let index = 1; index <= RUN_HISTORY_MAX_ENTRIES + 3; index += 1) {
      useRunHistoryStore.getState().addEntry(makeEntry(index))
    }

    expect(
      useRunHistoryStore.getState().entries.some((entry) => entry.id === pinned.id),
    ).toBe(true)
  })
})
