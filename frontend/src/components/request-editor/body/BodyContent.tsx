import { useMemo, useRef, useState } from 'react'
import { IconUpload } from '@tabler/icons-react'
import { EmptyState } from '@/components/common/EmptyState'
import { EnvironmentVariableInput } from '@/components/environment/EnvironmentVariableInput'
import type { BodyEncoding, FormRow } from '@/components/request-editor/types'

type BodyContentProps = {
  encoding: BodyEncoding
  value: string
  formRows: FormRow[]
  rawContentType: string
  binaryFile: string
  readOnly?: boolean
  onChange: (patch: {
    value?: string
    formRows?: FormRow[]
    rawContentType?: string
    binaryFile?: string
  }) => void
}

export function BodyContent({
  encoding,
  value,
  formRows,
  rawContentType,
  binaryFile,
  readOnly,
  onChange,
}: BodyContentProps) {
  const fileInputRef = useRef<HTMLInputElement>(null)

  if (encoding === 'JSON') {
    return (
      <JsonBodyEditor
        value={value}
        readOnly={readOnly}
        onChange={(next) => onChange({ value: next })}
      />
    )
  }

  if (encoding === 'Form data' || encoding === 'Multipart') {
    return (
      <div className="space-y-[var(--density-gap-sm)]" data-testid={`body-content-${encoding}`}>
        {encoding === 'Form data' ? (
          <p className="text-[length:var(--density-font-label)] text-text-muted">
            URL-encoded fields only. Use Multipart when uploading files.
          </p>
        ) : null}
        {encoding === 'Multipart' ? (
          <p className="rounded-ui border-[0.5px] border-info bg-info-bg p-2 text-[length:var(--density-font-label)] text-info-fg">
            Multipart sends each field as a separate part — use this for file
            uploads alongside text fields.
          </p>
        ) : null}
        <FormDataTable
          rows={formRows}
          readOnly={readOnly}
          onChange={(rows) => onChange({ formRows: rows })}
        />
      </div>
    )
  }

  if (encoding === 'Raw') {
    return (
      <div className="space-y-[var(--density-gap-sm)]" data-testid="body-content-Raw">
        <label className="flex items-center gap-2 text-[length:var(--density-font-label)] text-text-muted">
          Content-Type
          <select
            className="rounded-ui border-[0.5px] border-border bg-surface px-2 py-1 text-text"
            value={rawContentType}
            disabled={readOnly}
            onChange={(event) =>
              onChange({ rawContentType: event.currentTarget.value })
            }
          >
            <option value="text/plain">text/plain</option>
            <option value="text/xml">text/xml</option>
            <option value="application/xml">application/xml</option>
            <option value="text/html">text/html</option>
          </select>
        </label>
        <textarea
          aria-label="Raw request body"
          readOnly={readOnly}
          className="min-h-[156px] w-full resize-y rounded-ui border-[0.5px] border-border bg-surface-elevated p-[var(--density-padding-md)] font-mono text-[length:var(--density-font-mono)] leading-[var(--density-line-height)] text-text outline-none focus:border-primary"
          value={value}
          onChange={(event) => onChange({ value: event.currentTarget.value })}
        />
      </div>
    )
  }

  if (encoding === 'Binary') {
    return (
      <div data-testid="body-content-Binary">
        <input
          ref={fileInputRef}
          type="file"
          className="hidden"
          onChange={(event) => {
            const file = event.currentTarget.files?.[0]
            if (file) onChange({ binaryFile: file.name })
          }}
        />
        <button
          type="button"
          className="flex min-h-[156px] w-full flex-col items-center justify-center rounded-ui border-[0.5px] border-dashed border-border bg-surface-elevated text-center"
          onClick={() => fileInputRef.current?.click()}
        >
          <IconUpload size={22} stroke={1.75} className="text-text-muted" />
          <p className="mt-2 text-[length:var(--density-font-ui)] text-text">
            Drop a file here, or click to choose
          </p>
          <p className="mt-1 max-w-sm text-[length:var(--density-font-label)] text-text-muted">
            Binary sends one file as the entire request body (not mixed fields).
          </p>
          {binaryFile ? (
            <p className="mt-1 font-mono text-[length:var(--density-font-label)] text-text-muted">
              {binaryFile}
            </p>
          ) : null}
        </button>
      </div>
    )
  }

  return (
    <div data-testid="body-content-None" className="py-8">
      <EmptyState icon="file-off" text="No body sent with this request" />
    </div>
  )
}

