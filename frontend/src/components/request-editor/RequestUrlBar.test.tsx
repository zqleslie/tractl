import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { RequestUrlBar } from '@/components/request-editor/RequestUrlBar'

describe('RequestUrlBar', () => {
  it('renders method, url, send, and save controls', () => {
    render(
      <RequestUrlBar
        method="GET"
        url="https://api.example.com/health"
        canRun
        isRunning={false}
        onMethodChange={vi.fn()}
        onUrlChange={vi.fn()}
        onRun={vi.fn()}
        onSaveToCollection={vi.fn()}
      />,
    )

    expect(screen.getByLabelText('HTTP method')).toHaveValue('GET')
    expect(screen.getByLabelText('Request URL')).toHaveValue(
      'https://api.example.com/health',
    )
    expect(screen.getByTestId('request-run')).toBeEnabled()
    expect(screen.getByTestId('request-save-collection')).toBeInTheDocument()
  })

  it('disables run when canRun is false', () => {
    render(
      <RequestUrlBar
        method="POST"
        url=""
        canRun={false}
        isRunning={false}
        onMethodChange={vi.fn()}
        onUrlChange={vi.fn()}
        onRun={vi.fn()}
        onSaveToCollection={vi.fn()}
      />,
    )

    expect(screen.getByTestId('request-run')).toBeDisabled()
  })

  it('calls handlers when method, url, run, or save changes', () => {
    const onMethodChange = vi.fn()
    const onUrlChange = vi.fn()
    const onRun = vi.fn()
    const onSaveToCollection = vi.fn()

    render(
      <RequestUrlBar
        method="POST"
        url="https://api.example.com/users"
        canRun
        isRunning={false}
        onMethodChange={onMethodChange}
        onUrlChange={onUrlChange}
        onRun={onRun}
        onSaveToCollection={onSaveToCollection}
      />,
    )

    fireEvent.change(screen.getByLabelText('HTTP method'), {
      target: { value: 'DELETE' },
    })
    expect(onMethodChange).toHaveBeenCalledWith('DELETE')

    fireEvent.change(screen.getByLabelText('Request URL'), {
      target: { value: 'https://api.example.com/v2/users' },
    })
    expect(onUrlChange).toHaveBeenCalledWith('https://api.example.com/v2/users')

    fireEvent.click(screen.getByTestId('request-run'))
    expect(onRun).toHaveBeenCalledTimes(1)

    fireEvent.click(screen.getByTestId('request-save-collection'))
    expect(onSaveToCollection).toHaveBeenCalledTimes(1)
  })
})
