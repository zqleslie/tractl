import type {
  RequestRunResult,
  SaveRequestFileResponse,
} from '@/platform/localApi/types'
import type { RequestDef } from '@/types/requestDef'

export type SaveRequestInput = {
  id?: string | null
  name: string
  request: RequestDef
}

export type RunRequestInput = {
  request: RequestDef
  fileId?: string | null
  name: string
}

/** Platform transport for request editor save + run (ADR-016 platform abstraction). */
export type RequestExecutionRunner = {
  readonly supportsPersistence: boolean
  checkAvailable(): Promise<void>
  saveRequest(input: SaveRequestInput): Promise<SaveRequestFileResponse | null>
  runRequest(input: RunRequestInput): Promise<RequestRunResult>
}
