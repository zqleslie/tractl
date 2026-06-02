import type { HttpMethod } from '@/components/primitives'
import type { AssertionDef, ExtractDef, RequestDef } from '@/types/requestDef'
import type {
  FormRow,
  KVRow,
  RequestAuthState,
  RequestBodyState,
  RetryConfig,
} from '@/components/request-editor/types'

export type RequestState = Omit<
  RequestDef,
  'method' | 'params' | 'headers' | 'body' | 'auth' | 'assertions' | 'extracts' | 'settings'
> & {
  method: HttpMethod
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
