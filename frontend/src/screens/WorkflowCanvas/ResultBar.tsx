import type { KeyboardEvent, MouseEvent } from 'react'
import {
  IconCheck,
  IconChevronDown,
  IconChevronUp,
  IconLoader2,
  IconMinus,
  IconX,
} from '@tabler/icons-react'
import { cn } from '@/lib/cn'
import type { RunState, StepOutcome, WorkflowRunSummary, WorkflowStep } from '@/types/workflow'

export interface ResultBarProps {
  isExpanded: boolean
  onToggle: () => void
  runState: RunState
  summary: WorkflowRunSummary | null
  steps: WorkflowStep[]
  onStepChipClick: (stepId: string) => void
  onViewFullResults: () => void
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

function OutcomeIcon({ outcome }: { outcome: StepOutcome }) {
  if (outcome === 'passed') return <IconCheck size={13} stroke={1.5} />
  if (outcome === 'failed') return <IconX size={13} stroke={1.5} />
  if (outcome === 'running') return <IconLoader2 size={13} stroke={1.5} className="animate-spin" />
  return <IconMinus size={13} stroke={1.5} />
}

const chipToneClass: Record<StepOutcome, string> = {
  passed: 'bg-success-bg text-success-fg',
  failed: 'bg-danger-bg text-danger-fg',
  running: 'bg-warning-bg text-warning-fg',
  skipped: 'bg-slate-100 text-slate-500 dark:bg-slate-800 dark:text-slate-400',
  idle: 'bg-slate-100 text-slate-500 dark:bg-slate-800 dark:text-slate-400',
}

const badgeBase = 'inline-flex h-5 items-center rounded-full px-2 text-[11px] font-[500] leading-none'

function OutcomeBadge({
  runState,
  summary,
  steps,
}: {
  runState: RunState
  summary: WorkflowRunSummary | null
  steps: WorkflowStep[]
}) {
  if (runState === 'idle') {
    return <span className={cn(badgeBase, chipToneClass.idle)}>Not run</span>
  }
  if (runState === 'running') {
    return (
      <span className={cn(badgeBase, 'bg-warning-bg text-warning-fg')}>
        Running…
      </span>
    )
  }
  if (runState === 'error' || summary?.outcome === 'error') {
    return <span className={cn(badgeBase, chipToneClass.failed)}>Run failed</span>
  }
  const total = steps.length
  const passed = steps.filter((s) => s.result?.outcome === 'passed').length
  const failed = steps.filter((s) => s.result?.outcome === 'failed').length
  if (failed > 0) {
    return (
      <span className={cn(badgeBase, chipToneClass.failed)}>
        {failed}/{total} failed
      </span>
    )
  }
  return (
    <span className={cn(badgeBase, chipToneClass.passed)}>
      {passed}/{total} passed
    </span>
  )
}

export function ResultBar({
  isExpanded,
  onToggle,
  runState,
  summary,
  steps,
  onStepChipClick,
  onViewFullResults,
}: ResultBarProps) {
  const isPostRun = runState === 'complete' || runState === 'error'

  const handleHeaderKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault()
      onToggle()
    }
  }

  const handleViewFull = (event: MouseEvent<HTMLButtonElement>) => {
    event.stopPropagation()
    onViewFullResults()
  }

  return (
    <div className="shrink-0 border-t border-[0.5px] border-border bg-surface">
      <div
        role="button"
        tabIndex={0}
        onClick={onToggle}
        onKeyDown={handleHeaderKeyDown}
        className="flex h-8 cursor-pointer items-center gap-3 px-3"
      >
        <span className="text-[12px] font-[500] text-text">Results</span>
        <OutcomeBadge runState={runState} summary={summary} steps={steps} />
        {summary?.outcome === 'error' && summary.errorMessage ? (
          <span className="min-w-0 truncate text-[11px] font-[400] text-danger-fg">
            {summary.errorMessage}
          </span>
        ) : null}
        {summary && isPostRun && summary.outcome !== 'error' && (
          <span className="text-[11px] font-[400] text-text-muted">
            {formatDuration(summary.totalDurationMs)}
          </span>
        )}
        <span className="flex-1" />
        {isPostRun && (
          <button
            type="button"
            onClick={handleViewFull}
            className="text-[11px] font-[400] text-primary hover:underline"
          >
            View full results
          </button>
        )}
        {isExpanded ? (
          <IconChevronDown size={15} stroke={1.5} className="text-text-muted" />
        ) : (
          <IconChevronUp size={15} stroke={1.5} className="text-text-muted" />
        )}
      </div>

      {isExpanded && summary?.outcome === 'error' && summary.errorMessage && (
        <div className="mx-3 mb-2 rounded-card border border-[0.5px] border-danger-fg bg-danger-bg px-3 py-2 text-[11px] font-[400] text-danger-fg">
          {summary.errorMessage}
        </div>
      )}

      {isExpanded && (
        <div className="flex h-12 items-center gap-2 overflow-x-auto px-3 pb-2">
          {steps.map((step) => {
            const outcome: StepOutcome = step.result?.outcome ?? 'idle'
            const result = step.result
            const totalA = result ? result.assertions.length : step.assertionCount
            const passedA = result ? result.assertions.filter((a) => a.passed).length : 0
            return (
              <button
                key={step.id}
                type="button"
                onClick={() => onStepChipClick(step.id)}
                className={cn(
                  'inline-flex h-7 shrink-0 items-center gap-1.5 rounded-full px-3 text-[11px] font-[400]',
                  chipToneClass[outcome],
                )}
              >
                <OutcomeIcon outcome={outcome} />
                <span className="font-[500]">{step.id}</span>
                <span className="opacity-80">
                  {passedA}/{totalA} assert
                </span>
                {result?.durationMs !== undefined && (
                  <span className="opacity-80">{formatDuration(result.durationMs)}</span>
                )}
              </button>
            )
          })}
        </div>
      )}
    </div>
  )
}
