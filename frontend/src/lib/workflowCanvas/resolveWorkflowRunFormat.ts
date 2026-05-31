import type { RunHistorySourceFormat } from '@/stores/runHistoryStore'
import type { TractlWasmParseFormat } from '@/platform/web/wasm/loadTractlWasmRuntime'

/** Map workspace/history format to WASM run/parse format. */
export function resolveWorkflowRunFormat(
  sourceFormat?: RunHistorySourceFormat,
): TractlWasmParseFormat {
  if (sourceFormat === 'json') return 'json'
  if (sourceFormat === 'toon') return 'toon'
  return 'yaml'
}
