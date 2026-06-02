import { describe, it, expect } from 'vitest'
import { requestDefToDocument } from './requestDefToDocument'
import type { RequestDef } from '@/types/requestDef'

// minimal valid RequestDef
function base(overrides: Partial<RequestDef> = {}): RequestDef {
  return {
    id: 'step-1',
    name: 'Test Request',
    method: 'GET',
    url: 'https://api.example.com/users',
    ...overrides,
  }
}

function parse(json: string): ReturnType<typeof JSON.parse> {
  return JSON.parse(json) as ReturnType<typeof JSON.parse>
}

function step(doc: ReturnType<typeof JSON.parse>): ReturnType<typeof JSON.parse> {
  return doc.workflows[0].steps[0]
}

describe('requestDefToDocument', () => {
  it('defaultsProtocolToHttp — absent protocol field produces request.protocol === "http"', () => {
    const json = requestDefToDocument(base())
    const doc = parse(json)
    expect(step(doc).request.protocol).toBe('http')
  })

  it('explicitProtocolPassedThrough — explicit protocol is preserved', () => {
    const json = requestDefToDocument(base({ protocol: 'graphql' }))
    const doc = parse(json)
    expect(step(doc).request.protocol).toBe('graphql')
  })

  it('graphqlProtocol — graphql body produces JSON body, operation POST, Content-Type header', () => {
    const req = base({
      protocol: 'graphql',
      method: 'GET', // must be overridden
      body: {
        encoding: 'graphql',
        content: JSON.stringify({
          query: '{ users { id name } }',
          variables: { limit: 10 },
        }),
      },
    })
    const doc = parse(requestDefToDocument(req))
    const s = step(doc)

    expect(s.request.operation).toBe('POST')
    expect(s.request.headers?.['Content-Type']).toBe('application/json')
    expect(s.request.body.encoding).toBe('graphql')

    const bodyContent = JSON.parse(s.request.body.content as string) as Record<string, unknown>
    expect(bodyContent).toHaveProperty('query', '{ users { id name } }')
    expect(bodyContent).toHaveProperty('variables', { limit: 10 })
  })

  it('graphqlMissingQueryThrows — graphql body without query throws', () => {
    const req = base({
      protocol: 'graphql',
      body: {
        encoding: 'graphql',
        content: JSON.stringify({ variables: { limit: 10 } }),
      },
    })
    expect(() => requestDefToDocument(req)).toThrow(/required key "query" is missing/i)
  })

  it('graphqlInvalidJsonThrows — non-JSON graphql content throws', () => {
    const req = base({
      protocol: 'graphql',
      body: {
        encoding: 'graphql',
        content: '{not-json',
      },
    })
    expect(() => requestDefToDocument(req)).toThrow(/not valid JSON/i)
  })

  it('odataNormalizedToHttp — odata protocol is serialized as http', () => {
    const json = requestDefToDocument(base({ protocol: 'odata' }))
    const doc = parse(json)
    expect(step(doc).request.protocol).toBe('http')
  })

  it('disabledParamsExcluded — params with enabled:false are not included in the target URL', () => {
    const req = base({
      url: 'https://api.example.com/items',
      params: [
        { enabled: true, key: 'status', value: 'active' },
        { enabled: false, key: 'hidden', value: 'secret' },
      ],
    })
    const doc = parse(requestDefToDocument(req))
    const target: string = step(doc).request.target
    expect(target).toContain('status=active')
    expect(target).not.toContain('hidden')
    expect(target).not.toContain('secret')
  })

  it('disabledHeadersExcluded — headers with enabled:false are not in the output', () => {
    const req = base({
      headers: [
        { enabled: true, key: 'X-Tenant', value: 'acme' },
        { enabled: false, key: 'X-Internal', value: 'skip-me' },
      ],
    })
    const doc = parse(requestDefToDocument(req))
    const headers = step(doc).request.headers as Record<string, string> | undefined
    expect(headers?.['X-Tenant']).toBe('acme')
    expect(headers?.['X-Internal']).toBeUndefined()
  })

  it('timeoutConversion — settings.timeoutMs:30000 produces timeout:"PT30S"', () => {
    const req = base({ settings: { timeoutMs: 30000 } })
    const doc = parse(requestDefToDocument(req))
    expect(step(doc).timeout).toBe('PT30S')
  })

  it('timeoutConversionTruncates — 1500ms truncates to PT1S, not PT2S', () => {
    const req = base({ settings: { timeoutMs: 1500 } })
    const doc = parse(requestDefToDocument(req))
    expect(step(doc).timeout).toBe('PT1S')
  })

  it('timeoutConversionMinimum — sub-second timeout clamps to PT1S', () => {
    const req = base({ settings: { timeoutMs: 100 } })
    const doc = parse(requestDefToDocument(req))
    expect(step(doc).timeout).toBe('PT1S')
  })

  it('noTimeoutWhenZero — timeoutMs:0 produces no timeout field', () => {
    const req = base({ settings: { timeoutMs: 0 } })
    const doc = parse(requestDefToDocument(req))
    expect(step(doc).timeout).toBeUndefined()
  })

  it('retryConversion — RetryDef produces correct spec shape', () => {
    const req = base({
      settings: {
        retry: { strategy: 'exponential', maxAttempts: 3, delayMs: 2000, backoffFactor: 2 },
      },
    })
    const doc = parse(requestDefToDocument(req))
    const retry = step(doc).retry as Record<string, unknown>
    expect(retry.maxAttempts).toBe(3)
    expect(retry.backoff).toBe('exponential')
    expect(retry.delay).toBe('PT2S')
  })

  it('hooksWrapped — preScript and postScript map to hooks.beforeStep and hooks.afterStep', () => {
    const req = base({ preScript: 'console.log("pre")', postScript: 'console.log("post")' })
    const doc = parse(requestDefToDocument(req))
    const hooks = step(doc).hooks as Record<string, unknown>
    expect((hooks.beforeStep as Record<string, string>).source).toBe('console.log("pre")')
    expect((hooks.afterStep as Record<string, string>).source).toBe('console.log("post")')
  })

  it('schemaVersion — output document has schemaVersion:1', () => {
    const doc = parse(requestDefToDocument(base()))
    expect(doc.schemaVersion).toBe(1)
  })

  it('capabilitiesHttpOnly — http protocol produces only protocol.http', () => {
    const doc = parse(requestDefToDocument(base({ protocol: 'http' })))
    expect(doc.capabilities).toEqual(['protocol.http'])
  })

  it('capabilitiesHttpDefault — absent protocol defaults to protocol.http only', () => {
    const doc = parse(requestDefToDocument(base()))
    expect(doc.capabilities).toEqual(['protocol.http'])
  })

  it('capabilitiesGraphQL — graphql protocol produces both protocol.http and protocol.graphql', () => {
    const doc = parse(requestDefToDocument(base({ protocol: 'graphql' })))
    expect(doc.capabilities).toEqual(['protocol.http', 'protocol.graphql'])
  })

  it('workflowId — output workflow id is "request-flow"', () => {
    const doc = parse(requestDefToDocument(base()))
    expect(doc.workflows[0].id).toBe('request-flow')
  })

  it('workflowIdIsRequestFlow — workflow id is request-flow', () => {
    const doc = parse(requestDefToDocument(base()))
    expect(doc.workflows[0].id).toBe('request-flow')
  })

  it('authBearerInjected — bearer auth injects Authorization header', () => {
    const req = base({
      auth: { type: 'bearer', token: 'abc123' },
    })
    const doc = parse(requestDefToDocument(req))
    const headers = step(doc).request.headers as Record<string, string> | undefined
    expect(headers?.Authorization).toBe('Bearer abc123')
  })

  it('authBearerNotOverriddenWhenAlreadySet — explicit Authorization header is preserved', () => {
    const req = base({
      headers: [{ enabled: true, key: 'Authorization', value: 'Bearer explicit' }],
      auth: { type: 'bearer', token: 'abc123' },
    })
    const doc = parse(requestDefToDocument(req))
    const headers = step(doc).request.headers as Record<string, string> | undefined
    expect(headers?.Authorization).toBe('Bearer explicit')
  })

  it('authBasicInjected — basic auth injects base64 Authorization header', () => {
    const req = base({
      auth: { type: 'basic', username: 'user', password: 'pass' },
    })
    const doc = parse(requestDefToDocument(req))
    const headers = step(doc).request.headers as Record<string, string> | undefined
    expect(headers?.Authorization).toBe('Basic dXNlcjpwYXNz')
  })

  it('authApiKeyHeader — api key header placement injects named header', () => {
    const req = base({
      auth: { type: 'apikey', placement: 'header', keyName: 'X-API-Key', keyValue: 'secret' },
    })
    const doc = parse(requestDefToDocument(req))
    const headers = step(doc).request.headers as Record<string, string> | undefined
    expect(headers?.['X-API-Key']).toBe('secret')
  })

  it('authNoneProducesNoHeader — auth none does not inject Authorization', () => {
    const req = base({
      auth: { type: 'none' },
    })
    const doc = parse(requestDefToDocument(req))
    const headers = step(doc).request.headers as Record<string, string> | undefined
    expect(headers?.Authorization).toBeUndefined()
  })

  it('authAbsentProducesNoHeader — missing auth does not inject Authorization', () => {
    const doc = parse(requestDefToDocument(base()))
    const headers = step(doc).request.headers as Record<string, string> | undefined
    expect(headers?.Authorization).toBeUndefined()
  })

  it('formBodyEncoded — enabled text form rows serialize as urlencoded content', () => {
    const req = base({
      body: {
        encoding: 'form',
        formRows: [
          { enabled: true, key: 'name', value: 'Alice', type: 'text' },
          { enabled: true, key: 'city', value: 'New York', type: 'text' },
        ],
      },
    })
    const doc = parse(requestDefToDocument(req))
    const body = step(doc).request.body as { encoding: string; content: string }
    expect(body.encoding).toBe('form')
    expect(body.content).toBe('name=Alice&city=New%20York')
  })

  it('formBodyDisabledRowsExcluded — disabled form rows are excluded from content', () => {
    const req = base({
      body: {
        encoding: 'form',
        formRows: [
          { enabled: true, key: 'name', value: 'Alice', type: 'text' },
          { enabled: false, key: 'hidden', value: 'secret', type: 'text' },
        ],
      },
    })
    const doc = parse(requestDefToDocument(req))
    const body = step(doc).request.body as { encoding: string; content: string }
    expect(body.content).toBe('name=Alice')
  })

  it('formBodyInjectsContentType — form body injects x-www-form-urlencoded content type', () => {
    const req = base({
      body: {
        encoding: 'form',
        formRows: [{ enabled: true, key: 'name', value: 'Alice', type: 'text' }],
      },
    })
    const doc = parse(requestDefToDocument(req))
    const headers = step(doc).request.headers as Record<string, string> | undefined
    expect(headers?.['Content-Type']).toBe('application/x-www-form-urlencoded')
  })

  it('graphqlExtensionsPreserved — extra top-level keys in graphql body content are passed through', () => {
    const req = base({
      protocol: 'graphql',
      body: {
        encoding: 'graphql',
        content: JSON.stringify({
          query: '{ me { id } }',
          variables: { limit: 5 },
          extensions: { persistedQuery: { version: 1, sha256Hash: 'abc' } },
        }),
      },
    })
    const doc = parse(requestDefToDocument(req))
    const bodyContent = JSON.parse(step(doc).request.body.content as string) as Record<string, unknown>
    expect(bodyContent).toHaveProperty('query')
    expect(bodyContent).toHaveProperty('variables')
    expect(bodyContent).toHaveProperty('extensions')
    expect((bodyContent.extensions as Record<string, unknown>).persistedQuery).toBeDefined()
  })

  it('multipartBodyPassesThrough — multipart body is preserved without crashing', () => {
    const req = base({
      body: {
        encoding: 'multipart',
        content: '--boundary',
        formRows: [{ enabled: true, key: 'file', value: '', type: 'file', filename: 'a.txt' }],
      },
    })
    const doc = parse(requestDefToDocument(req))
    const body = step(doc).request.body as { encoding: string; content: string }
    expect(body.encoding).toBe('multipart')
    expect(body.content).toBe('--boundary')
  })
})
