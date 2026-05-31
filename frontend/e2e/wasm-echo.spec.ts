import { expect, test } from '@playwright/test'

async function wasmEcho(
  page: import('@playwright/test').Page,
  document: string,
): Promise<Record<string, unknown>> {
  await page.waitForFunction(
    () => typeof window.tractl?.echo === 'function',
    { timeout: 20_000 },
  )
  return page.evaluate(async (doc) => {
    const result = await window.tractl!.echo(doc)
    return result as Record<string, unknown>
  }, document)
}

function kbPayload(kilobytes: number): string {
  const unit = 'x'.repeat(1024)
  return Array.from({ length: kilobytes }, () => unit).join('')
}

test.describe('WASM echo boundary (UI-0.1C)', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  test('echo("hello") returns received: true', async ({ page }) => {
    const result = await wasmEcho(page, 'hello')
    expect(result).toMatchObject({ received: true, surface: 'web-wasm' })
    expect(result.length).toBe(5)
    expect(result.preview).toBe('hello')
  })

  test('echo round-trips sample workflow spacing', async ({ page }) => {
    const document = ' workflow:   id: sample '
    const result = await wasmEcho(page, document)
    expect(result).toMatchObject({
      received: true,
      length: 24,
      preview: document,
    })
  })

  test('unicode YAML names round-trip in preview', async ({ page }) => {
    const tamil = 'yaml name: தமிழ்\n'
    const japanese = 'yaml name: 日本語\n'

    const tamilResult = await wasmEcho(page, tamil)
    expect(tamilResult).toMatchObject({ received: true, preview: tamil })

    const japaneseResult = await wasmEcho(page, japanese)
    expect(japaneseResult).toMatchObject({ received: true, preview: japanese })
  })

  test('rejects non-string arguments with structured error', async ({ page }) => {
    await page.waitForFunction(
      () => typeof window.tractl?.echo === 'function',
      { timeout: 20_000 },
    )
    const result = await page.evaluate(async () => {
      const echo = window.tractl!.echo as (value: unknown) => Promise<Record<string, unknown>>
      return echo(42)
    })
    expect(result.error).toMatchObject({
      code: 'TRACTL_WASM_INVALID_ARGUMENT',
    })
  })

  test('large payloads up to 500 KB without freezing', async ({ page }) => {
    test.setTimeout(120_000)

    for (const kb of [1, 10, 100, 500] as const) {
      const document = kbPayload(kb)
      const started = Date.now()
      const result = await wasmEcho(page, document)
      const elapsedMs = Date.now() - started

      expect(result).toMatchObject({ received: true, length: document.length })
      expect(elapsedMs).toBeLessThan(30_000)
      expect(typeof result.preview).toBe('string')
    }
  })
})
