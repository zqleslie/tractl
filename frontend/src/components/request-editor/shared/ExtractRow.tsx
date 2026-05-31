import { IconTrash } from '@tabler/icons-react'
import type { ExtractRowModel } from '@/components/request-editor/types'

const sourceOptions = ['body', 'header', 'status', 'timing', 'metadata', 'extension']
const scopeOptions = ['workflow', 'spec', 'step']

export type ExtractRowProps = {
  row: ExtractRowModel
  onChange: (patch: Partial<Omit<ExtractRowModel, 'id'>>) => void
  onRemove: () => void
}

export function ExtractRow({ row, onChange, onRemove }: ExtractRowProps) {
  return (
    <div className="flex items-center gap-1.5 rounded-ui border-[0.5px] border-border bg-surface-elevated p-1.5">
      <select
        className="rounded-ui border-[0.5px] border-border bg-surface px-1.5 py-1 text-ui-xs text-text"
        value={row.source}
        onChange={(event) => onChange({ source: event.currentTarget.value })}
      >
        {sourceOptions.map((option) => (
          <option key={option} value={option}>
            {option}
          </option>
        ))}
      </select>
      <input
        className="min-w-0 flex-1 rounded-ui border-[0.5px] border-border bg-surface px-1.5 py-1 font-mono text-ui-xs text-text"
        value={row.path}
        onChange={(event) => onChange({ path: event.currentTarget.value })}
      />
      <span className="text-text-muted">-&gt;</span>
      <input
        className="min-w-0 flex-1 rounded-ui border-[0.5px] border-border bg-surface px-1.5 py-1 font-mono text-ui-xs text-text"
        value={row.variable}
        onChange={(event) => onChange({ variable: event.currentTarget.value })}
      />
      <select
        className="rounded-ui border-[0.5px] border-border bg-surface px-1.5 py-1 text-ui-xs text-text"
        value={row.scope}
        onChange={(event) => onChange({ scope: event.currentTarget.value })}
      >
        {scopeOptions.map((option) => (
          <option key={option} value={option}>
            {option}
          </option>
        ))}
      </select>
      <button
        type="button"
        aria-label={`Delete ${row.variable} extract`}
        className="rounded-ui p-1 text-text-muted hover:text-danger-fg"
        onClick={onRemove}
      >
        <IconTrash size={13} stroke={1.75} />
      </button>
    </div>
  )
}
