import { fireEvent, render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it } from 'vitest'
import { SidebarRequestsPanel } from '@/components/sidebar/SidebarRequestsPanel'
import { useRunHistoryStore } from '@/stores/runHistoryStore'
import { useUiStore } from '@/stores/uiStore'

describe('SidebarRequestsPanel', () => {
  beforeEach(() => {
    localStorage.clear()
    useRunHistoryStore.setState({ entries: [], pinnedIds: [] })
    useUiStore.setState({
      requestsSidebarView: 'history',
      requestsHistoryShowAll: false,
    })
  })

  it('defaults to the history tab with search', () => {
    render(<SidebarRequestsPanel />)

    expect(screen.getByTestId('sidebar-requests-tab-history')).toHaveAttribute(
      'aria-selected',
      'true',
    )
    expect(screen.getByTestId('sidebar-requests-search')).toBeInTheDocument()
  })

  it('switches to collections tab', () => {
    render(<SidebarRequestsPanel />)

    fireEvent.click(screen.getByTestId('sidebar-requests-tab-collections'))
    expect(screen.getByTestId('sidebar-create-folder')).toBeInTheDocument()
  })
})
