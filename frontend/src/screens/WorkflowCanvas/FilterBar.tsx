import { IconPlus } from '@tabler/icons-react'
import { cn } from '@/lib/cn'

export interface FilterBarProps {
  workflowNames: string[]
  activeFilterId: string | null
  onFilterChange: (id: string | null) => void
  onAddWorkflow: () => void
}

const chipBase =
  'inline-flex h-6 shrink-0 items-center rounded-full border border-[0.5px] px-3 text-[11px] font-[400] transition-colors'
const chipInactive = 'border-border text-text-muted hover:text-text'
const chipActive =
  'border-primary bg-primary-bg text-primary'

const OVERFLOW_THRESHOLD = 5

export function FilterBar({
  workflowNames,
  activeFilterId,
  onFilterChange,
  onAddWorkflow,
}: FilterBarProps) {
  const useDropdown = workflowNames.length > OVERFLOW_THRESHOLD

  return (
    <div className="flex h-9 shrink-0 items-center gap-2 overflow-x-auto border-b border-[0.5px] border-border bg-surface px-3">
      <span className="shrink-0 text-[11px] font-[400] text-text-muted">Show</span>

      <button
        type="button"
        onClick={onAddWorkflow}
        title="Add workflow"
        aria-label="Add workflow"
        className={cn(
          chipBase,
          'gap-1 border-dashed border-border text-text-muted hover:border-primary hover:text-primary',
        )}
      >
        <IconPlus size={12} stroke={1.5} />
        Add workflow
      </button>

      <button
        type="button"
        onClick={() => onFilterChange(null)}
        className={cn(chipBase, activeFilterId === null ? chipActive : chipInactive)}
      >
        All workflows
      </button>

      {useDropdown ? (
        <>
          {activeFilterId !== null && (
            <span className={cn(chipBase, chipActive)}>{activeFilterId}</span>
          )}
          <select
            value={activeFilterId ?? ''}
            onChange={(event) =>
              onFilterChange(event.target.value === '' ? null : event.target.value)
            }
            aria-label="Filter by workflow"
            className="h-6 shrink-0 rounded-[7px] border border-[0.5px] border-border bg-surface px-2 text-[11px] font-[400] text-text outline-none"
          >
            <option value="">All workflows</option>
            {workflowNames.map((name) => (
              <option key={name} value={name}>
                {name}
              </option>
            ))}
          </select>
        </>
      ) : (
        workflowNames.map((name) => (
          <button
            key={name}
            type="button"
            onClick={() => onFilterChange(name)}
            className={cn(chipBase, activeFilterId === name ? chipActive : chipInactive)}
          >
            {name}
          </button>
        ))
      )}
    </div>
  )
}
