type TimelineSegment = { label: string; ms: number }

export type RequestTimelineProps = {
  segments: TimelineSegment[]
  totalMs: number
}

export function RequestTimeline({ segments, totalMs }: RequestTimelineProps) {
  if (segments.length === 0) return null

  const totalSegmentMs = segments.reduce((sum, segment) => sum + segment.ms, 0)
  const basisTotal = totalSegmentMs > 0 ? totalSegmentMs : totalMs

  return (
    <div className="shrink-0 border-b-[0.5px] border-border px-3 py-2">
      <div className="mb-1 flex justify-between text-[10px] text-text-muted">
        <span>Request timeline</span>
        <span>{totalMs} ms total</span>
      </div>
      <div className="flex h-1.5 overflow-hidden rounded-full">
        {segments.map((segment) => {
          const width =
            basisTotal > 0
              ? `${Math.max((segment.ms / basisTotal) * 100, 4)}%`
              : '20%'
          return (
            <span
              key={segment.label}
              className="bg-primary"
              style={{ width, opacity: 0.35 + (segment.ms / (basisTotal || 1)) * 0.65 }}
            />
          )
        })}
      </div>
      <div className="mt-1 flex flex-wrap gap-x-3 gap-y-1 text-[10px] text-text-muted">
        {segments.map((segment) => (
          <span key={segment.label}>
            {segment.label} {segment.ms}ms
          </span>
        ))}
      </div>
    </div>
  )
}
