import { useWorkflowCanvasStore } from '@/stores/workflowCanvasStore'
import type { Workflow } from '@/types/workflow'

/** Load a parsed workflow onto the canvas without pre-seeded run results. */
export function loadWorkflowIntoCanvas(workflow: Workflow): void {
  useWorkflowCanvasStore.getState().loadCanvasWorkflow(workflow)
}
