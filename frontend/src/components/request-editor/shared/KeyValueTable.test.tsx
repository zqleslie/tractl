import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { KeyValueTable } from '@/components/request-editor/shared/KeyValueTable'

describe('KeyValueTable', () => {
  it('renders rows and add action', () => {
    render(
      <KeyValueTable
        rows={[
          { id: 'row-1', enabled: true, key: 'limit', value: '25' },
          { id: 'row-2', enabled: false, key: 'offset', value: '0' },
        ]}
        addLabel="Add parameter"
        onAdd={vi.fn()}
        onUpdate={vi.fn()}
        onRemove={vi.fn()}
      />,
    )

    expect(screen.getByLabelText('limit key')).toHaveValue('limit')
    expect(screen.getByLabelText('limit value')).toHaveValue('25')
    expect(screen.getByLabelText('offset enabled')).not.toBeChecked()
    expect(screen.getByRole('button', { name: 'Add parameter' })).toBeInTheDocument()
  })

  it('calls update and remove handlers', () => {
    const onUpdate = vi.fn()
    const onRemove = vi.fn()

    render(
      <KeyValueTable
        rows={[{ id: 'row-1', enabled: true, key: 'limit', value: '25' }]}
        onAdd={vi.fn()}
        onUpdate={onUpdate}
        onRemove={onRemove}
      />,
    )

    fireEvent.change(screen.getByLabelText('limit value'), {
      target: { value: '50' },
    })
    expect(onUpdate).toHaveBeenCalledWith('row-1', { value: '50' })

    fireEvent.click(screen.getByLabelText('Delete limit'))
    expect(onRemove).toHaveBeenCalledWith('row-1')
  })
})
