import { cn } from '@/lib/cn'
import type { StepDetailTab } from './types'

export interface StepDetailTabsProps {
  activeTab: StepDetailTab
  onTabChange: (tab: StepDetailTab) => void
  tabCounts: { assertions: number; extracts: number }
  hasResult: boolean
  resultPassed?: boolean
  failedAssertionCount?: number
}

const BASE_TABS: { id: StepDetailTab; label: string; script?: boolean }[] = [
  { id: 'request', label: 'Request' },
  { id: 'dependencies', label: 'Dependencies' },
  { id: 'pre-script', label: 'Pre-script', script: true },
  { id: 'post-script', label: 'Post-script', script: true },
  { id: 'assertions', label: 'Assertions' },
  { id: 'extracts', label: 'Extracts' },
]

function resultTabLabel(
  hasResult: boolean,
  resultPassed?: boolean,
  failedAssertionCount?: number,
): string {
  if (!hasResult) return 'Result'
  if (resultPassed) return 'Result ✓'
  const failed = failedAssertionCount ?? 0
  return failed > 0 ? `Result — ${failed} failed` : 'Result'
}

export function StepDetailTabs({
  activeTab,
  onTabChange,
  tabCounts,
  hasResult,
  resultPassed,
  failedAssertionCount,
}: StepDetailTabsProps) {
  const tabs = hasResult
    ? [...BASE_TABS, { id: 'result' as const, label: resultTabLabel(true, resultPassed, failedAssertionCount) }]
    : BASE_TABS

  return (
    <nav className="flex shrink-0 overflow-x-auto border-b border-[0.5px] border-border">
      {tabs.map((tab) => {
        const isActive = tab.id === activeTab
        const isScript = tab.script === true
        const isResult = tab.id === 'result'
        const count =
          tab.id === 'assertions'
            ? tabCounts.assertions
            : tab.id === 'extracts'
              ? tabCounts.extracts
              : 0

        return (
          <button
            key={tab.id}
            type="button"
            onClick={() => onTabChange(tab.id)}
            className={cn(
              'shrink-0 border-b-2 px-3 py-2 text-ui-xs font-medium',
              isActive && !isScript && !isResult && 'border-primary text-primary',
              !isActive && !isScript && !isResult && 'border-transparent text-text-muted hover:text-text',
              isScript &&
                isActive &&
                'border-info-fg bg-info-bg text-info-fg',
              isScript &&
                !isActive &&
                'border-transparent text-text-muted hover:text-text',
              isResult &&
                isActive &&
                resultPassed &&
                'border-success-fg text-success-fg',
              isResult &&
                isActive &&
                !resultPassed &&
                'border-danger-fg text-danger-fg',
              isResult &&
                !isActive &&
                'border-transparent text-text-muted hover:text-text',
            )}
          >
            {tab.label}
            {count > 0 && (tab.id === 'assertions' || tab.id === 'extracts') ? (
              <span className="ml-1 rounded-full bg-surface-elevated px-1.5 text-[10px]">
                {count}
              </span>
            ) : null}
          </button>
        )
      })}
    </nav>
  )
}
