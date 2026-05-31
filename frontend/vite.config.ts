import path from 'node:path'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { viteDevFixtures } from './viteDevFixtures'

const desktopDistDir = path.resolve(
  __dirname,
  '../cmd/desktop/frontend/dist',
)

export default defineConfig(({ mode, command }) => {
  const surface = mode === 'desktop' ? 'desktop' : 'web'

  return {
    plugins: [react(), ...(command === 'serve' ? [viteDevFixtures()] : [])],
    server: {
      host: '127.0.0.1',
      port: 5173,
      strictPort: true,
      proxy: {
        '/api': {
          target: 'http://127.0.0.1:7428',
          changeOrigin: true,
        },
      },
    },
    define: {
      __TRACTL_SURFACE__: JSON.stringify(surface),
    },
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
      },
    },
    build: {
      outDir: surface === 'desktop' ? desktopDistDir : path.resolve(__dirname, '../cmd/server/dist/web'),
      emptyOutDir: true,
    },
    test: {
      environment: 'jsdom',
      setupFiles: ['./src/test/setup.ts'],
      globals: true,
      css: true,
      include: ['src/**/*.{test,spec}.{ts,tsx}'],
      exclude: ['e2e/**', 'node_modules/**', 'dist/**'],
    },
  }
})
