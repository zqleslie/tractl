import type { ExtractResultItem as ExtractResultModel } from '@/stores/executionStore'

export type ExtractResultProps = {
  results: ExtractResultModel[]
}

export function ExtractResult({ results }: ExtractResultProps) {
  if (results.length === 0) {
    return <p className="text-ui-xs text-text-muted">No extracts defined</p>
  }

  const primary = results[0]!

  return (
    <div className="space-y-2" data-testid="extract-results">
      {results.map((row) => (
        <div
          key={row.id}
          className="flex items-center gap-2 rounded-[7px] border-[0.5px] border-border bg-surface-elevated px-[var(--density-padding-sm)] py-[var(--density-padding-xs)]"
        >
          <span className="flex-1 font-mono text-info-fg">
            steps.this.extracts.{row.variableName}
          </span>
          <span className="font-mono text-text">{row.resolvedValue ?? '—'}</span>
        </div>
      ))}
      <p className="text-[11px] text-text-muted">
        Written to {primary.scope} scope. Available in downstream steps as{' '}
        <span className="font-mono text-info-fg">
          {'${steps.this.extracts.'}
          {primary.variableName}
          {'}'}
        </span>
      </p>
    </div>
  )
}
