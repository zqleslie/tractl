export interface ConnectionPoint {
  x: number
  y: number
}

export interface ConnectionLineProps {
  from: ConnectionPoint
  to: ConnectionPoint
  edgeType: 'sequential' | 'fanout' | 'merge'
  dependencyKind?: 'explicit' | 'implicit'
}

const STROKE = 'var(--color-border)'
const STROKE_WIDTH = 0.5

export function ConnectionLine({
  from,
  to,
  edgeType,
  dependencyKind = 'explicit',
}: ConnectionLineProps) {
  const strokeProps = {
    stroke: STROKE,
    strokeWidth: STROKE_WIDTH,
    strokeDasharray: dependencyKind === 'implicit' ? '6 4' : undefined,
  }

  if (edgeType === 'sequential') {
    return (
      <line
        x1={from.x}
        y1={from.y}
        x2={to.x}
        y2={to.y}
        {...strokeProps}
      />
    )
  }

  const midY = (from.y + to.y) / 2
  const d = `M ${from.x} ${from.y} C ${from.x} ${midY} ${to.x} ${midY} ${to.x} ${to.y}`
  return <path d={d} fill="none" {...strokeProps} />
}
