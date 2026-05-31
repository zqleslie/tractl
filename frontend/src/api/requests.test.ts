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
        statusLabel: '201 Created',
        contentType: 'application/json',
        body: JSON.stringify({ userId: 'usr_01HXYZ9ABC' }),
        headers: [
          { id: 'header-1', enabled: true, key: 'content-type', value: 'application/json' },
        ],
        assertionResults: [],
        extractResults: [],
        passedCount: 0,
        totalCount: 0,
        timeline: [{ label: 'Total', ms: 234 }],
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
