import type { TractlEnvironment } from '@/stores/environmentStore'

export function isDangerousEnvironment(environment: TractlEnvironment | null): boolean {
  if (!environment) return false
  return /\b(prod|production)\b/i.test(environment.name)
}
