import { recordRunHistoryEntry } from '@/lib/runHistory/recordRunHistoryEntry'
import { runSelectedTraCtlFile } from '@/lib/tractlDocument/pickAndRunTraCtlFile'
import type { DocumentValidationIssue } from '@/lib/tractlDocument/parseValidateDocument'
import { useRunHistoryStore } from '@/stores/runHistoryStore'
import { useUiStore } from '@/stores/uiStore'

export type TraCtlFileActionError = {
  title: string
  message: string
  errors: DocumentValidationIssue[]
}

export const TRACTL_FILE_INPUT_ACCEPT =
  '.yaml,.yml,.json,.toon,application/json,text/yaml,text/x-yaml'

function errorTitle(stage: string | undefined): string {
  if (stage === 'parse') return 'Could not parse file'
  if (stage === 'validate') return 'Validation failed'
  return 'Could not run file'
}

/** Parse, validate, run a picked file, record history, and navigate to the result screen. */
export async function openTraCtlFileFromDisk(
  file: File,
  environmentVariables: Record<string, string>,
): Promise<{ ok: true } | { ok: false; error: TraCtlFileActionError }> {
  try {
    const outcome = await runSelectedTraCtlFile(file, environmentVariables)

    if (!outcome.ok) {
      return {
        ok: false,
        error: {
          title: errorTitle(outcome.stage),
          message: outcome.message,
          errors: outcome.errors,
        },
      }
    }

    recordRunHistoryEntry({
      requestName: outcome.summary.requestName,
      method: outcome.summary.method,
      url: outcome.summary.url,
      result: outcome.result,
      sourceType: outcome.summary.sourceType,
      sourceName: outcome.fileName,
      sourceFormat: outcome.sourceFormat,
      document: outcome.spec,
    })

    if (outcome.summary.sourceType === 'request') {
      const entry = useRunHistoryStore.getState().entries[0]
      if (entry) useUiStore.getState().openHistoryEntry(entry)
    } else {
      useUiStore.getState().setActiveScreen('run-history')
    }

    return { ok: true }
  } catch (error) {
    const message =
      error instanceof Error ? error.message : 'Could not open selected file'
    return {
      ok: false,
      error: {
        title: 'Could not open file',
        message,
        errors: [{ message }],
      },
    }
  }
}
