import { cn } from '@/lib/cn'

export type PanelTab = {
  id: string
  label: string
  count?: number
  variant?: 'default' | 'script'
}

export type PanelTabBarProps = {
  tabs: PanelTab[]
  activeTab: string
  onTabChange: (id: string) => void
  testIdPrefix?: string
}

export function PanelTabBar({
  tabs,
  activeTab,
  onTabChange,
  testIdPrefix = 'panel-tab',
}: PanelTabBarProps) {
  return (
    <div
      className="flex shrink-0 overflow-x-auto border-b-[0.5px] border-border px-[var(--density-padding-xs)]"
      role="tablist"
    >
      {tabs.map((tab) => {
        const isActive = activeTab === tab.id
        const isScript = tab.variant === 'script'

        return (
          <button
            key={tab.id}
            type="button"
            role="tab"
            aria-selected={isActive}
            data-testid={`${testIdPrefix}-${tab.id}`}
            className={cn(
              'shrink-0 border-b-2 border-transparent font-medium text-text-muted',
              'px-[var(--density-padding-sm)] py-[var(--density-padding-xs)]',
              'text-[length:var(--density-font-label)] leading-[var(--density-line-height)]',
              'min-h-[var(--density-row-height)]',
              'hover:text-text',
              isActive &&
                !isScript &&
                '-mb-[0.5px] border-primary bg-primary-bg text-text',
              isActive &&
                isScript &&
                '-mb-[0.5px] border-[#185FA5] bg-[#185FA5]/10 text-[#185FA5]',
            )}
            onClick={() => onTabChange(tab.id)}
          >
            {tab.label}
            {tab.count != null && tab.count > 0 ? (
              <span className="ml-[3px] rounded-[3px] bg-surface-elevated px-1 text-[9px] text-text-muted">
                {tab.count}
              </span>
            ) : null}
          </button>
        )
      })}
    </div>
  )
}
