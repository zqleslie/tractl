import { ResponseResultTabs } from '@/components/request-editor/results/ResponseResultTabs'
import { ResponseTabContent } from '@/components/request-editor/results/ResponseTabContent'
import { TimingBar } from '@/components/request-editor/results/TimingBar'
import type { ResultTab } from '@/components/request-editor/types'
import { cn } from '@/lib/cn'
import type { RunResult as PlatformRunResult } from '@/platform/types'
import type { RunResult, RunState } from '@/stores/executionStore'
import type { EditorLayout } from '@/stores/uiStore'

export type RequestResultsPanelProps = {
  layout: EditorLayout
  activeTab: ResultTab
  runState: RunState
  runResult: PlatformRunResult | null
  executionResult: RunResult | null
  isRunning: boolean
  onTabChange: (tab: ResultTab) => void
}

function outcomeBadge(runState: RunState, result: RunResult | null, isRunning: boolean) {
  if (isRunning || runState === 'running') {
    return { label: 'Sending…', className: 'bg-warning-bg text-warning-fg' }
  }
  if (runState === 'idle' || !result) {
    return { label: 'Not run', className: 'bg-surface-elevated text-text-muted' }
  }
  if (
    result.assertionsTotal > 0 &&
    result.assertionsPassed < result.assertionsTotal
  ) {
    return {
      label: `${result.assertionsPassed}/${result.assertionsTotal} failed`,
      className: 'bg-danger-bg text-danger-fg',
    }
  }
  return {
    label: `${result.assertionsPassed}/${result.assertionsTotal} passed`,
    className: 'bg-success-bg text-success-fg',
  }
}

export function RequestResultsPanel({
  layout: _layout,
  activeTab,
  runState,
  runResult,
  executionResult,
  isRunning,
  onTabChange,
}: RequestResultsPanelProps) {
  const badge = outcomeBadge(runState, executionResult, isRunning)

  return (
    <section
      className="flex min-w-0 flex-1 flex-col bg-surface"
      aria-label="Request response"
      data-testid="results-panel"
    >
      <div className="flex shrink-0 items-center gap-2 border-b-[0.5px] border-border px-[var(--density-padding-md)] py-[var(--density-padding-sm)]">
        <h2 className="text-[length:var(--density-font-ui)] font-medium text-text-muted">
          Response
        </h2>
        <span
          className={cn(
            'rounded-ui px-2 py-0.5 text-[length:var(--density-font-label)]',
            badge.className,
          )}
          data-testid="request-outcome-badge"
        >
          {badge.label}
        </span>
        <span className="ml-auto text-[11px] text-text-muted">
          {executionResult ? `${executionResult.durationMs} ms` : '—'}
        </span>
      </div>

      {executionResult ? <TimingBar timing={executionResult.timing} /> : null}

      <div className="min-h-0 flex-1 overflow-auto p-[var(--density-padding-md)]">
        <ResponseResultTabs activeTab={activeTab} onTabChange={onTabChange} />
        <ResponseTabContent
          tab={activeTab}
          runState={runState}
          runResult={runResult}
          executionResult={executionResult}
        />
      </div>
    </section>
  )
}
