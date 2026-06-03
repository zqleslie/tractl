/**
 * Local API GET endpoint integration tests.
 *
 * Tests the Go HTTP server at http://127.0.0.1:7428 directly using
 * Playwright's `request` fixture. Run these only when the desktop local
 * API server is available (TRACTL_API_URL env var must be set, or the
 * server must be running at the default address).
 *
 * Skip these in web-only CI by checking for the server via the status endpoint.
 *
 * Usage:
 *   TRACTL_API_URL=http://127.0.0.1:7428 npx playwright test localapi-get-endpoints
 */
import { expect, test } from '@playwright/test'

const API_BASE = process.env.TRACTL_API_URL ?? 'http://127.0.0.1:7428'

// Skip all tests if the API is not reachable
test.beforeEach(async ({ request }) => {
  try {
    const res = await request.get(`${API_BASE}/api/v1/status`, { timeout: 2_000 })
    if (!res.ok()) {
      test.skip(true, 'Local API server not reachable — skipping')
    }
  } catch {
    test.skip(true, 'Local API server not reachable — skipping')
  }
})

// ── GET /api/v1/status ────────────────────────────────────────────────────

test.describe('GET /api/v1/status', () => {
  test('returns 200 with ok:true and version string', async ({ request }) => {
    const res = await request.get(`${API_BASE}/api/v1/status`)
    expect(res.status()).toBe(200)

    const body = await res.json()
    expect(body).toMatchObject({ ok: true })
    expect(typeof body.version).toBe('string')
    expect(body.version.length).toBeGreaterThan(0)
  })

  test('returns application/json content-type', async ({ request }) => {
    const res = await request.get(`${API_BASE}/api/v1/status`)
    expect(res.headers()['content-type']).toContain('application/json')
  })

  test('rejects non-GET methods with 405', async ({ request }) => {
    const res = await request.post(`${API_BASE}/api/v1/status`, { data: {} })
    expect(res.status()).toBe(405)
  })

  test('responds to OPTIONS preflight with CORS headers', async ({ request }) => {
    const res = await request.fetch(`${API_BASE}/api/v1/status`, {
      method: 'OPTIONS',
      headers: {
        Origin: 'http://localhost:5173',
        'Access-Control-Request-Method': 'GET',
      },
    })
    expect(res.status()).toBe(204)
    expect(res.headers()['access-control-allow-origin']).toBe('http://localhost:5173')
  })
})

// ── GET /api/v1/engine/defaults ───────────────────────────────────────────

test.describe('GET /api/v1/engine/defaults', () => {
  test('returns canonical engine defaults shape', async ({ request }) => {
    const res = await request.get(`${API_BASE}/api/v1/engine/defaults`)
    expect(res.status()).toBe(200)

    const body = await res.json()
    expect(body).toMatchObject({
      failurePolicy: 'resilient',
      timeoutMs: 30_000,
      concurrency: 4,
      httpSuccessMin: 200,
      httpSuccessMax: 399,
      retry: {
        maxAttempts: 3,
        backoffFactor: 2,
      },
    })
  })

  test('rejects POST with 405', async ({ request }) => {
    const res = await request.post(`${API_BASE}/api/v1/engine/defaults`, { data: {} })
    expect(res.status()).toBe(405)
  })

  test('retry strategies array is present and non-empty', async ({ request }) => {
    const res = await request.get(`${API_BASE}/api/v1/engine/defaults`)
    const body = await res.json()
    expect(Array.isArray(body.retry.strategies)).toBe(true)
    expect(body.retry.strategies.length).toBeGreaterThan(0)
    expect(body.retry.strategies).toContain('none')
  })
})

// ── GET /api/v1/files ─────────────────────────────────────────────────────

