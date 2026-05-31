import { AddRow } from '@/components/common/AddRow'
import { EmptyState } from '@/components/common/EmptyState'
import { ExtractRow } from '@/components/request-editor/shared/ExtractRow'
import type { ExtractRowModel } from '@/components/request-editor/types'

export type ExtractsPanelProps = {
  value: ExtractRowModel[]
  onAdd: () => void
  onUpdate: (id: string, patch: Partial<Omit<ExtractRowModel, 'id'>>) => void
  onRemove: (id: string) => void
  readOnly?: boolean
}

export function ExtractsPanel({
  value,
  onAdd,
  onUpdate,
  onRemove,
  readOnly,
}: ExtractsPanelProps) {
  const primary = value[0]

  if (value.length === 0) {
    return (
      <div className="space-y-2">
        <EmptyState
          icon="variable"
          text="No extracts defined"
          sub="Extract values from the response to use in downstream steps"
        />
        {!readOnly ? <AddRow label="Add extract" onClick={onAdd} /> : null}
      </div>
    )
  }

  return (
    <div className="space-y-[var(--density-gap-sm)]">
      {value.map((row) => (
        <ExtractRow
          key={row.id}
          row={row}
          onChange={(patch) => onUpdate(row.id, patch)}
          onRemove={() => onRemove(row.id)}
        />
      ))}
      {primary ? (
        <p className="text-[length:var(--density-font-label)] text-text-muted">
          Available downstream as{' '}
          <span className="font-mono text-info-fg">
            {primary.variable.trim()
              ? `\${steps.this.extracts.${primary.variable}}`
              : '(set a variable name)'}
          </span>
        </p>
      ) : null}
      {!readOnly ? <AddRow label="Add extract" onClick={onAdd} /> : null}
    </div>
  )
}
