import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { RequestConfigTabs } from '@/components/request-editor/config/RequestConfigTabs'

describe('RequestConfigTabs', () => {
  it('marks the active tab selected', () => {
    render(
      <RequestConfigTabs
        activeTab="params"
        tabCounts={{ headers: 2, assertions: 2, extracts: 1 }}
        onTabChange={vi.fn()}
      />,
    )

    expect(screen.getByTestId('request-config-tab-params')).toHaveAttribute(
      'aria-selected',
      'true',
    )
    expect(screen.getByTestId('request-config-tab-headers')).toHaveAttribute(
      'aria-selected',
      'false',
    )
  })

  it('notifies parent when a tab is chosen', () => {
    const onTabChange = vi.fn()

    render(
      <RequestConfigTabs
        activeTab="params"
        tabCounts={{ headers: 2, assertions: 2, extracts: 1 }}
        onTabChange={onTabChange}
      />,
    )

    fireEvent.click(screen.getByTestId('request-config-tab-assertions'))
    expect(onTabChange).toHaveBeenCalledWith('assertions')
  })
})
