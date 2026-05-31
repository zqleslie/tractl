import { IconCheck, IconMinus, IconX } from '@tabler/icons-react'
import { cn } from '@/lib/cn'
import type { StepOutcome, WorkflowRunSummary, WorkflowStep } from '@/types/workflow'

export interface FullResultsPopupProps {
  summary: WorkflowRunSummary
  steps: WorkflowStep[]
  onClose: () => void
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

function StepOutcomeIcon({ outcome }: { outcome: StepOutcome }) {
  if (outcome === 'passed') {
    return <IconCheck size={14} stroke={1.5} className="text-success-fg" />
  }
  if (outcome === 'failed') {
    return <IconX size={14} stroke={1.5} className="text-danger-fg" />
  }
  return <IconMinus size={14} stroke={1.5} className="text-text-muted" />
}

function SummaryCard({
  label,
  value,
  tone,
}: {
  label: string
  value: string
  tone?: 'passed' | 'failed'
}) {
  return (
    <div
      className={cn(
        'flex flex-col items-center justify-center gap-1 rounded-card border border-[0.5px] border-border p-3',
        tone === 'passed' && 'bg-success-bg',
        tone === 'failed' && 'bg-danger-bg',
      )}
    >
      <span className="text-[11px] font-[400] text-text-muted">{label}</span>
      <span className="text-[14px] font-[500] text-text">{value}</span>
    </div>
  )
}

export function FullResultsPopup({ summary, steps, onClose }: FullResultsPopupProps) {
  const outcomeTone =
    summary.outcome === 'passed'
      ? 'passed'
      : summary.outcome === 'error'
        ? 'failed'
        : 'failed'

  return (
    <div className="fixed inset-0 z-30 flex items-center justify-center">
      <button
        type="button"
        aria-label="Close results"
        onClick={onClose}
        className="absolute inset-0 bg-black/30"
      />

      <div className="relative z-10 flex max-h-[80vh] w-[720px] max-w-[92vw] flex-col rounded-card border border-[0.5px] border-border bg-surface">
        <header className="flex h-12 shrink-0 items-center justify-between border-b border-[0.5px] border-border px-4">
          <span className="text-[13px] font-[500] text-text">Workflow results</span>
          <button
            type="button"
            onClick={onClose}
            title="Close"
            aria-label="Close"
            className="flex h-7 w-7 items-center justify-center rounded-[7px] text-text-muted hover:text-text"
          >
            <IconX size={15} stroke={1.5} />
          </button>
        </header>

        <div className="flex-1 overflow-auto p-4">
          <div className="grid grid-cols-4 gap-2">
            <SummaryCard label="Outcome" value={summary.outcome} tone={outcomeTone} />
            <SummaryCard label="Total time" value={formatDuration(summary.totalDurationMs)} />
            <SummaryCard
              label="Steps ran"
              value={`${summary.stepsRan}/${summary.totalSteps}`}
            />
            <SummaryCard
              label="Assertions"
              value={`${summary.assertionsPassed}/${summary.assertionsTotal}`}
            />
          </div>

          {summary.outcome === 'error' && summary.errorMessage && (
            <div className="mt-4 rounded-card border border-[0.5px] border-danger-fg bg-danger-bg px-3 py-2 text-[12px] font-[400] text-danger-fg">
              {summary.errorMessage}
            </div>
          )}

          <div className="mt-4">
            <p className="mb-1 text-[11px] font-[400] text-text-muted">Waterfall</p>
            <pre className="m-0 overflow-auto rounded-card border border-[0.5px] border-border bg-surface-elevated p-3 font-mono text-[11px] leading-relaxed text-text">
              {summary.waterfallText}
            </pre>
          </div>

          <div className="mt-4 flex flex-col gap-1.5">
            {steps.map((step) => {
              const outcome: StepOutcome = step.result?.outcome ?? 'idle'
              const result = step.result
              const totalA = result ? result.assertions.length : step.assertionCount
              const passedA = result ? result.assertions.filter((a) => a.passed).length : 0
              const failedAssertions = result
                ? result.assertions.filter((a) => !a.passed)
                : []
              return (
                <div
                  key={step.id}
                  className={cn(
                    'rounded-card border border-[0.5px] border-border px-3 py-2',
                    outcome === 'failed' && 'border-l-2 border-l-danger-fg',
                  )}
                >
                  <div className="flex items-center gap-2 text-[12px] font-[400] text-text">
                    <StepOutcomeIcon outcome={outcome} />
                    <span className="min-w-0 flex-1 truncate font-[500]">{step.id}</span>
                    {result?.statusCode !== undefined && (
                      <span className="shrink-0 font-mono text-[11px] text-text-muted">
                        {result.statusCode}
                      </span>
                    )}
                    {result?.durationMs !== undefined && (
                      <span className="shrink-0 text-[11px] text-text-muted">
                        {formatDuration(result.durationMs)}
                      </span>
                    )}
                    <span className="shrink-0 text-[11px] text-text-muted">
                      {passedA}/{totalA} assertions
                    </span>
                  </div>

                  {failedAssertions.length > 0 && (
                    <div className="mt-1.5 flex flex-col gap-1 pl-6">
                      {failedAssertions.map((assertion, index) => (
                        <div key={index} className="text-[11px] font-[400] text-text-muted">
                          <span className="text-text">{assertion.description}</span>
                          {(assertion.expected !== undefined ||
                            assertion.received !== undefined) && (
                            <span>
                              {' — '}
                              expected: {assertion.expected ?? '—'} · received:{' '}
                              {assertion.received ?? '—'}
                            </span>
                          )}
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        </div>
      </div>
    </div>
  )
}
