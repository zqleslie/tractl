import { fireEvent, render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it } from 'vitest'
import { ActivityBar } from '@/components/shell/ActivityBar'
import { ThemeSync } from '@/app/ThemeSync'
import { AppShell } from '@/components/shell/AppShell'
import { CommandPalette } from '@/components/shell/CommandPalette'
import { SidePanel } from '@/components/shell/SidePanel'
import { TabBar } from '@/components/shell/TabBar'
import { useUiStore } from '@/stores/uiStore'

describe('App shell v4', () => {
  beforeEach(() => {
    localStorage.clear()
    useUiStore.setState({
      activeScreen: 'fast-start',
      sidebarTab: 'requests',
      sidebarCollapsed: true,
      activityBarExpanded: false,
      commandPaletteOpen: false,
      keyboardShortcutsOpen: false,
      tabs: [],
      activeTabId: null,
      requestEditorKey: 0,
      theme: 'light',
      uiScale: 'default',
      editorLayout: 'stacked',
      wordWrap: true,
    })
  })

  it('renders shell chrome and content', () => {
    render(
      <AppShell>
        <div>Editor content</div>
      </AppShell>,
    )

    expect(screen.getByTestId('activity-bar')).toBeInTheDocument()
    expect(screen.queryByTestId('side-panel')).not.toBeInTheDocument()
    expect(screen.getByTestId('topbar')).toBeInTheDocument()
    expect(screen.getByTestId('tab-bar')).toBeInTheDocument()
    expect(screen.getByTestId('bottom-toolbar')).toBeInTheDocument()
    expect(screen.getByText('Editor content')).toBeInTheDocument()
  })

  it('collapses an active side panel and opens another panel', () => {
    render(
      <>
        <ActivityBar />
        <SidePanel />
      </>,
    )

    fireEvent.click(screen.getByTestId('activity-requests'))
    expect(screen.getByTestId('sidebar-new-request')).toBeInTheDocument()
    expect(screen.getByTestId('sidebar-requests-tab-history')).toBeInTheDocument()

    fireEvent.click(screen.getByTestId('activity-requests'))
    expect(screen.queryByTestId('side-panel')).not.toBeInTheDocument()

    fireEvent.click(screen.getByTestId('activity-workflows'))
    expect(screen.getByTestId('side-panel')).toHaveTextContent('Workflows')
  })

  it('expands the activity bar to show labels', () => {
    render(<ActivityBar />)

    expect(screen.queryByText('Requests')).not.toBeInTheDocument()

    fireEvent.click(screen.getByTestId('activity-expand'))

    expect(screen.getByText('Requests')).toBeInTheDocument()
  })

  it('closing the active tab activates the adjacent tab', () => {
    useUiStore.setState({
      tabs: [
        { id: 'tab-a', type: 'request', title: 'First', method: 'GET' },
        { id: 'tab-b', type: 'request', title: 'Second', method: 'POST' },
        { id: 'tab-c', type: 'workflow', title: 'Flow', method: 'WF' },
      ],
      activeTabId: 'tab-b',
    })

    render(<TabBar />)

    fireEvent.click(screen.getByTestId('close-tab-tab-b'))

    expect(useUiStore.getState().activeTabId).toBe('tab-c')
    expect(screen.getByRole('tab', { selected: true })).toHaveTextContent('Flow')
  })

  it('updates the active request tab after request metadata changes', () => {
    useUiStore.setState({
      tabs: [{ id: 'tab-a', type: 'request', title: 'Untitled request', method: 'GET' }],
      activeTabId: 'tab-a',
    })

    useUiStore.getState().updateActiveRequestTab({
      method: 'POST',
      title: 'api.example.test/users',
    })

    render(<TabBar />)

    expect(screen.getByRole('tab', { selected: true })).toHaveTextContent(
      'api.example.test/users',
    )
    expect(screen.getByText('POST')).toBeInTheDocument()
  })

  it('filters command palette results', () => {
    useUiStore.setState({ commandPaletteOpen: true })
    render(<CommandPalette />)

    fireEvent.change(screen.getByLabelText('Command search'), {
      target: { value: 'workflow' },
    })

    expect(screen.getByText('New workflow')).toBeInTheDocument()
    expect(screen.queryByText('New request')).not.toBeInTheDocument()
  })

  it('applies the selected density to the document root', () => {
    useUiStore.setState({ uiScale: 'comfortable' })

    render(<ThemeSync />)

    expect(document.documentElement.dataset.density).toBe('comfortable')
    expect(document.documentElement.dataset.uiScale).toBe('comfortable')
  })

  it('selects density from the bottom toolbar text-size dropdown', () => {
    render(
      <AppShell>
        <div>Editor content</div>
      </AppShell>,
    )

    const densitySelect = screen.getByTestId('density-toggle').querySelector('select')
    if (!densitySelect) throw new Error('density select not found')
    fireEvent.change(densitySelect, { target: { value: 'comfortable' } })

    expect(useUiStore.getState().uiScale).toBe('comfortable')
  })
})
