import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { ResultsPanel } from '@/components/results/ResultsPanel'

describe('ResultsPanel', () => {
  it('shows idle state when runState is idle', () => {
    render(
      <ResultsPanel
        layout="stacked"
        activeTab="body"
        runState="idle"
        runResult={null}
        executionResult={null}
        isRunning={false}
        onTabChange={() => undefined}
      />,
    )

    expect(screen.getByText('Hit Send to execute')).toBeInTheDocument()
    expect(screen.getByText('Not run')).toBeInTheDocument()
  })

  it('shows sending badge while running', () => {
    render(
      <ResultsPanel
        layout="stacked"
        activeTab="body"
        runState="running"
        runResult={null}
        executionResult={null}
        isRunning
        onTabChange={() => undefined}
      />,
    )

    expect(screen.getByText('Sending…')).toBeInTheDocument()
  })

  it('shows pass badge after success', () => {
    render(
      <ResultsPanel
        layout="stacked"
        activeTab="body"
        runState="success"
        runResult={null}
        executionResult={{
          statusCode: 200,
          statusText: 'OK',
          durationMs: 120,
          body: '{}',
          headers: {},
          timing: { dns: 1, tcp: 2, tls: 0, ttfb: 3, transfer: 4, total: 120 },
          assertionResults: [],
          extractResults: [],
          assertionsPassed: 2,
          assertionsTotal: 2,
        }}
        isRunning={false}
        onTabChange={() => undefined}
      />,
    )

    expect(screen.getByText('2/2 passed')).toBeInTheDocument()
  })
})