function JsonBodyEditor({
  value,
  readOnly,
  onChange,
}: {
  value: string
  readOnly?: boolean
  onChange: (value: string) => void
}) {
  const [view, setView] = useState<'raw' | 'parsed'>('raw')

  const parsedPreview = useMemo(() => {
    if (!value.trim()) return ''
    try {
      return JSON.stringify(JSON.parse(value), null, 2)
    } catch {
      return 'Invalid JSON — switch to Raw to edit.'
    }
  }, [value])

  return (
    <div className="space-y-2" data-testid="body-content-JSON">
      <div className="flex gap-1">
        {(['raw', 'parsed'] as const).map((mode) => (
          <button
            key={mode}
            type="button"
            className={
              view === mode
                ? 'rounded-ui bg-primary-bg px-2 py-0.5 text-ui-xs text-primary'
                : 'rounded-ui px-2 py-0.5 text-ui-xs text-text-muted hover:text-text'
            }
            onClick={() => setView(mode)}
          >
            {mode === 'raw' ? 'Raw' : 'Parsed'}
          </button>
        ))}
      </div>
      {view === 'raw' ? (
        <EnvironmentVariableInput
          aria-label="Request body"
          placeholder='{ "name": "{{userName}}" }'
          className="min-h-[156px] w-full resize-y rounded-ui border-[0.5px] border-border bg-surface-elevated p-[var(--density-padding-md)] font-mono text-[length:var(--density-font-mono)] leading-[var(--density-line-height)] text-text outline-none focus:border-primary"
          multiline
          rows={7}
          value={value}
          onChange={onChange}
        />
      ) : (
        <pre
          aria-label="Parsed JSON preview"
          className="min-h-[156px] overflow-auto rounded-ui border-[0.5px] border-border bg-surface-elevated p-[var(--density-padding-md)] font-mono text-[length:var(--density-font-mono)] leading-[var(--density-line-height)] text-text"
        >
          {parsedPreview}
        </pre>
      )}
    </div>
  )
}

function FormDataTable({
  rows,
  readOnly,
  onChange,
}: {
  rows: FormRow[]
  readOnly?: boolean
  onChange: (rows: FormRow[]) => void
}) {
  return (
    <div className="overflow-hidden rounded-ui border-[0.5px] border-border">
      <div
        className="grid grid-cols-[28px_1fr_1fr_84px_32px] border-b-[0.5px] border-border bg-surface-elevated text-[10px] uppercase tracking-[0.05em] text-text-muted"
        style={{ height: 'var(--density-row-height)' }}
      >
        <span className="px-2">On</span>
        <span className="px-2">Key</span>
        <span className="px-2">Value / file</span>
        <span className="px-2">Type</span>
        <span />
      </div>
      {rows.map((row) => (
        <div
          key={row.id}
          className="grid grid-cols-[28px_1fr_1fr_84px_32px] items-center border-b-[0.5px] border-border text-[length:var(--density-font-mono)]"
          style={{ height: 'var(--density-row-height)' }}
        >
          <input
            type="checkbox"
            className="mx-2"
            checked={row.enabled}
            disabled={readOnly}
            onChange={(event) =>
              onChange(
                rows.map((candidate) =>
                  candidate.id === row.id
                    ? { ...candidate, enabled: event.currentTarget.checked }
                    : candidate,
                ),
              )
            }
          />
          <input
            className="border-x-[0.5px] border-border bg-surface px-2 outline-none"
            value={row.key}
            readOnly={readOnly}
            onChange={(event) =>
              onChange(
                rows.map((candidate) =>
                  candidate.id === row.id
                    ? { ...candidate, key: event.currentTarget.value }
                    : candidate,
                ),
              )
            }
          />
          {row.type === 'file' ? (
            <button
              type="button"
              className="border-r-[0.5px] border-border bg-surface px-2 text-left text-[length:var(--density-font-label)]"
              onClick={() => {
                const input = document.createElement('input')
                input.type = 'file'
                input.onchange = () => {
                  const file = input.files?.[0]
                  if (!file) return
                  onChange(
                    rows.map((candidate) =>
                      candidate.id === row.id
                        ? {
                            ...candidate,
                            filename: file.name,
                            value: file.name,
                          }
                        : candidate,
                    ),
                  )
                }
                input.click()
              }}
            >
              {row.filename ? row.filename : 'Choose file'}
            </button>
          ) : (
            <input
              className="border-r-[0.5px] border-border bg-surface px-2 outline-none"
              value={row.value}
              readOnly={readOnly}
              onChange={(event) =>
                onChange(
                  rows.map((candidate) =>
                    candidate.id === row.id
                      ? { ...candidate, value: event.currentTarget.value }
                      : candidate,
                  ),
                )
              }
            />
          )}
          <select
            className="border-r-[0.5px] border-border bg-surface px-1 outline-none"
            value={row.type}
            disabled={readOnly}
            onChange={(event) =>
              onChange(
                rows.map((candidate) =>
                  candidate.id === row.id
                    ? {
                        ...candidate,
                        type: event.currentTarget.value as FormRow['type'],
                      }
                    : candidate,
                ),
              )
            }
          >
            <option value="text">Text</option>
            <option value="file">File</option>
          </select>
          <button
            type="button"
            className="text-text-muted hover:text-danger-fg"
            onClick={() => onChange(rows.filter((candidate) => candidate.id !== row.id))}
          >
            ×
          </button>
        </div>
      ))}
      <button
        type="button"
        className="w-full px-2 py-2 text-left text-[length:var(--density-font-label)] text-text-muted hover:text-text"
        onClick={() =>
          onChange([
            ...rows,
            {
              id: `field-${Date.now()}`,
              enabled: true,
              key: '',
              value: '',
              type: 'text',
            },
          ])
        }
      >
        + Add field
      </button>
    </div>
  )
}
