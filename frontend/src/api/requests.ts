import type { RunResult } from '@/stores/executionStore'
import { requestRunResultToExecutionResult } from '@/lib/execution/mapRunResults'
import type { RequestRunResult } from '@/platform/localApi/types'
import type { RequestDef } from '@/types/requestDef'
import { apiPost } from './client'

export async function runRequest(payload: RequestDef): Promise<RunResult> {
  const result = await apiPost<RequestRunResult>('/run', payload)
  return requestRunResultToExecutionResult(result)
}

export async function saveRequestFile(
  payload: { path: string; content: string },
): Promise<{ path: string; updatedAt: string }> {
  const saved = await apiPost<{ path: string; modifiedAt: string }>(
    `/files/${payload.path}`,
    { content: payload.content },
  )
  return { path: saved.path, updatedAt: saved.modifiedAt }
}
