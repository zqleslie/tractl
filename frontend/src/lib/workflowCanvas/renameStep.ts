import type { WorkflowStep } from '@/types/workflow'

// Pure function — no store access
export function renameStep(
  steps: WorkflowStep[],
  oldId: string,
  newId: string,
): WorkflowStep[] {
  return steps.map((step) => {
    if (step.id === oldId) {
      return { ...step, id: newId }
    }
    return {
      ...step,
      dependsOn: step.dependsOn.map((dep) => (dep === oldId ? newId : dep)),
      implicitDependsOn: step.implicitDependsOn?.map((dep) => (dep === oldId ? newId : dep)),
    }
  })
}
