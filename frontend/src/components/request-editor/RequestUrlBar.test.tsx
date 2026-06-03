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
        isSaved={false}
        onMethodChange={vi.fn()}
        onUrlChange={vi.fn()}
        onRun={vi.fn()}
        onSaveToCollection={vi.fn()}
        onUpdate={vi.fn()}
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
        isSaved={false}
        onMethodChange={vi.fn()}
        onUrlChange={vi.fn()}
        onRun={vi.fn()}
        onSaveToCollection={vi.fn()}
        onUpdate={vi.fn()}
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
        isSaved={false}
        onMethodChange={onMethodChange}
        onUrlChange={onUrlChange}
        onRun={onRun}
        onSaveToCollection={onSaveToCollection}
        onUpdate={vi.fn()}
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

  it('shows Update button and calls onUpdate when isSaved is true', () => {
    const onUpdate = vi.fn()
    const onSaveToCollection = vi.fn()

    render(
      <RequestUrlBar
        method="GET"
        url="https://api.example.com/users"
        canRun
        isRunning={false}
        isSaved={true}
        onMethodChange={vi.fn()}
        onUrlChange={vi.fn()}
        onRun={vi.fn()}
        onSaveToCollection={onSaveToCollection}
        onUpdate={onUpdate}
      />,
    )

    expect(screen.getByTestId('request-update-collection')).toBeInTheDocument()
    expect(screen.queryByTestId('request-save-collection')).not.toBeInTheDocument()

    fireEvent.click(screen.getByTestId('request-update-collection'))
    expect(onUpdate).toHaveBeenCalledTimes(1)
    expect(onSaveToCollection).not.toHaveBeenCalled()
  })
})
