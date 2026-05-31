import { ScriptPanel } from '@/components/request-editor/shared/ScriptPanel'
import type { StepDraftEditorSlice } from '@/screens/WorkflowCanvas/StepDetailPanel/useStepDraftEditor'

export function PostScriptTab({ editor }: { editor: StepDraftEditorSlice }) {
  return (
    <ScriptPanel
      hint="Runs after the response — maps to hooks.afterStep"
      code={editor.scripts.value.post}
      onChange={(post) => editor.scripts.update({ post })}
      chips={[
        'ctx.step.response',
        'ctx.step.timing',
        'ctx.step.extracts',
        'tractl.log()',
      ]}
    />
  )
}
