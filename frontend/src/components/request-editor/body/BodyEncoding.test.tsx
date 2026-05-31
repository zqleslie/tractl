import { fireEvent, render, screen, within } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { BodyContent } from '@/components/request-editor/body/BodyContent'
import { BodyEncodingBar } from '@/components/request-editor/body/BodyEncodingBar'

describe('Body encoding editor', () => {
  it('switches between all body modes', () => {
    const onChange = vi.fn()
    render(<BodyEncodingBar value="JSON" onChange={onChange} />)

    const select = screen.getByLabelText('Body encoding')
    for (const mode of ['Form data', 'Multipart', 'Raw', 'Binary', 'None']) {
      fireEvent.change(select, { target: { value: mode } })
      expect(onChange).toHaveBeenCalledWith(mode)
    }
  })

  it('renders mode-specific body content', () => {
    const onChange = vi.fn()
    const baseProps = {
      formRows: [],
      rawContentType: 'text/plain',
      binaryFile: '',
      onChange,
    }
    const { rerender } = render(
      <BodyContent encoding="JSON" value="{}" {...baseProps} />,
    )
    expect(screen.getByLabelText('Request body')).toBeInTheDocument()

    rerender(<BodyContent encoding="Form data" value="" {...baseProps} />)
    expect(screen.getByTestId('body-content-Form data')).toBeInTheDocument()

    rerender(<BodyContent encoding="Multipart" value="" {...baseProps} />)
    expect(screen.getByText(/separate part/i)).toBeInTheDocument()

    rerender(<BodyContent encoding="Raw" value="plain" {...baseProps} />)
    expect(screen.getByLabelText('Raw request body')).toHaveValue('plain')

    rerender(<BodyContent encoding="Binary" value="" {...baseProps} />)
    expect(screen.getByText(/Drop a file here/i)).toBeInTheDocument()

    rerender(<BodyContent encoding="None" value="" {...baseProps} />)
    expect(screen.getByText('No body sent with this request')).toBeInTheDocument()
  })

  it('shows file picker control when form row type is File', () => {
    const onChange = vi.fn()
    render(
      <BodyContent
        encoding="Form data"
        value=""
        formRows={[
          {
            id: 'f1',
            enabled: true,
            key: 'avatar',
            value: '',
            type: 'file',
          },
        ]}
        rawContentType="text/plain"
        binaryFile=""
        onChange={onChange}
      />,
    )

    const row = screen.getByTestId('body-content-Form data')
    expect(within(row).getByRole('button', { name: 'Choose file' })).toBeInTheDocument()
  })
})
