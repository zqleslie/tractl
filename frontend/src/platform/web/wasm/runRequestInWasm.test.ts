import { describe, expect, it } from 'vitest'
import { mapWasmRunResultToRequestRunResult } from '@/platform/web/wasm/runRequestInWasm'

// Engine PascalCase format returned by handleRun (cmd/wasm/handlers.go).
// handleRun now returns the raw engine.RunResult directly (not routed through
// localapi.MapRunResult), so the shape is PascalCase with Workflows/Steps nesting.
const engineSuccessResult = {
  surface: 'web-wasm',
  executionMode: 'browser',
  networkProvider: 'fetch',
  Passed: true,
  Workflows: [
    {
      WorkflowID: 'request-flow',
      Passed: true,
      Skipped: false,
      Steps: [
        {
          StepID: 'request-step',
          State: 'succeeded',
          CausesFailure: false,
          ResponseStatus: 200,
          ResponseHeaders: { 'content-type': 'application/json' },
          ResponseBody: '{"ok":true}',
          Error: '',
          AssertionResults: [
            {
              AssertionID: 'status-ok',
              Kind: 'status',
              Op: 'equals',
              Expected: '200',
              Received: '200',
              Outcome: 'pass',
              Severity: 'error',
            },
          ],
        },
      ],
    },
  ],
  diagnostics: {
    Workflows: [
      {
        WorkflowID: 'request-flow',
        Duration: 15_000_000,
        Steps: [
          {
            StepID: 'request-step',
            Duration: 15_000_000,
            Requests: [
              {
                URL: 'https://example.com',
                Timeline: { DNSMs: 1, TCPMs: 2, TLSMs: 3, TTFBMs: 4, TransferMs: 5, TotalMs: 15 },
              },
            ],
            Extracts: [{ ID: 'extract-1', As: 'token', Source: 'body' }],
          },
        ],
      },
    ],
  },
}

describe('mapWasmRunResultToRequestRunResult', () => {
  it('maps a successful WASM run into RequestRunResult', () => {
    const result = mapWasmRunResultToRequestRunResult(engineSuccessResult)

    expect(result.passed).toBe(true)
    expect(result.statusCode).toBe(200)
    expect(result.statusText).toBe('200 OK')
    expect(result.body).toBe('{"ok":true}')
    expect(result.assertionsPassed).toBe(1)
    expect(result.assertionsTotal).toBe(1)
    // timing comes from diagnostics Timeline
    expect(result.timing.total).toBe(15)
    // timeline is [] — timing segments are in result.timing, not timeline
    expect(result.timeline).toEqual([])
    // extractResults mapped from diagnostics Extracts (As → variable, Source → value)
    expect(result.extractResults[0]).toEqual({
      id: 'extract-1',
      variable: 'token',
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

  it('propagates engine execution errors (e.g. empty workflow result) as bridge errors', () => {
    // When the engine produces no workflows, handleRun in Go returns a
    // TRACTL_EXECUTION_ERROR bridge error — the empty-result guard runs on the Go side.
    const result = mapWasmRunResultToRequestRunResult({
      error: {
        code: 'TRACTL_EXECUTION_ERROR',
        message: 'run produced no workflow outcomes',
      },
    })

    expect(result.passed).toBe(false)
    expect(result.error).toBe(
      'TRACTL_EXECUTION_ERROR: run produced no workflow outcomes',
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
          WorkflowID: 'request-flow',
          Passed: true,
          Skipped: false,
          Steps: [
            {
              StepID: 'request-step',
              State: 'succeeded',
              CausesFailure: false,
              ResponseStatus: 200,
              ResponseHeaders: { 'content-type': 'text/html' },
              ResponseBody:
                '<!doctype html><html><script type="module" src="/@vite/client"></script></html>',
              Error: '',
              AssertionResults: [],
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
