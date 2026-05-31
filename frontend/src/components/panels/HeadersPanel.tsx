import { KeyValueTable } from '@/components/common/KeyValueTable'
import type { KVRow } from '@/components/request-editor/types'

export type HeadersPanelProps = {
  value: KVRow[]
  onAdd: () => void
  onUpdate: (id: string, patch: Partial<Omit<KVRow, 'id'>>) => void
  onRemove: (id: string) => void
  readOnly?: boolean
}

export function HeadersPanel({
  value,
  onAdd,
  onUpdate,
  onRemove,
  readOnly,
}: HeadersPanelProps) {
  return (
    <KeyValueTable
      rows={value}
      addLabel="Add header"
      valuePlaceholder="value or ${runtime.traceId}"
      readOnly={readOnly}
      onAdd={onAdd}
      onUpdate={onUpdate}
      onRemove={onRemove}
    />
  )
}
