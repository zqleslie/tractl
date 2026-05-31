import { useState } from 'react'
import { IconCheck, IconChevronRight, IconX } from '@tabler/icons-react'
import type { AssertionResult as AssertionResultModel } from '@/stores/executionStore'

export type AssertionResultProps = {
  results: AssertionResultModel[]
}

export function AssertionResult({ results }: AssertionResultProps) {
  const failed = results.filter((row) => !row.passed)
  const passed = results.filter((row) => row.passed)
  const [passedOpen, setPassedOpen] = useState(false)

  if (results.length === 0) {
    return <p className="text-ui-xs text-text-muted">No assertion results.</p>
  }

  const showCollapsedPassed =
    failed.length > 0 && passed.length > 0 && results.length > 3

  const summary =
    failed.length > 0
      ? `${failed.length} failed · ${passed.length} passed`
      : `All ${passed.length} passed`

  return (
    <div className="space-y-2" data-testid="assertion-results">
      <p className="text-[11px] text-text-muted">{summary}</p>

      {failed.map((row) => (
        <AssertionResultRow key={row.id} row={row} />
      ))}

      {showCollapsedPassed ? (
        <>
          <button
            type="button"
            className="text-[11px] text-text-muted hover:text-text"
            onClick={() => setPassedOpen((open) => !open)}
          >
            {passed.length} passed {passedOpen ? '▼' : '▶'}
          </button>
          {passedOpen
            ? passed.map((row) => <AssertionResultRow key={row.id} row={row} />)
            : null}
        </>
      ) : (
        passed.map((row) => <AssertionResultRow key={row.id} row={row} />)
      )}
    </div>
  )
}

function AssertionResultRow({ row }: { row: AssertionResultModel }) {
  const description = `${row.kind} ${row.op} ${row.expected}`

  return (
    <div
      className={[
        'flex items-start gap-[7px] rounded-[7px]',
        'px-[var(--density-padding-xs)] py-[var(--density-padding-sm)]',
        row.passed ? 'bg-success-bg text-success-fg' : 'bg-danger-bg text-danger-fg',
      ].join(' ')}
    >
      {row.passed ? (
        <IconCheck size={14} stroke={2} className="mt-0.5 shrink-0" />
      ) : (
        <IconX size={14} stroke={2} className="mt-0.5 shrink-0" />
      )}
      <div className="min-w-0 flex-1">
        <p className="text-[length:var(--density-font-ui)] text-text">{description}</p>
        <p className="text-[10px] text-text-muted">
          received {row.received} · {row.severity} severity
        </p>
        {!row.passed ? (
          <p className="mt-[3px] rounded-[4px] bg-danger-bg px-[5px] py-[2px] font-mono text-[10px] text-danger-fg">
            expected: {row.expected} · received: {row.received}
          </p>
        ) : null}
      </div>
      {!row.passed ? (
        <IconChevronRight size={12} className="invisible" aria-hidden />
      ) : null}
    </div>
  )
}
