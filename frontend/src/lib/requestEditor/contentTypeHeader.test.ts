import { describe, expect, it } from 'vitest'
import { syncContentTypeHeader } from '@/lib/requestEditor/contentTypeHeader'

describe('syncContentTypeHeader', () => {
  it('sets application/json when encoding changes to JSON', () => {
    const headers = syncContentTypeHeader(
      [],
      { encoding: 'JSON' },
      {
        encoding: 'None',
        value: '',
        formRows: [],
        rawContentType: 'text/plain',
        binaryFile: '',
      },
    )

    expect(headers).toHaveLength(1)
    expect(headers[0]?.key).toBe('Content-Type')
    expect(headers[0]?.value).toBe('application/json')
  })

  it('removes Content-Type when encoding is None', () => {
    const headers = syncContentTypeHeader(
      [{ id: '1', enabled: true, key: 'Content-Type', value: 'application/json' }],
      { encoding: 'None' },
      {
        encoding: 'JSON',
        value: '{}',
        formRows: [],
        rawContentType: 'text/plain',
        binaryFile: '',
      },
    )

    expect(headers).toHaveLength(0)
  })
})
