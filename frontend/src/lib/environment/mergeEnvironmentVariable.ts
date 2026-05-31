export function makeEnvironmentId(name: string, existingIds: Set<string>): string {
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

export function normalizeVariableKey(key: string): string {
  return key.trim()
}
