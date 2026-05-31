import type { ConfigTabDefinition, ResultTabDefinition } from '@/components/request-editor/types'

export const configTabs: ConfigTabDefinition[] = [
  { id: 'params', label: 'Params' },
  { id: 'headers', label: 'Headers' },
  { id: 'body', label: 'Body' },
  { id: 'auth', label: 'Auth' },
  { id: 'pre-script', label: 'Pre-script', script: true },
  { id: 'post-script', label: 'Post-script', script: true },
  { id: 'assertions', label: 'Assertions' },
  { id: 'extracts', label: 'Extracts' },
  { id: 'settings', label: 'Settings' },
]

export const resultTabs: ResultTabDefinition[] = [
  { id: 'body', label: 'Body' },
  { id: 'headers', label: 'Headers' },
  { id: 'assertions', label: 'Assertions' },
  { id: 'extracts', label: 'Extracts' },
]
