import { KeyValueTable } from '@/components/common/KeyValueTable'
import type { KVRow } from '@/components/request-editor/types'

export type ParamsPanelProps = {
  value: KVRow[]
  onAdd: () => void
  onUpdate: (id: string, patch: Partial<Omit<KVRow, 'id'>>) => void
  onRemove: (id: string) => void
  readOnly?: boolean
}

export function ParamsPanel({
  value,
  onAdd,
  onUpdate,
  onRemove,
  readOnly,
}: ParamsPanelProps) {
  return (
    <KeyValueTable
      rows={value}
      addLabel="Add parameter"
      keyPlaceholder="parameter"
      readOnly={readOnly}
      onAdd={onAdd}
      onUpdate={onUpdate}
      onRemove={onRemove}
    />
  )
}
