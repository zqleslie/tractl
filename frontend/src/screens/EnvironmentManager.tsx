import { useEffect, useMemo, useState } from 'react'
import {
  IconCheck,
  IconCircle,
  IconPlus,
  IconTrash,
} from '@tabler/icons-react'
import { Badge, Button } from '@/components/primitives'
import { cn } from '@/lib/cn'
import {
  isDuplicateEnvironmentName,
  useEnvironmentStore,
} from '@/stores/environmentStore'

function sortedVariables(variables: Record<string, string>) {
  return Object.entries(variables).sort(([left], [right]) =>
    left.localeCompare(right),
  )
}

export function EnvironmentManager() {
  const environments = useEnvironmentStore((state) => state.environments)
  const activeEnvironmentId = useEnvironmentStore(
    (state) => state.activeEnvironmentId,
  )
  const createEnvironment = useEnvironmentStore((state) => state.createEnvironment)
  const renameEnvironment = useEnvironmentStore((state) => state.renameEnvironment)
  const deleteEnvironment = useEnvironmentStore((state) => state.deleteEnvironment)
  const setActiveEnvironment = useEnvironmentStore(
    (state) => state.setActiveEnvironment,
  )
  const setVariable = useEnvironmentStore((state) => state.setVariable)
  const deleteVariable = useEnvironmentStore((state) => state.deleteVariable)

  const [selectedId, setSelectedId] = useState<string | null>(activeEnvironmentId)
  const [newEnvironmentName, setNewEnvironmentName] = useState('')
  const [newVariableKey, setNewVariableKey] = useState('')
  const [newVariableValue, setNewVariableValue] = useState('')

  const selectedEnvironment = useMemo(
    () =>
      environments.find((environment) => environment.id === selectedId) ??
      environments[0] ??
      null,
    [environments, selectedId],
  )

  useEffect(() => {
    if (selectedEnvironment || environments.length === 0) return
    setSelectedId(environments[0]?.id ?? null)
  }, [environments, selectedEnvironment])

  useEffect(() => {
    if (activeEnvironmentId) {
      setSelectedId(activeEnvironmentId)
    }
  }, [activeEnvironmentId])

  const duplicateNewName = isDuplicateEnvironmentName(
    environments,
    newEnvironmentName,
  )
  const canCreateEnvironment =
    newEnvironmentName.trim().length > 0 && !duplicateNewName

  const createNamedEnvironment = () => {
    if (!canCreateEnvironment) return
    const id = createEnvironment(newEnvironmentName)
    setSelectedId(id)
    setActiveEnvironment(id)
    setNewEnvironmentName('')
  }

  const duplicateVariableKey =
    !!selectedEnvironment &&
    !!newVariableKey.trim() &&
    Object.prototype.hasOwnProperty.call(
      selectedEnvironment.variables,
      newVariableKey.trim(),
    )

  const addVariable = () => {
    if (!selectedEnvironment || !newVariableKey.trim() || duplicateVariableKey) {
      return
    }
    setVariable(selectedEnvironment.id, newVariableKey, newVariableValue)
    setNewVariableKey('')
    setNewVariableValue('')
  }

  return (
    <div
      className="mx-auto flex h-full max-w-5xl flex-col px-8 py-7"
      data-testid="screen-environments"
    >
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-xl font-medium text-text">Environment Manager</h1>
          <p className="mt-1 text-ui-sm text-text-muted">
            Flat string variables for request URL, headers, query params, and body.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <input
            aria-label="New environment name"
            placeholder="Development"
            className="h-8 rounded-ui border-[0.5px] border-border bg-surface px-2.5 text-ui-sm text-text outline-none focus:border-primary"
            value={newEnvironmentName}
            onChange={(event) => setNewEnvironmentName(event.currentTarget.value)}
          />
          <Button
            variant="primary"
            className="h-8"
            data-testid="environment-create"
            disabled={!canCreateEnvironment}
            onClick={createNamedEnvironment}
          >
            <IconPlus size={14} stroke={1.75} />
            New environment
          </Button>
        </div>
      </div>
      {duplicateNewName ? (
        <p className="mt-2 text-ui-xs text-danger-fg">
          An environment with this name already exists.
        </p>
      ) : null}

      <div className="mt-6 grid min-h-0 flex-1 grid-cols-[260px_minmax(0,1fr)] gap-4">
        <section className="min-h-0 overflow-auto rounded-ui border-[0.5px] border-border bg-surface">
          <div className="border-b-[0.5px] border-border px-3 py-2 text-ui-xs font-medium uppercase tracking-wider text-text-muted">
            Configured
          </div>
          {environments.length === 0 ? (
            <p className="p-4 text-ui-sm text-text-muted">
              No environments yet. Create one to enable variable resolution.
            </p>
          ) : (
            <div className="p-1.5">
              {environments.map((environment) => {
                const active = environment.id === activeEnvironmentId
                const selected = environment.id === selectedEnvironment?.id
                return (
                  <button
                    key={environment.id}
                    type="button"
                    className={cn(
                      'flex w-full items-center gap-2 rounded-ui px-2 py-2 text-left text-ui-sm text-text hover:bg-surface-elevated',
                      selected && 'bg-surface-elevated',
                    )}
                    data-testid={`environment-row-${environment.id}`}
                    onClick={() => setSelectedId(environment.id)}
                  >
                    <span
                      className={cn(
                        'h-2 w-2 rounded-full',
                        active ? 'bg-success-fg' : 'bg-text-muted',
                      )}
                    />
                    <span className="min-w-0 flex-1 truncate">{environment.name}</span>
                    <span className="text-ui-xs text-text-muted">
                      {Object.keys(environment.variables).length}
                    </span>
                    {active ? (
                      <IconCheck
                        size={14}
                        stroke={2}
                        className="shrink-0 text-success-fg"
                      />
                    ) : null}
                  </button>
                )
              })}
            </div>
          )}
        </section>

        <section className="min-h-0 overflow-auto rounded-ui border-[0.5px] border-border bg-surface">
          {selectedEnvironment ? (
            <div className="flex min-h-full flex-col">
              <div className="flex items-center gap-2 border-b-[0.5px] border-border p-3">
                <IconCircle
                  size={12}
                  fill={
                    selectedEnvironment.id === activeEnvironmentId
                      ? 'currentColor'
                      : 'none'
                  }
                  className={cn(
                    selectedEnvironment.id === activeEnvironmentId
                      ? 'text-success-fg'
                      : 'text-text-muted',
                  )}
                />
                <input
                  aria-label="Environment name"
                  className="min-w-0 flex-1 rounded-ui border-[0.5px] border-transparent bg-transparent px-2 py-1 text-ui-sm font-medium text-text outline-none hover:border-border focus:border-primary focus:bg-surface-elevated"
                  value={selectedEnvironment.name}
                  onChange={(event) =>
                    renameEnvironment(
                      selectedEnvironment.id,
                      event.currentTarget.value,
                    )
                  }
                />
                <Badge tone="neutral">
                  {Object.keys(selectedEnvironment.variables).length} variables
                </Badge>
                <Button
                  variant="secondary"
                  className="h-8"
                  disabled={selectedEnvironment.id === activeEnvironmentId}
                  onClick={() => setActiveEnvironment(selectedEnvironment.id)}
                >
                  Set active
                </Button>
                <Button
                  variant="danger"
                  className="h-8 px-2"
                  aria-label="Delete environment"
                  onClick={() => deleteEnvironment(selectedEnvironment.id)}
                >
                  <IconTrash size={14} stroke={1.75} />
                </Button>
              </div>

              <div className="p-3">
                <div className="grid grid-cols-[minmax(0,1fr)_minmax(0,1.4fr)_32px] gap-2 border-b-[0.5px] border-border pb-2 text-[10px] font-medium uppercase tracking-wider text-text-muted">
                  <span>Key</span>
                  <span>Value</span>
                  <span />
                </div>

                <div className="divide-y-[0.5px] divide-border">
                  {sortedVariables(selectedEnvironment.variables).map(
                    ([key, value]) => (
                      <div
                        key={key}
                        className="grid grid-cols-[minmax(0,1fr)_minmax(0,1.4fr)_32px] gap-2 py-2"
                      >
                        <input
                          aria-label={`${key} key`}
                          className="rounded-ui border-[0.5px] border-border bg-surface px-2 py-1 font-mono text-ui-xs text-text outline-none focus:border-primary"
                          value={key}
                          onChange={(event) => {
                            const nextKey = event.currentTarget.value
                            deleteVariable(selectedEnvironment.id, key)
                            setVariable(selectedEnvironment.id, nextKey, value)
                          }}
                        />
                        <input
                          aria-label={`${key} value`}
                          className="rounded-ui border-[0.5px] border-border bg-surface px-2 py-1 font-mono text-ui-xs text-text outline-none focus:border-primary"
                          value={value}
                          onChange={(event) =>
                            setVariable(
                              selectedEnvironment.id,
                              key,
                              event.currentTarget.value,
                            )
                          }
                        />
                        <Button
                          variant="ghost"
                          className="h-7 w-7 p-0 text-text-muted hover:text-danger-fg"
                          aria-label={`Delete ${key}`}
                          onClick={() => deleteVariable(selectedEnvironment.id, key)}
                        >
                          <IconTrash size={13} stroke={1.75} />
                        </Button>
                      </div>
                    ),
                  )}
                </div>

                <div className="mt-3 grid grid-cols-[minmax(0,1fr)_minmax(0,1.4fr)_auto] gap-2">
                  <input
                    aria-label="New variable key"
                    placeholder="baseUrl"
                    className="rounded-ui border-[0.5px] border-border bg-surface px-2 py-1 font-mono text-ui-xs text-text outline-none focus:border-primary"
                    value={newVariableKey}
                    onChange={(event) => setNewVariableKey(event.currentTarget.value)}
                  />
                  <input
                    aria-label="New variable value"
                    placeholder="https://api.example.com"
                    className="rounded-ui border-[0.5px] border-border bg-surface px-2 py-1 font-mono text-ui-xs text-text outline-none focus:border-primary"
                    value={newVariableValue}
                    onChange={(event) =>
                      setNewVariableValue(event.currentTarget.value)
                    }
                  />
                  <Button
                    variant="secondary"
                    className="h-7"
                    disabled={!newVariableKey.trim() || duplicateVariableKey}
                    onClick={addVariable}
                  >
                    Add variable
                  </Button>
                </div>
                {duplicateVariableKey ? (
                  <p className="mt-2 text-ui-xs text-danger-fg">
                    This variable key already exists in the environment.
                  </p>
                ) : null}
              </div>
            </div>
          ) : (
            <div className="flex h-full items-center justify-center p-8 text-ui-sm text-text-muted">
              Select or create an environment to edit variables.
            </div>
          )}
        </section>
      </div>
    </div>
  )
}
