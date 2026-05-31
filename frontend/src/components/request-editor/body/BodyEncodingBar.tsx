import type { BodyEncoding } from '@/components/request-editor/types'
import { bodyEncodings } from '@/components/request-editor/body/bodyEncodingOptions'

type BodyEncodingBarProps = {
  value: BodyEncoding
  onChange: (encoding: BodyEncoding) => void
}

export function BodyEncodingBar({ value, onChange }: BodyEncodingBarProps) {
  return (
    <div
      className="flex flex-wrap items-center gap-2"
      data-testid="body-encoding-bar"
    >
      <label className="flex items-center gap-2 text-ui-xs font-medium text-text-muted">
        Encoding
        <select
          className="rounded-ui border-[0.5px] border-border bg-surface px-2 py-1 text-ui-xs text-text"
          value={value}
          aria-label="Body encoding"
          onChange={(event) => onChange(event.currentTarget.value as BodyEncoding)}
        >
          {bodyEncodings.map((encoding) => (
            <option key={encoding} value={encoding}>
              {encoding}
            </option>
          ))}
        </select>
      </label>
    </div>
  )
}
