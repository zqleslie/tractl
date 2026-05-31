import { resultTabs } from '@/components/request-editor/constants'
import type { ResultTab } from '@/components/request-editor/types'
import { cn } from '@/lib/cn'

export type ResponseResultTabsProps = {
  activeTab: ResultTab
  onTabChange: (tab: ResultTab) => void
}

export function ResponseResultTabs({
  activeTab,
  onTabChange,
}: ResponseResultTabsProps) {
  return (
    <div
      className="mb-3 flex border-b-[0.5px] border-border"
      role="tablist"
      aria-label="Response tabs"
    >
      {resultTabs.map((tab) => (
        <button
          key={tab.id}
          type="button"
          role="tab"
          aria-selected={activeTab === tab.id}
          data-testid={`request-result-tab-${tab.id}`}
          className={cn(
            'border-b-2 border-transparent px-2.5 py-1.5 text-ui-xs font-medium text-text-muted',
            activeTab === tab.id && 'border-text text-text',
          )}
          onClick={() => onTabChange(tab.id)}
        >
          {tab.label}
        </button>
      ))}
    </div>
  )
}
