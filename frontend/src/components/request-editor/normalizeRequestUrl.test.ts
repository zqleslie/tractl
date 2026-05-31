import { describe, expect, it } from 'vitest'
import {
  looksLikeViteDevShell,
  normalizeRequestTargetUrl,
  validateRequestRunUrl,
} from '@/components/request-editor/normalizeRequestUrl'

describe('normalizeRequestTargetUrl', () => {
  it('preserves absolute http and https URLs', () => {
    expect(normalizeRequestTargetUrl('https://httpbin.org/get')).toBe(
      'https://httpbin.org/get',
    )
    expect(normalizeRequestTargetUrl('http://127.0.0.1:5173/fixture/get')).toBe(
      'http://127.0.0.1:5173/fixture/get',
    )
  })

  it('prepends https when scheme is missing', () => {
    expect(normalizeRequestTargetUrl('httpbin.org/get')).toBe('https://httpbin.org/get')
  })

  it('leaves relative paths unchanged', () => {
    expect(normalizeRequestTargetUrl('/request-editor-fixture/get')).toBe(
      '/request-editor-fixture/get',
    )
  })
})

describe('validateRequestRunUrl', () => {
  it('rejects relative paths', () => {
    expect(validateRequestRunUrl('/request-editor-fixture/get')).toMatch(/absolute URL/i)
  })

  it('allows dev fixture URLs on the app origin', () => {
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: { origin: 'http://127.0.0.1:5173' },
    })

    expect(
      validateRequestRunUrl('http://127.0.0.1:5173/request-editor-fixture/get'),
    ).toBeNull()
  })

  it('rejects same-origin non-fixture URLs', () => {
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: { origin: 'http://127.0.0.1:5173' },
    })

    expect(validateRequestRunUrl('http://127.0.0.1:5173/')).toMatch(/dev server/i)
  })
})

describe('looksLikeViteDevShell', () => {
  it('detects Vite index.html', () => {
    const html = `<!doctype html><script type="module" src="/@vite/client"></script>`
    expect(looksLikeViteDevShell(html)).toBe(true)
  })

  it('ignores normal JSON bodies', () => {
    expect(looksLikeViteDevShell('{"ok":true}')).toBe(false)
  })
})
