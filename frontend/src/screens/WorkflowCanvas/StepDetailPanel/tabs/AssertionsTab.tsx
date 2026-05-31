import { AddRowButton } from '@/components/request-editor/shared/AddRowButton'
import { AssertionRow } from '@/components/request-editor/shared/AssertionRow'
import type { StepDraftEditorSlice } from '@/screens/WorkflowCanvas/StepDetailPanel/useStepDraftEditor'

export function AssertionsTab({ editor }: { editor: StepDraftEditorSlice }) {
  return (
    <div className="space-y-2">
      {editor.assertions.rows.map((row) => (
        <AssertionRow
          key={row.id}
          row={row}
          onChange={(patch) => editor.assertions.update(row.id, patch)}
          onRemove={() => editor.assertions.remove(row.id)}
        />
      ))}
      <AddRowButton label="Add assertion" onClick={editor.assertions.add} />
    </div>
  )
}
