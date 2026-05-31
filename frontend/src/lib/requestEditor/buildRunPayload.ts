import type { RequestFormState } from '@/components/request-editor/types'
import type { HttpMethod } from '@/components/primitives'
import type { RequestDef } from '@/types/requestDef'

function enabledHeaders(
  headers: RequestFormState['headers'],
): Record<string, string> {
  const result: Record<string, string> = {}
  for (const row of headers) {
    if (!row.enabled || !row.key.trim()) continue
    result[row.key] = row.value
  }
  return result
}

export function buildRunRequestPayload(input: {
  method: HttpMethod
  url: string
  draft: RequestFormState
  environmentId: string | null
}): RequestDef {
  const { draft, method, url } = input
  const authType = draft.auth.type

  return {
    method,
    url,
    id: crypto.randomUUID(),
    name: url || 'Untitled request',
    headers: Object.entries(enabledHeaders(draft.headers)).map(([key, value]) => ({
      enabled: true,
      key,
      value,
    })),
    body:
      draft.body.encoding === 'None'
        ? undefined
        : {
            encoding: draft.body.encoding,
            content: draft.body.value,
          },
    auth:
      authType === 'None'
        ? undefined
        : {
            type: authType,
            token: draft.auth.token || draft.auth.tokenRef || undefined,
          },
    assertions: draft.assertions.map((row) => ({
      id: row.id,
      kind: row.kind,
      op: row.operator,
      expected: row.expected,
      severity: row.severity,
    })),
    extracts: draft.extracts.map((row) => ({
      id: row.id,
      source: row.source,
      path: row.path,
      variableName: row.variable,
      scope: row.scope,
    })),
    settings: {
      timeoutMs:
        Number.parseFloat(draft.settings.timeoutValue) > 0
          ? Number.parseFloat(draft.settings.timeoutValue) * 1000
          : 30_000,
      failurePolicy: draft.settings.failurePolicy,
    },
  } as unknown as RequestDef
}
