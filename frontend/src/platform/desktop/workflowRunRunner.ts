import { ensureApiAvailable, runWorkflowDocument } from '@/platform/localApi/client'
import type { TractlWasmRunResult } from '@/platform/web/wasm/loadTractlWasmRuntime'
import type { WorkflowRunRunner } from '@/platform/workflowRun/types'

export const desktopWorkflowRunRunner: WorkflowRunRunner = {
  async checkAvailable() {
    await ensureApiAvailable()
  },

  async runWorkflow(input) {
    const result = await runWorkflowDocument({
      document: input.document,
      format: input.format,
    })
    return result as TractlWasmRunResult
  },
}
