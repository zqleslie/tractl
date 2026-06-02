const ABSOLUTE_HTTP_PATTERN = /^https?:\/\//i

export function normalizeRequestTargetUrl(url: string): string {
  const trimmed = url.trim()
  if (!trimmed) return ''
  if (ABSOLUTE_HTTP_PATTERN.test(trimmed)) return trimmed
  if (trimmed.startsWith('/')) return trimmed
  return `https://${trimmed}`
}

export function validateRequestRunUrl(url: string): string | null {
  const trimmed = url.trim()
  if (!trimmed) return 'Enter a request URL'

  if (trimmed.startsWith('/')) {
    return 'Use an absolute URL (https://…). Relative paths resolve to this app and return the Vite page instead of an API response.'
  }

  const normalized = normalizeRequestTargetUrl(trimmed)

  let parsed: URL
  try {
    parsed = new URL(normalized)
  } catch {
    return 'Enter a valid absolute URL'
  }

  if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') {
    return 'Request URL must use http:// or https://'
  }

  if (typeof window !== 'undefined' && parsed.origin === window.location.origin) {
    const isDevFixture = /\/(?:request-editor|wasm-run)-fixture\//.test(parsed.pathname)
    if (!isDevFixture) {
      return `This URL points at the traCtl dev server (${parsed.origin}). Use an external API such as https://httpbin.org/get.`
    }
  }

  return null
}

export function looksLikeViteDevShell(body: string | undefined | null): boolean {
  if (!body) return false
  const sample = body.slice(0, 2048).toLowerCase()
  return (
    sample.includes('<!doctype html>') &&
    (sample.includes('/@vite/client') || sample.includes('react-refresh'))
  )
}
