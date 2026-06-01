import type { RunHistorySourceFormat } from '@/stores/runHistoryStore'
import type { TractlWasmParseFormat } from '@/platform/web/wasm/loadTractlWasmRuntime'
import { engine } from '@/platform/engine'
import type { Workflow } from '@/types/workflow'
import { workflowToExportRequest } from './workflowToExportRequest'
import { resolveWorkflowRunFormat } from './resolveWorkflowRunFormat'

export type WorkflowRunDocumentSource = 'opened-file' | 'canvas-serializer'

export async function buildWorkflowRunPayload(
  workflow: Workflow,
  options?: { sourceFormat?: RunHistorySourceFormat },
): Promise<{
  document: string
  format: TractlWasmParseFormat
  source: WorkflowRunDocumentSource
}> {
  const hasOpenedYaml = Boolean(workflow.yaml?.trim())
  if (hasOpenedYaml) {
    return {
      document: workflow.yaml!,
      format: resolveWorkflowRunFormat(options?.sourceFormat),
      source: 'opened-file',
    }
  }
  const req = workflowToExportRequest(workflow)
  const yaml = await engine.exportWorkflow(req)
  return {
    document: yaml,
    format: resolveWorkflowRunFormat(options?.sourceFormat),
    source: 'canvas-serializer',
  }
}
