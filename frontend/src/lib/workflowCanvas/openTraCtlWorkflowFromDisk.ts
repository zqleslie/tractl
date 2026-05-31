import { inferDocumentFormat } from '@/lib/tractlDocument/inferDocumentFormat'
import type { TraCtlFileActionError } from '@/lib/tractlDocument/openTraCtlFileFromDisk'
import { parseAndValidateDocument } from '@/lib/tractlDocument/parseValidateDocument'
import { readFileAsText } from '@/lib/tractlDocument/pickTextFile'
import { loadWorkflowIntoCanvas } from '@/lib/workflowCanvas/loadWorkflowIntoCanvas'
import { workflowDocumentToCanvasWorkflow } from '@/lib/workflowCanvas/workflowDocumentToCanvasWorkflow'
import { useWorkflowWorkspaceStore } from '@/stores/workflowWorkspaceStore'
import { useUiStore } from '@/stores/uiStore'

/** Parse and validate a workflow file, then open it on the canvas (no execution). */
export async function openTraCtlWorkflowFromDisk(
  file: File,
): Promise<{ ok: true; workflowId: string } | { ok: false; error: TraCtlFileActionError }> {
  const format = inferDocumentFormat(file.name)
  if (!format) {
    return {
      ok: false,
      error: {
        title: 'Could not open file',
        message: 'Unsupported file type. Use YAML, JSON, or TOON.',
        errors: [{ message: 'Unsupported file type. Use YAML, JSON, or TOON.' }],
      },
    }
  }

  try {
    const raw = await readFileAsText(file)
    const validated = await parseAndValidateDocument(raw, format)
    if (!validated.ok) {
      const title =
        validated.stage === 'parse'
          ? 'Could not parse file'
          : validated.stage === 'validate'
            ? 'Validation failed'
            : 'Could not open file'
      return {
        ok: false,
        error: {
          title,
          message: validated.message,
          errors: validated.errors,
        },
      }
    }

    if (validated.spec.workflows.length === 0) {
      return {
        ok: false,
        error: {
          title: 'Could not open file',
          message: 'Document does not contain any workflows.',
          errors: [{ message: 'Document does not contain any workflows.' }],
        },
      }
    }

    const workflow = workflowDocumentToCanvasWorkflow(validated.spec, {
      sourceName: file.name,
      raw,
      lastModified: new Date(file.lastModified).toISOString(),
    })

    const sourceFormat = format === 'yml' ? 'yaml' : format
    useWorkflowWorkspaceStore.getState().upsertWorkflow(workflow, {
      sourceName: file.name,
      sourceFormat,
    })

    const ui = useUiStore.getState()
    ui.setSidebarTab('workflows')
    ui.openWorkflow(workflow.id)
    loadWorkflowIntoCanvas(workflow)

    return { ok: true, workflowId: workflow.id }
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
