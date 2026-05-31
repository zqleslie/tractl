import { useMemo } from 'react'
import { IconPlus } from '@tabler/icons-react'
import { useWorkflowCanvasStore } from '@/stores/workflowCanvasStore'
import { computeDagLayout } from './dagLayout'
import type { LayoutStep } from './dagLayout'
import { DagCanvas } from './DagCanvas'

export function GraphView() {
  const workflow = useWorkflowCanvasStore((s) => s.workflow)
  const activeFilterId = useWorkflowCanvasStore((s) => s.activeFilterId)
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
    const layoutSteps: LayoutStep[] = visibleSteps.map((step) => ({
      id: step.id,
      dependsOn: step.dependsOn,
      implicitDependsOn: step.implicitDependsOn,
    }))
    return computeDagLayout(layoutSteps)
  }, [visibleSteps])

  if (visibleSteps.length === 0) {
    return (
      <div className="flex h-full w-full flex-col items-center justify-center gap-4 bg-surface">
        <p className="text-[13px] font-[400] text-text-muted">No steps yet</p>
        <button
          type="button"
          onClick={() => insertStepAfter('')}
          title="Add first step"
          aria-label="Add first step"
          className="flex h-12 w-12 items-center justify-center rounded-full border border-[0.5px] border-slate-300 text-slate-400 hover:border-[#185FA5] hover:text-[#185FA5] dark:border-slate-600 dark:text-slate-500 dark:hover:border-[#B5D4F4] dark:hover:text-[#B5D4F4]"
        >
          <IconPlus size={22} stroke={1.5} />
        </button>
      </div>
    )
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
