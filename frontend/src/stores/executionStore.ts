import { create } from 'zustand'

export type RunState = 'idle' | 'running' | 'success' | 'error'

export type AssertionResult = {
  id: string
  kind: string
  op: string
  expected: string
  received: string
  passed: boolean
  severity: 'error' | 'warning'
}

export type ExtractResult = {
  id: string
  variableName: string
  scope: 'workflow' | 'spec' | 'step'
  resolvedValue: string | null
}

export type RunTiming = {
  dns: number
  tcp: number
  tls: number
  ttfb: number
  transfer: number
  total: number
}

export type RunResult = {
  statusCode: number
  statusText: string
  durationMs: number
  body: string
  headers: Record<string, string>
  timing: RunTiming
  assertionResults: AssertionResult[]
  extractResults: ExtractResult[]
  assertionsPassed: number
  assertionsTotal: number
}

type ExecutionStore = {
  runState: RunState
  result: RunResult | null
  error: string | null
  startRun: () => void
  setResult: (result: RunResult) => void
  setError: (error: string) => void
  reset: () => void
}

export const useExecutionStore = create<ExecutionStore>((set) => ({
  runState: 'idle',
  result: null,
  error: null,
  startRun: () => set({ runState: 'running', error: null }),
  setResult: (result) =>
    set({
      runState:
        result.assertionsTotal > 0 &&
        result.assertionsPassed < result.assertionsTotal
          ? 'error'
          : 'success',
      result,
      error: null,
    }),
  setError: (error) => set({ runState: 'error', error, result: null }),
  reset: () => set({ runState: 'idle', result: null, error: null }),
}))
