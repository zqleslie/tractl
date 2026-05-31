export interface BranchBarProps {
  x: number
  y: number
  width: number
}

const STROKE = 'var(--color-border)'

/** Visual fan-out connector only — parallel steps are added via the parent step card +. */
export function BranchBar({ x, y, width }: BranchBarProps) {
  const mid = width / 2
  return (
    <div className="pointer-events-none absolute" style={{ left: x, top: y, width }}>
      <svg width={width} height={8} className="block overflow-visible">
        <line x1={0} y1={4} x2={width} y2={4} stroke={STROKE} strokeWidth={0.5} />
        <circle cx={mid} cy={4} r={3} fill={STROKE} />
      </svg>
    </div>
  )
}
