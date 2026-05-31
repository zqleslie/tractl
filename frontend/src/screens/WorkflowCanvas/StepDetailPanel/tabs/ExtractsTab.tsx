import { AddRowButton } from '@/components/request-editor/shared/AddRowButton'
import { ExtractRow } from '@/components/request-editor/shared/ExtractRow'
import type { StepDraftEditorSlice } from '@/screens/WorkflowCanvas/StepDetailPanel/useStepDraftEditor'

export function ExtractsTab({ editor }: { editor: StepDraftEditorSlice }) {
  const primaryExtract = editor.extracts.rows[0]

  return (
    <div className="space-y-2">
      {editor.extracts.rows.map((row) => (
        <ExtractRow
          key={row.id}
          row={row}
          onChange={(patch) => editor.extracts.update(row.id, patch)}
          onRemove={() => editor.extracts.remove(row.id)}
        />
      ))}
      {primaryExtract ? (
        <p className="rounded-ui border-[0.5px] border-border bg-surface-elevated p-2 text-ui-xs text-text-muted">
          Downstream reference:{' '}
          {primaryExtract.variable.trim()
            ? `\${steps.this.extracts.${primaryExtract.variable}}`
            : '(set a variable name)'}
        </p>
      ) : null}
      <p className="rounded-ui border-[0.5px] border-info bg-info-bg p-2 text-ui-xs text-info-fg">
        Extracted values are written to workflow scope and available to downstream steps.
      </p>
      <AddRowButton label="Add extract" onClick={editor.extracts.add} />
    </div>
  )
}
