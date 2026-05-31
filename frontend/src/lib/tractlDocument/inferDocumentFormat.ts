import type { RunHistorySourceFormat } from '@/stores/runHistoryStore'
import type { TractlWasmParseFormat } from '@/platform/web/wasm/loadTractlWasmRuntime'

export function inferDocumentFormat(fileName: string): TractlWasmParseFormat | null {
  const lower = fileName.toLowerCase()
  if (lower.endsWith('.yaml') || lower.endsWith('.yml')) return 'yaml'
  if (lower.endsWith('.json')) return 'json'
  if (lower.endsWith('.toon')) return 'toon'
  return null
}

export function toHistorySourceFormat(
  format: TractlWasmParseFormat,
): RunHistorySourceFormat {
  return format === 'yml' ? 'yaml' : format
}
