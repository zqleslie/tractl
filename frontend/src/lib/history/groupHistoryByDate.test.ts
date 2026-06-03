import { describe, expect, it } from 'vitest'
import { groupHistoryByDate } from '@/lib/history/groupHistoryByDate'
import { makeHistoryEntry } from '@/test/historyFixtures'

const entry = (id: string, timestamp: string) => makeHistoryEntry({ id, timestamp })

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
