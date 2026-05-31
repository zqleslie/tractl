import type { RecentItem } from '@/types/recent'
import { formatTimeAgo } from '@/lib/fastStart/formatTimeAgo'
import type { RunHistoryEntry } from '@/stores/runHistoryStore'

function displayName(entry: RunHistoryEntry): string {
  if (entry.sourceName?.trim()) return entry.sourceName.trim()
  if (entry.requestName.trim() && entry.requestName !== 'Untitled request') {
    return entry.requestName
  }
  return `${entry.method} ${entry.url}`
}

function recentType(entry: RunHistoryEntry): RecentItem['type'] {
  return entry.sourceType === 'workflow' ? 'workflow' : 'request'
}

export function buildRecentItems(
  entries: RunHistoryEntry[],
  limit = 5,
): RecentItem[] {
  return entries.slice(0, limit).map((entry) => ({
    id: entry.id,
    historyEntryId: entry.id,
    name: displayName(entry),
    type: recentType(entry),
    timeAgo: formatTimeAgo(entry.timestamp),
  }))
}
