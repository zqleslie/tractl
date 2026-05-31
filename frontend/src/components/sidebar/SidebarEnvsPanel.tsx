import { IconCheck, IconPlus } from '@tabler/icons-react'
import { SidebarAction } from '@/components/sidebar/SidebarAction'
import { SidebarListItem } from '@/components/sidebar/SidebarListItem'
import { SidebarSection } from '@/components/sidebar/SidebarSection'
import { cn } from '@/lib/cn'
import { useEnvironmentStore } from '@/stores/environmentStore'
import { useUiStore } from '@/stores/uiStore'

export function SidebarEnvsPanel() {
  const environments = useEnvironmentStore((state) => state.environments)
  const activeEnvironmentId = useEnvironmentStore(
    (state) => state.activeEnvironmentId,
  )
  const setActiveEnvironment = useEnvironmentStore(
    (state) => state.setActiveEnvironment,
  )
  const setActiveScreen = useUiStore((state) => state.setActiveScreen)

  const activeEnvironment =
    environments.find((environment) => environment.id === activeEnvironmentId) ??
    null
  const activeVariables = Object.entries(activeEnvironment?.variables ?? {}).sort(
    ([left], [right]) => left.localeCompare(right),
  )

  return (
    <div data-testid="sidebar-panel-envs">
      <SidebarAction
        primary
        testId="sidebar-new-environment"
        onClick={() => setActiveScreen('environments')}
      >
        <IconPlus size={14} stroke={1.75} />
        New environment
      </SidebarAction>

      <SidebarSection>Configured</SidebarSection>
      {environments.length === 0 ? (
        <p className="px-2 py-2 text-ui-xs text-text-muted">
          No environments configured.
        </p>
      ) : (
        environments.map((env) => {
          const active = env.id === activeEnvironmentId
          return (
            <SidebarListItem
              key={env.id}
              testId={`sidebar-env-${env.id}`}
              active={active}
              onClick={() => {
                setActiveEnvironment(env.id)
                setActiveScreen('environments')
              }}
              icon={
                <span
                  className={cn(
                    'inline-block h-2 w-2 rounded-full',
                    active ? 'bg-success-fg' : 'bg-text-muted',
                  )}
                />
              }
              label={
                <span className="flex w-full items-center gap-1">
                  <span className="min-w-0 truncate">{env.name}</span>
                  <span className="ml-auto text-[10px] text-text-muted">
                    {Object.keys(env.variables).length}
                  </span>
                  {active ? (
                    <IconCheck
                      size={13}
                      stroke={2}
                      className="shrink-0 text-success-fg"
                    />
                  ) : null}
                </span>
              }
            />
          )
        })
      )}

      <SidebarSection>Active variables</SidebarSection>
      <div className="px-2">
        {activeVariables.length === 0 ? (
          <p className="py-1 text-ui-xs text-text-muted">
            {activeEnvironment ? 'No variables set.' : 'No active environment.'}
          </p>
        ) : (
          activeVariables.map(([key, value], index) => (
            <div
              key={key}
              className={cn(
                'flex justify-between gap-2 py-1 text-ui-xs',
                index < activeVariables.length - 1 &&
                  'border-b-[0.5px] border-border',
              )}
            >
              <span className="text-text-muted">{key}</span>
              <span className="truncate font-mono text-[10px] text-text">
                {value}
              </span>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
