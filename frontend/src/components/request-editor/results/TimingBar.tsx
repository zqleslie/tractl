import type { RunTiming } from '@/stores/executionStore'

export type TimingBarProps = {
  timing: RunTiming
}

const segmentStyles = [
  { key: 'dns' as const, label: 'DNS', color: 'var(--color-info-bg)' },
  { key: 'tcp' as const, label: 'TCP', color: 'var(--color-primary-bg)' },
  { key: 'tls' as const, label: 'TLS', color: 'var(--color-primary)' },
  { key: 'ttfb' as const, label: 'TTFB', color: 'var(--color-info-fg)' },
  { key: 'transfer' as const, label: 'Transfer', color: 'var(--color-primary)' },
]

export function TimingBar({ timing }: TimingBarProps) {
  const total = timing.total || 1

  return (
    <div
      className="shrink-0 border-b-[0.5px] border-border px-[var(--density-padding-md)] py-[var(--density-padding-sm)]"
      data-testid="timing-bar"
    >
      <div className="mb-1 flex justify-between text-[10px] text-text-muted">
        <span>Request timeline</span>
        <span>{timing.total} ms</span>
      </div>
      <div className="flex h-[5px] gap-px overflow-hidden rounded-[3px]">
        {segmentStyles.map((segment) => (
          <span
            key={segment.key}
            className="timing-segment"
            data-segment={segment.key}
            style={{
              flex: Math.max(timing[segment.key], 1),
              backgroundColor: segment.color,
            }}
          />
        ))}
      </div>
      <div className="mt-1 flex flex-wrap gap-2 text-[10px] text-text-muted">
        {segmentStyles.map((segment) => (
          <span key={segment.key} className="inline-flex items-center gap-1">
            <span
              className="inline-block size-[7px] rounded-[2px]"
              style={{ backgroundColor: segment.color }}
            />
            {segment.label} {timing[segment.key]}ms
          </span>
        ))}
      </div>
      <span className="sr-only">Total duration {total} ms</span>
    </div>
  )
}
