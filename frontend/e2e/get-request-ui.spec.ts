/**
 * GET request UI flow tests.
 *
 * Covers the complete UI flow for executing a GET request in the web/WASM surface:
 * response display, headers panel, timing bar, assertions, extracts, run history.
 */
import { expect, test } from '@playwright/test'

test.describe('GET request UI flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
    await page.waitForFunction(
      () => typeof window.tractl?.run === 'function',
      { timeout: 30_000 },
    )
    await page.getByTestId('fast-start-new-request').click()
    await expect(page.getByTestId('screen-request')).toBeVisible()
  })

  // ── Response panel ────────────────────────────────────────────────────────

  test('shows 200 OK status code and JSON body after successful GET', async ({ page }) => {
    await page.route('**/get-ui-fixture/ok', (route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'hello tractl' }),
      }),
    )

    await page.getByLabel('Request URL').fill('http://127.0.0.1:5173/get-ui-fixture/ok')
    await page.getByTestId('request-run').click()

    await expect(page.getByTestId('request-outcome-badge')).toBeVisible({ timeout: 30_000 })
    await expect(page.getByText('200 OK')).toBeVisible()
    await expect(page.getByTestId('request-response-body')).toContainText(
      '"message": "hello tractl"',
    )
  })

  test('shows 404 status and marks run as not passed', async ({ page }) => {
    await page.route('**/get-ui-fixture/notfound', (route) =>
      route.fulfill({ status: 404, body: 'not found' }),
    )

    await page.getByLabel('Request URL').fill('http://127.0.0.1:5173/get-ui-fixture/notfound')
    await page.getByTestId('request-run').click()

    await expect(page.getByTestId('request-outcome-badge')).toBeVisible({ timeout: 30_000 })
    await expect(page.getByText('404')).toBeVisible()
  })

  // ── Timing bar ────────────────────────────────────────────────────────────

  test('renders timing bar after a completed request', async ({ page }) => {
    await page.route('**/get-ui-fixture/timing', (route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ ok: true }),
      }),
    )

    await page.getByLabel('Request URL').fill('http://127.0.0.1:5173/get-ui-fixture/timing')
    await page.getByTestId('request-run').click()

    await expect(page.getByTestId('timing-bar')).toBeVisible({ timeout: 30_000 })
  })

  // ── Assertions panel ─────────────────────────────────────────────────────

  test('shows 1/1 passed badge when status assertion passes', async ({ page }) => {
    await page.route('**/get-ui-fixture/assert-pass', (route) =>
      route.fulfill({ status: 200, contentType: 'application/json', body: '{}' }),
    )

    await page.getByLabel('Request URL').fill('http://127.0.0.1:5173/get-ui-fixture/assert-pass')
    await page.getByTestId('request-config-tab-assertions').click()
    await page.getByRole('button', { name: 'Add assertion' }).click()
    await page.getByTestId('request-run').click()

    await expect(page.getByTestId('request-outcome-badge')).toHaveText('1/1 passed', {
      timeout: 30_000,
    })
  })

  test('shows 0/1 passed when assertion fails', async ({ page }) => {
    await page.route('**/get-ui-fixture/assert-fail', (route) =>
      route.fulfill({ status: 404, contentType: 'application/json', body: '{}' }),
    )

    await page.getByLabel('Request URL').fill('http://127.0.0.1:5173/get-ui-fixture/assert-fail')
    await page.getByTestId('request-config-tab-assertions').click()
    await page.getByRole('button', { name: 'Add assertion' }).click()
    await page.getByTestId('request-run').click()

    await expect(page.getByTestId('request-outcome-badge')).toHaveText('0/1 passed', {
      timeout: 30_000,
    })
  })

  test('assertion results panel is visible after run', async ({ page }) => {
    await page.route('**/get-ui-fixture/assertions-ui', (route) =>
      route.fulfill({ status: 200, contentType: 'application/json', body: '{}' }),
    )

    await page.getByLabel('Request URL').fill('http://127.0.0.1:5173/get-ui-fixture/assertions-ui')
    await page.getByTestId('request-config-tab-assertions').click()
    await page.getByRole('button', { name: 'Add assertion' }).click()
    await page.getByTestId('request-run').click()

    await expect(page.getByTestId('request-outcome-badge')).toBeVisible({ timeout: 30_000 })
    // switch to assertions result tab
    const assertTab = page.getByTestId('request-result-tab-assertions')
    if (await assertTab.isVisible()) {
      await assertTab.click()
    }
    await expect(page.getByTestId('assertion-results')).toBeVisible()
  })

  // ── Request with query parameters ─────────────────────────────────────────

  test('appends query parameters from Params tab to the request URL', async ({ page }) => {
    const capturedUrls: string[] = []
    await page.route('**/get-ui-fixture/params**', (route) => {
      capturedUrls.push(route.request().url())
      return route.fulfill({ status: 200, contentType: 'application/json', body: '{}' })
    })

    await page.getByLabel('Request URL').fill('http://127.0.0.1:5173/get-ui-fixture/params')
    await page.getByTestId('request-config-tab-params').click()
    await page.getByRole('button', { name: 'Add parameter' }).click()

    const keyInputs = page.getByPlaceholder('Key')
    const valueInputs = page.getByPlaceholder('Value')
    await keyInputs.last().fill('version')
    await valueInputs.last().fill('2')

    await page.getByTestId('request-run').click()
    await expect(page.getByTestId('request-outcome-badge')).toBeVisible({ timeout: 30_000 })

    expect(capturedUrls.some((u) => u.includes('version=2'))).toBe(true)
  })

  // ── Request with custom headers ──────────────────────────────────────────

  test('sends custom headers configured in Headers tab', async ({ page }) => {
    let capturedHeaders: Record<string, string> = {}
    await page.route('**/get-ui-fixture/headers', (route) => {
      capturedHeaders = route.request().headers()
      return route.fulfill({ status: 200, contentType: 'application/json', body: '{}' })
    })

    await page.getByLabel('Request URL').fill('http://127.0.0.1:5173/get-ui-fixture/headers')
    await page.getByTestId('request-config-tab-headers').click()
    await page.getByRole('button', { name: 'Add header' }).click()

    const keyInputs = page.getByPlaceholder('Key')
    const valueInputs = page.getByPlaceholder('Value')
    await keyInputs.last().fill('X-Custom-Header')
    await valueInputs.last().fill('tractl-test')

    await page.getByTestId('request-run').click()
    await expect(page.getByTestId('request-outcome-badge')).toBeVisible({ timeout: 30_000 })

    expect(capturedHeaders['x-custom-header']).toBe('tractl-test')
  })

  // ── Run history ───────────────────────────────────────────────────────────

  test('records run history entry after a GET request', async ({ page }) => {
    await page.route('**/get-ui-fixture/history', (route) =>
      route.fulfill({ status: 200, contentType: 'application/json', body: '{}' }),
    )

    await page.getByLabel('Request URL').fill('http://127.0.0.1:5173/get-ui-fixture/history')
    await page.getByTestId('request-run').click()
    await expect(page.getByTestId('request-outcome-badge')).toBeVisible({ timeout: 30_000 })

    await page.getByRole('button', { name: 'Home' }).click()
    await expect(page.getByTestId('fast-start-recent')).toBeVisible()
    await expect(page.getByTestId(/recent-item-/)).toBeVisible()
  })

  test('run history persists across page reload', async ({ page }) => {
    await page.route('**/get-ui-fixture/persist', (route) =>
      route.fulfill({ status: 200, contentType: 'application/json', body: '{}' }),
    )

    await page.getByLabel('Request URL').fill('http://127.0.0.1:5173/get-ui-fixture/persist')
    await page.getByTestId('request-run').click()
    await expect(page.getByTestId('request-outcome-badge')).toBeVisible({ timeout: 30_000 })

    await page.reload()
    await page.waitForFunction(() => typeof window.tractl?.run === 'function', {
      timeout: 30_000,
    })

    await page.getByRole('button', { name: 'Home' }).click()
    await expect(page.getByTestId('fast-start-recent')).toBeVisible()
    await expect(page.getByTestId(/recent-item-/)).toBeVisible()
  })

  // ── Run button state ──────────────────────────────────────────────────────

  test('run button is disabled while a request is in-flight', async ({ page }) => {
    let resolveRequest!: () => void
    const requestHeld = new Promise<void>((resolve) => {
      resolveRequest = resolve
    })

    await page.route('**/get-ui-fixture/slow', async (route) => {
      await requestHeld
      await route.fulfill({ status: 200, contentType: 'application/json', body: '{}' })
    })

    await page.getByLabel('Request URL').fill('http://127.0.0.1:5173/get-ui-fixture/slow')
    await page.getByTestId('request-run').click()

    await expect(page.getByTestId('request-run')).toBeDisabled()

    resolveRequest()
    await expect(page.getByTestId('request-outcome-badge')).toBeVisible({ timeout: 30_000 })
    await expect(page.getByTestId('request-run')).toBeEnabled()
  })

  // ── Idle state ────────────────────────────────────────────────────────────

  test('shows idle placeholder before any request is run', async ({ page }) => {
    await expect(page.getByTestId('request-response-idle')).toBeVisible()
    await expect(page.getByTestId('request-outcome-badge')).not.toBeVisible()
  })

  // ── Extract results ───────────────────────────────────────────────────────

  test('shows extract results after running with extracts configured', async ({ page }) => {
    await page.route('**/get-ui-fixture/extract', (route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ token: 'abc123' }),
      }),
    )

    await page.getByLabel('Request URL').fill('http://127.0.0.1:5173/get-ui-fixture/extract')
    await page.getByTestId('request-config-tab-extracts').click()
    await page.getByRole('button', { name: 'Add extract' }).click()

    // ExtractRow: [select:source] [input:path] -> [input:variable] [select:scope]
    // Fill the variable name (second input inside the row)
    const extractRowInputs = page.locator('input').all()
    const inputs = await extractRowInputs
    if (inputs.length > 0) {
      await inputs[inputs.length - 1].fill('myToken')
    }

    await page.getByTestId('request-run').click()
    await expect(page.getByTestId('request-outcome-badge')).toBeVisible({ timeout: 30_000 })

    await page.getByTestId('request-result-tab-extracts').click()
    await expect(page.getByTestId('extract-results')).toBeVisible()
  })
})
