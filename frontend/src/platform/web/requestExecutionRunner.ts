import type { RequestExecutionRunner } from '@/platform/requestExecution/types'
import { draftToRequestDef } from '@/lib/requestEditor/draftToRequestDef'
import type { RequestState } from '@/components/request-editor/requestState'
import { requestStateToDraft } from '@/lib/requestEditor/workspaceMapping'
import { engine } from '@/platform/engine'

export class WasmRuntimeUnavailableError extends Error {
  constructor(message = 'WASM runtime is not ready') {
    super(message)
    this.name = 'WasmRuntimeUnavailableError'
  }
}

export const webRequestExecutionRunner: RequestExecutionRunner = {
  supportsPersistence: false,

  async checkAvailable() {
    if (!window.tractl?.run) {
      throw new WasmRuntimeUnavailableError()
    }
  },

  async saveRequest() {
    return null
  },

  async runRequest(input) {
    const request = input.request as RequestState
    const draft = requestStateToDraft(request)
    const def = draftToRequestDef(request.method, request.url, draft)
    return engine.runRequest(def)
  },
}
