import { createEmptyRequestFormState } from '@/components/request-editor/createEmptyRequestFormState'
import type { RequestFormState } from '@/components/request-editor/types'

/** Example request used in docs/tests only — not the default new-request state. */
export function createExampleRequestFormState(): RequestFormState {
  const draft = createEmptyRequestFormState()
  draft.headers = [
    {
      id: 'header-1',
      enabled: true,
      key: 'Content-Type',
      value: 'application/json',
    },
  ]
  draft.body.value = `{
  "email": "dev@example.com"
}`
  draft.assertions = [
    {
      id: 'assertion-1',
      kind: 'status',
      operator: 'equals',
      expected: '200',
      severity: 'error',
    },
  ]
  return draft
}
