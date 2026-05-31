import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { HttpMethod } from '@/components/primitives'
import type { TraCtlSpecDocument } from '@/components/request-editor/tractlSpecDocument'
import type { RequestRunResult } from '@/platform/localApi/types'

export type RunOutcome = 'success' | 'error'
export type RunHistorySourceType = 'request' | 'workflow'
export type RunHistorySourceFormat = 'json' | 'yaml' | 'yml' | 'toon'

export type RunHistoryEntry = {
  id: string
  timestamp: string
  sourceType?: RunHistorySourceType
  sourceName?: string
  sourceFormat?: RunHistorySourceFormat
  document?: TraCtlSpecDocument
  requestName: string
  method: HttpMethod
  url: string
  statusCode: number
  durationMs: number
  outcome: RunOutcome
  result: RequestRunResult
}

export const RUN_HISTORY_MAX_ENTRIES = 50
export const RUN_HISTORY_SIDEBAR_PREVIEW = 25

function trimHistoryEntries(
  entries: RunHistoryEntry[],
  pinnedIds: string[],
  max: number,
): RunHistoryEntry[] {
  if (entries.length <= max) return entries

  const next = [...entries]
  while (next.length > max) {
    let removed = false
    for (let index = next.length - 1; index >= 0; index -= 1) {
      const entry = next[index]
      if (entry && !pinnedIds.includes(entry.id)) {
        next.splice(index, 1)
        removed = true
        break
      }
    }
    if (!removed) break
  }
  return next
}

type RunHistoryState = {
  entries: RunHistoryEntry[]
  pinnedIds: string[]
  addEntry: (entry: RunHistoryEntry) => void
  removeEntry: (id: string) => void
  pinEntry: (id: string) => void
  unpinEntry: (id: string) => void
  clearHistory: () => void
}

export const useRunHistoryStore = create<RunHistoryState>()(
  persist(
    (set) => ({
      entries: [],
      pinnedIds: [],
      addEntry: (entry) =>
        set((state) => {
          const withoutDuplicate = state.entries.filter(
            (candidate) => candidate.id !== entry.id,
          )
          const entries = trimHistoryEntries(
            [entry, ...withoutDuplicate],
            state.pinnedIds,
            RUN_HISTORY_MAX_ENTRIES,
          )
          return { entries }
        }),
      removeEntry: (id) =>
        set((state) => ({
          entries: state.entries.filter((e) => e.id !== id),
          pinnedIds: state.pinnedIds.filter((p) => p !== id),
        })),
      pinEntry: (id) =>
        set((state) =>
          state.pinnedIds.includes(id)
            ? {}
            : { pinnedIds: [...state.pinnedIds, id] },
        ),
      unpinEntry: (id) =>
        set((state) => ({ pinnedIds: state.pinnedIds.filter((p) => p !== id) })),
      clearHistory: () => set({ entries: [], pinnedIds: [] }),
    }),
    { name: 'tractl-run-history' },
  ),
)
