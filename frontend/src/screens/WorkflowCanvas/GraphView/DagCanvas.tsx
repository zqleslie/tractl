import { useMemo } from 'react'
import type { WorkflowStep } from '@/types/workflow'
import type { DagLayout, DagNode } from './dagLayout'
import { StepCard } from './StepCard'
import { ConnectionLine } from './ConnectionLine'
import { BranchBar } from './BranchBar'
import { IndependentDivider } from './IndependentDivider'
import { PlusButton } from './PlusButton'

const CARD_W = 260
const CARD_H = 110

export interface DagCanvasProps {
  layout: DagLayout
  steps: WorkflowStep[]
  onStepClick: (stepId: string) => void
  onInsertAfter: (stepId: string) => void
  onInsertBetween: (fromStepId: string, toStepId: string) => void
}

export function DagCanvas({
  layout,
  steps,
  onStepClick,
  onInsertAfter,
  onInsertBetween,
}: DagCanvasProps) {
  const stepById = useMemo(() => {
    const map = new Map<string, WorkflowStep>()
    for (const step of steps) map.set(step.id, step)
    return map
  }, [steps])

  const stepNumberById = useMemo(() => {
    const map = new Map<string, number>()
    steps.forEach((step, index) => map.set(step.id, index + 1))
    return map
  }, [steps])

  const nodeById = useMemo(() => {
    const map = new Map<string, DagNode>()
    for (const node of layout.nodes) map.set(node.stepId, node)
    return map
  }, [layout.nodes])

  return (
    <div className="h-full w-full overflow-auto bg-surface">
      <div
        className="relative"
        style={{ width: layout.canvasWidth, height: layout.canvasHeight }}
      >
        {layout.groupDividers.map((divider) => (
          <IndependentDivider
            key={`divider-${divider.leftGroup}-${divider.rightGroup}`}
            x={divider.x}
            height={layout.canvasHeight}
          />
        ))}

        <svg
          width={layout.canvasWidth}
          height={layout.canvasHeight}
          className="pointer-events-none absolute inset-0 overflow-visible"
        >
          {layout.edges.map((edge) => {
            const from = nodeById.get(edge.fromId)
            const to = nodeById.get(edge.toId)
            if (!from || !to) return null
            return (
              <ConnectionLine
                key={`edge-${edge.fromId}-${edge.toId}-${edge.dependencyKind}`}
                from={{ x: from.x + CARD_W / 2, y: from.y + CARD_H }}
                to={{ x: to.x + CARD_W / 2, y: to.y }}
                edgeType={edge.edgeType}
                dependencyKind={edge.dependencyKind}
              />
            )
          })}
        </svg>

        {layout.nodes.map((node) => {
          const step = stepById.get(node.stepId)
          if (!step) return null
          return (
            <div key={node.stepId} className="absolute" style={{ left: node.x, top: node.y }}>
              <StepCard
                step={step}
                stepNumber={stepNumberById.get(node.stepId) ?? 0}
                onClick={() => onStepClick(node.stepId)}
                onInsertBelow={() => onInsertAfter(node.stepId)}
              />
            </div>
          )
        })}

        {layout.branchBars.map((bar) => (
          <BranchBar
            key={`bar-${bar.parentId}`}
            x={bar.xStart}
            y={bar.y}
            width={bar.xEnd - bar.xStart}
          />
        ))}

        {layout.edges
          .filter(
            (edge) => edge.edgeType === 'sequential' && edge.dependencyKind === 'explicit',
          )
          .map((edge) => {
            const from = nodeById.get(edge.fromId)
            const to = nodeById.get(edge.toId)
            if (!from || !to) return null
            const midX = (from.x + to.x) / 2 + CARD_W / 2
            const midY = (from.y + CARD_H + to.y) / 2
            return (
              <div
                key={`conn-plus-${edge.fromId}-${edge.toId}`}
                className="group/conn absolute"
                style={{ left: midX - 12, top: midY - 12, width: 24, height: 24 }}
              >
                <PlusButton
                  x={12}
                  y={12}
                  onClick={() => onInsertBetween(edge.fromId, edge.toId)}
                  revealClassName="group-hover/conn:opacity-100"
                  title="Insert step between"
                />
              </div>
            )
          })}
      </div>
    </div>
  )
}
