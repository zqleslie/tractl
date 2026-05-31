import { describe, expect, it } from 'vitest'
import { stripAuthoringSystemFields } from '@/lib/workflowCanvas/stripAuthoringSystemFields'

describe('stripAuthoringSystemFields', () => {
  it('removes _ulid and reserved-prefix keys recursively', () => {
    const input = {
      id: 'step-1',
      _ulid: '01ARZ3NDEKTSV4RRFFQ69G5FAV',
      assertions: [{ id: 'a1', _ulid: '01ARZ3NDEKTSV4RRFFQ69G5FAV', kind: 'status' }],
      '_traCtl.debug': true,
    }

    expect(stripAuthoringSystemFields(input)).toEqual({
      id: 'step-1',
      assertions: [{ id: 'a1', kind: 'status' }],
    })
  })
})
