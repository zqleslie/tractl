import { createEmptyRequestFormState } from '@/components/request-editor/createEmptyRequestFormState'
import type { RequestState } from '@/components/request-editor/requestState'
import type {
  AuthType,
  BodyEncoding,
  BodyEncodingId,
  RequestAuthDraft,
  RequestBodyDraft,
  RequestFormState,
  RequestSettingsDraft,
  RetryStrategy,
} from '@/components/request-editor/types'
import type { HttpMethod } from '@/components/primitives'

const encodingToUi: Record<BodyEncodingId, BodyEncoding> = {
  json: 'JSON',
  form: 'Form data',
  multipart: 'Multipart',
  raw: 'Raw',
  binary: 'Binary',
  none: 'None',
}

const encodingFromUi: Record<BodyEncoding, BodyEncodingId> = {
  JSON: 'json',
  'Form data': 'form',
  Multipart: 'multipart',
  Raw: 'raw',
  Binary: 'binary',
  None: 'none',
}

const authToUi: Record<RequestState['auth']['type'], AuthType> = {
  none: 'None',
  bearer: 'Bearer token',
  basic: 'Basic auth',
  apikey: 'API key',
}

const authFromUi: Record<AuthType, RequestState['auth']['type']> = {
  None: 'none',
  'Bearer token': 'bearer',
  'Basic auth': 'basic',
  'API key': 'apikey',
}

const retryToUi: Record<
  NonNullable<RequestState['settings']['retry']>['strategy'],
  RetryStrategy
> = {
  fixed: 'Fixed',
  linear: 'Linear',
  exponential: 'Exponential',
}

const retryFromUi: Record<
  RetryStrategy,
  NonNullable<RequestState['settings']['retry']>['strategy'] | null
> = {
  None: null,
  Fixed: 'fixed',
  Linear: 'linear',
  Exponential: 'exponential',
}

function timeoutMsToDraft(timeoutMs: number): Pick<RequestSettingsDraft, 'timeoutValue' | 'timeoutUnit'> {
  if (timeoutMs % 60_000 === 0 && timeoutMs >= 60_000) {
    return { timeoutValue: String(timeoutMs / 60_000), timeoutUnit: 'minutes' }
  }
  if (timeoutMs % 1000 === 0 && timeoutMs >= 1000) {
    return { timeoutValue: String(timeoutMs / 1000), timeoutUnit: 'seconds' }
  }
  return { timeoutValue: String(timeoutMs), timeoutUnit: 'milliseconds' }
}

function timeoutDraftToMs(settings: RequestSettingsDraft): number {
  const value = Number.parseFloat(settings.timeoutValue)
  if (!Number.isFinite(value) || value <= 0) return 30_000
  switch (settings.timeoutUnit) {
    case 'minutes':
      return Math.round(value * 60_000)
    case 'seconds':
      return Math.round(value * 1000)
    default:
      return Math.round(value)
  }
}

function bodyStateToDraft(body: RequestState['body']): RequestBodyDraft {
  return {
    encoding: encodingToUi[body.encoding],
    value: body.content,
    formRows: body.formRows ?? [],
    rawContentType: body.rawContentType ?? 'text/plain',
    binaryFile: body.binaryFile ?? '',
  }
}

function bodyDraftToState(body: RequestBodyDraft): RequestState['body'] {
  return {
    encoding: encodingFromUi[body.encoding],
    content: body.value,
    formRows: body.formRows,
    rawContentType: body.rawContentType,
    binaryFile: body.binaryFile,
  }
}

function authStateToDraft(auth: RequestState['auth']): RequestAuthDraft {
  return {
    type: authToUi[auth.type],
    tokenRef: '',
    token: auth.token ?? '',
    username: auth.username ?? '',
    password: auth.password ?? '',
    keyName: auth.keyName ?? '',
    keyValue: auth.keyValue ?? '',
    placement: auth.placement ?? 'header',
  }
}

function authDraftToState(auth: RequestAuthDraft): RequestState['auth'] {
  return {
    type: authFromUi[auth.type],
    token: auth.token || undefined,
    username: auth.username || undefined,
    password: auth.password || undefined,
    keyName: auth.keyName || undefined,
    keyValue: auth.keyValue || undefined,
    placement: auth.placement,
  }
}

export function requestStateToDraft(state: RequestState): RequestFormState {
  const retryStrategy = state.settings.retry
    ? retryToUi[state.settings.retry.strategy]
    : 'None'

  return {
    params: state.params,
    headers: state.headers,
    body: bodyStateToDraft(state.body),
    auth: authStateToDraft(state.auth),
    scripts: { pre: state.preScript, post: state.postScript },
    assertions: state.assertions.map((row) => ({
      id: row.id,
      kind: row.kind,
      operator: row.op,
      expected: row.expected,
      severity: row.severity,
    })),
    extracts: state.extracts.map((row) => ({
      id: row.id,
      source: row.source,
      path: row.path,
      variable: row.variableName,
      scope: row.scope,
    })),
    settings: {
      ...timeoutMsToDraft(state.settings.timeoutMs),
      retry: retryStrategy,
      retryConfig: state.settings.retry,
      failurePolicy: state.settings.failurePolicy,
    },
  }
}

export function draftToRequestState(
  id: string,
  name: string,
  method: HttpMethod,
  url: string,
  draft: RequestFormState,
): RequestState {
  const retryStrategy = draft.settings.retry
  const strategy = retryFromUi[retryStrategy]

  return {
    id,
    name,
    method,
    url,
    params: draft.params,
    headers: draft.headers,
    body: bodyDraftToState(draft.body),
    auth: authDraftToState(draft.auth),
    preScript: draft.scripts.pre,
    postScript: draft.scripts.post,
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
      timeoutMs: timeoutDraftToMs(draft.settings),
      retry:
        strategy && draft.settings.retryConfig
          ? {
              strategy,
              maxAttempts: draft.settings.retryConfig.maxAttempts,
              delayMs: draft.settings.retryConfig.delayMs,
              backoffFactor: draft.settings.retryConfig.backoffFactor,
            }
          : null,
      failurePolicy: draft.settings.failurePolicy,
    },
  }
}

export function createWorkspaceRequestState(
  id: string,
  method: HttpMethod = 'GET',
): RequestState {
  return draftToRequestState(
    id,
    'Untitled request',
    method,
    '',
    createEmptyRequestFormState(),
  )
}
