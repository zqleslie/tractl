import type { RunResult } from '@/platform/types'
import { requestRunResultToExecutionResult } from '@/lib/execution/mapRunResults'
import type { RequestDef } from '@/types/requestDef'
import { apiPost } from './client'
import type { RunResult as ExecutionRunResult } from '@/stores/executionStore'

export async function runRequest(payload: RequestDef): Promise<ExecutionRunResult> {
  const result = await apiPost<RunResult>('/run', payload)
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
