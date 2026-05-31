import type { RunTiming } from '@/stores/executionStore'

export type TimingBarProps = {
  timing: RunTiming
}

const segmentStyles = [
  { key: 'dns' as const, label: 'DNS', light: '#B5D4F4', dark: '#0C447C' },
  { key: 'tcp' as const, label: 'TCP', light: '#85B7EB', dark: '#185FA5' },
  { key: 'tls' as const, label: 'TLS', light: '#378ADD', dark: '#378ADD' },
  { key: 'ttfb' as const, label: 'TTFB', light: '#185FA5', dark: '#85B7EB' },
  {
    key: 'transfer' as const,
    label: 'Transfer',
    light: '#0C447C',
    dark: '#B5D4F4',
  },
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
              backgroundColor: segment.light,
            }}
          />
        ))}
      </div>
      <div className="mt-1 flex flex-wrap gap-2 text-[10px] text-text-muted">
        {segmentStyles.map((segment) => (
          <span key={segment.key} className="inline-flex items-center gap-1">
            <span
              className="inline-block size-[7px] rounded-[2px]"
              style={{ backgroundColor: segment.light }}
            />
            {segment.label} {timing[segment.key]}ms
          </span>
        ))}
      </div>
      <style>{`
        [data-theme='dark'] .timing-segment[data-segment='dns'] { background: #B5D4F4 !important; }
        [data-theme='dark'] .timing-segment[data-segment='tcp'] { background: #85B7EB !important; }
        [data-theme='dark'] .timing-segment[data-segment='tls'] { background: #378ADD !important; }
        [data-theme='dark'] .timing-segment[data-segment='ttfb'] { background: #185FA5 !important; }
        [data-theme='dark'] .timing-segment[data-segment='transfer'] { background: #0C447C !important; }
      `}</style>
      <span className="sr-only">Total duration {total} ms</span>
    </div>
  )
}
