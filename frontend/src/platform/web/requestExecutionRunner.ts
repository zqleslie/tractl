import type { RequestExecutionRunner } from '@/platform/requestExecution/types'
import type { RequestState } from '@/components/request-editor/requestState'
import { requestDefToDocumentObject } from '@/platform/mappers/requestDefToDocument'
import {
  isTractlWasmRuntimeReady,
  loadTractlWasmRuntime,
} from '@/platform/web/wasm/loadTractlWasmRuntime'
import { runRequestInWasm } from '@/platform/web/wasm/runRequestInWasm'

export class WasmRuntimeUnavailableError extends Error {
  constructor(message = 'WASM runtime is not ready') {
    super(message)
    this.name = 'WasmRuntimeUnavailableError'
  }
}

export const webRequestExecutionRunner: RequestExecutionRunner = {
  supportsPersistence: false,

  async checkAvailable() {
    await loadTractlWasmRuntime()
    if (!isTractlWasmRuntimeReady()) {
      throw new WasmRuntimeUnavailableError()
    }
  },

  async saveRequest() {
    return null
  },

  async runRequest(input) {
    const request = input.request as RequestState
    const document = requestDefToDocumentObject(request)
    return runRequestInWasm(document)
  },
}
