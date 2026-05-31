import { create } from 'zustand'
import { persist } from 'zustand/middleware'

type Density = 'comfortable' | 'default' | 'compact'
type Theme = 'light' | 'dark' | 'system'
type ResultLayout = 'stacked' | 'side-by-side'

type SettingsStore = {
  serverUrl: string | null
  density: Density
  theme: Theme
  resultLayout: ResultLayout
  historyRetentionLimit: number
  setServerUrl: (url: string | null) => void
  setDensity: (density: Density) => void
  setTheme: (theme: Theme) => void
  setResultLayout: (resultLayout: ResultLayout) => void
}

export const useSettingsStore = create<SettingsStore>()(
  persist(
    (set) => ({
      serverUrl: import.meta.env.VITE_API_URL ?? null,
      density: 'default',
      theme: 'system',
      resultLayout: 'stacked',
      historyRetentionLimit: 100,
      setServerUrl: (serverUrl) => set({ serverUrl }),
      setDensity: (density) => set({ density }),
      setTheme: (theme) => set({ theme }),
      setResultLayout: (resultLayout) => set({ resultLayout }),
    }),
    { name: 'tractl.settings' },
  ),
)
