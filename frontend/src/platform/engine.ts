// Routes engine calls to WASM or HTTP based on platform mode.
// No business logic. No data transformation. Only dispatch.

import { detectSurface } from '@/platform'
import type {
  RequestDef, RunResult, WorkflowLayoutResponse, InferDepsResponse,
  WorkflowExportRequest, WorkflowExportResponse, EngineDefaults,
  StepRef, StepScanDef,
} from '@/platform/types'

import { requestJson } from '@/platform/localApi/client'

function httpPost<T>(path: string, body: unknown): Promise<T> {
  return requestJson<T>(path, { method: 'POST', body: JSON.stringify(body) })
}

function httpGet<T>(path: string): Promise<T> {
  return requestJson<T>(path)
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function wasmBridge(): Record<string, (...args: string[]) => Promise<unknown>> {
  return (window as unknown as Record<string, unknown>)['tractl'] as Record<string, (...args: string[]) => Promise<unknown>>
}

async function wasmCall<T>(fn: string, ...args: string[]): Promise<T> {
  const bridge = wasmBridge()
  if (!bridge?.[fn]) throw new Error(`WASM not ready: ${fn}`)
  const result = await bridge[fn](...args)
  const r = result as { error?: { code: string; message: string } }
  if (r?.error?.code) throw new Error(`${r.error.code}: ${r.error.message}`)
  return result as T
}

function isWasm(): boolean {
  return detectSurface() === 'web'
}

export const engine = {
  async runRequest(def: RequestDef): Promise<RunResult> {
    if (isWasm()) return wasmCall<RunResult>('runRequest', JSON.stringify(def))
    return httpPost<RunResult>('/api/v1/run', def)
  },

  async runWorkflow(document: string, format = 'yaml'): Promise<Record<string, unknown>> {
    if (isWasm()) return wasmCall('run', document, format)
    return httpPost('/api/v1/workflows/run', { document, format })
  },

  async computeLayout(steps: StepRef[]): Promise<WorkflowLayoutResponse> {
    if (isWasm()) return wasmCall<WorkflowLayoutResponse>('layout', JSON.stringify(steps))
    return httpPost<WorkflowLayoutResponse>('/api/v1/workflow/layout', steps)
  },

  async inferDeps(steps: StepScanDef[]): Promise<InferDepsResponse> {
    if (isWasm()) return wasmCall<InferDepsResponse>('inferDeps', JSON.stringify(steps))
    return httpPost<InferDepsResponse>('/api/v1/workflow/infer-deps', steps)
  },

  async exportWorkflow(req: WorkflowExportRequest): Promise<string> {
    const res = isWasm()
      ? await wasmCall<WorkflowExportResponse>('exportWorkflow', JSON.stringify(req))
      : await httpPost<WorkflowExportResponse>('/api/v1/workflow/export', req)
    if (res.error) throw new Error(res.error)
    return res.yaml
  },

  async defaults(): Promise<EngineDefaults> {
    if (isWasm()) return wasmCall<EngineDefaults>('defaults')
    return httpGet<EngineDefaults>('/api/v1/engine/defaults')
  },
}
