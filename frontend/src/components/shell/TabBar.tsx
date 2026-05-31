import { IconPlus, IconX } from '@tabler/icons-react'
import { MethodBadge } from '@/components/primitives'
import { cn } from '@/lib/cn'
import { useUiStore } from '@/stores/uiStore'

export function TabBar() {
  const tabs = useUiStore((state) => state.tabs)
  const activeTabId = useUiStore((state) => state.activeTabId)
  const activateTab = useUiStore((state) => state.activateTab)
  const closeTab = useUiStore((state) => state.closeTab)
  const openNewRequest = useUiStore((state) => state.openNewRequest)

  return (
    <div
      className="flex h-9 shrink-0 items-end gap-0.5 border-b-[0.5px] border-border bg-surface-elevated px-3 pt-1"
      role="tablist"
      aria-label="Open workspace tabs"
      data-testid="tab-bar"
    >
      {tabs.map((tab) => {
        const active = tab.id === activeTabId
        return (
          <button
            key={tab.id}
            type="button"
            role="tab"
            aria-selected={active}
            className={cn(
              'group flex h-8 max-w-[240px] items-center gap-2 rounded-t-ui border border-b-0 px-3 text-ui-xs',
              active
                ? 'border-border bg-surface text-text shadow-sm'
                : 'border-transparent text-text-muted hover:bg-surface/80 hover:text-text',
            )}
            data-testid={`workspace-tab-${tab.id}`}
            onClick={() => activateTab(tab.id)}
          >
            <MethodBadge method={tab.method} />
            <span className="min-w-0 flex-1 truncate">{tab.title}</span>
            <span
              role="button"
              tabIndex={0}
              aria-label={`Close ${tab.title}`}
              className="rounded p-0.5 opacity-60 hover:bg-surface-elevated hover:opacity-100"
              data-testid={`close-tab-${tab.id}`}
              onClick={(event) => {
                event.stopPropagation()
                closeTab(tab.id)
              }}
              onKeyDown={(event) => {
                if (event.key !== 'Enter' && event.key !== ' ') return
                event.preventDefault()
                event.stopPropagation()
                closeTab(tab.id)
              }}
            >
              <IconX size={12} stroke={1.75} />
            </span>
          </button>
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
