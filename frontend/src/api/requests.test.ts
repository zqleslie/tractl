import { afterEach, describe, expect, it, vi } from 'vitest'
import { runRequest } from '@/api/requests'

describe('runRequest', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('returns mapped local API response payload', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: true,
      json: async () => ({
        passed: true,
        durationMs: 234,
        statusCode: 201,
        statusText: '201 Created',
        body: JSON.stringify({ userId: 'usr_01HXYZ9ABC' }),
        headers: { 'content-type': 'application/json' },
        timing: { dns: 0, tcp: 0, tls: 0, ttfb: 0, transfer: 0, total: 234, unit: 'ms' },
        assertionResults: [],
        extractResults: [],
        assertionsPassed: 0,
        assertionsTotal: 0,
        error: '',
      }),
    } as Response)

    const result = await runRequest({
      id: 'users',
      name: 'Create user',
      method: 'POST',
      url: 'https://api.example.com/users',
      headers: [],
      assertions: [],
      extracts: [],
      settings: { timeoutMs: 30_000, failurePolicy: 'resilient' },
    })

    expect(result.statusCode).toBe(201)
    expect(result.timing.total).toBe(234)
    expect(result.body).toContain('userId')
  })
})
