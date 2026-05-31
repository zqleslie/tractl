import { loadTractlWasmRuntime } from '@/platform/web/wasm/loadTractlWasmRuntime'
import type { WorkflowRunRunner } from '@/platform/workflowRun/types'

export const webWorkflowRunRunner: WorkflowRunRunner = {
  async checkAvailable() {
    await loadTractlWasmRuntime()
    if (!window.tractl?.run) {
      throw new Error('WASM runtime is not ready')
    }
  },

  async runWorkflow(input) {
    return window.tractl!.run(input.document, input.format)
  },
}
