import { fireEvent, render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it } from 'vitest'
import { Topbar } from '@/components/shell/Topbar'
import { useEnvironmentStore } from '@/stores/environmentStore'
import { useUiStore } from '@/stores/uiStore'

describe('Topbar', () => {
  beforeEach(() => {
    localStorage.clear()
    useEnvironmentStore.setState({
      environments: [
        {
          id: 'env-development',
          name: 'Development',
          variables: { baseUrl: 'https://dev.example.test' },
        },
        {
          id: 'env-staging',
          name: 'Staging',
          variables: {},
        },
      ],
      activeEnvironmentId: 'env-development',
    })
    useUiStore.setState({
      activeScreen: 'fast-start',
      theme: 'light',
      uiScale: 'default',
      commandPaletteOpen: false,
    })
  })

  it('shows the active environment and switches from the popover', () => {
    render(<Topbar />)

    expect(screen.getByTestId('env-pill')).toHaveTextContent('Development')

    fireEvent.click(screen.getByTestId('env-pill'))
    fireEvent.click(screen.getByRole('menuitem', { name: /Staging/ }))

    expect(useEnvironmentStore.getState().activeEnvironmentId).toBe('env-staging')
    expect(screen.getByTestId('env-pill')).toHaveTextContent('Staging')
  })

  it('creates a new environment from the quick switcher', () => {
    render(<Topbar />)

    fireEvent.click(screen.getByTestId('env-pill'))
    fireEvent.change(screen.getByLabelText('Quick new environment name'), {
      target: { value: 'Production' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'New' }))

    const state = useEnvironmentStore.getState()
    expect(state.environments.map((environment) => environment.name)).toContain(
      'Production',
    )
    expect(screen.getByTestId('env-pill')).toHaveTextContent('Production')
    expect(useUiStore.getState().activeScreen).toBe('environments')
  })

  it('closes the environment popover when clicking outside', () => {
    render(
      <div>
        <Topbar />
        <button type="button">Outside target</button>
      </div>,
    )

    fireEvent.click(screen.getByTestId('env-pill'))
    expect(screen.getByTestId('env-switcher-popover')).toBeInTheDocument()

    fireEvent.mouseDown(screen.getByText('Outside target'))
    expect(screen.queryByTestId('env-switcher-popover')).not.toBeInTheDocument()
  })

  it('toggles from light to dark theme', () => {
    render(<Topbar />)

    fireEvent.click(screen.getByTestId('theme-toggle'))

    expect(useUiStore.getState().theme).toBe('dark')
  })

})
