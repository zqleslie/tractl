import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export type EnvironmentVariables = Record<string, string>

export type TractlEnvironment = {
  id: string
  name: string
  variables: EnvironmentVariables
}

type EnvironmentState = {
  environments: TractlEnvironment[]
  activeEnvironmentId: string | null
  createEnvironment: (name: string) => string
  renameEnvironment: (id: string, name: string) => void
  deleteEnvironment: (id: string) => void
  setActiveEnvironment: (id: string | null) => void
  setVariable: (environmentId: string, key: string, value: string) => void
  deleteVariable: (environmentId: string, key: string) => void
}

function environmentIdFromName(name: string, existingIds: Set<string>): string {
  const base =
    name
      .trim()
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-+|-+$/g, '') || 'environment'

  let candidate = `env-${base}`
  let index = 2
  while (existingIds.has(candidate)) {
    candidate = `env-${base}-${index}`
    index += 1
  }
  return candidate
}

function cleanedName(name: string): string {
  return name.trim() || 'Untitled environment'
}

function namesMatch(left: string, right: string): boolean {
  return cleanedName(left).toLowerCase() === cleanedName(right).toLowerCase()
}

export function isDuplicateEnvironmentName(
  environments: TractlEnvironment[],
  name: string,
  exceptId?: string,
): boolean {
  return environments.some(
    (environment) =>
      environment.id !== exceptId && namesMatch(environment.name, name),
  )
}

function cleanedVariableKey(key: string): string {
  return key.trim()
}

export const useEnvironmentStore = create<EnvironmentState>()(
  persist(
    (set, get) => ({
      environments: [],
      activeEnvironmentId: null,
      createEnvironment: (name) => {
        const current = get().environments
        const existing = current.find((environment) =>
          namesMatch(environment.name, name),
        )
        if (existing) return existing.id

        const id = environmentIdFromName(
          name,
          new Set(current.map((environment) => environment.id)),
        )
        const environment: TractlEnvironment = {
          id,
          name: cleanedName(name),
          variables: {},
        }

        set({
          environments: [...current, environment],
          activeEnvironmentId: get().activeEnvironmentId ?? id,
        })

        return id
      },
      renameEnvironment: (id, name) =>
        set((state) => {
          const duplicate = state.environments.some(
            (environment) =>
              environment.id !== id && namesMatch(environment.name, name),
          )
          if (duplicate) return state

          return {
            environments: state.environments.map((environment) =>
              environment.id === id
                ? { ...environment, name: cleanedName(name) }
                : environment,
            ),
          }
        }),
      deleteEnvironment: (id) =>
        set((state) => {
          const environments = state.environments.filter(
            (environment) => environment.id !== id,
          )
          const activeEnvironmentId =
            state.activeEnvironmentId === id
              ? (environments[0]?.id ?? null)
              : state.activeEnvironmentId

          return { environments, activeEnvironmentId }
        }),
      setActiveEnvironment: (id) =>
        set((state) => ({
          activeEnvironmentId:
            id && state.environments.some((environment) => environment.id === id)
              ? id
              : null,
        })),
      setVariable: (environmentId, key, value) =>
        set((state) => {
          const variableKey = cleanedVariableKey(key)
          if (!variableKey) return state

          return {
            environments: state.environments.map((environment) =>
              environment.id === environmentId
                ? {
                    ...environment,
                    variables: {
                      ...environment.variables,
                      [variableKey]: String(value),
                    },
                  }
                : environment,
            ),
          }
        }),
      deleteVariable: (environmentId, key) =>
        set((state) => ({
          environments: state.environments.map((environment) => {
            if (environment.id !== environmentId) return environment

            const variables = { ...environment.variables }
            delete variables[key]
            return { ...environment, variables }
          }),
        })),
    }),
    {
      name: 'tractl-environments',
      partialize: (state) => ({
        environments: state.environments,
        activeEnvironmentId: state.activeEnvironmentId,
      }),
    },
  ),
)
