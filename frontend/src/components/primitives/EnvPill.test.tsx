import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { EnvPill } from '@/components/primitives/EnvPill'
import { isDangerousEnvironment } from '@/components/primitives/envPillModel'

describe('EnvPill', () => {
  it('marks production environments as dangerous', () => {
    const environment = {
      id: 'env-production',
      name: 'Production',
      variables: {},
    }

    render(<EnvPill environment={environment} />)

    expect(isDangerousEnvironment(environment)).toBe(true)
    expect(screen.getByTestId('env-pill')).toHaveAttribute('data-dangerous', 'true')
    expect(screen.getByTestId('env-pill')).toHaveClass('env-pill-danger')
  })
})
