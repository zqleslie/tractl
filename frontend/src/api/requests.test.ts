import { describe, expect, it } from 'vitest'
import { runRequest } from '@/api/requests'

describe('runRequest stub', () => {
  it('returns mock timing and response payload', async () => {
    const result = await runRequest({
      method: 'POST',
      url: 'https://api.example.com/users',
      headers: {},
      assertions: [],
      extracts: [],
      settings: { timeoutMs: 30_000, failurePolicy: 'resilient' },
      environmentId: null,
    })

    expect(result.statusCode).toBe(201)
    expect(result.timing.total).toBe(234)
    expect(result.body).toContain('userId')
  })
})
