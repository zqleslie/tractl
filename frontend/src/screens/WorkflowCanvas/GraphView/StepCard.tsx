import type { KeyboardEvent } from 'react'
import { IconGripVertical, IconLoader2 } from '@tabler/icons-react'
import { cn } from '@/lib/cn'
import type { HttpMethod, StepOutcome, WorkflowStep } from '@/types/workflow'
import { PlusButton } from './PlusButton'

export interface StepCardProps {
  step: WorkflowStep
  stepNumber: number
  onClick: () => void
  onInsertBelow: () => void
}

const methodBadgeClass: Record<HttpMethod, string> = {
  GET: 'method-badge-get',
  POST: 'method-badge-post',
  PUT: 'method-badge-put',
  PATCH: 'method-badge-patch',
  DELETE: 'method-badge-delete',
}

const outcomeBorderClass: Record<StepOutcome, string> = {
  idle: 'border-slate-200 dark:border-slate-700',
  running: 'border-slate-200 dark:border-slate-700',
  passed: 'border-success-fg',
  failed: 'border-danger-fg',
  skipped: 'border-dashed border-slate-300 dark:border-slate-600',
}

const pillBase =
  'inline-flex h-5 items-center rounded-full px-2 text-[10px] font-[500] leading-none'

function OutcomeBadge({ outcome }: { outcome: StepOutcome }) {
  if (outcome === 'idle') return null
  if (outcome === 'running') {
    return (
      <IconLoader2
        size={14}
        stroke={1.5}
        className="animate-spin text-slate-400 dark:text-slate-500"
        aria-label="running"
      />
    )
  }
  if (outcome === 'passed') {
    return (
      <span className={cn(pillBase, 'bg-success-bg text-success-fg')}>
        passed
      </span>
    )
  }
  if (outcome === 'failed') {
    return (
      <span className={cn(pillBase, 'bg-danger-bg text-danger-fg')}>
        failed
      </span>
    )
  }
  return (
    <span className={cn(pillBase, 'bg-slate-100 text-slate-500 dark:bg-slate-800 dark:text-slate-400')}>
      skipped
    </span>
  )
}

function AssertionPill({ step }: { step: WorkflowStep }) {
  const result = step.result
  if (!result) {
    if (step.assertionCount <= 0) return null
    return (
      <span className={cn(pillBase, 'border border-[0.5px] border-slate-300 text-slate-500 dark:border-slate-600 dark:text-slate-400')}>
        Assertions: {step.assertionCount}
      </span>
    )
  }
  const total = result.assertions.length
  const passed = result.assertions.filter((a) => a.passed).length
  const failed = total - passed
  if (failed > 0) {
    return (
      <span className={cn(pillBase, 'bg-danger-bg text-danger-fg')}>
        {failed}/{total} failed
      </span>
    )
  }
  return (
    <span className={cn(pillBase, 'bg-success-bg text-success-fg')}>
      {passed}/{total} passed
    </span>
  )
}

const extractPillClass = cn(
  pillBase,
  'bg-info-bg text-info-fg',
)

function ExtractPills({ step }: { step: WorkflowStep }) {
  const extracts = step.result?.extracts ?? []
  if (extracts.length > 0) {
    const shown = extracts.slice(0, 2)
    const remaining = extracts.length - shown.length
    return (
      <>
        {shown.map((extract) => (
          <span key={extract.variable} className={extractPillClass}>
            → {extract.variable}
          </span>
        ))}
        {remaining > 0 && <span className={extractPillClass}>+{remaining} more</span>}
      </>
    )
  }
  if (step.extractCount > 0) {
    return (
      <span className={extractPillClass}>
        → {step.extractCount} extract{step.extractCount === 1 ? '' : 's'}
      </span>
    )
  }
  return null
}

export function StepCard({ step, stepNumber, onClick, onInsertBelow }: StepCardProps) {
  const outcome: StepOutcome = step.result?.outcome ?? 'idle'

  const handleKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault()
      onClick()
    }
  }

  return (
    <div className="group/card relative w-[260px]">
      <div
        role="button"
        tabIndex={0}
        onClick={onClick}
        onKeyDown={handleKeyDown}
        className={cn(
          'flex min-h-[110px] cursor-pointer flex-col gap-2 rounded-card border-[1.5px] bg-surface p-3',
          outcomeBorderClass[outcome],
        )}
      >
        <div className="flex items-center gap-2">
          <IconGripVertical
            size={14}
            stroke={1.5}
            className="shrink-0 cursor-grab text-slate-300 dark:text-slate-600"
            aria-hidden
          />
          <span className="flex h-[18px] w-[18px] shrink-0 items-center justify-center rounded-full bg-slate-100 text-[11px] font-[500] text-slate-600 dark:bg-slate-800 dark:text-slate-300">
            {stepNumber}
          </span>
          <span className="min-w-0 flex-1 truncate text-[12px] font-[400] text-text">
            {step.id}
          </span>
          <span
            className={cn(
              'inline-flex h-5 shrink-0 items-center rounded-full px-2 text-[10px] font-[500] leading-none',
              methodBadgeClass[step.method],
            )}
          >
            {step.method}
          </span>
          <span className="flex shrink-0 items-center">
            <OutcomeBadge outcome={outcome} />
          </span>
        </div>

        <div className="w-full truncate font-mono text-[11px] text-text-muted">{step.url}</div>

        <div className="flex flex-wrap items-center gap-1">
          {step.hasAuth && (
            <span className={cn(pillBase, 'border border-[0.5px] border-slate-300 text-slate-500 dark:border-slate-600 dark:text-slate-400')}>
              Auth
            </span>
          )}
          <AssertionPill step={step} />
          {step.hasPreScript && (
            <span className="text-[10px] leading-none text-info-fg" title="Has pre-script">
              ●
            </span>
          )}
          <ExtractPills step={step} />
        </div>
      </div>

      <div className="absolute -bottom-2 left-1/2 -translate-x-1/2">
        <PlusButton
          x={0}
          y={0}
          onClick={onInsertBelow}
          revealClassName="group-hover/card:opacity-100"
          title="Insert step below"
        />
      </div>
    </div>
  )
}
