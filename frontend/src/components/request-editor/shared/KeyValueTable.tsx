import { IconTrash } from '@tabler/icons-react'
import { EnvironmentVariableInput } from '@/components/environment/EnvironmentVariableInput'
import { AddRowButton } from '@/components/request-editor/shared/AddRowButton'
import { VariableReferenceHints } from '@/components/request-editor/shared/VariableReferenceHints'
import type { KeyValueRow } from '@/components/request-editor/types'

export type KeyValueTableProps = {
  rows: KeyValueRow[]
  addLabel?: string
  readOnly?: boolean
  onAdd?: () => void
  onUpdate?: (id: string, patch: Partial<Omit<KeyValueRow, 'id'>>) => void
  onRemove?: (id: string) => void
}

export function KeyValueTable({
  rows,
  addLabel = 'Add row',
  readOnly = false,
  onAdd,
  onUpdate,
  onRemove,
}: KeyValueTableProps) {
  return (
    <div>
      <table className="w-full border-collapse text-ui-sm">
        <thead>
          <tr className="border-b-[0.5px] border-border text-left text-[10px] uppercase tracking-wider text-text-muted">
            <th className="w-12 py-1 font-medium">On</th>
            <th className="py-1 font-medium">Key</th>
            <th className="py-1 font-medium">Value</th>
            {!readOnly ? <th className="w-8 py-1" /> : null}
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr key={row.id} className="border-b-[0.5px] border-border">
              <td className="py-1">
                <input
                  aria-label={`${row.key} enabled`}
                  type="checkbox"
                  checked={row.enabled}
                  disabled={readOnly}
                  className="h-3.5 w-3.5"
                  onChange={(event) =>
                    onUpdate?.(row.id, { enabled: event.currentTarget.checked })
                  }
                />
              </td>
              <td className="py-1 pr-2">
                <input
                  aria-label={`${row.key} key`}
                  placeholder="key"
                  className="w-full rounded-ui border-[0.5px] border-transparent bg-transparent px-1.5 py-1 font-mono text-ui-xs text-text hover:border-border focus:border-border focus:bg-surface-elevated focus:outline-none read-only:cursor-default"
                  value={row.key}
                  readOnly={readOnly}
                  onChange={(event) =>
                    onUpdate?.(row.id, { key: event.currentTarget.value })
                  }
                />
              </td>
              <td className="py-1 pr-2">
                <div>
                  {readOnly ? (
                    <input
                      aria-label={`${row.key} value`}
                      placeholder="value"
                      className="w-full rounded-ui border-[0.5px] border-transparent bg-transparent px-1.5 py-1 font-mono text-ui-xs text-text read-only:cursor-default"
                      value={row.value}
                      readOnly
                    />
                  ) : (
                    <EnvironmentVariableInput
                      aria-label={`${row.key} value`}
                      placeholder="value or {{token}}"
                      className="w-full rounded-ui border-[0.5px] border-transparent bg-transparent px-1.5 py-1 font-mono text-ui-xs text-text hover:border-border focus:border-border focus:bg-surface-elevated focus:outline-none"
                      value={row.value}
                      onChange={(value) => onUpdate?.(row.id, { value })}
                    />
                  )}
                  <VariableReferenceHints value={row.value} />
                </div>
              </td>
              {!readOnly ? (
                <td className="py-1 text-right">
                  <button
                    type="button"
                    aria-label={`Delete ${row.key}`}
                    className="rounded-ui p-1 text-text-muted hover:bg-surface-elevated hover:text-danger-fg"
                    onClick={() => onRemove?.(row.id)}
                  >
                    <IconTrash size={13} stroke={1.75} />
                  </button>
                </td>
              ) : null}
            </tr>
          ))}
        </tbody>
      </table>
      {!readOnly ? <AddRowButton label={addLabel} onClick={onAdd} /> : null}
    </div>
  )
}
