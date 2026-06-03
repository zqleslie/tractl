import type { RunResult } from '@/stores/executionStore'
import { requestRunResultToExecutionResult } from '@/lib/execution/mapRunResults'
import type { RequestRunResult } from '@/platform/localApi/types'
import type { RequestDef } from '@/types/requestDef'
import { apiPost } from './client'

export async function runRequest(payload: RequestDef): Promise<RunResult> {
  const result = await apiPost<RequestRunResult>('/run', payload)
  return requestRunResultToExecutionResult(result)
}

