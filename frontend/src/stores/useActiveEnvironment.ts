import { useEnvironmentStore } from '@/stores/environmentStore'

export function useActiveEnvironment() {
  const environments = useEnvironmentStore((state) => state.environments)
  const activeEnvironmentId = useEnvironmentStore((state) => state.activeEnvironmentId)
  const setActiveEnvironment = useEnvironmentStore((state) => state.setActiveEnvironment)
  const createEnvironment = useEnvironmentStore((state) => state.createEnvironment)

  const activeEnvironment = environments.find((env) => env.id === activeEnvironmentId) ?? null

  return { environments, activeEnvironmentId, activeEnvironment, setActiveEnvironment, createEnvironment }
}
