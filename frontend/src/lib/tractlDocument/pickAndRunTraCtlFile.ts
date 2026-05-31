import { inferDocumentFormat } from '@/lib/tractlDocument/inferDocumentFormat'
import { pickTextFile, readFileAsText } from '@/lib/tractlDocument/pickTextFile'
import { runTraCtlDocument } from '@/lib/tractlDocument/runDocument'
import type { RunDocumentResult } from '@/lib/tractlDocument/runDocument'
import { parseAndValidateDocument } from '@/lib/tractlDocument/parseValidateDocument'

export type PickRunFileResult =
  | { ok: false; cancelled: true }
  | RunDocumentResult & { cancelled?: false; fileName?: string }

export async function runSelectedTraCtlFile(
  file: File,
  environmentVariables: Record<string, string>,
): Promise<Exclude<PickRunFileResult, { cancelled: true }>> {
  const format = inferDocumentFormat(file.name)
  if (!format) {
    return {
      ok: false,
      stage: 'parse',
      message: 'Unsupported file type. Use YAML, JSON, or TOON.',
      errors: [{ message: 'Unsupported file type. Use YAML, JSON, or TOON.' }],
    }
  }

  const raw = await readFileAsText(file)

  const validated = await parseAndValidateDocument(raw, format)
  if (!validated.ok) {
    return validated
  }

  const stepCount = validated.spec.workflows[0]?.steps?.length ?? 0
  const sourceType = stepCount > 1 ? 'workflow' : 'request'

  const run = await runTraCtlDocument({
    raw,
    format,
    sourceName: file.name,
    sourceType,
    environmentVariables,
  })

  if (!run.ok) {
    return run
  }

  return {
    ...run,
    fileName: file.name,
  }
}

export async function pickAndRunTraCtlFile(
  environmentVariables: Record<string, string>,
): Promise<PickRunFileResult> {
  const file = await pickTextFile()
  if (!file) {
    return { ok: false, cancelled: true }
  }

  return runSelectedTraCtlFile(file, environmentVariables)
}
