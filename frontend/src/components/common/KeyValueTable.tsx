import { IconX } from '@tabler/icons-react'
import { EnvironmentVariableInput } from '@/components/environment/EnvironmentVariableInput'
import { AddRow } from '@/components/common/AddRow'
import type { KVRow } from '@/components/request-editor/types'

export type KeyValueTableProps = {
  rows: KVRow[]
  keyPlaceholder?: string
  valuePlaceholder?: string
  monoValue?: boolean
  addLabel?: string
  readOnly?: boolean
  onAdd?: () => void
  onChange?: (rows: KVRow[]) => void
  onUpdate?: (id: string, patch: Partial<Omit<KVRow, 'id'>>) => void
  onRemove?: (id: string) => void
}

export function KeyValueTable({
  rows,
  keyPlaceholder = 'key',
  valuePlaceholder = 'value',
  monoValue = true,
  addLabel = 'Add row',
  readOnly = false,
  onAdd,
  onChange,
  onUpdate,
  onRemove,
}: KeyValueTableProps) {
  const updateRow = (id: string, patch: Partial<Omit<KVRow, 'id'>>) => {
    if (onUpdate) {
      onUpdate(id, patch)
      return
    }
    if (!onChange) return
    onChange(
      rows.map((row) => (row.id === id ? { ...row, ...patch } : row)),
    )
  }

  const removeRow = (id: string) => {
    if (onRemove) {
      onRemove(id)
      return
    }
    onChange?.(rows.filter((row) => row.id !== id))
  }

  return (
    <div>
      <table className="w-full border-collapse">
        <thead>
          <tr className="border-b-[0.5px] border-border text-left text-[10px] uppercase tracking-[0.05em] text-text-muted">
            <th className="w-10 py-[var(--density-padding-xs)] font-medium" />
            <th className="py-[var(--density-padding-xs)] font-medium">Key</th>
            <th className="py-[var(--density-padding-xs)] font-medium">Value</th>
            {!readOnly ? <th className="w-8 py-[var(--density-padding-xs)]" /> : null}
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr
              key={row.id}
              className="border-b-[0.5px] border-border"
              style={{ height: 'var(--density-row-height)' }}
            >
              <td className="px-[var(--density-padding-xs)]">
                <input
                  aria-label={`${row.key || 'row'} enabled`}
                  type="checkbox"
                  checked={row.enabled}
                  disabled={readOnly}
                  className="h-3.5 w-3.5"
                  onChange={(event) =>
                    updateRow(row.id, { enabled: event.currentTarget.checked })
                  }
                />
              </td>
              <td className="px-[var(--density-padding-xs)]">
                <input
                  aria-label={`${row.key || 'row'} key`}
                  placeholder={keyPlaceholder}
                  className="w-full bg-transparent px-1 text-[length:var(--density-font-mono)] text-text outline-none"
                  value={row.key}
                  readOnly={readOnly}
                  onChange={(event) =>
                    updateRow(row.id, { key: event.currentTarget.value })
                  }
                />
              </td>
              <td className="px-[var(--density-padding-xs)]">
                {readOnly ? (
                  <input
                    readOnly
                    value={row.value}
                    className={cnInput(monoValue)}
                  />
                ) : (
                  <EnvironmentVariableInput
                    aria-label={`${row.key || 'row'} value`}
                    placeholder={valuePlaceholder}
                    className={cnInput(monoValue)}
                    value={row.value}
                    onChange={(value) => updateRow(row.id, { value })}
                  />
                )}
              </td>
              {!readOnly ? (
                <td className="px-[var(--density-padding-xs)] text-right">
                  <button
                    type="button"
                    aria-label="Delete row"
                    className="text-text-muted hover:text-danger-fg"
                    onClick={() => removeRow(row.id)}
                  >
                    <IconX size={14} stroke={1.75} />
                  </button>
                </td>
              ) : null}
            </tr>
          ))}
        </tbody>
      </table>
      {!readOnly ? <AddRow label={addLabel} onClick={onAdd} /> : null}
    </div>
  )
}

function cnInput(mono: boolean) {
  return [
    'w-full bg-transparent px-1 outline-none',
    mono
      ? 'font-mono text-[length:var(--density-font-mono)]'
      : 'text-[length:var(--density-font-ui)]',
    'text-text',
  ].join(' ')
}
