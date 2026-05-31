import type { Plugin } from 'vite'

/**
 * Dev-only HTTP fixtures for manual browser testing and e2e parity.
 * Playwright tests may still mock these routes; without mocks, Vite would
 * return the SPA shell (index.html) for unknown paths.
 */
export function viteDevFixtures(): Plugin {
  return {
    name: 'tractl-dev-fixtures',
    configureServer(server) {
      server.middlewares.use((req, res, next) => {
        const pathname = (req.url ?? '').split('?')[0] ?? ''

        if (pathname === '/request-editor-fixture/get') {
          res.statusCode = 200
          res.setHeader('Content-Type', 'application/json')
          res.end(JSON.stringify({ ok: true }))
          return
        }

        if (pathname === '/wasm-run-fixture/get') {
          res.statusCode = 200
          res.setHeader('Content-Type', 'application/json')
          res.end(JSON.stringify({ ok: true }))
          return
        }

        if (pathname === '/wasm-run-fixture/status') {
          res.statusCode = 404
          res.setHeader('Content-Type', 'text/plain')
          res.end('not found')
          return
        }

        next()
      })
    },
  }
}
