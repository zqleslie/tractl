import { AssertionsPanel } from '@/components/panels/AssertionsPanel'
import { AuthPanel } from '@/components/panels/AuthPanel'
import { BodyPanel } from '@/components/panels/BodyPanel'
import { ExtractsPanel } from '@/components/panels/ExtractsPanel'
import { HeadersPanel } from '@/components/panels/HeadersPanel'
import { ParamsPanel } from '@/components/panels/ParamsPanel'
import { ScriptPanel } from '@/components/request-editor/shared/ScriptPanel'
import { SettingsPanel } from '@/components/panels/SettingsPanel'
import type { RequestEditorController } from '@/components/request-editor/useRequestEditor'
import type { ConfigTab } from '@/components/request-editor/types'

export type RequestConfigTabContentProps = {
  tab: ConfigTab
  editor: RequestEditorController
}

export function RequestConfigTabContent({
  tab,
  editor,
}: RequestConfigTabContentProps) {
  if (tab === 'params') {
    return (
      <ParamsPanel
        value={editor.params.rows}
        onAdd={editor.params.add}
        onUpdate={editor.params.update}
        onRemove={editor.params.remove}
      />
    )
  }

  if (tab === 'headers') {
    return (
      <HeadersPanel
        value={editor.headers.rows}
        onAdd={editor.headers.add}
        onUpdate={editor.headers.update}
        onRemove={editor.headers.remove}
      />
    )
  }

  if (tab === 'body') {
    return (
      <BodyPanel
        value={editor.body.value}
        headers={editor.headers.rows}
        onChange={(body) => editor.body.update(body)}
        onHeadersChange={(headers) => editor.headers.replace(headers)}
      />
    )
  }

  if (tab === 'auth') {
    return <AuthPanel value={editor.auth.value} onChange={(auth) => editor.auth.update(auth)} />
  }

  if (tab === 'pre-script') {
    return (
      <ScriptPanel
        hook="beforeStep"
        value={editor.scripts.value.pre}
        onChange={(pre) => editor.scripts.update({ pre })}
      />
    )
  }

  if (tab === 'post-script') {
    return (
      <ScriptPanel
        hook="afterStep"
        value={editor.scripts.value.post}
        onChange={(post) => editor.scripts.update({ post })}
      />
    )
  }

  if (tab === 'assertions') {
    return (
      <AssertionsPanel
        value={editor.assertions.rows}
        onAdd={editor.assertions.add}
        onUpdate={editor.assertions.update}
        onRemove={editor.assertions.remove}
      />
    )
  }

  if (tab === 'extracts') {
    return (
      <ExtractsPanel
        value={editor.extracts.rows}
        onAdd={editor.extracts.add}
        onUpdate={editor.extracts.update}
        onRemove={editor.extracts.remove}
      />
    )
  }

  return (
    <SettingsPanel
      value={editor.settings.value}
      onChange={(settings) => editor.settings.update(settings)}
    />
  )
}
