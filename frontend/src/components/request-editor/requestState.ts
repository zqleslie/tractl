import type { HttpMethod } from '@/components/primitives'
import type {
  AssertionDef,
  ExtractDef,
  FormRow,
  KVRow,
  RequestAuthState,
  RequestBodyState,
  RetryConfig,
} from '@/components/request-editor/types'

export type RequestState = {
  id: string
  name: string
  method: HttpMethod
  url: string
  params: KVRow[]
  headers: KVRow[]
  body: RequestBodyState
  auth: RequestAuthState
  preScript: string
  postScript: string
  assertions: AssertionDef[]
  extracts: ExtractDef[]
  settings: {
    timeoutMs: number
    retry: RetryConfig | null
    failurePolicy: 'resilient' | 'failFast'
  }
}

export type { AssertionDef, ExtractDef, FormRow, KVRow, RetryConfig }
