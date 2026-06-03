import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { ScriptPanel } from '@/components/request-editor/shared/ScriptPanel'

describe('ScriptPanel', () => {
  it('copies context chip token to clipboard', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined)
    Object.assign(navigator, {
      clipboard: { writeText },
    })

    render(<ScriptPanel hook="beforeStep" value="" onChange={() => undefined} />)

    fireEvent.click(screen.getByRole('button', { name: 'ctx.spec' }))

    await waitFor(() => {
      expect(writeText).toHaveBeenCalledWith('ctx.spec')
    })
    expect(screen.getByText('Copied')).toBeInTheDocument()
  })
})
