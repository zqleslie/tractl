import { render } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { TimingBar } from '@/components/request-editor/results/TimingBar'

describe('TimingBar', () => {
  it('renders segments proportional to timing values', () => {
    const { container } = render(
      <TimingBar
        timing={{
          dns: 4,
          tcp: 12,
          tls: 23,
          ttfb: 187,
          transfer: 8,
          total: 234,
        }}
      />,
    )

    const segments = container.querySelectorAll('.timing-segment')
    expect(segments).toHaveLength(5)
    expect(segments[3]?.getAttribute('style')).toContain('flex: 187')
  })
})
