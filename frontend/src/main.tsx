import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import App from '@/app/App'
import { Providers } from '@/app/providers'
import { detectSurface } from '@/platform'
import { loadTractlWasmRuntime } from '@/platform/web/wasm/loadTractlWasmRuntime'
import '@/index.css'

if (detectSurface() === 'web') {
  void loadTractlWasmRuntime()
}

const root = document.getElementById('root')
if (!root) {
  throw new Error('Root element #root not found')
}

createRoot(root).render(
  <StrictMode>
    <Providers>
      <App />
    </Providers>
  </StrictMode>,
)
