import { useMemo, useState } from 'react'
import { Badge } from '@/components/primitives'
import { EmptyState } from '@/components/common/EmptyState'
import { KeyValueTable } from '@/components/common/KeyValueTable'
import { AssertionResult } from '@/components/request-editor/results/AssertionResult'
import { ExtractResult } from '@/components/request-editor/results/ExtractResult'
import type { ResultTab } from '@/components/request-editor/types'
import { cn } from '@/lib/cn'
import type { RequestRunResult } from '@/platform/localApi/types'
import type { RunResult, RunState } from '@/stores/executionStore'
import { useUiStore } from '@/stores/uiStore'

export type ResponseTabContentProps = {
  tab: ResultTab
  runState: RunState
  runResult: RequestRunResult | null
  executionResult: RunResult | null
}

type BodyViewMode = 'pretty' | 'raw'

function prettifyJsonBody(body: string): string | null {
  if (!body.trim()) return ''
  try {
    return JSON.stringify(JSON.parse(body), null, 2)
  } catch {
    return null
  }
}

export function ResponseTabContent({
  tab,
  runState,
  runResult,
  executionResult,
}: ResponseTabContentProps) {
  const wordWrap = useUiStore((state) => state.wordWrap)
  const [bodyViewMode, setBodyViewMode] = useState<BodyViewMode>('pretty')
  const prettyBody = useMemo(
    () => (runResult ? prettifyJsonBody(runResult.body) : null),
    [runResult],
  )
  const displayedBody =
    runResult && prettyBody !== null && bodyViewMode === 'pretty'
      ? prettyBody
      : runResult?.body

  if (runState === 'idle' || !runResult) {
    return (
      <div className="flex min-h-[200px] items-center justify-center" data-testid="request-response-idle">
        <EmptyState
          icon="player-play"
          text="Hit Send to execute"
          shortcut="⌘↵"
        />
      </div>
    )
  }

  if (tab === 'headers') {
    return <KeyValueTable rows={runResult.headers} readOnly addLabel="Add header" />
  }

  if (tab === 'assertions') {
    return (
      <AssertionResult results={executionResult?.assertionResults ?? []} />
    )
  }

  if (tab === 'extracts') {
    return <ExtractResult results={executionResult?.extractResults ?? []} />
  }

  return (
    <div data-testid="request-response-body">
      <div className="mb-2 flex flex-wrap items-center gap-2">
        <Badge
          tone={
            runResult.statusCode >= 200 && runResult.statusCode < 300
              ? 'success'
              : 'danger'
          }
        >
          {runResult.statusLabel || String(runResult.statusCode)}
        </Badge>
        <Badge tone="neutral">{runResult.durationMs} ms</Badge>
        <Badge tone="neutral">{runResult.contentType}</Badge>
        {prettyBody !== null ? (
          <div className="ml-auto inline-flex rounded-ui border-[0.5px] border-border bg-surface-elevated p-0.5">
            <BodyModeButton
              active={bodyViewMode === 'pretty'}
              onClick={() => setBodyViewMode('pretty')}
            >
              Pretty
            </BodyModeButton>
            <BodyModeButton
              active={bodyViewMode === 'raw'}
              onClick={() => setBodyViewMode('raw')}
            >
              Raw
            </BodyModeButton>
          </div>
        ) : null}
      </div>
      <pre
        className={cn(
          'overflow-auto rounded-ui border-[0.5px] border-border bg-surface-elevated p-[var(--density-padding-md)] font-mono text-[length:var(--density-font-mono)] leading-[var(--density-line-height)] text-text',
          wordWrap && 'whitespace-pre-wrap break-words',
        )}
      >
        {displayedBody || '(empty body)'}
      </pre>
    </div>
  )
}

function BodyModeButton({
  active,
  children,
  onClick,
}: {
  active: boolean
  children: string
  onClick: () => void
}) {
  return (
    <button
      type="button"
      className={cn(
        'rounded px-2 py-0.5 text-ui-xs text-text-muted hover:text-text',
        active && 'bg-surface text-text',
      )}
      aria-pressed={active}
      onClick={onClick}
    >
      {children}
    </button>
  )
}
