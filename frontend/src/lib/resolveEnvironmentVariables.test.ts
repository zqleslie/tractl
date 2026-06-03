import { describe, expect, it } from 'vitest'
import {
  formatMissingEnvironmentVariables,
  resolveEnvironmentString,
  resolveEnvironmentUrl,
  resolveRequestDefEnvironmentVariables,
  resolveTraCtlSpecEnvironmentVariables,
} from '@/lib/resolveEnvironmentVariables'
import type { TraCtlSpecDocument } from '@/components/request-editor/tractlSpecDocument'
import type { RequestDef } from '@/types/requestDef'

function minimalDef(overrides: Partial<RequestDef> = {}): RequestDef {
  return {
    id: 'test',
    name: 'Test',
    method: 'GET',
    url: 'https://api.example.test',
    ...overrides,
  }
}

describe('resolveRequestDefEnvironmentVariables', () => {
  it('resolvesUrlPlaceholder', () => {
    const def = minimalDef({ url: '{{baseUrl}}/users' })
    const result = resolveRequestDefEnvironmentVariables(def, { baseUrl: 'https://api.example.test' })
    expect(result.value.url).toBe('https://api.example.test/users')
    expect(result.missing).toEqual([])
  })

  it('resolvesParamValue', () => {
    const def = minimalDef({
      params: [{ enabled: true, key: 'limit', value: '{{limit}}' }],
    })
    const result = resolveRequestDefEnvironmentVariables(def, { limit: '10' })
    expect(result.value.params?.[0]?.value).toBe('10')
    expect(result.missing).toEqual([])
  })

  it('resolvesHeaderValue', () => {
    const def = minimalDef({
      headers: [{ enabled: true, key: 'Authorization', value: 'Bearer {{token}}' }],
    })
    const result = resolveRequestDefEnvironmentVariables(def, { token: 'abc' })
    expect(result.value.headers?.[0]?.value).toBe('Bearer abc')
    expect(result.missing).toEqual([])
  })

  it('resolvesBodyContent', () => {
    const def = minimalDef({
      body: { encoding: 'json', content: '{"id":"{{userId}}"}' },
    })
    const result = resolveRequestDefEnvironmentVariables(def, { userId: '42' })
    expect(result.value.body?.content).toBe('{"id":"42"}')
    expect(result.missing).toEqual([])
  })

  it('collectsMissingAcrossFields', () => {
    const def = minimalDef({
      url: '{{baseUrl}}/path',
      params: [{ enabled: true, key: 'q', value: '{{query}}' }],
      headers: [{ enabled: true, key: 'X-Token', value: '{{token}}' }],
      body: { encoding: 'json', content: '{"id":"{{userId}}"}' },
    })
    const result = resolveRequestDefEnvironmentVariables(def, {})
    expect(result.missing).toEqual(['baseUrl', 'query', 'token', 'userId'])
  })

  it('disabledRowsNotResolved', () => {
    const def = minimalDef({
      params: [{ enabled: false, key: 'limit', value: '{{limit}}' }],
      headers: [{ enabled: false, key: 'Authorization', value: 'Bearer {{token}}' }],
    })
    const result = resolveRequestDefEnvironmentVariables(def, {})
    // Disabled rows are not resolved — missing list stays empty
    expect(result.missing).toEqual([])
    // Values are left unchanged
    expect(result.value.params?.[0]?.value).toBe('{{limit}}')
    expect(result.value.headers?.[0]?.value).toBe('Bearer {{token}}')
  })

  it('noDoubleParamMerge', () => {
    // Params remain as rows — not merged into URL. Go's mapper does the merge exactly once.
    const def = minimalDef({
      url: 'https://api.example.test',
      params: [{ enabled: true, key: 'limit', value: '10' }],
    })
    const result = resolveRequestDefEnvironmentVariables(def, {})
    expect(result.value.url).toBe('https://api.example.test')
    expect(result.value.params).toEqual([{ enabled: true, key: 'limit', value: '10' }])
  })
})

describe('resolveEnvironmentVariables', () => {
  it('resolves placeholders in plain strings', () => {
    const result = resolveEnvironmentString('{{baseUrl}}/users/{{userId}}', {
      baseUrl: 'https://api.example.test',
      userId: 'u_123',
    })

    expect(result.value).toBe('https://api.example.test/users/u_123')
    expect(result.missing).toEqual([])
  })

  it('returns every missing variable without substituting placeholders', () => {
    const result = resolveEnvironmentString('{{baseUrl}}/users/{{missing}}', {
      baseUrl: 'https://api.example.test',
    })

    expect(result.value).toBe('https://api.example.test/users/{{missing}}')
    expect(result.missing).toEqual(['missing'])
    expect(formatMissingEnvironmentVariables(result.missing)).toBe(
      'TRACTL_ENV_VARIABLE_MISSING: Missing environment variable: missing',
    )
  })

  it('resolves repeated placeholders and reports unique missing names', () => {
    const result = resolveEnvironmentString(
      '{{baseUrl}}/{{team}}/{{team}}/{{missing}}/{{missing}}',
      { baseUrl: 'https://api.example.test', team: 'platform' },
    )

    expect(result.value).toBe(
      'https://api.example.test/platform/platform/{{missing}}/{{missing}}',
    )
    expect(result.missing).toEqual(['missing'])
  })

  it('resolves encoded query parameter placeholders in canonical targets', () => {
    const result = resolveEnvironmentUrl('/search?q=%7B%7Bquery%7D%7D', {
      query: 'hello world',
    })

    expect(result.value).toBe('/search?q=hello%20world')
    expect(result.missing).toEqual([])
  })

  it('resolves request URL, headers, and nested body strings in a spec document', () => {
    const document: TraCtlSpecDocument = {
      schemaVersion: 1,
      capabilities: ['protocol.http'],
      workflows: [
        {
          id: 'request-flow',
          steps: [
            {
              id: 'request-step',
              kind: 'request',
              request: {
                protocol: 'http',
                target: '{{baseUrl}}/accounts?tenant=%7B%7Btenant%7D%7D',
                operation: 'POST',
                headers: { 'x-tenant': '{{tenant}}' },
                body: {
                  encoding: 'json',
                  content: { account: '{{accountId}}' },
                },
              },
            },
          ],
        },
      ],
    }

    const result = resolveTraCtlSpecEnvironmentVariables(document, {
      accountId: 'acct_123',
      baseUrl: 'https://api.example.test',
      tenant: 'qa team',
    })

    const request = result.value.workflows[0]?.steps[0]?.request
    expect(request?.target).toBe(
      'https://api.example.test/accounts?tenant=qa%20team',
    )
    expect(request?.headers).toEqual({ 'x-tenant': 'qa team' })
    expect(request?.body?.content).toEqual({ account: 'acct_123' })
    expect(result.missing).toEqual([])
  })
})
