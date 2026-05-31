import { computeDagLayout } from './dagLayout'
import type { WorkflowLayoutResponse } from '@/platform/types'

const CARD_W = 260
const CARD_H = 110

const MOCK_LAYOUT: WorkflowLayoutResponse = {
  rows: [
    { rowIndex: 0, stepIds: ['step-auth'] },
    { rowIndex: 1, stepIds: ['step-billing', 'step-health'] },
    { rowIndex: 2, stepIds: ['step-metrics'] },
  ],
  edges: [
    { from: 'step-auth', to: 'step-billing', kind: 'fanout', implicit: false },
    { from: 'step-auth', to: 'step-health', kind: 'fanout', implicit: false },
    { from: 'step-billing', to: 'step-metrics', kind: 'merge', implicit: false },
    { from: 'step-health', to: 'step-metrics', kind: 'merge', implicit: false },
  ],
  groups: [{ groupIndex: 0, rootStepId: 'step-auth', stepIds: ['step-auth', 'step-billing', 'step-health', 'step-metrics'] }],
  topologySummary: '1 root · 1 fan-out · 1 merge',
}

const EDGE_STROKE: Record<string, string> = {
  sequential: 'var(--color-border)',
  fanout: 'var(--color-primary)',
  merge: 'var(--color-info-fg)',
}

/**
 * Standalone visual harness for {@link computeDagLayout}. Renders the layout of a
 * mock 4-step workflow as absolutely-positioned cards with an SVG overlay drawing
 * edges, branch bars and group dividers. This component is for manual inspection
 * only and has no dependency on app state.
 */
export function DagLayoutPreview() {
  const layout = computeDagLayout(MOCK_LAYOUT)

  const centreOf = (stepId: string) => {
    const node = layout.nodes.find((n) => n.stepId === stepId)!
    return { cx: node.x + CARD_W / 2, cy: node.y + CARD_H / 2 }
  }

  return (
    <div style={{ padding: 24, background: 'var(--color-surface-elevated)', overflow: 'auto' }}>
      <h2 style={{ fontFamily: 'sans-serif', fontSize: 16, marginBottom: 16 }}>
        DAG Layout Preview
      </h2>
      <div
        style={{
          position: 'relative',
          width: layout.canvasWidth,
          height: layout.canvasHeight,
          background: 'var(--color-surface)',
          border: '1px solid var(--color-border)',
        }}
      >
        <svg
          width={layout.canvasWidth}
          height={layout.canvasHeight}
          style={{ position: 'absolute', inset: 0, pointerEvents: 'none' }}
        >
          {layout.groupDividers.map((divider) => (
            <line
              key={`divider-${divider.leftGroup}-${divider.rightGroup}`}
              x1={divider.x}
              y1={0}
              x2={divider.x}
              y2={layout.canvasHeight}
              stroke="var(--color-border)"
              strokeDasharray="6 6"
            />
          ))}

          {layout.edges.map((edge) => {
            const from = centreOf(edge.fromId)
            const to = centreOf(edge.toId)
            return (
              <line
                key={`edge-${edge.fromId}-${edge.toId}`}
                x1={from.cx}
                y1={from.cy}
                x2={to.cx}
                y2={to.cy}
                stroke={EDGE_STROKE[edge.edgeType] ?? 'var(--color-border)'}
                strokeWidth={2}
              />
            )
          })}

          {layout.branchBars.map((bar) => (
            <line
              key={`bar-${bar.parentId}`}
              x1={bar.xStart}
              y1={bar.y}
              x2={bar.xEnd}
              y2={bar.y}
              stroke="var(--color-primary)"
              strokeWidth={3}
            />
          ))}
        </svg>

        {layout.nodes.map((node) => (
          <div
            key={node.stepId}
            style={{
              position: 'absolute',
              left: node.x,
              top: node.y,
              width: CARD_W,
              height: CARD_H,
              boxSizing: 'border-box',
              border: '1px solid var(--color-border)',
              borderRadius: 8,
              background: 'var(--color-surface)',
              padding: 12,
              fontFamily: 'sans-serif',
            }}
          >
            <div style={{ fontSize: 14, fontWeight: 600 }}>{node.stepId}</div>
            <div style={{ fontSize: 12, color: 'var(--color-text-muted)', marginTop: 8 }}>
              row {node.row} · col {node.col} · group {node.columnGroup}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
