import { expect, test } from '@playwright/test'

test.describe('Request editor (S2)', () => {
  test('edits draft and runs through browser WASM runtime', async ({ page }) => {
    await page.route('**/request-editor-fixture/get', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ ok: true }),
      })
    })

    await page.goto('/')
    await page.waitForFunction(
      () => typeof window.tractl?.run === 'function',
      { timeout: 30_000 },
    )

    await page.getByTestId('fast-start-new-request').click()
    await expect(page.getByTestId('screen-request')).toBeVisible()

    await page
      .getByLabel('Request URL')
      .fill('http://127.0.0.1:5173/request-editor-fixture/get')
    await expect(page.getByText('Draft')).toBeVisible()

    await page.getByTestId('request-config-tab-assertions').click()
    await page.getByRole('button', { name: 'Add assertion' }).click()

    await page.getByTestId('request-run').click()
    await expect(page.getByTestId('request-outcome-badge')).toHaveText('1/1 passed', {
      timeout: 30_000,
    })
    await expect(page.getByTestId('request-response-body')).toBeVisible()
    await expect(page.getByText('200 OK')).toBeVisible()
    await expect(page.getByText('{"ok":true}')).toBeVisible()
  })

  test('resolves active environment variables before running', async ({ page }) => {
    await page.addInitScript(() => {
      localStorage.setItem(
        'tractl-environments',
        JSON.stringify({
          state: {
            environments: [
              {
                id: 'env-development',
                name: 'Development',
                variables: {
                  baseUrl: 'http://127.0.0.1:5173/request-editor-fixture/dev',
                },
              },
              {
                id: 'env-staging',
                name: 'Staging',
                variables: {
                  baseUrl:
                    'http://127.0.0.1:5173/request-editor-fixture/staging',
                },
              },
            ],
            activeEnvironmentId: 'env-development',
          },
          version: 0,
        }),
      )
    })

    await page.route('**/request-editor-fixture/dev/get', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ env: 'development' }),
      })
    })
    await page.route('**/request-editor-fixture/staging/get', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ env: 'staging' }),
      })
    })

    await page.goto('/')
    await page.waitForFunction(
      () => typeof window.tractl?.run === 'function',
      { timeout: 30_000 },
    )

    await expect(page.getByTestId('env-badge')).toHaveText('Development')
    await page.getByTestId('fast-start-new-request').click()
    await page.getByLabel('Request URL').fill('{{baseUrl}}/get')
    await page.getByTestId('request-run').click()
    await expect(page.getByText('{"env":"development"}')).toBeVisible({
      timeout: 30_000,
    })

    await page.getByTestId('env-badge').click()
    await page.getByRole('menuitem', { name: /Staging/ }).click()
    await expect(page.getByTestId('env-badge')).toHaveText('Staging')
    await page.getByTestId('request-run').click()
    await expect(page.getByText('{"env":"staging"}')).toBeVisible({
      timeout: 30_000,
    })
  })

  test('reports missing environment variables without running', async ({ page }) => {
    await page.goto('/')
    await page.getByTestId('fast-start-new-request').click()
    await page.getByLabel('Request URL').fill('{{missing}}/get')
    await page.getByTestId('request-run').click()

    await expect(
      page.getByText(
        'TRACTL_ENV_VARIABLE_MISSING: Missing environment variable: missing',
      ),
    ).toBeVisible()
  })
})
