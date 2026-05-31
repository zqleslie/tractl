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
      document: input.document,
    })
  },

  async runRequest(input) {
    if (!input.fileId) {
      throw new Error('Request must be saved before running')
    }
    return runRequestFile({ fileId: input.fileId })
  },
}
