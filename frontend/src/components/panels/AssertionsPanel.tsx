import { AddRow } from '@/components/common/AddRow'
import { EmptyState } from '@/components/common/EmptyState'
import { AssertionRow } from '@/components/request-editor/shared/AssertionRow'
import type { AssertionRowModel } from '@/components/request-editor/types'

export type AssertionsPanelProps = {
  value: AssertionRowModel[]
  onAdd: () => void
  onUpdate: (id: string, patch: Partial<Omit<AssertionRowModel, 'id'>>) => void
  onRemove: (id: string) => void
  readOnly?: boolean
}

export function AssertionsPanel({
  value,
  onAdd,
  onUpdate,
  onRemove,
  readOnly,
}: AssertionsPanelProps) {
  if (value.length === 0) {
    return (
      <div className="space-y-2">
        <EmptyState
          icon="shield-check"
          text="No assertions yet"
          sub="Add an assertion to validate this response automatically"
        />
        {!readOnly ? <AddRow label="Add assertion" onClick={onAdd} /> : null}
      </div>
    )
  }

  return (
    <div className="space-y-[var(--density-gap-sm)]">
      {value.map((row) => (
        <AssertionRow
          key={row.id}
          row={row}
          onChange={(patch) => onUpdate(row.id, patch)}
          onRemove={() => onRemove(row.id)}
        />
      ))}
      {!readOnly ? <AddRow label="Add assertion" onClick={onAdd} /> : null}
    </div>
  )
}
