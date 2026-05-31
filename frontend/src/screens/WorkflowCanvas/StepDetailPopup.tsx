import { useWorkflowCanvasStore } from '@/stores/workflowCanvasStore'
import { StepDetailPanel } from '@/screens/WorkflowCanvas/StepDetailPanel'

export function StepDetailPopup() {
  const openStepId = useWorkflowCanvasStore((s) => s.openStepId)
  const openStepDefaultTab = useWorkflowCanvasStore((s) => s.openStepDefaultTab)
  const workflow = useWorkflowCanvasStore((s) => s.workflow)
  const closeDetail = useWorkflowCanvasStore((s) => s.closeStep)

  if (!openStepId || !workflow) return null

  const step = workflow.steps.find((s) => s.id === openStepId)
  if (!step) return null

  return (
    <StepDetailPanel
      key={step.id}
      step={step}
      allSteps={workflow.steps}
      workflowVariables={workflow.variables}
      defaultTab={openStepDefaultTab}
      onClose={closeDetail}
    />
  )
}
