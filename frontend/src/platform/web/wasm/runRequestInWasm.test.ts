import { describe, expect, it } from 'vitest'
import { mapWasmRunResultToRequestRunResult } from '@/platform/web/wasm/runRequestInWasm'

describe('mapWasmRunResultToRequestRunResult', () => {
  it('maps a successful WASM run into RunResult', () => {
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
    expect(result.statusText).toBe('200 OK')
    expect(result.body).toBe('{"ok":true}')
    expect(result.assertionsPassed).toBe(1)
    expect(result.assertionsTotal).toBe(1)
    expect(result.timing).toEqual({
      dns: 1, tcp: 2, tls: 3, ttfb: 4, transfer: 5, total: 15, unit: 'ms',
    })
    expect(result.extractResults[0]).toEqual({
      id: 'extract-1',
      variableName: 'token',
      scope: 'workflow',
      resolvedValue: 'body',
    })
  })

  it('maps bridge failures into structured RunResult errors', () => {
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
