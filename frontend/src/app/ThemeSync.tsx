import { useEffect } from 'react'
import { useUiStore } from '@/stores/uiStore'

function resolveTheme(theme: 'light' | 'dark' | 'system'): 'light' | 'dark' {
  if (theme !== 'system') return theme
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

export function ThemeSync() {
  const theme = useUiStore((s) => s.theme)
  const uiScale = useUiStore((s) => s.uiScale)

  useEffect(() => {
    const applyTheme = () => {
      const resolvedTheme = resolveTheme(theme)
      document.documentElement.dataset.theme = resolvedTheme
      document.documentElement.classList.toggle('dark', resolvedTheme === 'dark')
    }

    applyTheme()

    if (theme !== 'system') return undefined

    const media = window.matchMedia('(prefers-color-scheme: dark)')
    media.addEventListener('change', applyTheme)
    return () => media.removeEventListener('change', applyTheme)
  }, [theme])

  useEffect(() => {
    document.documentElement.dataset.density = uiScale
    document.documentElement.dataset.uiScale = uiScale
  }, [uiScale])

  return null
}
