import type { TraCtlSpecDocument } from '@/components/request-editor/tractlSpecDocument'
import { loadWorkflowIntoCanvas } from '@/lib/workflowCanvas/loadWorkflowIntoCanvas'
import { workflowDocumentToCanvasWorkflow } from '@/lib/workflowCanvas/workflowDocumentToCanvasWorkflow'
import type { RunHistorySourceFormat } from '@/stores/runHistoryStore'
import { useWorkflowWorkspaceStore } from '@/stores/workflowWorkspaceStore'
import { useUiStore } from '@/stores/uiStore'

export function openWorkflowFromHistoryDocument(input: {
  document: TraCtlSpecDocument
  sourceName?: string
  sourceFormat?: RunHistorySourceFormat
}): string {
  const workflow = workflowDocumentToCanvasWorkflow(input.document, {
    sourceName: input.sourceName,
  })

  useWorkflowWorkspaceStore.getState().upsertWorkflow(workflow, {
    sourceName: input.sourceName ?? workflow.name,
    sourceFormat: input.sourceFormat ?? 'yaml',
  })

  useUiStore.getState().setSidebarTab('workflows')
  useUiStore.getState().openWorkflow(workflow.id)
  loadWorkflowIntoCanvas(workflow)

  return workflow.id
}
