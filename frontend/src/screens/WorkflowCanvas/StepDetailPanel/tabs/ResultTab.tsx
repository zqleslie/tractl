import { cn } from '@/lib/cn'
import type { StepOutcome, WorkflowStep } from '@/types/workflow'

export interface ResultTabProps {
  step: WorkflowStep
}

function HttpStatusChip({ statusCode }: { statusCode?: number }) {
  const tone =
    statusCode === undefined
      ? 'border-border bg-surface-elevated text-text-muted'
      : statusCode >= 200 && statusCode < 300
        ? 'border-success bg-success-bg text-success-fg'
        : statusCode >= 400
          ? 'border-danger bg-danger-bg text-danger-fg'
          : 'border-border bg-surface-elevated text-text-muted'

  return (
    <span
      className={cn(
        'rounded-full border border-[0.5px] px-2.5 py-0.5 text-ui-xs font-medium',
        tone,
      )}
    >
      {statusCode ?? '—'}
    </span>
  )
}

function OutcomeChip({ outcome }: { outcome: StepOutcome }) {
  const tone =
    outcome === 'passed'
      ? 'border-success bg-success-bg text-success-fg'
      : outcome === 'failed'
        ? 'border-danger bg-danger-bg text-danger-fg'
        : 'border-border bg-surface-elevated text-text-muted'

  return (
    <span
      className={cn(
        'rounded-full border border-[0.5px] px-2.5 py-0.5 text-ui-xs font-medium capitalize',
        tone,
      )}
    >
      {outcome}
    </span>
  )
}

export function ResultTab({ step }: ResultTabProps) {
  if (!step.result) {
    return (
      <div className="flex h-full items-center justify-center text-ui-xs text-text-muted">
        Run the workflow to see results here.
      </div>
    )
  }

  const { outcome, statusCode, durationMs, assertions, extracts, responseBody, requestUrl } =
    step.result
  const templateUrl = step.url
  const resolvedUrl = requestUrl ?? templateUrl
  const showTemplate =
    requestUrl !== undefined && requestUrl.trim() !== '' && requestUrl !== templateUrl

  return (
    <div className="space-y-4">
      <div className="space-y-1">
        <p className="text-ui-xs font-medium text-text-muted">URL</p>
        {showTemplate ? (
          <div className="space-y-0.5 font-mono text-ui-xs">
            <p className="text-text-muted">
              <span className="font-sans text-[10px] uppercase tracking-wide">template </span>
              {templateUrl}
            </p>
            <p className="break-all text-text">
              <span className="font-sans text-[10px] uppercase tracking-wide">resolved </span>
              {requestUrl}
            </p>
          </div>
        ) : (
          <p className="break-all font-mono text-ui-xs text-text">{resolvedUrl}</p>
        )}
      </div>

      <div className="flex flex-wrap gap-2">
        <HttpStatusChip statusCode={statusCode} />
        {durationMs !== undefined ? (
          <span className="rounded-full border border-[0.5px] border-border px-2.5 py-0.5 text-ui-xs text-text-muted">
            {durationMs}ms
          </span>
        ) : null}
        <OutcomeChip outcome={outcome} />
      </div>

      {assertions.length > 0 ? (
        <div className="space-y-1">
          <p className="text-ui-xs font-medium text-text-muted">
            {assertions.filter((a) => !a.passed).length} failed ·{' '}
            {assertions.filter((a) => a.passed).length} passed
          </p>
          {assertions.map((a, i) => (
            <div
              key={`${a.description}-${i}`}
              className={cn(
                'rounded-ui border-[0.5px] p-2 text-ui-xs',
                a.passed
                  ? 'border-success bg-success-bg text-success-fg'
                  : 'border-danger bg-danger-bg text-danger-fg',
              )}
            >
              <span className="mr-1">{a.passed ? '✓' : '✗'}</span>
              {a.description}
              {!a.passed && a.received ? (
                <span className="ml-2 text-text-muted">received: {a.received}</span>
              ) : null}
            </div>
          ))}
        </div>
      ) : null}

      {extracts.length > 0 ? (
        <div className="space-y-1">
          <p className="text-ui-xs font-medium text-text-muted">Extracts</p>
          {extracts.map((e, i) => (
            <div key={`${e.variable}-${i}`} className="flex items-center gap-2 text-ui-xs font-mono">
              <span className="text-text-muted">{e.variable}</span>
              <span>→</span>
              <span className="text-text">{e.value ?? '(empty)'}</span>
            </div>
          ))}
        </div>
      ) : null}

      {responseBody ? (
        <div>
          <p className="mb-1 text-ui-xs font-medium text-text-muted">Response body</p>
          <pre className="overflow-auto rounded-ui border-[0.5px] border-border bg-surface-elevated p-3 font-mono text-ui-xs text-text">
            {responseBody}
          </pre>
        </div>
      ) : null}
    </div>
  )
}
