import { expect, test } from '@playwright/test'

const validWorkflowYaml = `schemaVersion: 1
capabilities:
  - protocol.http
workflows:
  - id: wf-fast-start
    steps:
      - id: step-get
        kind: request
        request:
          protocol: http
          target: https://httpbin.org/get
          operation: GET
        assertions:
          - id: status-ok
            kind: status
            op: equals
            expected: 200
`

test.describe('Fast Start (S1)', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
    await page.evaluate(() => {
      localStorage.removeItem('tractl-run-history')
    })
    await page.reload()
    await page.waitForFunction(
      () => typeof window.tractl?.run === 'function',
      { timeout: 30_000 },
    )
  })

  test('shows action cards and empty recent state', async ({ page }) => {
    await expect(page.getByTestId('screen-fast-start')).toBeVisible()
    await expect(page.getByTestId('fast-start-new-request')).toBeVisible()
    await expect(page.getByTestId('fast-start-recent-empty')).toBeVisible()
  })

  test('creates a new request in the editor', async ({ page }) => {
    await page.getByTestId('fast-start-new-request').click()
    await expect(page.getByTestId('screen-request')).toBeVisible()
    await expect(page.getByTestId('sidebar-tab-requests')).toHaveAttribute(
      'aria-selected',
      'true',
    )
  })

  test('navigates to run history via View all', async ({ page }) => {
    await page.getByTestId('fast-start-run-history').click()
    await expect(page.getByTestId('screen-run-history')).toBeVisible()
  })

  test('opens workflow placeholder from new workflow card', async ({ page }) => {
    await page.getByTestId('fast-start-new-workflow').click()
    await expect(page.getByTestId('screen-workflow')).toBeVisible()
    await expect(page.getByTestId('sidebar-tab-workflows')).toHaveAttribute(
      'aria-selected',
      'true',
    )
  })

  test('records recent item after running a request', async ({ page }) => {
    await page.getByTestId('fast-start-new-request').click()
    await page.getByLabel('Request URL').fill('https://httpbin.org/get')
    await page.getByRole('button', { name: 'Run' }).click()
    await expect(page.getByTestId('request-outcome-badge')).toBeVisible({
      timeout: 30_000,
    })

    await page.getByRole('button', { name: 'Home' }).click()
    await expect(page.getByTestId('fast-start-recent')).toBeVisible()
    await expect(page.getByTestId(/recent-item-/)).toBeVisible()
  })

  test('runs a traCtl file from Open file', async ({ page }) => {
    await page.getByTestId('fast-start-file-input').setInputFiles({
      name: 'fast-start-run.yaml',
      mimeType: 'text/yaml',
      buffer: Buffer.from(validWorkflowYaml, 'utf-8'),
    })

    await expect(page.getByTestId('screen-run-history')).toBeVisible({
      timeout: 30_000,
    })
  })

  test('sidebar new request and view run history', async ({ page }) => {
    await page.getByTestId('sidebar-new-request').click()
    await expect(page.getByTestId('screen-request')).toBeVisible()

    await page.getByRole('button', { name: 'Home' }).click()
    await expect(page.getByTestId('screen-fast-start')).toBeVisible()

    await page.getByTestId('sidebar-view-run-history').click()
    await expect(page.getByTestId('screen-run-history')).toBeVisible()
  })
})
