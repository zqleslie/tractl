import type { RequestDef } from '@/platform/types'
import type { RequestFormState } from '@/components/request-editor/types'
import type { HttpMethod } from '@/components/primitives'
import { getEngineDefaults } from '@/stores/engineDefaultsStore'

// Maps form state to RequestDef (flat field mapping only).
// No spec construction. No ISO-8601 formatting. No JSON parsing.
// The engine handles normalization internally.
export function draftToRequestDef(
  method: HttpMethod,
  url: string,
  draft: RequestFormState,
  env?: Record<string, string>,
): RequestDef {
  return {
    method,
    url,
    headers: draft.headers
      .filter((r) => r.enabled && r.key.trim())
      .map((r) => ({ enabled: true, key: r.key, value: r.value })),
    body: draft.body.encoding === 'None'
      ? undefined
      : { encoding: draft.body.encoding.toLowerCase(), content: draft.body.value },
    auth: draft.auth.type === 'None'
      ? undefined
      : { type: draft.auth.type, token: draft.auth.token || undefined },
    assertions: draft.assertions
      .filter((r) => r.kind.trim() && r.expected.trim())
      .map((r) => ({ id: r.id, kind: r.kind, op: r.operator, expected: r.expected, severity: r.severity })),
    extracts: draft.extracts
      .filter((r) => r.variable.trim())
      .map((r) => ({ id: r.id, source: r.source, path: r.path, variableName: r.variable, scope: r.scope })),
    settings: {
      timeoutMs: Number.parseFloat(draft.settings.timeoutValue) * 1000
        || getEngineDefaults().timeoutMs,
      failurePolicy: draft.settings.failurePolicy,
      retry: draft.settings.retry === 'None' ? undefined : draft.settings.retryConfig
        ? {
            strategy: draft.settings.retryConfig.strategy,
            maxAttempts: draft.settings.retryConfig.maxAttempts,
            delayMs: draft.settings.retryConfig.delayMs,
            backoffFactor: draft.settings.retryConfig.backoffFactor,
          }
        : undefined,
    },
    env,
  }
}
