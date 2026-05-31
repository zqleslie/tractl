import { MethodBadge } from '@/components/primitives/MethodBadge'
import { Badge } from '@/components/primitives/Badge'
import { useRunHistoryStore, type RunHistoryEntry } from '@/stores/runHistoryStore'
import { useUiStore } from '@/stores/uiStore'
import { cn } from '@/lib/cn'

function formatTimestamp(iso: string): string {
  const d = new Date(iso)
  const today = new Date()
  const isToday =
    d.getFullYear() === today.getFullYear() &&
    d.getMonth() === today.getMonth() &&
    d.getDate() === today.getDate()
  if (isToday) {
    return d.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit', second: '2-digit' })
  }
  return d.toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function RunHistoryScreen() {
  const entries = useRunHistoryStore((s) => s.entries)
  const clearHistory = useRunHistoryStore((s) => s.clearHistory)
  const openHistoryEntry = useUiStore((s) => s.openHistoryEntry)

  return (
    <div className="flex h-full min-h-0 flex-col" data-testid="screen-run-history">
      <div className="flex shrink-0 items-center gap-2 border-b-[0.5px] border-border px-3 py-2">
        <h2 className="text-ui-sm font-medium text-text">Run History</h2>
        {entries.length > 0 && (
          <button
            type="button"
            className="ml-auto text-ui-xs text-text-muted hover:text-text"
            onClick={clearHistory}
          >
            Clear
          </button>
        )}
      </div>

      {entries.length === 0 ? (
        <div className="flex flex-1 flex-col items-center justify-center gap-2 text-center">
          <p className="text-ui-sm text-text-muted">No runs yet</p>
          <p className="text-ui-xs text-text-muted">
            Execute a request to see its history here.
          </p>
        </div>
      ) : (
        <ul className="min-h-0 flex-1 overflow-y-auto" role="list">
          {entries.map((entry) => (
            <HistoryRow
              key={entry.id}
              entry={entry}
              onOpen={() => openHistoryEntry(entry)}
            />
          ))}
        </ul>
      )}
    </div>
  )
}

function HistoryRow({
  entry,
  onOpen,
}: {
  entry: RunHistoryEntry
  onOpen: () => void
}) {
  const isError = entry.outcome === 'error'
  const statusTone = isError
    ? 'danger'
    : entry.result.statusCode >= 400
      ? 'warning'
      : 'success'

  return (
    <li>
      <button
        type="button"
        className={cn(
          'flex w-full items-center gap-2 border-b-[0.5px] border-border px-3 py-2 text-left',
          'hover:bg-surface-elevated',
        )}
        onClick={onOpen}
        data-testid="history-row"
      >
        <MethodBadge method={entry.method} />

        <span className="min-w-0 flex-1 truncate text-ui-xs text-text">
          {entry.url}
        </span>

        <Badge tone={statusTone} data-testid="history-outcome-badge">
          {isError ? 'Error' : entry.result.statusText || String(entry.statusCode)}
        </Badge>

        <span className="shrink-0 text-ui-xs text-text-muted">
          {entry.durationMs} ms
        </span>

        <span className="shrink-0 text-ui-xs text-text-muted">
          {formatTimestamp(entry.timestamp)}
        </span>
      </button>
    </li>
  )
}
