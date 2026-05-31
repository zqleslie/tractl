import { useState } from 'react'
import { IconMoon, IconSun } from '@tabler/icons-react'
import { Badge, Button } from '@/components/primitives'
import { cn } from '@/lib/cn'
import {
  isDuplicateEnvironmentName,
  useEnvironmentStore,
} from '@/stores/environmentStore'
import { useUiStore } from '@/stores/uiStore'

export function Titlebar() {
  const theme = useUiStore((s) => s.theme)
  const toggleTheme = useUiStore((s) => s.toggleTheme)
  const setActiveScreen = useUiStore((s) => s.setActiveScreen)
  const environments = useEnvironmentStore((s) => s.environments)
  const activeEnvironmentId = useEnvironmentStore((s) => s.activeEnvironmentId)
  const setActiveEnvironment = useEnvironmentStore((s) => s.setActiveEnvironment)
  const createEnvironment = useEnvironmentStore((s) => s.createEnvironment)
  const [envMenuOpen, setEnvMenuOpen] = useState(false)
  const [newEnvironmentName, setNewEnvironmentName] = useState('')

  const activeEnvironment =
    environments.find((environment) => environment.id === activeEnvironmentId) ??
    null

  const duplicateNewName = isDuplicateEnvironmentName(
    environments,
    newEnvironmentName,
  )
  const canCreateEnvironment =
    newEnvironmentName.trim().length > 0 && !duplicateNewName

  const createFromPopover = () => {
    if (!canCreateEnvironment) return
    const id = createEnvironment(newEnvironmentName)
    setActiveEnvironment(id)
    setNewEnvironmentName('')
    setEnvMenuOpen(false)
    setActiveScreen('environments')
  }

  return (
    <header
      className="flex h-10 shrink-0 items-center justify-between border-b-[0.5px] border-border px-4"
      data-testid="app-titlebar"
    >
      <div className="flex items-center gap-2">
        <span className="text-ui-sm font-medium text-text">traCtl</span>
        <Badge tone="info">alpha</Badge>
      </div>

      <div className="flex items-center gap-2">
        <div className="relative">
          <button
            type="button"
            className={cn(
              'inline-flex items-center gap-1 rounded-ui border-[0.5px] px-1.5 py-0.5 text-ui-xs font-medium',
              activeEnvironment
                ? 'border-success bg-success-bg text-success-fg'
                : 'border-border bg-surface-elevated text-text-muted',
            )}
            data-testid="env-badge"
            aria-expanded={envMenuOpen}
            aria-haspopup="menu"
            onClick={() => setEnvMenuOpen((open) => !open)}
          >
            <span
              className={cn(
                'h-1.5 w-1.5 rounded-full',
                activeEnvironment ? 'bg-success-fg' : 'bg-text-muted',
              )}
            />
            {activeEnvironment?.name ?? 'No environment'}
          </button>

          {envMenuOpen ? (
            <div
              className="absolute right-0 top-7 z-20 w-64 rounded-ui border-[0.5px] border-border bg-surface p-2 shadow-none"
              role="menu"
              data-testid="env-switcher-popover"
            >
              <div className="px-1 pb-1 text-[10px] font-medium uppercase tracking-wider text-text-muted">
                Switch environment
              </div>
              {environments.length === 0 ? (
                <p className="px-1 py-2 text-ui-xs text-text-muted">
                  No environments configured.
                </p>
              ) : (
                <div className="max-h-40 overflow-auto">
                  {environments.map((environment) => {
                    const active = environment.id === activeEnvironmentId
                    return (
                      <button
                        key={environment.id}
                        type="button"
                        className={cn(
                          'flex w-full items-center gap-2 rounded-ui px-2 py-1.5 text-left text-ui-sm hover:bg-surface-elevated',
                          active ? 'text-text' : 'text-text-muted',
                        )}
                        role="menuitem"
                        onClick={() => {
                          setActiveEnvironment(environment.id)
                          setEnvMenuOpen(false)
                        }}
                      >
                        <span
                          className={cn(
                            'h-2 w-2 rounded-full',
                            active ? 'bg-success-fg' : 'bg-text-muted',
                          )}
                        />
                        <span className="min-w-0 flex-1 truncate">
                          {environment.name}
                        </span>
                        <Badge tone="neutral">
                          {Object.keys(environment.variables).length}
                        </Badge>
                      </button>
                    )
                  })}
                </div>
              )}

              <div className="mt-2 border-t-[0.5px] border-border pt-2">
                <label className="block px-1 pb-1 text-[10px] font-medium uppercase tracking-wider text-text-muted">
                  New environment
                </label>
                <div className="flex gap-1">
                  <input
                    aria-label="Quick new environment name"
                    placeholder="Development"
                    className="min-w-0 flex-1 rounded-ui border-[0.5px] border-border bg-surface px-2 py-1 text-ui-xs text-text outline-none focus:border-primary"
                    value={newEnvironmentName}
                    onChange={(event) =>
                      setNewEnvironmentName(event.currentTarget.value)
                    }
                  />
                  <Button
                    variant="secondary"
                    className="px-2 py-1 text-ui-xs"
                    disabled={!canCreateEnvironment}
                    onClick={createFromPopover}
                  >
                    New
                  </Button>
                </div>
              </div>
            </div>
          ) : null}
        </div>
        <Button
          variant="ghost"
          className="px-2"
          aria-label={
            theme === 'light' ? 'Switch to dark theme' : 'Switch to light theme'
          }
          data-testid="theme-toggle"
          onClick={toggleTheme}
        >
          {theme === 'light' ? (
            <IconMoon size={16} stroke={1.75} />
          ) : (
            <IconSun size={16} stroke={1.75} />
          )}
        </Button>
      </div>
    </header>
  )
}
