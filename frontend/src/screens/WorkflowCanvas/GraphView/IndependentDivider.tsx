export interface IndependentDividerProps {
  x: number
  height: number
}

export function IndependentDivider({ x, height }: IndependentDividerProps) {
  return (
    <div className="pointer-events-none absolute" style={{ left: x, top: 0, height }}>
      <svg width={1} height={height} className="block overflow-visible">
        <line
          x1={0}
          y1={0}
          x2={0}
          y2={height}
          stroke="var(--color-border)"
          strokeWidth={0.5}
          strokeDasharray="4 4"
        />
      </svg>
      <span className="absolute left-0 top-1/2 -translate-x-1/2 -translate-y-1/2 rotate-90 whitespace-nowrap bg-surface px-1 text-[10px] font-[400] text-text-muted">
        independent
      </span>
    </div>
  )
}
