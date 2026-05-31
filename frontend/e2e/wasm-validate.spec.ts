import { expect, test } from '@playwright/test'

type ValidateFormat = 'yaml' | 'yml' | 'json' | 'toon'

async function wasmValidate(
  page: import('@playwright/test').Page,
  document: string,
  format: ValidateFormat,
): Promise<Record<string, unknown>> {
  await page.waitForFunction(
    () => typeof window.tractl?.validate === 'function',
    { timeout: 20_000 },
  )
  return page.evaluate(
    async ({ doc, fmt }) => {
      const result = await window.tractl!.validate(doc, fmt)
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

const invalidYamlFormat = `schemaVersion: 1
capabilities:
\t- protocol.http
workflows: []
`

const semanticInvalidYaml = `schemaVersion: 0
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

const invalidJsonFormat = `{
  "schemaVersion": 1,
  "capabilities": ["protocol.http"],
  "workflows": [],
}`

const semanticInvalidJson = JSON.stringify(
  {
    schemaVersion: 0,
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

const invalidToonFormat = `schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wf-001
    steps: []
---
extra: data
`

const semanticInvalidToon = `schemaVersion: 0
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

test.describe('WASM validation pipeline (UI-0.1E)', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  test('valid YAML returns valid: true', async ({ page }) => {
    const result = await wasmValidate(page, validYaml, 'yaml')
    expect(result).toMatchObject({ valid: true, surface: 'web-wasm' })
  })

  test('invalid YAML format returns valid: false with errors', async ({ page }) => {
    const result = await wasmValidate(page, invalidYamlFormat, 'yaml')
    expect(result).toMatchObject({ valid: false, surface: 'web-wasm' })
    const errors = result.errors as Array<{ message: string }>
    expect(errors.length).toBeGreaterThan(0)
    expect(errors[0].message).toContain('YAML_TAB_INDENTATION')
  })

  test('YAML semantic validation failure returns structured errors', async ({ page }) => {
    const result = await wasmValidate(page, semanticInvalidYaml, 'yaml')
    expect(result).toMatchObject({ valid: false, surface: 'web-wasm' })
    const errors = result.errors as Array<{ code?: string; field?: string; message: string }>
    expect(errors.some((e) => e.code === 'INVALID_SCHEMA_VERSION')).toBe(true)
  })

  test('valid JSON returns valid: true', async ({ page }) => {
    const result = await wasmValidate(page, validJson, 'json')
    expect(result).toMatchObject({ valid: true, surface: 'web-wasm' })
  })

  test('invalid JSON format returns valid: false with errors', async ({ page }) => {
    const result = await wasmValidate(page, invalidJsonFormat, 'json')
    expect(result).toMatchObject({ valid: false, surface: 'web-wasm' })
    const errors = result.errors as Array<{ message: string }>
    expect(errors.length).toBeGreaterThan(0)
  })

  test('JSON semantic validation failure returns structured errors', async ({ page }) => {
    const result = await wasmValidate(page, semanticInvalidJson, 'json')
    expect(result).toMatchObject({ valid: false, surface: 'web-wasm' })
    const errors = result.errors as Array<{ code?: string }>
    expect(errors.some((e) => e.code === 'INVALID_SCHEMA_VERSION')).toBe(true)
  })

  test('valid TOON returns valid: true', async ({ page }) => {
    const result = await wasmValidate(page, validToon, 'toon')
    expect(result).toMatchObject({ valid: true, surface: 'web-wasm' })
  })

  test('invalid TOON format returns valid: false with errors', async ({ page }) => {
    const result = await wasmValidate(page, invalidToonFormat, 'toon')
    expect(result).toMatchObject({ valid: false, surface: 'web-wasm' })
    const errors = result.errors as Array<{ message: string }>
    expect(errors.length).toBeGreaterThan(0)
  })

  test('TOON semantic validation failure returns structured errors', async ({ page }) => {
    const result = await wasmValidate(page, semanticInvalidToon, 'toon')
    expect(result).toMatchObject({ valid: false, surface: 'web-wasm' })
    const errors = result.errors as Array<{ code?: string }>
    expect(errors.some((e) => e.code === 'INVALID_SCHEMA_VERSION')).toBe(true)
  })

  test('unicode metadata names validate successfully', async ({ page }) => {
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

    const tamilResult = await wasmValidate(page, tamilDoc, 'yaml')
    expect(tamilResult).toMatchObject({ valid: true, surface: 'web-wasm' })

    const japaneseResult = await wasmValidate(page, japaneseDoc, 'yaml')
    expect(japaneseResult).toMatchObject({ valid: true, surface: 'web-wasm' })
  })

  test('500 KB document validates without freezing the runtime', async ({ page }) => {
    test.setTimeout(120_000)

    const document = largeValidYaml(500 * 1024)
    expect(Buffer.byteLength(document, 'utf8')).toBeGreaterThanOrEqual(500 * 1024)

    const started = Date.now()
    const result = await wasmValidate(page, document, 'yaml')
    const elapsedMs = Date.now() - started

    expect(result).toMatchObject({ valid: true, surface: 'web-wasm' })
    expect(elapsedMs).toBeLessThan(60_000)
  })

  test('failed validation does not break subsequent successful validation', async ({ page }) => {
    const first = await wasmValidate(page, semanticInvalidYaml, 'yaml')
    expect(first).toMatchObject({ valid: false })

    const second = await wasmValidate(page, validYaml, 'yaml')
    expect(second).toMatchObject({ valid: true, surface: 'web-wasm' })
  })

  test('rejects invalid argument types with structured bridge error', async ({ page }) => {
    await page.waitForFunction(
      () => typeof window.tractl?.validate === 'function',
      { timeout: 20_000 },
    )
    const result = await page.evaluate(async () => {
      const validate = window.tractl!.validate as (
        doc: unknown,
        format: unknown,
      ) => Promise<Record<string, unknown>>
      return validate(42, 'yaml')
    })
    expect(result.error).toMatchObject({
      code: 'TRACTL_WASM_INVALID_ARGUMENT',
    })
  })
})
