import { expect, test } from '@playwright/test'

test.describe('App shell', () => {
  test('renders sidebar, status bar, and fast start', async ({ page }) => {
    await page.goto('/')
    await expect(page.getByTestId('app-shell')).toBeVisible()
    await expect(page.getByTestId('app-sidebar')).toBeVisible()
    await expect(page.getByTestId('app-statusbar')).toBeVisible()
    await expect(page.getByTestId('screen-fast-start')).toBeVisible()
    await expect(page.getByText('Engine ready')).toBeVisible()
    await expect(page.getByTestId('web-desktop-banner')).toBeVisible()
    await expect(page.getByTestId('surface-label')).toHaveText(
      'Web · v0.1.0-alpha',
    )
  })

  test('switches theme via titlebar toggle', async ({ page }) => {
    await page.goto('/')
    const html = page.locator('html')
    await expect(html).toHaveAttribute('data-theme', 'light')
    await page.getByTestId('theme-toggle').click()
    await expect(html).toHaveAttribute('data-theme', 'dark')
  })

  test('navigates from fast start to request placeholder', async ({ page }) => {
    await page.goto('/')
    await page.getByTestId('fast-start-new-request').click()
    await expect(page.getByTestId('screen-request')).toBeVisible()
    await expect(page.getByTestId('sidebar-tab-requests')).toHaveAttribute(
      'aria-selected',
      'true',
    )
  })
})
