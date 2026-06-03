/**
 * requestDefToDocument — spec preview utility (ADR-017 §3 amendment 2026-06-02 B).
 *
 * Converts a RequestDef to a spec-shaped JSON object for display in the code view.
 * This is a PRESENTATION utility only. It MUST NOT be used on the execution path.
 *
 * Execution: RequestDef → window.tractl.runRequest() → Go: RunRequestDef() → engine.
 * The Go mapper (internal/localapi/request_mapper.go) is the sole transformation
 * path for all delivery surfaces including WASM.
 */

import type {
  TraCtlAssertion,
  TraCtlExtract,
  TraCtlSpecDocument,
  TraCtlStep,
} from '@/components/request-editor/tractlSpecDocument'
import type { AuthDef, BodyDef, RequestDef, KVRow, AssertionDef, ExtractDef } from '@/types/requestDef'

// ─── Helpers ─────────────────────────────────────────────────────────────────

function enabledRows(rows: KVRow[] | undefined): KVRow[] {
  return (rows ?? []).filter((r) => r.enabled && r.key.trim() !== '')
}

function buildTarget(req: RequestDef): string {
  const base = req.url.trim()
  if (!base) {
    throw new Error('url required')
  }
  const active = enabledRows(req.params)
  if (active.length === 0) return base
  const qs = active.map((r) => `${encodeURIComponent(r.key)}=${encodeURIComponent(r.value)}`).join('&')
  const sep = base.includes('?') ? '&' : '?'
  return `${base}${sep}${qs}`
}

function buildHeaders(req: RequestDef, forceContentType?: string): Record<string, string> | undefined {
  const active = enabledRows(req.headers)
  const out: Record<string, string> = {}
  for (const row of active) {
    out[row.key.trim()] = row.value
  }
  injectAuthHeaders(req.auth, out)
  if (forceContentType && !Object.keys(out).some((k) => k.toLowerCase() === 'content-type')) {
    out['Content-Type'] = forceContentType
  }
  return Object.keys(out).length > 0 ? out : undefined
}

function hasHeader(headers: Record<string, string>, key: string): boolean {
  const lower = key.toLowerCase()
  return Object.keys(headers).some((k) => k.toLowerCase() === lower)
}

function injectAuthHeaders(auth: AuthDef | undefined, headers: Record<string, string>): void {
  if (!auth || auth.type === 'none') return

  if (auth.type === 'bearer') {
    const token = auth.token?.trim() ?? ''
    if (token !== '' && !hasHeader(headers, 'Authorization')) {
      headers['Authorization'] = `Bearer ${token}`
    }
    return
  }

  if (auth.type === 'basic') {
    const username = auth.username?.trim() ?? ''
    const password = auth.password?.trim() ?? ''
    if (username !== '' && password !== '' && !hasHeader(headers, 'Authorization')) {
      headers['Authorization'] = `Basic ${btoa(`${username}:${password}`)}`
    }
    return
  }

  if (auth.type === 'apikey') {
    if (auth.placement === 'query') {
      // TODO: support auth query parameter injection in target URL builder.
      return
    }
    const keyName = auth.keyName?.trim() ?? ''
    const keyValue = auth.keyValue?.trim() ?? ''
    if (keyName !== '' && keyValue !== '' && !hasHeader(headers, keyName)) {
      headers[keyName] = keyValue
    }
  }
}

/**
 * Converts timeoutMs to ISO 8601 duration. Minimum 1 second.
 * 30000 → "PT30S"
 */
function toIso8601Duration(ms: number): string {
  const secs = Math.max(1, Math.trunc(ms / 1000))
  return `PT${secs}S`
}

type BodyBuildResult = {
  body: TraCtlStep['request']['body'] | undefined
  forcePost: boolean
  forceContentType?: string
}

type BodyBuilder = (req: RequestDef) => BodyBuildResult
type EncodingHandler = (body: BodyDef) => BodyBuildResult
type ProtocolBodyFamily = 'http'

