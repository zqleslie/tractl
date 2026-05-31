import { describe, expect, it } from 'vitest'
import { mapWasmRunResultToRequestRunResult } from '@/platform/web/wasm/runRequestInWasm'

describe('mapWasmRunResultToRequestRunResult', () => {
  it('maps a successful WASM run into RequestRunResult', () => {
    const result = mapWasmRunResultToRequestRunResult({
      surface: 'web-wasm',
      executionMode: 'browser',
      networkProvider: 'fetch',
      Passed: true,
      Workflows: [
        {
          Passed: true,
          Steps: [
            {
              StepID: 'request-step',
              ResponseStatus: 200,
              ResponseHeaders: {
                'content-type': 'application/json',
              },
              ResponseBody: '{"ok":true}',
              AssertionResults: [
                {
                  AssertionID: 'status-ok',
                  Kind: 'status',
                  Outcome: 'pass',
                  Message: '',
                },
              ],
            },
          ],
        },
      ],
      diagnostics: {
        Workflows: [
          {
            Steps: [
              {
                Requests: [
                  {
                    Timeline: {
                      DNSMs: 1,
                      TCPMs: 2,
                      TLSMs: 3,
                      TTFBMs: 4,
                      TransferMs: 5,
                      TotalMs: 15,
                    },
                  },
                ],
                Extracts: [
                  {
                    ID: 'extract-1',
                    Source: 'body',
                    As: 'token',
                  },
                ],
              },
            ],
          },
        ],
      },
    })

    expect(result.passed).toBe(true)
    expect(result.statusCode).toBe(200)
    expect(result.statusLabel).toBe('200 OK')
    expect(result.body).toBe('{"ok":true}')
    expect(result.passedCount).toBe(1)
    expect(result.totalCount).toBe(1)
    expect(result.timeline).toEqual([
      { label: 'DNS', ms: 1 },
      { label: 'TCP', ms: 2 },
      { label: 'TLS', ms: 3 },
      { label: 'TTFB', ms: 4 },
      { label: 'Transfer', ms: 5 },
    ])
    expect(result.extractResults[0]).toEqual({
      variable: 'steps.this.extracts.token',
      value: 'body',
      scope: 'workflow',
    })
  })

  it('maps bridge failures into structured RequestRunResult errors', () => {
    const result = mapWasmRunResultToRequestRunResult({
      error: {
        code: 'TRACTL_EXECUTION_ERROR',
        message: 'invalid URL',
      },
    })

    expect(result.passed).toBe(false)
    expect(result.statusCode).toBe(0)
    expect(result.error).toBe('TRACTL_EXECUTION_ERROR: invalid URL')
    expect(result.body).toBe('TRACTL_EXECUTION_ERROR: invalid URL')
  })

  it('maps empty workflow outcomes into structured errors', () => {
    const result = mapWasmRunResultToRequestRunResult({
      surface: 'web-wasm',
      executionMode: 'browser',
      networkProvider: 'fetch',
      Passed: false,
      Workflows: [],
    })

    expect(result.error).toBe(
      'TRACTL_WASM_EMPTY_RESULT: Run produced no request result',
    )
  })

  it('flags Vite dev shell responses as misconfigured URLs', () => {
    const result = mapWasmRunResultToRequestRunResult({
      surface: 'web-wasm',
      executionMode: 'browser',
      networkProvider: 'fetch',
      Passed: true,
      Workflows: [
        {
          Passed: true,
          Steps: [
            {
              StepID: 'request-step',
              ResponseStatus: 200,
              ResponseHeaders: { 'content-type': 'text/html' },
              ResponseBody:
                '<!doctype html><html><script type="module" src="/@vite/client"></script></html>',
            },
          ],
        },
      ],
    })

    expect(result.passed).toBe(false)
    expect(result.error).toMatch(/dev server/i)
    expect(result.body).toMatch(/external URL/i)
  })
})
