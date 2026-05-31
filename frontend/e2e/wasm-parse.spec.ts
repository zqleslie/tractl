import { expect, test } from '@playwright/test'

type ParseFormat = 'yaml' | 'yml' | 'json' | 'toon'

async function wasmParse(
  page: import('@playwright/test').Page,
  document: string,
  format: ParseFormat,
): Promise<Record<string, unknown>> {
  await page.waitForFunction(
    () => typeof window.tractl?.parse === 'function',
    { timeout: 20_000 },
  )
  return page.evaluate(
    async ({ doc, fmt }) => {
      const result = await window.tractl!.parse(doc, fmt)
      return result as Record<string, unknown>
    },
    { doc: document, fmt: format },
  )
}

const validYaml = `schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wf-001
    steps:
      - id: step-001
        kind: request
        request:
          protocol: http
          target: https://example.com
          operation: GET
`

const invalidYaml = `schemaVersion: 1
capabilities:
\t- protocol.http
workflows: []
`

const validJson = JSON.stringify(
  {
    schemaVersion: 1,
    capabilities: ['protocol.http'],
    workflows: [
      {
        id: 'wf-001',
        steps: [
          {
            id: 'step-001',
            kind: 'request',
            request: {
              protocol: 'http',
              target: 'https://example.com',
              operation: 'GET',
            },
          },
        ],
      },
    ],
  },
  null,
  2,
)

const invalidJson = `{
  "schemaVersion": 1,
  "capabilities": ["protocol.http"],
  "workflows": [],
}`

const validToon = `schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wf-001
    steps:
      - id: step-001
        kind: request
        request:
          protocol: http
          target: "https://example.com"
          operation: GET
`

const invalidToon = `schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wf-001
    steps: []
---
extra: data
`

function largeValidYaml(targetBytes: number): string {
  const header = `schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wf-large
    steps:
`
  const stepTemplate = (index: number) => `      - id: step-${index}
        kind: request
        request:
          protocol: http
          target: "https://example.com/items/${index}"
          operation: GET
`
  let body = ''
  let index = 0
  while (Buffer.byteLength(header + body, 'utf8') < targetBytes) {
    body += stepTemplate(index)
    index += 1
  }
  return header + body
}

test.describe('WASM parser pipeline (UI-0.1D)', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  test('valid YAML returns parsed canonical spec', async ({ page }) => {
    const result = await wasmParse(page, validYaml, 'yaml')
    expect(result).toMatchObject({ parsed: true, surface: 'web-wasm' })
    const spec = result.spec as Record<string, unknown>
    expect(spec.schemaVersion).toBe(1)
    expect(spec.metadata).toMatchObject({
      sourceFormat: 'yaml',
      sourceRef: 'web-wasm',
    })
  })

  test('invalid YAML returns structured parse error', async ({ page }) => {
    const result = await wasmParse(page, invalidYaml, 'yaml')
    expect(result.error).toMatchObject({ code: 'TRACTL_PARSE_ERROR' })
    expect(String((result.error as { message: string }).message)).toContain(
      'YAML_TAB_INDENTATION',
    )
  })

  test('valid JSON returns parsed canonical spec', async ({ page }) => {
    const result = await wasmParse(page, validJson, 'json')
    expect(result).toMatchObject({ parsed: true, surface: 'web-wasm' })
    const spec = result.spec as Record<string, unknown>
    expect(spec.schemaVersion).toBe(1)
    expect(spec.metadata).toMatchObject({
      sourceFormat: 'json',
      sourceRef: 'web-wasm',
    })
  })

  test('invalid JSON returns structured parse error', async ({ page }) => {
    const result = await wasmParse(page, invalidJson, 'json')
    expect(result.error).toMatchObject({ code: 'TRACTL_PARSE_ERROR' })
  })

  test('valid TOON returns parsed canonical spec', async ({ page }) => {
    const result = await wasmParse(page, validToon, 'toon')
    expect(result).toMatchObject({ parsed: true, surface: 'web-wasm' })
    const spec = result.spec as Record<string, unknown>
    expect(spec.schemaVersion).toBe(1)
    expect(spec.metadata).toMatchObject({
      sourceFormat: 'toon',
      sourceRef: 'web-wasm',
    })
  })

  test('invalid TOON returns structured parse error', async ({ page }) => {
    const result = await wasmParse(page, invalidToon, 'toon')
    expect(result.error).toMatchObject({ code: 'TRACTL_PARSE_ERROR' })
  })

  test('unicode metadata names parse successfully', async ({ page }) => {
    const tamilDoc = `schemaVersion: 1
capabilities:
  - protocol.http
metadata:
  name: தமிழ்
workflows:
  - id: wf-001
    steps:
      - id: step-001
        kind: request
        request:
          protocol: http
          target: https://example.com
          operation: GET
`
    const japaneseDoc = `schemaVersion: 1
capabilities:
  - protocol.http
metadata:
  name: 日本語
workflows:
  - id: wf-001
    steps:
      - id: step-001
        kind: request
        request:
          protocol: http
          target: https://example.com
          operation: GET
`

    const tamilResult = await wasmParse(page, tamilDoc, 'yaml')
    expect(tamilResult).toMatchObject({ parsed: true })
    expect((tamilResult.spec as { metadata: { name: string } }).metadata.name).toBe(
      'தமிழ்',
    )

    const japaneseResult = await wasmParse(page, japaneseDoc, 'yaml')
    expect(japaneseResult).toMatchObject({ parsed: true })
    expect(
      (japaneseResult.spec as { metadata: { name: string } }).metadata.name,
    ).toBe('日本語')
  })

  test('500 KB document parses without freezing the runtime', async ({ page }) => {
    test.setTimeout(120_000)

    const document = largeValidYaml(500 * 1024)
    expect(Buffer.byteLength(document, 'utf8')).toBeGreaterThanOrEqual(500 * 1024)

    const started = Date.now()
    const result = await wasmParse(page, document, 'yaml')
    const elapsedMs = Date.now() - started

    expect(result).toMatchObject({ parsed: true, surface: 'web-wasm' })
    expect(elapsedMs).toBeLessThan(60_000)

    const workflows = (result.spec as { workflows: unknown[] }).workflows
    expect(workflows.length).toBeGreaterThan(0)
  })

  test('failed parse does not break subsequent successful parse', async ({ page }) => {
    const first = await wasmParse(page, invalidYaml, 'yaml')
    expect(first.error).toMatchObject({ code: 'TRACTL_PARSE_ERROR' })

    const second = await wasmParse(page, validYaml, 'yaml')
    expect(second).toMatchObject({ parsed: true, surface: 'web-wasm' })
  })

  test('rejects invalid argument types with structured error', async ({ page }) => {
    await page.waitForFunction(
      () => typeof window.tractl?.parse === 'function',
      { timeout: 20_000 },
    )
    const result = await page.evaluate(async () => {
      const parse = window.tractl!.parse as (
        doc: unknown,
        format: unknown,
      ) => Promise<Record<string, unknown>>
      return parse(42, 'yaml')
    })
    expect(result.error).toMatchObject({
      code: 'TRACTL_WASM_INVALID_ARGUMENT',
    })
  })
})
