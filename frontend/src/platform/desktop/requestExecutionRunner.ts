import {
  ensureApiAvailable,
  runRequestFile,
  saveRequestFile,
} from '@/platform/localApi/client'
import type { RequestExecutionRunner } from '@/platform/requestExecution/types'

export const desktopRequestExecutionRunner: RequestExecutionRunner = {
  supportsPersistence: true,

  async checkAvailable() {
    await ensureApiAvailable()
  },

  async saveRequest(input) {
    return saveRequestFile({
      id: input.id,
      name: input.name,
      request: input.request,
    })
  },

  async runRequest(input) {
    return runRequestFile({ request: input.request })
  },
}
