import { beforeEach, describe, expect, it } from 'vitest'
import { RUN_HISTORY_MAX_ENTRIES, useRunHistoryStore } from '@/stores/runHistoryStore'
import { makeHistoryEntry } from '@/test/historyFixtures'

const makeEntry = (index: number) =>
  makeHistoryEntry({
    id: `entry-${index}`,
    timestamp: new Date().toISOString(),
    requestName: `Request ${index}`,
    url: `https://example.com/${index}`,
  })

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
