const RESERVED_AUTHORED_KEY_PREFIXES = ['_traCtl.', 'traCtl.'] as const

function isProhibitedAuthoredKey(key: string): boolean {
  if (key === '_ulid') return true
  return RESERVED_AUTHORED_KEY_PREFIXES.some((prefix) => key.startsWith(prefix))
}

/** Remove system identity fields before emitting or persisting an authored document shape. */
export function stripAuthoringSystemFields<T>(value: T): T {
  if (Array.isArray(value)) {
    return value.map((entry) => stripAuthoringSystemFields(entry)) as T
  }
  if (typeof value !== 'object' || value === null) {
    return value
  }

  const record = value as Record<string, unknown>
  const sanitized = Object.fromEntries(
    Object.entries(record)
      .filter(([key]) => !isProhibitedAuthoredKey(key))
      .map(([key, entry]) => [key, stripAuthoringSystemFields(entry)]),
  )
  return sanitized as T
}