const protocolToBodyFamily: Partial<Record<NonNullable<RequestDef['protocol']>, ProtocolBodyFamily>> = {
  http: 'http',
  graphql: 'http',
  odata: 'http',
  soap: 'http',
}

const bodyBuildersByFamily: Record<ProtocolBodyFamily, BodyBuilder> = {
  http: buildHttpBody,
}

const httpEncodingHandlers: Record<string, EncodingHandler> = {
  graphql: buildGraphqlBody,
  json: buildJsonBody,
  form: buildFormBody,
  multipart: buildMultipartBody,
}

function normalizeProtocol(protocol: RequestDef['protocol']): string {
  const inputProtocol = protocol ?? 'http'
  // ADR-017: OData is plain HTTP, not a distinct runtime protocol.
  return inputProtocol === 'odata' ? 'http' : inputProtocol
}

function buildBody(req: RequestDef): BodyBuildResult {
  const family = protocolToBodyFamily[req.protocol ?? 'http'] ?? 'http'
  return bodyBuildersByFamily[family](req)
}

function buildHttpBody(req: RequestDef): BodyBuildResult {
  const body = req.body
  if (!body || body.encoding === 'none') return { body: undefined, forcePost: false }

  const enc = body.encoding.toLowerCase()
  const handler = httpEncodingHandlers[enc]
  if (handler) return handler(body)

  return {
    body: { encoding: enc, content: body.content },
    forcePost: false,
  }
}

function buildGraphqlBody(body: BodyDef): BodyBuildResult {
  // Parse content as JSON; enforce ADR-017 GraphQL envelope semantics.
  let parsed: Record<string, unknown>
  try {
    parsed = JSON.parse(body.content ?? '{}') as Record<string, unknown>
  } catch {
    throw new Error('graphql body: content is not valid JSON')
  }

  if (!Object.prototype.hasOwnProperty.call(parsed, 'query')) {
    throw new Error('graphql body: required key "query" is missing')
  }
  const query = parsed['query']
  if (typeof query !== 'string' || query.trim() === '') {
    throw new Error('graphql body: query must be a non-empty string')
  }

  const gqlBody: Record<string, unknown> = { ...parsed }
  gqlBody['query'] = query
  if ('variables' in parsed) gqlBody['variables'] = parsed['variables']
  if ('operationName' in parsed) {
    const op = parsed['operationName']
    if (op !== undefined && op !== null && typeof op !== 'string') {
      throw new Error('graphql body: operationName must be a string when provided')
    }
    gqlBody['operationName'] = op
  }
  return {
    body: { encoding: 'graphql', content: JSON.stringify(gqlBody) },
    forcePost: true,
    forceContentType: 'application/json',
  }
}

function buildCapabilities(protocol: string): string[] {
  switch (protocol) {
    case 'graphql':
      return ['protocol.http', 'protocol.graphql']
    default:
      return ['protocol.http']
  }
}

function buildJsonBody(body: BodyDef): BodyBuildResult {
  const content = body.content?.trim()
  if (!content) return { body: undefined, forcePost: false }
  try {
    return { body: { encoding: 'json', content: JSON.parse(content) }, forcePost: false }
  } catch {
    // not valid JSON — pass as raw string
  }
  return {
    body: { encoding: 'json', content },
    forcePost: false,
  }
}

function buildFormBody(body: BodyDef): BodyBuildResult {
  const pairs = (body.formRows ?? [])
    .filter((row) => row.enabled && row.type === 'text' && row.key.trim() !== '')
    .map((row) => `${encodeURIComponent(row.key)}=${encodeURIComponent(row.value ?? '')}`)
  return {
    body: { encoding: 'form', content: pairs.join('&') },
    forcePost: false,
    forceContentType: 'application/x-www-form-urlencoded',
  }
}

