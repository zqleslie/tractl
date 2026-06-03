import { useRef, useState } from 'react'
import { useClickOutside } from '@/hooks/useClickOutside'
import { IconCommand, IconMoon, IconSun } from '@tabler/icons-react'
import { Button, EnvPill } from '@/components/primitives'
import { isDuplicateEnvironmentName } from '@/stores/environmentStore'
import { useActiveEnvironment } from '@/stores/useActiveEnvironment'
import { useUiStore } from '@/stores/uiStore'

export function Topbar() {
  const theme = useUiStore((state) => state.theme)
  const setTheme = useUiStore((state) => state.setTheme)
  const setActiveScreen = useUiStore((state) => state.setActiveScreen)
  const setCommandPaletteOpen = useUiStore((state) => state.setCommandPaletteOpen)
  const { environments, activeEnvironment, setActiveEnvironment, createEnvironment } = useActiveEnvironment()
  const [envMenuOpen, setEnvMenuOpen] = useState(false)
  const [newEnvironmentName, setNewEnvironmentName] = useState('')
  const envMenuRef = useRef<HTMLDivElement | null>(null)

  const duplicateNewName = isDuplicateEnvironmentName(environments, newEnvironmentName)
  const canCreateEnvironment =
    newEnvironmentName.trim().length > 0 && !duplicateNewName

  useClickOutside(envMenuRef, () => setEnvMenuOpen(false), envMenuOpen)

  const cycleTheme = () => {
    setTheme(theme === 'light' ? 'dark' : theme === 'dark' ? 'system' : 'light')
  }

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
      className="grid h-[42px] shrink-0 grid-cols-[1fr_auto_1fr] items-center gap-3 border-b-[0.5px] border-border bg-surface px-3"
      data-testid="topbar"
    >
      <div aria-hidden="true" />
      <button
        type="button"
        className="flex h-7 w-[min(420px,42vw)] min-w-[240px] items-center gap-2 rounded-ui border-[0.5px] border-border bg-surface-elevated px-2 text-left text-ui-xs text-text-muted hover:border-primary hover:text-text"
        data-testid="command-search"
        onClick={() => setCommandPaletteOpen(true)}
      >
        <IconCommand size={14} stroke={1.75} />
        <span className="min-w-0 flex-1 truncate">Search and commands...</span>
        <kbd className="font-mono text-[10px]">Cmd K</kbd>
      </button>

      <div ref={envMenuRef} className="relative ml-auto flex items-center gap-2">
        <EnvPill
          environment={activeEnvironment}
          aria-expanded={envMenuOpen}
          aria-haspopup="menu"
          onClick={() => setEnvMenuOpen((open) => !open)}
        />
        {envMenuOpen ? (
          <div
            className="absolute right-0 top-9 z-30 w-64 rounded-ui border-[0.5px] border-border bg-surface p-2"
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
                {environments.map((environment) => (
                  <button
                    key={environment.id}
                    type="button"
                    className="flex w-full items-center gap-2 rounded-ui px-2 py-1.5 text-left text-ui-sm text-text-muted hover:bg-surface-elevated hover:text-text"
                    role="menuitem"
                    onClick={() => {
                      setActiveEnvironment(environment.id)
                      setEnvMenuOpen(false)
                    }}
                  >
                    <span className="min-w-0 flex-1 truncate">
                      {environment.name}
                    </span>
                    <span className="font-mono text-[10px]">
                      {Object.keys(environment.variables).length}
                    </span>
                  </button>
                ))}
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
        <Button
          variant="ghost"
          className="h-7 px-2"
          aria-label={`Theme: ${theme}`}
          data-testid="theme-toggle"
          onClick={cycleTheme}
        >
          {theme === 'dark' ? (
            <IconSun size={16} stroke={1.75} />
          ) : (
            <IconMoon size={16} stroke={1.75} />
          )}
        </Button>
      </div>
    </header>
  )
}
