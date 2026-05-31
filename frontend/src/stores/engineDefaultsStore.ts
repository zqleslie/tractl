import { create } from 'zustand'
import type { EngineDefaults } from '@/platform/types'

// Built-in values match Go constants in internal/localapi/mapper.go (DefaultEngineSettings).
// Used as fallback before engine.defaults() resolves — startup, offline, WASM loading.
const BUILT_IN: EngineDefaults = {
  failurePolicy: 'resilient',
  timeoutMs: 30_000,
  concurrency: 4,
  httpSuccessMin: 200,
  httpSuccessMax: 399,
  retry: {
    maxAttempts: 3,
    backoffFactor: 2,
    strategies: ['none', 'fixed', 'linear', 'exponential'],
  },
}

interface EngineDefaultsStore {
  defaults: EngineDefaults
  setDefaults: (d: EngineDefaults) => void
}

export const useEngineDefaultsStore = create<EngineDefaultsStore>((set) => ({
  defaults: BUILT_IN,
  setDefaults: (defaults) => set({ defaults }),
}))

export function getEngineDefaults(): EngineDefaults {
  return useEngineDefaultsStore.getState().defaults
}

export function isSuccessStatus(code: number): boolean {
  const { httpSuccessMin, httpSuccessMax } = getEngineDefaults()
  return code >= httpSuccessMin && code <= httpSuccessMax
}
