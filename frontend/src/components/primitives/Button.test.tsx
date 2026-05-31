import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { Button } from './Button'

describe('Button', () => {
  it('renders label and primary variant styles', () => {
    render(<Button variant="primary">Run</Button>)
    const button = screen.getByRole('button', { name: 'Run' })
    expect(button).toBeInTheDocument()
    expect(button.className).toContain('bg-primary')
  })
})
