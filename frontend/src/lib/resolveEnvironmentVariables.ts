import type { TraCtlSpecDocument } from '@/components/request-editor/tractlSpecDocument'
import type { EnvironmentVariables } from '@/stores/environmentStore'

export const ENV_VARIABLE_MISSING_CODE = 'TRACTL_ENV_VARIABLE_MISSING'

export type EnvironmentResolutionResult<T> = {
  value: T
  missing: string[]
}

const PLACEHOLDER_PATTERN = /\{\{\s*([A-Za-z_][A-Za-z0-9_.-]*)\s*\}\}/g
const ENCODED_PLACEHOLDER_PATTERN =
  /%7B%7B\s*([A-Za-z_][A-Za-z0-9_.-]*)\s*%7D%7D/gi

function uniqueSorted(values: Iterable<string>): string[] {
  return Array.from(new Set(values)).sort((left, right) => left.localeCompare(right))
}

function resolveWithPattern(
  input: string,
  variables: EnvironmentVariables,
  pattern: RegExp,
  encodeReplacement: boolean,
): EnvironmentResolutionResult<string> {
  const missing = new Set<string>()
  const value = input.replace(pattern, (placeholder: string, name: string) => {
    if (!Object.prototype.hasOwnProperty.call(variables, name)) {
      missing.add(name)
      return placeholder
    }

    const resolved = variables[name] ?? ''
    return encodeReplacement ? encodeURIComponent(resolved) : resolved
  })

  return { value, missing: uniqueSorted(missing) }
}

function mergeMissing(...groups: string[][]): string[] {
  return uniqueSorted(groups.flat())
}

export function resolveEnvironmentString(
  input: string,
  variables: EnvironmentVariables,
): EnvironmentResolutionResult<string> {
  return resolveWithPattern(input, variables, PLACEHOLDER_PATTERN, false)
}

export function resolveEnvironmentUrl(
  input: string,
  variables: EnvironmentVariables,
): EnvironmentResolutionResult<string> {
  const plain = resolveEnvironmentString(input, variables)
  const encoded = resolveWithPattern(
    plain.value,
    variables,
    ENCODED_PLACEHOLDER_PATTERN,
    true,
  )

  return {
    value: encoded.value,
    missing: mergeMissing(plain.missing, encoded.missing),
  }
}

function resolveUnknownStrings(
  value: unknown,
  variables: EnvironmentVariables,
): EnvironmentResolutionResult<unknown> {
  if (typeof value === 'string') {
    return resolveEnvironmentString(value, variables)
  }

  if (Array.isArray(value)) {
    const missing: string[] = []
    const resolved = value.map((item) => {
      const result = resolveUnknownStrings(item, variables)
      missing.push(...result.missing)
      return result.value
    })

    return { value: resolved, missing: uniqueSorted(missing) }
  }

  if (value && typeof value === 'object') {
    const missing: string[] = []
    const resolved: Record<string, unknown> = {}

    for (const [key, child] of Object.entries(value)) {
      const keyResult = resolveEnvironmentString(key, variables)
      const valueResult = resolveUnknownStrings(child, variables)
      missing.push(...keyResult.missing, ...valueResult.missing)
      resolved[keyResult.value] = valueResult.value
    }

    return { value: resolved, missing: uniqueSorted(missing) }
  }

  return { value, missing: [] }
}

function resolveHeaders(
  headers: Record<string, string> | undefined,
  variables: EnvironmentVariables,
): EnvironmentResolutionResult<Record<string, string> | undefined> {
  if (!headers) return { value: undefined, missing: [] }

  const missing: string[] = []
  const resolved: Record<string, string> = {}

  for (const [key, value] of Object.entries(headers)) {
    const keyResult = resolveEnvironmentString(key, variables)
    const valueResult = resolveEnvironmentString(value, variables)
    missing.push(...keyResult.missing, ...valueResult.missing)
    resolved[keyResult.value] = valueResult.value
  }

  return { value: resolved, missing: uniqueSorted(missing) }
}

export function resolveTraCtlSpecEnvironmentVariables(
  document: TraCtlSpecDocument,
  variables: EnvironmentVariables,
): EnvironmentResolutionResult<TraCtlSpecDocument> {
  const missing: string[] = []
  const workflows = document.workflows.map((workflow) => ({
    ...workflow,
    steps: workflow.steps.map((step) => {
      const target = resolveEnvironmentUrl(step.request.target, variables)
      const headers = resolveHeaders(step.request.headers, variables)
      const bodyContent = step.request.body
        ? resolveUnknownStrings(step.request.body.content, variables)
        : undefined

      missing.push(
        ...target.missing,
        ...headers.missing,
        ...(bodyContent?.missing ?? []),
      )

      return {
        ...step,
        request: {
          ...step.request,
          target: target.value,
          ...(headers.value ? { headers: headers.value } : {}),
          ...(step.request.body
            ? {
                body: {
                  ...step.request.body,
                  content: bodyContent?.value,
                },
              }
            : {}),
        },
      }
    }),
  }))

  return {
    value: { ...document, workflows },
    missing: uniqueSorted(missing),
  }
}

export function formatMissingEnvironmentVariables(names: string[]): string {
  const label = names.length === 1 ? 'variable' : 'variables'
  return `${ENV_VARIABLE_MISSING_CODE}: Missing environment ${label}: ${names.join(', ')}`
}
