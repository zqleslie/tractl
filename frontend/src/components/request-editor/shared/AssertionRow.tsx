import { IconX } from '@tabler/icons-react'
import { SeverityBadge } from '@/components/common/SeverityBadge'
import { VariableReferenceHints } from '@/components/request-editor/shared/VariableReferenceHints'
import type { AssertionRowModel } from '@/components/request-editor/types'

const operatorsByKind: Record<string, string[]> = {
  status: ['equals', 'inRange', 'exists'],
  header: ['equals', 'contains', 'exists', 'matches'],
  body: ['equals', 'contains', 'exists', 'matches', 'jsonpath'],
  schema: ['exists'],
  script: ['custom'],
}

export type AssertionRowProps = {
  row: AssertionRowModel
  onChange: (patch: Partial<Omit<AssertionRowModel, 'id'>>) => void
  onRemove: () => void
}

export function AssertionRow({ row, onChange, onRemove }: AssertionRowProps) {
  const operators = operatorsByKind[row.kind] ?? ['equals']

  return (
    <div className="rounded-ui border-[0.5px] border-border bg-surface-elevated p-[var(--density-padding-xs)]">
      <div className="flex flex-wrap items-center gap-[var(--density-gap-sm)]">
        <select
          className="rounded-ui border-[0.5px] border-border bg-surface px-1.5 py-1 text-[length:var(--density-font-label)] text-text"
          value={row.kind}
          onChange={(event) => onChange({ kind: event.currentTarget.value })}
        >
          {Object.keys(operatorsByKind).map((option) => (
            <option key={option} value={option}>
              {option}
            </option>
          ))}
        </select>
        <select
          className="rounded-ui border-[0.5px] border-border bg-surface px-1.5 py-1 text-[length:var(--density-font-label)] text-text"
          value={row.operator}
          onChange={(event) => onChange({ operator: event.currentTarget.value })}
        >
          {operators.map((option) => (
            <option key={option} value={option}>
              {option}
            </option>
          ))}
        </select>
        <input
          className="min-w-0 flex-1 rounded-ui border-[0.5px] border-border bg-surface px-1.5 py-1 font-mono text-[length:var(--density-font-mono)] text-text"
          style={
            row.kind === 'status' || row.kind === 'inRange'
              ? { width: '52px', flex: '0 0 auto' }
              : undefined
          }
          value={row.expected}
          onChange={(event) => onChange({ expected: event.currentTarget.value })}
        />
        <SeverityBadge
          severity={row.severity}
          onToggle={() =>
            onChange({
              severity: row.severity === 'error' ? 'warning' : 'error',
            })
          }
        />
        <button
          type="button"
          aria-label="Delete assertion"
          className="text-text-muted hover:text-danger-fg"
          onClick={onRemove}
        >
          <IconX size={14} stroke={1.75} />
        </button>
      </div>
      <VariableReferenceHints value={row.expected} className="px-0.5" />
    </div>
  )
}
