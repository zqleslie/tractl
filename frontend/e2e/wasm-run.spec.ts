import { expect, test } from '@playwright/test'

type RunFormat = 'yaml' | 'json'

async function wasmRun(
  page: import('@playwright/test').Page,
  document: string,
  format: RunFormat = 'yaml',
): Promise<Record<string, unknown>> {
  await page.waitForFunction(
    () => typeof window.tractl?.run === 'function',
    { timeout: 30_000 },
  )
  return page.evaluate(
    async ({ doc, fmt }) => {
      const result = await window.tractl!.run(doc, fmt)
      return result as Record<string, unknown>
    },
    { doc: document, fmt: format },
  )
}

const httpbin = 'https://httpbin.org'

function baseCapabilitiesYaml(): string {
  return `schemaVersion: 1
capabilities:
  - protocol.http
`
}

function largeValidYaml(targetBytes: number): string {
  const header = `${baseCapabilitiesYaml()}metadata:
  padding: "
`
  const footer = `"
workflows:
  - id: wf-large
    steps:
      - id: step-one
        kind: request
        request:
          protocol: http
          target: "${httpbin}/get"
          operation: GET
        assertions:
          - id: status-one
            kind: status
            op: equals
            expected: 200
`
  const padSize = Math.max(0, targetBytes - Buffer.byteLength(header + footer, 'utf8'))
  const padding = 'x'.repeat(padSize)
  return header + padding + footer
}

