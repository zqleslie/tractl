import { describe, expect, it } from 'vitest'
import { mapRunResult } from '@/platform/web/workflowRunAdapter'

describe('mapRunResult', () => {
  it('maps a successful workflow run into canvas step results', () => {
    const outcome = mapRunResult(
      {
        surface: 'web-wasm',
        executionMode: 'browser',
        networkProvider: 'fetch',
        Passed: true,
        Workflows: [
          {
            WorkflowID: 'wf-auth-billing',
            Passed: true,
            Steps: [
              {
                StepID: 'step-auth',
                State: 'succeeded',
                ResponseStatus: 200,
                AssertionResults: [
                  {
                    AssertionID: 'status-ok',
                    Kind: 'status',
                    Outcome: 'pass',
                    Message: '',
                    Severity: 'error',
                  },
                ],
              },
            ],
          },
        ],
        diagnostics: {
          Duration: 120,
          Workflows: [
            {
              WorkflowID: 'wf-auth-billing',
              Duration: 120,
              Steps: [
                {
                  StepID: 'step-auth',
                  Requests: [
                    {
                      URL: 'https://api.example.com/health',
                      Timeline: { TotalMs: 120 },
                    },
                  ],
                },
              ],
            },
          ],
        },
      },
      'wf-auth-billing',
    )

    expect(outcome.workflowOutcome).toBe('passed')
    expect(outcome.duration).toBe(120)
    expect(outcome.stepOutcomes['step-auth']?.outcome).toBe('passed')
    expect(outcome.stepOutcomes['step-auth']?.statusCode).toBe(200)
    expect(outcome.stepOutcomes['step-auth']?.assertions).toHaveLength(1)
    expect(outcome.stepOutcomes['step-auth']?.requestUrl).toBe(
      'https://api.example.com/health',
    )
  })

  it('maps bridge failures without throwing', () => {
    const outcome = mapRunResult(
      {
        error: {
          code: 'TRACTL_EXECUTION_ERROR',
          message: 'invalid document',
        },
      },
      'wf-auth-billing',
    )

    expect(outcome.workflowOutcome).toBe('error')
    expect(outcome.stepOutcomes).toEqual({})
    expect(outcome.errorMessage).toContain('TRACTL_EXECUTION_ERROR')
  })

  it('defaults missing steps to skipped', () => {
    const outcome = mapRunResult(
      {
        surface: 'web-wasm',
        executionMode: 'browser',
        networkProvider: 'fetch',
        Passed: false,
        Workflows: [
          {
            WorkflowID: 'wf-auth-billing',
            Passed: false,
            Steps: [
              {
                StepID: 'step-auth',
                State: 'failed',
                CausesFailure: true,
                ResponseStatus: 404,
                AssertionResults: [
                  {
                    AssertionID: 'status-ok',
                    Kind: 'status',
                    Outcome: 'fail',
                    Message: 'expected 200, got 404',
                    Severity: 'error',
                  },
                ],
              },
            ],
          },
        ],
      },
      'wf-auth-billing',
    )

    expect(outcome.workflowOutcome).toBe('failed')
    expect(outcome.stepOutcomes['step-auth']?.outcome).toBe('failed')
    expect(outcome.stepOutcomes['step-auth']?.assertions[0]?.expected).toBe('200')
    expect(outcome.stepOutcomes['step-auth']?.assertions[0]?.received).toBe('404')
    expect(outcome.stepOutcomes['step-billing']).toBeUndefined()
  })
})
