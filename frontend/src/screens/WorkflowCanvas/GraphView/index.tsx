import { useMemo } from 'react'
import { IconPlus } from '@tabler/icons-react'
import { useWorkflowCanvasStore } from '@/stores/workflowCanvasStore'
import { computeDagLayout } from './dagLayout'
import { DagCanvas } from './DagCanvas'

export function GraphView() {
  const workflow = useWorkflowCanvasStore((s) => s.workflow)
  const activeFilterId = useWorkflowCanvasStore((s) => s.activeFilterId)
  const engineLayout = useWorkflowCanvasStore((s) => s.layout)
  const openStep = useWorkflowCanvasStore((s) => s.openStep)
  const insertStepAfter = useWorkflowCanvasStore((s) => s.insertStepAfter)
  const insertStepBetween = useWorkflowCanvasStore((s) => s.insertStepBetween)

  const visibleSteps = useMemo(() => {
    if (!workflow) return []
    const showWorkflow =
      activeFilterId === null ||
      activeFilterId === workflow.name ||
      activeFilterId === workflow.id
    return showWorkflow ? workflow.steps : []
  }, [workflow, activeFilterId])

  const layout = useMemo(() => {
    if (!engineLayout) return null
    return computeDagLayout(engineLayout)
  }, [engineLayout])

  if (visibleSteps.length === 0) {
    return (
      <div className="flex h-full w-full flex-col items-center justify-center gap-4 bg-surface">
        <p className="text-[13px] font-[400] text-text-muted">No steps yet</p>
        <button
          type="button"
          onClick={() => insertStepAfter('')}
          title="Add first step"
          aria-label="Add first step"
          className="flex h-12 w-12 items-center justify-center rounded-full border border-[0.5px] border-border text-text-muted hover:border-primary hover:text-primary"
        >
          <IconPlus size={22} stroke={1.5} />
        </button>
      </div>
    )
  }

  if (!layout) {
    return <div className="flex h-full w-full bg-surface" />
  }

  return (
    <DagCanvas
      layout={layout}
      steps={visibleSteps}
      onStepClick={openStep}
      onInsertAfter={insertStepAfter}
      onInsertBetween={insertStepBetween}
    />
  )
}
