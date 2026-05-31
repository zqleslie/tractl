import { render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { RequestConfigTabContent } from '@/components/request-editor/config/RequestConfigTabContent'
import { RequestConfigTabs } from '@/components/request-editor/config/RequestConfigTabs'
import { createEmptyRequestFormState } from '@/components/request-editor/createEmptyRequestFormState'
import { ThemeSync } from '@/app/ThemeSync'
import { useExecutionStore } from '@/stores/executionStore'
import { useUiStore } from '@/stores/uiStore'

const editorStub = {
  method: 'GET' as const,
  url: 'https://api.example.com',
  draft: createEmptyRequestFormState(),
  params: {
    rows: [],
    add: vi.fn(),
    update: vi.fn(),
    remove: vi.fn(),
  },
  headers: {
    rows: [{ id: 'h1', enabled: true, key: 'X-Test', value: '1' }],
    add: vi.fn(),
    update: vi.fn(),
    remove: vi.fn(),
    replace: vi.fn(),
  },
  body: {
    value: createEmptyRequestFormState().body,
    update: vi.fn(),
  },
  auth: {
    value: createEmptyRequestFormState().auth,
    update: vi.fn(),
  },
  scripts: {
    value: createEmptyRequestFormState().scripts,
    update: vi.fn(),
  },
  assertions: {
    rows: [],
    add: vi.fn(),
    update: vi.fn(),
    remove: vi.fn(),
  },
  extracts: {
    rows: [],
    add: vi.fn(),
    update: vi.fn(),
    remove: vi.fn(),
  },
  settings: {
    value: createEmptyRequestFormState().settings,
    update: vi.fn(),
  },
}

describe('Request editor panels', () => {
  beforeEach(() => {
    useUiStore.setState({ uiScale: 'default' })
    useExecutionStore.getState().reset()
  })

  it('renders all nine config tabs', () => {
    render(
      <RequestConfigTabs
        activeTab="params"
        tabCounts={{ headers: 1, assertions: 0, extracts: 0 }}
        onTabChange={vi.fn()}
      />,
    )

    for (const id of [
      'params',
      'headers',
      'body',
      'auth',
      'pre-script',
      'post-script',
      'assertions',
      'extracts',
      'settings',
    ]) {
      expect(screen.getByTestId(`request-config-tab-${id}`)).toBeInTheDocument()
    }
  })

  it('switches tab content to auth panel fields', () => {
    const { rerender } = render(
      <RequestConfigTabContent tab="auth" editor={editorStub as never} />,
    )
    expect(screen.getByText('Type')).toBeInTheDocument()

    rerender(<RequestConfigTabContent tab="body" editor={editorStub as never} />)
    expect(screen.getByTestId('body-encoding-bar')).toBeInTheDocument()
  })

  it('applies density attribute on document root', () => {
    useUiStore.setState({ uiScale: 'compact' })
    render(<ThemeSync />)
    expect(document.documentElement.dataset.density).toBe('compact')
  })

  it('shows retry fields when retry strategy is selected', () => {
    const update = vi.fn()
    render(
      <RequestConfigTabContent
        tab="settings"
        editor={
          {
            ...editorStub,
            settings: {
              value: {
                ...createEmptyRequestFormState().settings,
                retry: 'Fixed',
                retryConfig: {
                  strategy: 'fixed',
                  maxAttempts: 3,
                  delayMs: 1000,
                },
              },
              update,
            },
          } as never
        }
      />,
    )

    expect(screen.getByText('Max attempts')).toBeInTheDocument()
  })
})

describe('AssertionsPanel', () => {
  it('increments tab count badge when a row is added', async () => {
    const { AssertionsPanel } = await import('@/components/panels/AssertionsPanel')
    const onAdd = vi.fn()
    const { rerender } = render(
      <>
        <RequestConfigTabs
          activeTab="assertions"
          tabCounts={{ headers: 0, assertions: 0, extracts: 0 }}
          onTabChange={vi.fn()}
        />
        <AssertionsPanel value={[]} onAdd={onAdd} onUpdate={vi.fn()} onRemove={vi.fn()} />
      </>,
    )

    expect(screen.queryByText('1')).not.toBeInTheDocument()
    onAdd.mockImplementation(() => undefined)
    rerender(
      <>
        <RequestConfigTabs
          activeTab="assertions"
          tabCounts={{ headers: 0, assertions: 1, extracts: 0 }}
          onTabChange={vi.fn()}
        />
        <AssertionsPanel
          value={[
            {
              id: 'a1',
              kind: 'status',
              operator: 'equals',
              expected: '200',
              severity: 'error',
            },
          ]}
          onAdd={onAdd}
          onUpdate={vi.fn()}
          onRemove={vi.fn()}
        />
      </>,
    )

    expect(screen.getByText('1')).toBeInTheDocument()
  })
})

describe('ExtractsPanel', () => {
  it('updates downstream hint when variable name changes', async () => {
    const { ExtractsPanel } = await import('@/components/panels/ExtractsPanel')
    const row = {
      id: 'e1',
      source: 'body',
      path: '$.id',
      variable: 'userId',
      scope: 'workflow',
    }
    const { rerender } = render(
      <ExtractsPanel
        value={[row]}
        onAdd={vi.fn()}
        onUpdate={vi.fn()}
        onRemove={vi.fn()}
      />,
    )

    expect(screen.getByText(/\$\{steps\.this\.extracts\.userId\}/)).toBeInTheDocument()

    rerender(
      <ExtractsPanel
        value={[{ ...row, variable: 'accountId' }]}
        onAdd={vi.fn()}
        onUpdate={vi.fn()}
        onRemove={vi.fn()}
      />,
    )

    expect(screen.getByText(/\$\{steps\.this\.extracts\.accountId\}/)).toBeInTheDocument()
  })
})

describe('Auth panel', () => {
  it('reveals bearer fields when bearer is selected', async () => {
    const { AuthPanel } = await import('@/components/panels/AuthPanel')
    const onChange = vi.fn()
    const { rerender } = render(
      <AuthPanel
        value={{ ...createEmptyRequestFormState().auth, type: 'None' }}
        onChange={onChange}
      />,
    )
    expect(screen.getByText('No authentication')).toBeInTheDocument()

    rerender(
      <AuthPanel
        value={{ ...createEmptyRequestFormState().auth, type: 'Bearer token', token: '' }}
        onChange={onChange}
      />,
    )
    expect(screen.getByLabelText('Bearer token')).toBeInTheDocument()
  })
})
