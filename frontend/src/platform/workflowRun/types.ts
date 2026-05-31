import type { TractlWasmParseFormat, TractlWasmRunResult } from '@/platform/web/wasm/loadTractlWasmRuntime'

export type WorkflowRunInput = {
  document: string
  format: TractlWasmParseFormat
}

export interface WorkflowRunRunner {
  checkAvailable(): Promise<void>
  runWorkflow(input: WorkflowRunInput): Promise<TractlWasmRunResult>
}
