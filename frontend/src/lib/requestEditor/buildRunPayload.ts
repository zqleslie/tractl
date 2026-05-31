import type { RunRequestPayload } from '@/api/requests'
import type { RequestDraft } from '@/components/request-editor/types'
import type { HttpMethod } from '@/components/primitives'

function enabledHeaders(
  headers: RequestDraft['headers'],
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
  draft: RequestDraft
  environmentId: string | null
}): RunRequestPayload {
  const { draft, method, url, environmentId } = input
  const authType = draft.auth.type

  return {
    method,
    url,
    headers: enabledHeaders(draft.headers),
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
    environmentId,
  }
}
