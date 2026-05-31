import { computeDagLayout } from './dagLayout';
import type { LayoutStep } from './dagLayout';

const CARD_W = 260;
const CARD_H = 110;

const MOCK_WORKFLOW: LayoutStep[] = [
  { id: 'step-auth', dependsOn: [] },
  { id: 'step-billing', dependsOn: ['step-auth'] },
  { id: 'step-health', dependsOn: ['step-auth'] },
  { id: 'step-metrics', dependsOn: ['step-billing', 'step-health'] },
];

const EDGE_STROKE: Record<string, string> = {
  sequential: '#64748b',
  fanout: '#2563eb',
  merge: '#9333ea',
};

/**
 * Standalone visual harness for {@link computeDagLayout}. Renders the layout of a
 * mock 4-step workflow as absolutely-positioned cards with an SVG overlay drawing
 * edges, branch bars and group dividers. This component is for manual inspection
 * only and has no dependency on app state.
 */
export function DagLayoutPreview() {
  const layout = computeDagLayout(MOCK_WORKFLOW);

  const centreOf = (stepId: string) => {
    const node = layout.nodes.find((n) => n.stepId === stepId)!;
    return { cx: node.x + CARD_W / 2, cy: node.y + CARD_H / 2 };
  };

  return (
    <div style={{ padding: 24, background: '#f8fafc', overflow: 'auto' }}>
      <h2 style={{ fontFamily: 'sans-serif', fontSize: 16, marginBottom: 16 }}>
        DAG Layout Preview
      </h2>
      <div
        style={{
          position: 'relative',
          width: layout.canvasWidth,
          height: layout.canvasHeight,
          background: '#ffffff',
          border: '1px solid #e2e8f0',
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
              stroke="#cbd5e1"
              strokeDasharray="6 6"
            />
          ))}

          {layout.edges.map((edge) => {
            const from = centreOf(edge.fromId);
            const to = centreOf(edge.toId);
            return (
              <line
                key={`edge-${edge.fromId}-${edge.toId}`}
                x1={from.cx}
                y1={from.cy}
                x2={to.cx}
                y2={to.cy}
                stroke={EDGE_STROKE[edge.edgeType] ?? '#64748b'}
                strokeWidth={2}
              />
            );
          })}

          {layout.branchBars.map((bar) => (
            <line
              key={`bar-${bar.parentId}`}
              x1={bar.xStart}
              y1={bar.y}
              x2={bar.xEnd}
              y2={bar.y}
              stroke="#2563eb"
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
              border: '1px solid #cbd5e1',
              borderRadius: 8,
              background: '#ffffff',
              padding: 12,
              fontFamily: 'sans-serif',
            }}
          >
            <div style={{ fontSize: 14, fontWeight: 600 }}>{node.stepId}</div>
            <div style={{ fontSize: 12, color: '#64748b', marginTop: 8 }}>
              row {node.row} · col {node.col} · group {node.columnGroup}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
