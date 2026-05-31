import type { RequestFormState } from '@/components/request-editor/types'
import { normalizeRequestTargetUrl } from '@/components/request-editor/normalizeRequestUrl'
import type {
  TraCtlAssertion,
  TraCtlExtract,
  TraCtlSpecDocument,
  TraCtlStep,
} from '@/components/request-editor/tractlSpecDocument'
import type { HttpMethod } from '@/components/primitives'

const WORKFLOW_ID = 'request-flow'
const STEP_ID = 'request-step'

function enabledRows(
  rows: RequestFormState['headers'],
): Record<string, string> | undefined {
  const headers: Record<string, string> = {}
  for (const row of rows) {
    if (!row.enabled) continue
    const key = row.key.trim()
    if (!key) continue
    headers[key] = row.value
  }
  return Object.keys(headers).length > 0 ? headers : undefined
}

function queryFromParams(rows: RequestFormState['params']): string {
  const parts: string[] = []
  for (const row of rows) {
    if (!row.enabled) continue
    const key = row.key.trim()
    if (!key) continue
    parts.push(
      `${encodeURIComponent(key)}=${encodeURIComponent(row.value)}`,
    )
  }
  return parts.join('&')
}

function resolveTarget(_method: HttpMethod, url: string, draft: RequestFormState): string {
  const trimmed = normalizeRequestTargetUrl(url)
  if (!trimmed) return ''
  const query = queryFromParams(draft.params)
  if (!query) return trimmed
  const separator = trimmed.includes('?') ? '&' : '?'
  return `${trimmed}${separator}${query}`
}

function parseBodyContent(encoding: RequestFormState['body']['encoding'], value: string) {
  const trimmed = value.trim()
  if (!trimmed || encoding === 'None') return undefined

  if (encoding === 'JSON') {
    try {
      return JSON.parse(trimmed) as unknown
    } catch {
      return trimmed
    }
  }

  return trimmed
}

function bodyDescriptor(draft: RequestFormState) {
  const content = parseBodyContent(draft.body.encoding, draft.body.value)
  if (content === undefined) return undefined

  const encoding =
    draft.body.encoding === 'Form data'
      ? 'form'
      : draft.body.encoding === 'Raw'
        ? 'raw'
        : draft.body.encoding === 'JSON'
          ? 'json'
          : 'none'

  return { encoding, content }
}

function toAssertions(rows: RequestFormState['assertions']): TraCtlAssertion[] {
  return rows
    .filter((row) => row.kind.trim() && row.expected.trim())
    .map((row, index) => {
      const expected =
        row.kind === 'status' ? Number.parseInt(row.expected, 10) : row.expected

      return {
        id: row.id || `assertion-${index + 1}`,
        kind: row.kind,
        op: row.operator,
        expected: Number.isNaN(expected as number) ? row.expected : expected,
        severity: row.severity,
      }
    })
}

function toExtracts(rows: RequestFormState['extracts']): TraCtlExtract[] {
  return rows
    .filter((row) => row.variable.trim())
    .map((row, index) => ({
      id: row.id || `extract-${index + 1}`,
      source: row.source,
      path: row.path.trim() || undefined,
      as: row.variable.trim(),
      scope: row.scope,
    }))
}

function toTimeout(settings: RequestFormState['settings']): string | undefined {
  const value = settings.timeoutValue.trim()
  if (!value) return undefined

  const unit = settings.timeoutUnit
  if (unit === 'milliseconds') return `PT${value}ms`
  if (unit === 'minutes') return `PT${value}M`
  return `PT${value}S`
}

function toRetry(settings: RequestFormState['settings']): TraCtlStep['retry'] | undefined {
  if (settings.retry === 'None' || !settings.retryConfig) return undefined
  return {
    maxAttempts: settings.retryConfig.maxAttempts,
    backoff: settings.retryConfig.strategy,
    delay: `PT${Math.max(1, Math.round(settings.retryConfig.delayMs / 1000))}S`,
  }
}

function toHooks(draft: RequestFormState): TraCtlStep['hooks'] | undefined {
  const before = draft.scripts.pre.trim()
  const after = draft.scripts.post.trim()
  if (!before && !after) return undefined

  return {
    ...(before
      ? { beforeStep: { language: 'js', source: before } }
      : {}),
    ...(after ? { afterStep: { language: 'js', source: after } } : {}),
  }
}

export function requestFormStateToTraCtlSpec(
  method: HttpMethod,
  url: string,
  draft: RequestFormState,
  requestName = 'Untitled request',
): TraCtlSpecDocument {
  const target = resolveTarget(method, url, draft)
  const headers = enabledRows(draft.headers)
  const body = bodyDescriptor(draft)
  const assertions = toAssertions(draft.assertions)
  const extracts = toExtracts(draft.extracts)

  const step: TraCtlStep = {
    id: STEP_ID,
    kind: 'request',
    request: {
      protocol: 'http',
      target,
      operation: method,
      ...(headers ? { headers } : {}),
      ...(body ? { body } : {}),
    },
    ...(assertions.length > 0 ? { assertions } : {}),
    ...(extracts.length > 0 ? { extracts } : {}),
    ...(toHooks(draft) ? { hooks: toHooks(draft) } : {}),
    ...(toTimeout(draft.settings) ? { timeout: toTimeout(draft.settings) } : {}),
    ...(toRetry(draft.settings) ? { retry: toRetry(draft.settings) } : {}),
  }

  return {
    schemaVersion: 1,
    capabilities: ['protocol.http'],
    metadata: { name: requestName },
    workflows: [
      {
        id: WORKFLOW_ID,
        name: requestName,
        failurePolicy: draft.settings.failurePolicy,
        steps: [step],
      },
    ],
  }
}
