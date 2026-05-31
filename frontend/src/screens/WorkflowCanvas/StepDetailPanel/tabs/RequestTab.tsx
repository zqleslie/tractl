import { AddRowButton } from '@/components/request-editor/shared/AddRowButton'
import { KeyValueTable } from '@/components/request-editor/shared/KeyValueTable'
import {
  LabeledInput,
  LabeledSelect,
} from '@/components/request-editor/shared/LabeledField'
import type { AuthType, BodyEncoding } from '@/components/request-editor/types'
import type { HttpMethod } from '@/components/primitives'
import { VariableReferenceHints } from '@/components/request-editor/shared/VariableReferenceHints'
import type { StepDraftEditorSlice } from '@/screens/WorkflowCanvas/StepDetailPanel/useStepDraftEditor'

const HTTP_METHODS: HttpMethod[] = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE']

export interface RequestTabProps {
  editor: StepDraftEditorSlice
}

function SectionHeading({ children }: { children: string }) {
  return (
    <h3 className="mb-2 mt-4 text-ui-xs font-medium uppercase tracking-wide text-text-muted first:mt-0">
      {children}
    </h3>
  )
}

export function RequestTab({ editor }: RequestTabProps) {
  const showTokenRef = editor.auth.value.type !== 'None'

  return (
    <div>
      <SectionHeading>URL</SectionHeading>
      <div className="flex items-center gap-2">
        <select
          aria-label="HTTP method"
          className="rounded-ui border-[0.5px] border-border bg-surface px-2 py-1 text-ui-xs"
          value={editor.method}
          onChange={(event) =>
            editor.setMethod(event.currentTarget.value as HttpMethod)
          }
        >
          {HTTP_METHODS.map((option) => (
            <option key={option} value={option}>
              {option}
            </option>
          ))}
        </select>
        <div className="min-w-0 flex-1">
          <input
            aria-label="Request URL"
            type="text"
            className="w-full rounded-ui border-[0.5px] border-border bg-surface px-2 py-1 font-mono text-ui-xs text-text outline-none focus:border-primary"
            value={editor.url}
            onChange={(event) => editor.setUrl(event.currentTarget.value)}
          />
          <VariableReferenceHints value={editor.url} />
        </div>
      </div>

      <SectionHeading>Parameters</SectionHeading>
      <KeyValueTable
        rows={editor.params.rows}
        onUpdate={editor.params.update}
        onRemove={editor.params.remove}
      />
      <AddRowButton label="Add parameter" onClick={editor.params.add} />

      <SectionHeading>Headers</SectionHeading>
      <KeyValueTable
        rows={editor.headers.rows}
        onUpdate={editor.headers.update}
        onRemove={editor.headers.remove}
      />
      <AddRowButton label="Add header" onClick={editor.headers.add} />

      <SectionHeading>Body</SectionHeading>
      <div className="space-y-2">
        <LabeledSelect
          label="Encoding"
          options={['JSON', 'Form data', 'Raw', 'None']}
          value={editor.body.value.encoding}
          onChange={(encoding) =>
            editor.body.update({ encoding: encoding as BodyEncoding })
          }
        />
        <textarea
          aria-label="Request body"
          className="min-h-[132px] w-full resize-y rounded-ui border-[0.5px] border-border bg-surface-elevated p-3 font-mono text-ui-xs leading-6 text-text outline-none focus:border-primary"
          value={editor.body.value.value}
          onChange={(event) =>
            editor.body.update({ value: event.currentTarget.value })
          }
        />
        <VariableReferenceHints value={editor.body.value.value} />
      </div>

      <SectionHeading>Auth</SectionHeading>
      <div className="space-y-2">
        <LabeledSelect
          label="Type"
          options={['None', 'Bearer token', 'Basic auth', 'API key']}
          value={editor.auth.value.type}
          onChange={(type) => editor.auth.update({ type: type as AuthType })}
        />
        {showTokenRef ? (
          <LabeledInput
            label="Token ref"
            value={editor.auth.value.tokenRef}
            onChange={(tokenRef) => editor.auth.update({ tokenRef })}
          />
        ) : null}
      </div>
    </div>
  )
}
