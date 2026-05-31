import type { RunHistorySourceFormat } from '@/stores/runHistoryStore'
import type { TractlWasmParseFormat } from '@/platform/web/wasm/loadTractlWasmRuntime'
import { serializeWorkflow } from '@/platform/web/workflowSerializer'
import type { Workflow } from '@/types/workflow'
import { resolveWorkflowRunFormat } from './resolveWorkflowRunFormat'

export type WorkflowRunDocumentSource = 'opened-file' | 'canvas-serializer'

export function buildWorkflowRunPayload(
  workflow: Workflow,
  options?: { sourceFormat?: RunHistorySourceFormat },
): {
  document: string
  format: TractlWasmParseFormat
  source: WorkflowRunDocumentSource
} {
  const hasOpenedYaml = Boolean(workflow.yaml?.trim())
  const document = hasOpenedYaml ? workflow.yaml! : serializeWorkflow(workflow)
  return {
    document,
    format: resolveWorkflowRunFormat(options?.sourceFormat),
    source: hasOpenedYaml ? 'opened-file' : 'canvas-serializer',
  }
}
