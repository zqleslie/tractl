import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { ResponseTabContent } from '@/components/request-editor/results/ResponseTabContent'
import type { RequestRunResult } from '@/platform/localApi/types'

const mockRunResult: RequestRunResult = {
  passed: true,
  durationMs: 120,
  statusCode: 200,
  statusLabel: '200 OK',
  contentType: 'application/json',
  body: '{"id":"1"}',
  headers: [],
  assertionResults: [],
  extractResults: [],
  passedCount: 0,
  totalCount: 0,
  timeline: [],
}

describe('ResponseTabContent', () => {
  it('shows idle state before a run', () => {
    render(
      <ResponseTabContent
        tab="body"
        runState="idle"
        runResult={null}
        executionResult={null}
      />,
    )

    expect(screen.getByTestId('request-response-idle')).toBeInTheDocument()
    expect(screen.getByText('Hit Send to execute')).toBeInTheDocument()
  })

  it('shows response body after a run', () => {
    render(
      <ResponseTabContent
        tab="body"
        runState="success"
        runResult={mockRunResult}
        executionResult={null}
      />,
    )

    expect(screen.getByTestId('request-response-body')).toBeInTheDocument()
    expect(screen.getByText('200 OK')).toBeInTheDocument()
    expect(screen.getByText(/"id": "1"/)).toBeInTheDocument()
  })

  it('switches JSON response body between pretty and raw views', () => {
    render(
      <ResponseTabContent
        tab="body"
        runState="success"
        runResult={mockRunResult}
        executionResult={null}
      />,
    )

    fireEvent.click(screen.getByRole('button', { name: 'Raw' }))

    expect(screen.getByText('{"id":"1"}')).toBeInTheDocument()
  })
})