test.describe('GET /api/v1/files', () => {
  test('returns an array of file entries', async ({ request }) => {
    const res = await request.get(`${API_BASE}/api/v1/files`)
    expect(res.status()).toBe(200)

    const body = await res.json()
    expect(Array.isArray(body)).toBe(true)
  })

  test('returns only .yaml files', async ({ request }) => {
    const res = await request.get(`${API_BASE}/api/v1/files`)
    const body = (await res.json()) as Array<{ path: string }>
    for (const entry of body) {
      expect(entry.path).toMatch(/\.yaml$/)
    }
  })

  test('each entry has path, name, modifiedAt, size fields', async ({ request }) => {
    const res = await request.get(`${API_BASE}/api/v1/files`)
    const body = (await res.json()) as Array<unknown>
    if (body.length === 0) {
      test.skip(true, 'workspace is empty — skipping field structure check')
    }
    const entry = body[0] as Record<string, unknown>
    expect(typeof entry.path).toBe('string')
    expect(typeof entry.name).toBe('string')
    expect(typeof entry.modifiedAt).toBe('string')
    expect(typeof entry.size).toBe('number')
  })

  test('prefix filter returns only matching files', async ({ request }) => {
    const res = await request.get(`${API_BASE}/api/v1/files?prefix=requests/`)
    const body = (await res.json()) as Array<{ path: string }>
    for (const entry of body) {
      expect(entry.path).toMatch(/^requests\//)
    }
  })

  test('prefix filter with no matches returns empty array', async ({ request }) => {
    const res = await request.get(`${API_BASE}/api/v1/files?prefix=nonexistent-prefix-xyz/`)
    expect(res.status()).toBe(200)
    const body = await res.json()
    expect(body).toEqual([])
  })

  test('rejects POST with 405', async ({ request }) => {
    const res = await request.post(`${API_BASE}/api/v1/files`, { data: {} })
    expect(res.status()).toBe(405)
  })

  test('responds to CORS preflight from localhost', async ({ request }) => {
    const res = await request.fetch(`${API_BASE}/api/v1/files`, {
      method: 'OPTIONS',
      headers: {
        Origin: 'http://localhost:5173',
        'Access-Control-Request-Method': 'GET',
      },
    })
    expect(res.status()).toBe(204)
    expect(res.headers()['access-control-allow-origin']).toBe('http://localhost:5173')
  })
})

// ── GET /api/v1/files/{path} ──────────────────────────────────────────────

test.describe('GET /api/v1/files/{path}', () => {
  test('returns file content for a valid path', async ({ request }) => {
    const res = await request.get(`${API_BASE}/api/v1/files/requests/sample.yaml`)
    // May 404 if workspace isn't the test dir; skip if so
    if (res.status() === 404) {
      test.skip(true, 'sample.yaml not present in active workspace — skipping')
    }
    expect(res.status()).toBe(200)
    const body = await res.json()
    expect(typeof body.path).toBe('string')
    expect(typeof body.content).toBe('string')
  })

  test('returns 404 for a non-existent file', async ({ request }) => {
    const res = await request.get(
      `${API_BASE}/api/v1/files/requests/does-not-exist-xyz.yaml`,
    )
    expect(res.status()).toBe(404)
    const body = await res.json()
    expect(typeof body.error).toBe('string')
  })

  test('rejects path traversal attempt with 400', async ({ request }) => {
    const traversal = encodeURIComponent('../../../etc/passwd')
    const res = await request.get(`${API_BASE}/api/v1/files/${traversal}`)
    expect(res.status()).toBe(400)
    const body = await res.json()
    expect(body.error).toMatch(/path escapes workspace|relative file path required/i)
  })

  test('rejects double-encoded traversal attempt with 400', async ({ request }) => {
    // %2F = /, %2E%2E = ..
    const res = await request.get(`${API_BASE}/api/v1/files/..%2F..%2Fetc%2Fpasswd`)
    expect([400, 404]).toContain(res.status())
  })

  test('rejects absolute path with 400', async ({ request }) => {
    const res = await request.get(`${API_BASE}/api/v1/files//etc/passwd`)
    expect([400, 404]).toContain(res.status())
  })

  test('rejects empty path segment gracefully', async ({ request }) => {
    // An empty path-value results in no segment — handled by resolvePath
    const res = await request.get(`${API_BASE}/api/v1/files/`)
    // Could 404 (no handler match) or 400; either is acceptable
    expect([400, 404, 405]).toContain(res.status())
  })
})

// ── GET /api/v1/workspace/status ─────────────────────────────────────────

test.describe('GET /api/v1/workspace/status', () => {
  test('returns workspace status with rootDir', async ({ request }) => {
    const res = await request.get(`${API_BASE}/api/v1/workspace/status`)
    expect(res.status()).toBe(200)
    const body = await res.json()
    expect(typeof body.rootDir).toBe('string')
    expect(body.rootDir.length).toBeGreaterThan(0)
  })

  test('isGitRepo is a boolean', async ({ request }) => {
    const res = await request.get(`${API_BASE}/api/v1/workspace/status`)
    const body = await res.json()
    expect(typeof body.isGitRepo).toBe('boolean')
  })

  test('branch is a string when isGitRepo is true', async ({ request }) => {
    const res = await request.get(`${API_BASE}/api/v1/workspace/status`)
    const body = await res.json()
    if (body.isGitRepo) {
      expect(typeof body.branch).toBe('string')
    }
  })

  test('dirtyFiles is a non-negative integer', async ({ request }) => {
    const res = await request.get(`${API_BASE}/api/v1/workspace/status`)
    const body = await res.json()
    if (body.isGitRepo) {
      expect(typeof body.dirtyFiles).toBe('number')
      expect(body.dirtyFiles).toBeGreaterThanOrEqual(0)
    }
  })

  test('rejects POST with 405', async ({ request }) => {
    const res = await request.post(`${API_BASE}/api/v1/workspace/status`, { data: {} })
    expect(res.status()).toBe(405)
  })
})

// ── CORS ──────────────────────────────────────────────────────────────────

test.describe('CORS middleware', () => {
  test('allows requests from http://localhost origin', async ({ request }) => {
    const res = await request.get(`${API_BASE}/api/v1/status`, {
      headers: { Origin: 'http://localhost:5173' },
    })
    expect(res.headers()['access-control-allow-origin']).toBe('http://localhost:5173')
  })

  test('allows requests from http://127.0.0.1 origin', async ({ request }) => {
    const res = await request.get(`${API_BASE}/api/v1/status`, {
      headers: { Origin: 'http://127.0.0.1:5173' },
    })
    expect(res.headers()['access-control-allow-origin']).toBe('http://127.0.0.1:5173')
  })

  test('allows wails://wails.localhost origin (desktop)', async ({ request }) => {
    const res = await request.get(`${API_BASE}/api/v1/status`, {
      headers: { Origin: 'wails://wails.localhost' },
    })
    expect(res.headers()['access-control-allow-origin']).toBe('wails://wails.localhost')
  })

  test('does not set CORS headers for untrusted origins', async ({ request }) => {
    const res = await request.get(`${API_BASE}/api/v1/status`, {
      headers: { Origin: 'https://evil.example.com' },
    })
    expect(res.headers()['access-control-allow-origin']).toBeUndefined()
  })

  test('does not set CORS headers when Origin is absent', async ({ request }) => {
    const res = await request.get(`${API_BASE}/api/v1/status`)
    expect(res.headers()['access-control-allow-origin']).toBeUndefined()
  })
})
