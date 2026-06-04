import {
  IconBolt,
  IconChevronLeft,
  IconChevronRight,
  IconClock,
  IconLayersIntersect,
  IconSettings,
  IconStack2,
  IconTopologyStar3,
} from '@tabler/icons-react'
import { cn } from '@/lib/cn'
import { type SidebarTab, useUiStore } from '@/stores/uiStore'

type ActivityItem = {
  id: SidebarTab
  label: string
  icon: typeof IconBolt
  bottom?: boolean
}

const items: ActivityItem[] = [
  { id: 'requests', label: 'Requests', icon: IconBolt },
  { id: 'workflows', label: 'Workflows', icon: IconTopologyStar3 },
  { id: 'overlays', label: 'Overlays', icon: IconStack2 },
  { id: 'history', label: 'History', icon: IconClock },
  { id: 'envs', label: 'Environments', icon: IconLayersIntersect },
  { id: 'settings', label: 'Settings', icon: IconSettings, bottom: true },
]

export function ActivityBar() {
  const sidebarTab = useUiStore((state) => state.sidebarTab)
  const sidebarCollapsed = useUiStore((state) => state.sidebarCollapsed)
  const expanded = useUiStore((state) => state.activityBarExpanded)
  const toggleSidebarTab = useUiStore((state) => state.toggleSidebarTab)
  const toggleExpanded = useUiStore((state) => state.toggleActivityBarExpanded)

  return (
    <nav
      className={cn(
        'flex shrink-0 flex-col border-r-[0.5px] border-border bg-surface-sidebar transition-[width] duration-200',
        expanded ? 'w-40' : 'w-12',
      )}
      aria-label="Primary"
      data-testid="activity-bar"
    >
      <div className="flex h-11 items-center justify-center">
        <span className="h-2 w-2 rounded-full bg-danger-fg" aria-hidden="true" />
      </div>
      <div className="flex w-full flex-1 flex-col gap-1 py-1">
        {items.map((item, index) => {
          const Icon = item.icon
          const active = sidebarTab === item.id && !sidebarCollapsed
          return (
            <button
              key={item.id}
              type="button"
              className={cn(
                'relative flex h-9 w-full items-center gap-2 px-3 text-text-muted hover:bg-surface-elevated hover:text-text',
                expanded ? 'justify-start' : 'justify-center',
                active && 'text-primary',
                item.bottom && 'mt-auto',
                index === 4 && 'mt-2 border-t-[0.5px] border-border pt-2',
              )}
              title={item.label}
              aria-label={item.label}
              aria-pressed={active}
              data-testid={`activity-${item.id}`}
              onClick={() => toggleSidebarTab(item.id)}
            >
              {active ? (
                <span className="absolute left-0 h-5 w-[3px] rounded-r bg-primary" />
              ) : null}
              <Icon size={18} stroke={1.75} className="shrink-0" />
              {expanded ? (
                <span className="truncate text-ui-sm font-medium">{item.label}</span>
              ) : null}
            </button>
          )
        })}
      </div>
      <button
        type="button"
        className={cn(
          'm-1 flex h-8 items-center gap-2 rounded-ui px-2 text-text-muted hover:bg-surface-elevated hover:text-text',
          expanded ? 'justify-start' : 'justify-center',
        )}
        aria-label={expanded ? 'Collapse activity bar' : 'Expand activity bar'}
        aria-expanded={expanded}
        data-testid="activity-expand"
        onClick={toggleExpanded}
      >
        {expanded ? (
          <IconChevronLeft size={16} stroke={1.75} />
        ) : (
          <IconChevronRight size={16} stroke={1.75} />
        )}
        {expanded ? <span className="text-ui-xs">Collapse</span> : null}
      </button>
    </nav>
  )
}
