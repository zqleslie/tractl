import { describe, expect, it } from 'vitest'
import { createEmptyRequestFormState } from '@/components/request-editor/createEmptyRequestFormState'
import { requestFormStateToTraCtlSpec } from '@/components/request-editor/requestFormStateToTraCtlSpec'

describe('requestFormStateToTraCtlSpec', () => {
  it('serializes a minimal GET request', () => {
    const draft = createEmptyRequestFormState()
    draft.headers.push({
      id: 'h1',
      enabled: true,
      key: 'Accept',
      value: 'application/json',
    })

    const spec = requestFormStateToTraCtlSpec(
      'GET',
      'https://jsonplaceholder.typicode.com/todos/1',
      draft,
      'Health check',
    )

    expect(spec.schemaVersion).toBe(1)
    expect(spec.workflows[0]?.steps[0]?.request.operation).toBe('GET')
    expect(spec.workflows[0]?.steps[0]?.request.target).toBe(
      'https://jsonplaceholder.typicode.com/todos/1',
    )
    expect(spec.workflows[0]?.steps[0]?.request.headers?.Accept).toBe(
      'application/json',
    )
  })

  it('appends enabled query params to the target URL', () => {
    const draft = createEmptyRequestFormState()
    draft.params.push(
      { id: 'p1', enabled: true, key: 'limit', value: '10' },
      { id: 'p2', enabled: false, key: 'offset', value: '5' },
    )

    const spec = requestFormStateToTraCtlSpec(
      'GET',
      'https://api.example.com/items',
      draft,
    )

    expect(spec.workflows[0]?.steps[0]?.request.target).toBe(
      'https://api.example.com/items?limit=10',
    )
  })
})
