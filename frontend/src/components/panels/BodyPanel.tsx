import { BodyContent } from '@/components/request-editor/body/BodyContent'
import { BodyEncodingBar } from '@/components/request-editor/body/BodyEncodingBar'
import { syncContentTypeHeader } from '@/lib/requestEditor/contentTypeHeader'
import type { BodyEncoding, KeyValueRow, RequestBodyDraft } from '@/components/request-editor/types'

export type BodyPanelProps = {
  value: RequestBodyDraft
  onChange: (value: RequestBodyDraft) => void
  headers: KeyValueRow[]
  onHeadersChange: (headers: KeyValueRow[]) => void
  readOnly?: boolean
}

export function BodyPanel({
  value,
  onChange,
  headers,
  onHeadersChange,
  readOnly,
}: BodyPanelProps) {
  const applyBodyChange = (patch: Partial<RequestBodyDraft>) => {
    onChange({ ...value, ...patch })
  }

  const applyContentTypeToHeaders = () => {
    onHeadersChange(syncContentTypeHeader(headers, {}, value))
  }

  return (
    <div className="space-y-[var(--density-gap-md)]">
      <div className="flex flex-wrap items-center gap-2">
        <BodyEncodingBar
          value={value.encoding}
          onChange={(encoding: BodyEncoding) => applyBodyChange({ encoding })}
        />
        {!readOnly ? (
          <button
            type="button"
            className="rounded-ui border-[0.5px] border-border bg-surface-elevated px-2 py-1 text-ui-xs text-text-muted hover:text-text"
            data-testid="apply-content-type-header"
            onClick={applyContentTypeToHeaders}
          >
            Update Content-Type header
          </button>
        ) : null}
      </div>
      <BodyContent
        encoding={value.encoding}
        value={value.value}
        formRows={value.formRows}
        rawContentType={value.rawContentType}
        binaryFile={value.binaryFile}
        readOnly={readOnly}
        onChange={applyBodyChange}
      />
    </div>
  )
}
