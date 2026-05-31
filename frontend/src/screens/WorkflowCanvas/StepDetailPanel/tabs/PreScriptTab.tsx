import { ScriptPanel } from '@/components/request-editor/shared/ScriptPanel'
import type { StepDraftEditorSlice } from '@/screens/WorkflowCanvas/StepDetailPanel/useStepDraftEditor'

export function PreScriptTab({ editor }: { editor: StepDraftEditorSlice }) {
  return (
    <ScriptPanel
      hint="Runs before the request — maps to hooks.beforeStep"
      code={editor.scripts.value.pre}
      onChange={(pre) => editor.scripts.update({ pre })}
      chips={[
        'ctx.spec',
        'ctx.workflow',
        'ctx.step',
        'ctx.environment',
        'ctx.runtime',
        'tractl.now()',
        'tractl.log()',
      ]}
    />
  )
}