test.describe('WASM engine execution (UI-0.1F)', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  test('GET against public endpoint returns engine result with wasm metadata', async ({
    page,
  }) => {
    await page.route('**/wasm-run-fixture/get', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ ok: true }),
      })
    })

    const doc = `${baseCapabilitiesYaml()}workflows:
  - id: wf-get
    steps:
      - id: step-get
        kind: request
        request:
          protocol: http
          target: http://127.0.0.1:5173/wasm-run-fixture/get
          operation: GET
        assertions:
          - id: assert-status
            kind: status
            op: equals
            expected: 200
`
    const requests: string[] = []
    page.on('request', (req) => {
      if (req.url().includes('/wasm-run-fixture/get')) {
        requests.push(req.url())
      }
    })

    const result = await wasmRun(page, doc, 'yaml')
    expect(result).toMatchObject({
      surface: 'web-wasm',
      executionMode: 'browser',
      networkProvider: 'fetch',
      Passed: true,
    })
    expect(requests.some((u) => u.includes('/wasm-run-fixture/get'))).toBe(true)
  })

  test('POST sends JSON body', async ({ page }) => {
    const doc = `${baseCapabilitiesYaml()}workflows:
  - id: wf-post
    steps:
      - id: step-post
        kind: request
        request:
          protocol: http
          target: ${httpbin}/post
          operation: POST
          headers:
            Content-Type: application/json
          body:
            encoding: json
            content:
              probe: tractl-wasm
        assertions:
          - id: assert-status
            kind: status
            op: equals
            expected: 200
`
    const result = await wasmRun(page, doc, 'yaml')
    expect(result).toMatchObject({
      surface: 'web-wasm',
      Passed: true,
    })
  })

  test('assertion evaluation marks failed status', async ({ page }) => {
    await page.route('**/wasm-run-fixture/status', async (route) => {
      await route.fulfill({ status: 404, body: 'not found' })
    })

    const doc = `${baseCapabilitiesYaml()}workflows:
  - id: wf-404
    steps:
      - id: step-404
        kind: request
        request:
          protocol: http
          target: http://127.0.0.1:5173/wasm-run-fixture/status
          operation: GET
        assertions:
          - id: assert-status
            kind: status
            op: equals
            expected: 200
`
    const result = await wasmRun(page, doc, 'yaml')
    expect(result).toMatchObject({ surface: 'web-wasm', Passed: false })
    expect(result.error).toBeUndefined()
    const steps = (result.Workflows as Array<{ Steps: Array<{ CausesFailure: boolean }> }>)[0]
      .Steps
    expect(steps[0].CausesFailure).toBe(true)
  })

  test('extracts flow across dependent steps', async ({ page }) => {
    const doc = `${baseCapabilitiesYaml()}workflows:
  - id: wf-extract
    steps:
      - id: step-a
        kind: request
        request:
          protocol: http
          target: ${httpbin}/response-headers?X-Probe=wasm-token
          operation: GET
        extracts:
          - id: hdr-extract
            source: header
            path: X-Probe
            as: probe
            scope: workflow
      - id: step-b
        kind: request
        dependsOn:
          - step-a
        request:
          protocol: http
          target: ${httpbin}/get?token=\${vars.probe}
          operation: GET
        assertions:
          - id: assert-status
            kind: status
            op: equals
            expected: 200
`
    const result = await wasmRun(page, doc, 'yaml')
    expect(result).toMatchObject({ Passed: true, surface: 'web-wasm' })
  })

  test('multi-step workflow executes', async ({ page }) => {
    const doc = `${baseCapabilitiesYaml()}workflows:
  - id: wf-multi
    steps:
      - id: step-one
        kind: request
        request:
          protocol: http
          target: ${httpbin}/get?step=1
          operation: GET
        assertions:
          - id: a1
            kind: status
            op: equals
            expected: 200
      - id: step-two
        kind: request
        dependsOn:
          - step-one
        request:
          protocol: http
          target: ${httpbin}/get?step=2
          operation: GET
        assertions:
          - id: a2
            kind: status
            op: equals
            expected: 200
`
    const result = await wasmRun(page, doc, 'yaml')
    expect(result).toMatchObject({ Passed: true })
    const workflows = result.Workflows as Array<{ Steps: unknown[] }>
    expect(workflows[0].Steps.length).toBe(2)
  })

  test('failure paths return structured results without crashing runtime', async ({
    page,
  }) => {
    const cases = [
      {
        name: '500',
        target: `${httpbin}/status/500`,
      },
      {
        name: 'dns',
        target: 'https://tractl-wasm-invalid.invalid/',
      },
      {
        name: 'timeout',
        target: `${httpbin}/delay/10`,
        timeout: 'PT2S',
      },
    ] as const

    for (const tc of cases) {
      const timeoutLine = 'timeout' in tc ? `        timeout: ${tc.timeout}\n` : ''
      const doc = `${baseCapabilitiesYaml()}workflows:
  - id: wf-fail-${tc.name}
    steps:
      - id: step-fail
        kind: request
${timeoutLine}        request:
          protocol: http
          target: ${tc.target}
          operation: GET
        assertions:
          - id: assert-status
            kind: status
            op: equals
            expected: 200
`
      const result = await wasmRun(page, doc, 'yaml')
      expect(result.surface).toBe('web-wasm')
      expect(result.error).toBeUndefined()
    }

    const recovery = `${baseCapabilitiesYaml()}workflows:
  - id: wf-ok
    steps:
      - id: step-ok
        kind: request
        request:
          protocol: http
          target: ${httpbin}/get
          operation: GET
        assertions:
          - id: assert-status
            kind: status
            op: equals
            expected: 200
`
    const after = await wasmRun(page, recovery, 'yaml')
    expect(after).toMatchObject({ Passed: true })
  })

  test('plan failure returns structured error without terminating wasm', async ({ page }) => {
    const doc = `${baseCapabilitiesYaml().replace('protocol.http', 'playwright')}workflows:
  - id: wf-plan
    steps:
      - id: step-one
        kind: request
        request:
          protocol: http
          target: https://example.com
          operation: GET
`
    const result = await wasmRun(page, doc, 'yaml')
    expect(result.error).toMatchObject({ code: 'TRACTL_EXECUTION_ERROR' })

    const recovery = `${baseCapabilitiesYaml()}workflows:
  - id: wf-ok
    steps:
      - id: step-one
        kind: request
        request:
          protocol: http
          target: https://example.com
          operation: GET
`
    const second = await page.evaluate(async () => {
      const r = await window.tractl!.validate(
        `schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wf-ok
    steps:
      - id: step-one
        kind: request
        request:
          protocol: http
          target: https://example.com
          operation: GET
`,
        'yaml',
      )
      return r as Record<string, unknown>
    })
    expect(second).toMatchObject({ valid: true })
  })

  test('pipeline parse failure returns TRACTL_EXECUTION_ERROR', async ({ page }) => {
    const result = await wasmRun(page, 'not: valid: yaml: [', 'yaml')
    expect(result.error).toMatchObject({ code: 'TRACTL_EXECUTION_ERROR' })
  })

  test('unicode metadata and variables execute', async ({ page }) => {
    const doc = `${baseCapabilitiesYaml()}metadata:
  name: தமிழ்
variables:
  label: 日本語
workflows:
  - id: wf-unicode
    steps:
      - id: step-unicode
        kind: request
        request:
          protocol: http
          target: ${httpbin}/get?label=\${vars.label}
          operation: GET
        assertions:
          - id: assert-status
            kind: status
            op: equals
            expected: 200
`
    const result = await wasmRun(page, doc, 'yaml')
    expect(result).toMatchObject({ Passed: true, surface: 'web-wasm' })
  })

  test('500 KB document runs without freezing', async ({ page }) => {
    test.setTimeout(60_000)
    const document = largeValidYaml(500 * 1024)
    expect(Buffer.byteLength(document, 'utf8')).toBeGreaterThanOrEqual(500 * 1024)

    const started = Date.now()
    const result = await wasmRun(page, document, 'yaml')
    const elapsedMs = Date.now() - started

    expect(result.surface).toBe('web-wasm')
    expect(elapsedMs).toBeLessThan(120_000)
    expect(result.error).toBeUndefined()
    expect(typeof result.Passed).toBe('boolean')
  })
})
