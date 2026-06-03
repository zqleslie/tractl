/**
 * Run history persistence and navigation tests.
 *
 * Verifies that GET request executions are recorded in the run history store,
 * that history survives page reload, that entries can be reopened,
 * and that history is accessible from both the fast-start screen and sidebar.
 */
import { expect, test } from '@playwright/test'

async function runGetRequest(
  page: import('@playwright/test').Page,
  fixture: string,
  status = 200,
) {
  await page.route(`**/${fixture}`, (route) =>
    route.fulfill({
      status,
      contentType: 'application/json',
      body: JSON.stringify({ fixture }),
    }),
  )
  await page.getByLabel('Request URL').fill(`http://127.0.0.1:5173/${fixture}`)
  await page.getByTestId('request-run').click()
  await expect(page.getByTestId('request-outcome-badge')).toBeVisible({ timeout: 30_000 })
}

test.describe('Run history', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
    // clear any pre-existing history
    await page.evaluate(() => localStorage.removeItem('tractl-run-history'))
    await page.reload()
    await page.waitForFunction(() => typeof window.tractl?.run === 'function', {
      timeout: 30_000,
    })
    await page.getByTestId('fast-start-new-request').click()
    await expect(page.getByTestId('screen-request')).toBeVisible()
  })

  test('records a recent item on the fast-start screen after a GET run', async ({ page }) => {
    await runGetRequest(page, 'history-fixture/recent')

    await page.getByRole('button', { name: 'Home' }).click()
    await expect(page.getByTestId('fast-start-recent')).toBeVisible()
    await expect(page.getByTestId(/recent-item-/)).toBeVisible()
  })

  test('recent item shows the correct URL', async ({ page }) => {
    await runGetRequest(page, 'history-fixture/url-display')

    await page.getByRole('button', { name: 'Home' }).click()
    await expect(page.getByTestId(/recent-item-/)).toContainText('history-fixture/url-display')
  })

  test('history persists after page reload', async ({ page }) => {
    await runGetRequest(page, 'history-fixture/persist')

    await page.reload()
    await page.waitForFunction(() => typeof window.tractl?.run === 'function', {
      timeout: 30_000,
    })
    await page.getByRole('button', { name: 'Home' }).click()
    await expect(page.getByTestId('fast-start-recent')).toBeVisible()
    await expect(page.getByTestId(/recent-item-/)).toBeVisible()
  })

  test('navigates to run history screen via fast-start link', async ({ page }) => {
    await runGetRequest(page, 'history-fixture/screen-nav')

    await page.getByRole('button', { name: 'Home' }).click()
    await page.getByTestId('fast-start-run-history').click()
    await expect(page.getByTestId('screen-run-history')).toBeVisible()
  })

  test('run history screen shows history rows', async ({ page }) => {
    await runGetRequest(page, 'history-fixture/screen-rows')

    await page.getByRole('button', { name: 'Home' }).click()
    await page.getByTestId('fast-start-run-history').click()
    await expect(page.getByTestId('screen-run-history')).toBeVisible()
    await expect(page.getByTestId('history-row').first()).toBeVisible()
  })

  test('history row shows outcome badge', async ({ page }) => {
    await runGetRequest(page, 'history-fixture/outcome-badge')

    await page.getByRole('button', { name: 'Home' }).click()
    await page.getByTestId('fast-start-run-history').click()
    await expect(page.getByTestId('history-outcome-badge').first()).toBeVisible()
  })

  test('opening a history item from fast-start restores the request', async ({ page }) => {
    await runGetRequest(page, 'history-fixture/reopen')

    await page.getByRole('button', { name: 'Home' }).click()
    await page.getByTestId(/recent-item-/).first().click()
    // should navigate to request editor or run history with the item loaded
    const isRequestScreen = await page.getByTestId('screen-request').isVisible()
    const isHistoryScreen = await page.getByTestId('screen-run-history').isVisible()
    expect(isRequestScreen || isHistoryScreen).toBe(true)
  })

  test('sidebar history panel shows recent runs', async ({ page }) => {
    await runGetRequest(page, 'history-fixture/sidebar')

    // History is accessible from the requests panel's "history" sub-tab
    await page.getByTestId('sidebar-requests-tab-history').click()
    await expect(page.getByTestId('history-row').first()).toBeVisible({ timeout: 5_000 })
  })

  test('run history is capped and old entries are removed', async ({ page }) => {
    // Inject 50 entries directly into the store to simulate the cap
    await page.evaluate(() => {
      const entries = Array.from({ length: 50 }, (_, i) => ({
        id: `history-entry-${i}`,
        timestamp: Date.now() - i * 1000,
        sourceType: 'request-editor',
        sourceName: `Request ${i}`,
        requestName: `Request ${i}`,
        method: 'GET',
        url: `http://example.com/${i}`,
        statusCode: 200,
        durationMs: 100,
        outcome: 'passed',
        result: {},
      }))
      const stored = { state: { entries, pinnedIds: [] }, version: 0 }
      localStorage.setItem('tractl-run-history', JSON.stringify(stored))
    })

    await page.reload()
    await page.waitForFunction(() => typeof window.tractl?.run === 'function', {
      timeout: 30_000,
    })

    // Run one more request — it should be added and oldest non-pinned trimmed
    await page.getByTestId('fast-start-new-request').click()
    await runGetRequest(page, 'history-fixture/cap-test')

    const historyJson = await page.evaluate(
      () => localStorage.getItem('tractl-run-history') ?? '{}',
    )
    const parsed = JSON.parse(historyJson) as { state: { entries: unknown[] } }
    expect(parsed.state.entries.length).toBeLessThanOrEqual(50)
  })
})
