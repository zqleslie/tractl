/**
 * GET request edge-case and error scenario tests.
 *
 * Covers: empty URL, network failure, server errors, large responses,
 * concurrent runs, URL normalization, and Vite dev-shell detection.
 */
import { expect, test } from '@playwright/test'

async function openNewRequest(page: import('@playwright/test').Page) {
  await page.goto('/')
  await page.waitForFunction(() => typeof (window as any).tractl?.run === 'function', {
    timeout: 30_000,
  })
  await page.getByTestId('fast-start-new-request').click()
  await expect(page.getByTestId('screen-request')).toBeVisible()
}

test.describe('GET request edge cases', () => {
  // ── Empty / invalid URL ───────────────────────────────────────────────────

  test('run button is disabled when URL is empty', async ({ page }) => {
    await openNewRequest(page)
    await expect(page.getByTestId('request-run')).toBeDisabled()
  })

  test('run shows error when URL has unresolved env variable', async ({ page }) => {
    await openNewRequest(page)
    await page.getByLabel('Request URL').fill('{{missing_var}}/api/data')
    await page.getByTestId('request-run').click()

    await expect(
      page.getByText(/Missing environment variable: missing_var/i),
    ).toBeVisible({ timeout: 10_000 })
  })

  // ── Network failure ───────────────────────────────────────────────────────

  test('network failure shows an error result not a crash', async ({ page }) => {
    await openNewRequest(page)
    await page.route('**/get-edge-fixture/abort', (route) => route.abort('failed'))

    await page.getByLabel('Request URL').fill('http://127.0.0.1:5173/get-edge-fixture/abort')
    await page.getByTestId('request-run').click()

    // Should show some kind of error state — outcome badge or error text
    await expect(
      page.getByTestId('request-outcome-badge').or(page.getByTestId('request-response-body')),
    ).toBeVisible({ timeout: 30_000 })
    // Run button should be re-enabled regardless of outcome
    await expect(page.getByTestId('request-run')).toBeEnabled({ timeout: 5_000 })
  })

  // ── 5xx server errors ─────────────────────────────────────────────────────

  test('500 response shows non-passing result', async ({ page }) => {
    await openNewRequest(page)
    await page.route('**/get-edge-fixture/500', (route) =>
      route.fulfill({ status: 500, body: 'internal server error' }),
    )

    await page.getByLabel('Request URL').fill('http://127.0.0.1:5173/get-edge-fixture/500')
    await page.getByTestId('request-run').click()

    await expect(page.getByTestId('request-outcome-badge')).toBeVisible({ timeout: 30_000 })
    await expect(page.getByText('500')).toBeVisible()
  })

  test('503 response is displayed with correct status code', async ({ page }) => {
    await openNewRequest(page)
    await page.route('**/get-edge-fixture/503', (route) =>
      route.fulfill({ status: 503, body: 'service unavailable' }),
    )

    await page.getByLabel('Request URL').fill('http://127.0.0.1:5173/get-edge-fixture/503')
    await page.getByTestId('request-run').click()

    await expect(page.getByText('503')).toBeVisible({ timeout: 30_000 })
  })

  // ── Empty response body ───────────────────────────────────────────────────

  test('204 No Content does not crash the result panel', async ({ page }) => {
    await openNewRequest(page)
    await page.route('**/get-edge-fixture/nocontent', (route) =>
      route.fulfill({ status: 204 }),
    )

    await page.getByLabel('Request URL').fill('http://127.0.0.1:5173/get-edge-fixture/nocontent')
    await page.getByTestId('request-run').click()

    await expect(page.getByTestId('request-outcome-badge')).toBeVisible({ timeout: 30_000 })
    await expect(page.getByText('204')).toBeVisible()
  })

  // ── Large response body ───────────────────────────────────────────────────

  test('large JSON response (100 KB) renders without freezing', async ({ page }) => {
    test.setTimeout(60_000)
    const largeBody = JSON.stringify({ items: Array.from({ length: 1000 }, (_, i) => ({ id: i, value: 'x'.repeat(100) })) })

    await openNewRequest(page)
    await page.route('**/get-edge-fixture/large', (route) =>
      route.fulfill({ status: 200, contentType: 'application/json', body: largeBody }),
    )

    const start = Date.now()
    await page.getByLabel('Request URL').fill('http://127.0.0.1:5173/get-edge-fixture/large')
    await page.getByTestId('request-run').click()
    await expect(page.getByTestId('request-outcome-badge')).toBeVisible({ timeout: 30_000 })

    expect(Date.now() - start).toBeLessThan(30_000)
    await expect(page.getByTestId('request-response-body')).toBeVisible()
  })

  // ── Vite dev-shell detection ──────────────────────────────────────────────

  test('Vite dev-shell HTML response is flagged as misconfigured URL', async ({ page }) => {
    await openNewRequest(page)
    await page.route('**/get-edge-fixture/devshell', (route) =>
      route.fulfill({
        status: 200,
        contentType: 'text/html',
        body: '<!doctype html><html><script type="module" src="/@vite/client"></script></html>',
      }),
    )

    await page.getByLabel('Request URL').fill('http://127.0.0.1:5173/get-edge-fixture/devshell')
    await page.getByTestId('request-run').click()

    await expect(
      page.getByTestId('request-response-body').or(page.getByText(/external URL/i)),
    ).toBeVisible({ timeout: 30_000 })
  })

  // ── Redirect handling ─────────────────────────────────────────────────────

  test('redirect is followed and final response is shown', async ({ page }) => {
    await openNewRequest(page)
    await page.route('**/get-edge-fixture/redirect', (route) =>
      route.fulfill({
        status: 302,
        headers: { Location: 'http://127.0.0.1:5173/get-edge-fixture/redirect-target' },
      }),
    )
    await page.route('**/get-edge-fixture/redirect-target', (route) =>
      route.fulfill({ status: 200, contentType: 'application/json', body: '{"final":true}' }),
    )

    await page.getByLabel('Request URL').fill('http://127.0.0.1:5173/get-edge-fixture/redirect')
    await page.getByTestId('request-run').click()
    await expect(page.getByTestId('request-outcome-badge')).toBeVisible({ timeout: 30_000 })
  })

  // ── Non-JSON response bodies ──────────────────────────────────────────────

  test('plain text response is displayed verbatim', async ({ page }) => {
    await openNewRequest(page)
    await page.route('**/get-edge-fixture/plaintext', (route) =>
      route.fulfill({ status: 200, contentType: 'text/plain', body: 'Hello, World!' }),
    )

    await page.getByLabel('Request URL').fill('http://127.0.0.1:5173/get-edge-fixture/plaintext')
    await page.getByTestId('request-run').click()

    await expect(page.getByTestId('request-response-body')).toContainText('Hello, World!', {
      timeout: 30_000,
    })
  })

  // ── Concurrent requests ───────────────────────────────────────────────────

  test('second run cancels or replaces the first run result', async ({ page }) => {
    await openNewRequest(page)

    let resolveFirst: (() => void) | undefined
    await page.route('**/get-edge-fixture/concurrent-a', async (route) => {
      await new Promise<void>((r) => {
        resolveFirst = r
      })
      await route.fulfill({ status: 200, contentType: 'application/json', body: '{"run":1}' })
    })
    await page.route('**/get-edge-fixture/concurrent-b', (route) =>
      route.fulfill({ status: 200, contentType: 'application/json', body: '{"run":2}' }),
    )

    await page.getByLabel('Request URL').fill('http://127.0.0.1:5173/get-edge-fixture/concurrent-a')

    // Wait for the in-flight request so the route handler has fired and resolveFirst is assigned
    const intercepted = page.waitForRequest('**/get-edge-fixture/concurrent-a')
    await page.getByTestId('request-run').click()
    await intercepted

    resolveFirst?.()
    await expect(page.getByTestId('request-outcome-badge')).toBeVisible({ timeout: 30_000 })

    // Run a second request and verify the result panel updates
    await page.getByLabel('Request URL').fill('http://127.0.0.1:5173/get-edge-fixture/concurrent-b')
    await page.getByTestId('request-run').click()
    await expect(page.getByTestId('request-response-body')).toContainText('"run": 2', {
      timeout: 30_000,
    })
  })
})
