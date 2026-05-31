import type { ReactNode } from 'react'
import {
  IconColumns2,
  IconKeyboard,
  IconLayoutRows,
  IconTextWrap,
} from '@tabler/icons-react'
import { cn } from '@/lib/cn'
import { type UiScale, useUiStore } from '@/stores/uiStore'

const densityOptions: Array<{ value: UiScale; label: string }> = [
  { value: 'comfortable', label: 'Comfortable' },
  { value: 'default', label: 'Default' },
  { value: 'compact', label: 'Compact' },
]

export function BottomToolbar() {
  const editorLayout = useUiStore((state) => state.editorLayout)
  const uiScale = useUiStore((state) => state.uiScale)
  const wordWrap = useUiStore((state) => state.wordWrap)
  const toggleEditorLayout = useUiStore((state) => state.toggleEditorLayout)
  const setDensity = useUiStore((state) => state.setDensity)
  const toggleWordWrap = useUiStore((state) => state.toggleWordWrap)
  const setKeyboardShortcutsOpen = useUiStore(
    (state) => state.setKeyboardShortcutsOpen,
  )

  return (
    <div
      className="flex h-[26px] shrink-0 items-center border-t-[0.5px] border-border bg-surface px-3 text-ui-xs text-text-muted"
      data-testid="bottom-toolbar"
    >
      <div className="ml-auto flex items-center gap-1">
        <ToolbarButton
          active
          label={`Layout: ${editorLayout === 'stacked' ? 'stacked' : 'side by side'}`}
          onClick={toggleEditorLayout}
        >
          {editorLayout === 'stacked' ? (
            <IconLayoutRows size={14} stroke={1.75} />
          ) : (
            <IconColumns2 size={14} stroke={1.75} />
          )}
          {editorLayout === 'stacked' ? 'Stacked' : 'Side by side'}
        </ToolbarButton>
        <label
          className="inline-flex h-5 items-center gap-1 rounded border-[0.5px] border-border bg-surface-elevated px-1.5 text-ui-xs text-text-muted"
          aria-label="Text size"
          data-testid="density-toggle"
        >
          <span
            className="inline-flex items-baseline gap-[1px] font-semibold text-text-muted"
            aria-hidden="true"
          >
            <span className="text-[13px] leading-none">T</span>
            <span className="text-[9px] leading-none">T</span>
          </span>
          <select
            className="h-full bg-transparent text-ui-xs text-text outline-none"
            value={uiScale}
            onChange={(event) => setDensity(event.currentTarget.value as UiScale)}
          >
            {densityOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <ToolbarButton
          active={wordWrap}
          label="Toggle word wrap"
          onClick={toggleWordWrap}
        >
          <IconTextWrap size={14} stroke={1.75} />
          Wrap
        </ToolbarButton>
        <span className="mx-1 h-3.5 border-l-[0.5px] border-border" />
        <ToolbarButton
          active={false}
          label="Keyboard shortcuts"
          onClick={() => setKeyboardShortcutsOpen(true)}
        >
          <IconKeyboard size={14} stroke={1.75} />
          Shortcuts
        </ToolbarButton>
      </div>
    </div>
  )
}

function ToolbarButton({
  active,
  label,
  children,
  onClick,
  testId,
}: {
  active: boolean
  label: string
  children: ReactNode
  onClick: () => void
  testId?: string
}) {
  return (
    <button
      type="button"
      aria-label={label}
      aria-pressed={active}
      data-testid={testId}
      className={cn(
        'inline-flex h-5 items-center gap-1 rounded px-1.5 text-ui-xs hover:bg-surface-elevated hover:text-text',
        active && 'bg-primary-bg text-primary',
      )}
      onClick={onClick}
    >
      {children}
    </button>
  )
}
