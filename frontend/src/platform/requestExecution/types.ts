import type { TraCtlSpecDocument } from '@/components/request-editor/tractlSpecDocument'
import type {
  RequestRunResult,
  SaveRequestFileResponse,
} from '@/platform/localApi/types'

export type SaveRequestInput = {
  id?: string | null
  name: string
  document: TraCtlSpecDocument
}

export type RunRequestInput = {
  document: TraCtlSpecDocument
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
