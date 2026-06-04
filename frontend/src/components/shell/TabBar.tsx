import { useState } from 'react'
import { IconPlus, IconX } from '@tabler/icons-react'
import { MethodBadge } from '@/components/primitives'
import { cn } from '@/lib/cn'
import { useUiStore, type WorkspaceTab } from '@/stores/uiStore'

type WorkspaceTabButtonProps = {
  tab: WorkspaceTab
  active: boolean
  onActivate: () => void
  onClose: () => void
}

function WorkspaceTabButton({
  tab,
  active,
  onActivate,
  onClose,
}: WorkspaceTabButtonProps) {
  const [isHovered, setIsHovered] = useState(false)

  return (
    <button
      type="button"
      role="tab"
      aria-selected={active}
      className={cn(
        'flex h-8 max-w-[240px] items-center gap-2 rounded-t-ui border border-b-0 px-3 text-ui-xs',
        active
          ? 'border-border bg-surface text-text shadow-sm'
          : 'border-transparent text-text-muted hover:bg-surface/80 hover:text-text',
      )}
      data-testid={`workspace-tab-${tab.id}`}
      onClick={onActivate}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      <MethodBadge method={tab.method} />
      <span className="min-w-0 flex-1 truncate">{tab.title}</span>
      {tab.isDirty && !isHovered ? (
        <span
          aria-label="Unsaved changes"
          style={{
            width: 7,
            height: 7,
            borderRadius: '50%',
            background: 'var(--dot-unsaved, var(--color-warning-fg))',
            flexShrink: 0,
            display: 'inline-block',
          }}
        />
      ) : (
        <span
          role="button"
          tabIndex={0}
          aria-label={`Close ${tab.title}`}
          className="rounded p-0.5 hover:bg-surface-elevated"
          data-testid={`close-tab-${tab.id}`}
          style={{ opacity: isHovered ? 1 : active ? 0.4 : 0 }}
          onClick={(event) => {
            event.stopPropagation()
            onClose()
          }}
          onKeyDown={(event) => {
            if (event.key !== 'Enter' && event.key !== ' ') return
            event.preventDefault()
            event.stopPropagation()
            onClose()
          }}
        >
          <IconX size={12} stroke={1.75} />
        </span>
      )}
    </button>
  )
}

export function TabBar() {
  const tabs = useUiStore((state) => state.tabs)
  const activeTabId = useUiStore((state) => state.activeTabId)
  const activateTab = useUiStore((state) => state.activateTab)
  const closeTab = useUiStore((state) => state.closeTab)
  const openNewRequest = useUiStore((state) => state.openNewRequest)

  return (
    <div
      className="flex h-9 shrink-0 items-end gap-0.5 border-b-[0.5px] border-border bg-surface-sidebar px-3 pt-1"
      role="tablist"
      aria-label="Open workspace tabs"
      data-testid="tab-bar"
    >
      {tabs.map((tab) => {
        const active = tab.id === activeTabId
        return (
          <WorkspaceTabButton
            key={tab.id}
            tab={tab}
            active={active}
            onActivate={() => activateTab(tab.id)}
            onClose={() => closeTab(tab.id)}
          />
        )
      })}
      <button
        type="button"
        className="mb-0.5 ml-1 flex h-7 w-8 items-center justify-center rounded-ui text-text-muted hover:bg-surface hover:text-text"
        aria-label="New tab"
        data-testid="new-tab"
        onClick={openNewRequest}
      >
        <IconPlus size={16} stroke={1.75} />
      </button>
    </div>
  )
}
