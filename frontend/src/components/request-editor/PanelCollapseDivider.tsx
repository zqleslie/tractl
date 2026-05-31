import {
  IconChevronDown,
  IconChevronLeft,
  IconChevronRight,
  IconChevronUp,
} from '@tabler/icons-react'
import { cn } from '@/lib/cn'
import type { EditorLayout } from '@/stores/uiStore'

export type PanelCollapseDividerProps = {
  layout: EditorLayout
  resultsOpen: boolean
  onToggle: () => void
}

export function PanelCollapseDivider({
  layout,
  resultsOpen,
  onToggle,
}: PanelCollapseDividerProps) {
  const isStacked = layout === 'stacked'

  return (
    <div
      className={cn(
        'relative z-10 shrink-0 bg-border',
        isStacked
          ? 'flex h-px w-full items-center justify-center'
          : 'flex w-3 self-stretch items-center justify-center',
      )}
      data-testid="request-results-divider"
    >
      <button
        type="button"
        className={cn(
          'absolute flex h-8 w-6 items-center justify-center rounded-ui border-[0.5px] border-border bg-surface text-text-muted shadow-sm hover:text-text',
          isStacked
            ? 'left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2'
            : 'left-0 top-1/2 -translate-x-[calc(100%-2px)] -translate-y-1/2',
        )}
        aria-label={resultsOpen ? 'Collapse results panel' : 'Expand results panel'}
        data-testid="request-results-toggle"
        onClick={onToggle}
      >
        {isStacked ? (
          resultsOpen ? (
            <IconChevronDown size={13} stroke={2} />
          ) : (
            <IconChevronUp size={13} stroke={2} />
          )
        ) : resultsOpen ? (
          <IconChevronLeft size={13} stroke={2} />
        ) : (
          <IconChevronRight size={13} stroke={2} />
        )}
      </button>
    </div>
  )
}
