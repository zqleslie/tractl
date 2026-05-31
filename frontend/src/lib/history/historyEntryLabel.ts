import type { RunHistoryEntry } from '@/stores/runHistoryStore'

export function historyEntryTitle(entry: RunHistoryEntry): string {
  if (entry.requestName.trim() && entry.requestName !== 'Untitled request') {
    return entry.requestName
  }
  try {
    const parsed = new URL(entry.url)
    const path = parsed.pathname === '/' ? '' : parsed.pathname
    return `${parsed.host}${path}`.slice(0, 48) || entry.url
  } catch {
    return entry.url.slice(0, 48) || 'Untitled request'
  }
}
