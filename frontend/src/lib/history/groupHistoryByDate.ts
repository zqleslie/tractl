import type { RunHistoryEntry } from '@/stores/runHistoryStore'

export type HistoryDateGroup = {
  label: string
  entries: RunHistoryEntry[]
}

function startOfDay(date: Date): Date {
  const copy = new Date(date)
  copy.setHours(0, 0, 0, 0)
  return copy
}

function groupLabelForDate(date: Date, today: Date, yesterday: Date): string {
  const day = startOfDay(date).getTime()
  if (day === today.getTime()) return 'Today'
  if (day === yesterday.getTime()) return 'Yesterday'
  return date.toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    year: date.getFullYear() !== today.getFullYear() ? 'numeric' : undefined,
  })
}

export function groupHistoryByDate(entries: RunHistoryEntry[]): HistoryDateGroup[] {
  const now = new Date()
  const today = startOfDay(now)
  const yesterday = new Date(today)
  yesterday.setDate(yesterday.getDate() - 1)

  const groups = new Map<string, RunHistoryEntry[]>()

  for (const entry of entries) {
    const label = groupLabelForDate(new Date(entry.timestamp), today, yesterday)
    const bucket = groups.get(label) ?? []
    bucket.push(entry)
    groups.set(label, bucket)
  }

  return Array.from(groups.entries()).map(([label, groupEntries]) => ({
    label,
    entries: groupEntries,
  }))
}