function buildMultipartBody(body: BodyDef): BodyBuildResult {
  // multipart formRows serialization is handled by the engine for WASM; raw content is passed through.
  return {
    body: { encoding: 'multipart', content: body.content ?? '' },
    forcePost: false,
  }
}

function buildAssertions(defs: AssertionDef[] | undefined): TraCtlAssertion[] | undefined {
  if (!defs || defs.length === 0) return undefined
  return defs.map((a, i): TraCtlAssertion => {
    const id = a.id.trim() !== '' ? a.id : `assertion-${i + 1}`
    let expected: string | number | boolean = a.expected
    if (a.kind === 'status') {
      const n = parseInt(a.expected, 10)
      if (!isNaN(n)) expected = n
    }
    return { id, kind: a.kind, op: a.op, expected, severity: a.severity || 'error' }
  })
}

function buildExtracts(defs: ExtractDef[] | undefined): TraCtlExtract[] | undefined {
  if (!defs || defs.length === 0) return undefined
  return defs.map((e, i): TraCtlExtract => ({
    id: e.id.trim() !== '' ? e.id : `extract-${i + 1}`,
    source: e.source,
    path: e.path,
    as: e.variableName || e.id,
    scope: e.scope || 'workflow',
  }))
}

function buildHooks(req: RequestDef): TraCtlStep['hooks'] | undefined {
  const pre = req.preScript?.trim()
  const post = req.postScript?.trim()
  if (!pre && !post) return undefined
  const hooks: TraCtlStep['hooks'] = {}
  if (pre) hooks.beforeStep = { language: 'js', source: pre }
  if (post) hooks.afterStep = { language: 'js', source: post }
  return hooks
}

function buildRetry(req: RequestDef): TraCtlStep['retry'] | undefined {
  const retry = req.settings?.retry
  if (!retry || retry.maxAttempts <= 0) return undefined
  return {
    maxAttempts: retry.maxAttempts,
    backoff: retry.strategy,
    delay: toIso8601Duration(retry.delayMs),
  }
}

// ─── Public API ───────────────────────────────────────────────────────────────

/**
 * Converts a RequestDef to a JSON string representing a single-step TraCtlSpec
 * document, suitable for use as WorkflowRunRequest.document (ADR-017 §3).
 *
 * @returns JSON string
 */
export function requestDefToDocument(req: RequestDef): string {
  return JSON.stringify(requestDefToDocumentObject(req))
}

/**
 * Converts a RequestDef to a single-step TraCtlSpec object.
 */
export function requestDefToDocumentObject(req: RequestDef): TraCtlSpecDocument {
  const protocol = normalizeProtocol(req.protocol)

  const { body, forcePost, forceContentType } = buildBody(req)

  let method = (req.method || 'GET').toUpperCase()
  if (forcePost) method = 'POST'

  const target = buildTarget(req)

  const headers = buildHeaders(req, forceContentType)

  const request: TraCtlStep['request'] = {
    protocol,
    target,
    operation: method,
  }
  if (headers) request.headers = headers
  if (body) request.body = body

  const name = req.name?.trim() || target
  const stepId = req.id?.trim() || 'request-step'
  const failurePolicy = req.settings?.failurePolicy || 'resilient'

  const step: TraCtlStep = { id: stepId, kind: 'request', request }

  const assertions = buildAssertions(req.assertions)
  if (assertions) step.assertions = assertions

  const extracts = buildExtracts(req.extracts)
  if (extracts) step.extracts = extracts

  const hooks = buildHooks(req)
  if (hooks) step.hooks = hooks

  const timeoutMs = req.settings?.timeoutMs
  if (timeoutMs && timeoutMs > 0) {
    step.timeout = toIso8601Duration(timeoutMs)
  }

  const retry = buildRetry(req)
  if (retry) step.retry = retry

  return {
    schemaVersion: 1,
    capabilities: buildCapabilities(protocol),
    metadata: { name },
    workflows: [
      {
        id: 'request-flow',
        name,
        failurePolicy,
        steps: [step],
      },
    ],
  }
}
