import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { AssertionResult } from '@/components/request-editor/results/AssertionResult'

describe('AssertionResult', () => {
  it('shows failed assertions first and collapses passed when mixed', () => {
    render(
      <AssertionResult
        results={[
          {
            id: '1',
            kind: 'status',
            op: 'equals',
            expected: '200',
            received: '404',
            passed: false,
            severity: 'error',
          },
          {
            id: '2',
            kind: 'status',
            op: 'equals',
            expected: '201',
            received: '201',
            passed: true,
            severity: 'error',
          },
          {
            id: '3',
            kind: 'status',
            op: 'equals',
            expected: '201',
            received: '201',
            passed: true,
            severity: 'error',
          },
          {
            id: '4',
            kind: 'status',
            op: 'equals',
            expected: '201',
            received: '201',
            passed: true,
            severity: 'error',
          },
        ]}
      />,
    )

    expect(screen.getByText('1 failed · 3 passed')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: /3 passed/i }))
    expect(screen.getAllByText(/status equals/).length).toBeGreaterThan(1)
  })
})
