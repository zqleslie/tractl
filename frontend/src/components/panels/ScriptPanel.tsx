import {
  ScriptPanel as BaseScriptPanel,
  type ScriptPanelProps as BaseScriptPanelProps,
} from '@/components/request-editor/shared/ScriptPanel'

export type ScriptPanelProps = BaseScriptPanelProps

export function ScriptPanel(props: ScriptPanelProps) {
  return <BaseScriptPanel {...props} />
}
