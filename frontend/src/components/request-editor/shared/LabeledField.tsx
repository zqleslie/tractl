import { VariableReferenceHints } from '@/components/request-editor/shared/VariableReferenceHints'

export type LabeledInputProps = {
  label: string
  value: string
  onChange: (value: string) => void
}

export function LabeledInput({ label, value, onChange }: LabeledInputProps) {
  return (
    <label className="flex items-start gap-3 text-ui-sm">
      <span className="w-24 shrink-0 pt-1.5 text-text-muted">{label}</span>
      <div className="min-w-0 flex-1">
        <input
          className="w-full rounded-ui border-[0.5px] border-border bg-surface px-2 py-1.5 font-mono text-ui-xs text-text"
          placeholder={label === 'Token ref' ? '${secrets.env.name}' : undefined}
          value={value}
          onChange={(event) => onChange(event.currentTarget.value)}
        />
        <VariableReferenceHints value={value} />
      </div>
    </label>
  )
}

export type LabeledSelectProps = {
  label: string
  options: string[]
  value: string
  onChange: (value: string) => void
}

export function LabeledSelect({
  label,
  options,
  value,
  onChange,
}: LabeledSelectProps) {
  return (
    <label className="flex items-center gap-3 text-ui-sm">
      <span className="w-24 shrink-0 text-text-muted">{label}</span>
      <select
        className="min-w-0 flex-1 rounded-ui border-[0.5px] border-border bg-surface px-2 py-1.5 text-text"
        value={value}
        onChange={(event) => onChange(event.currentTarget.value)}
      >
        {options.map((option) => (
          <option key={option} value={option}>
            {option}
          </option>
        ))}
      </select>
    </label>
  )
}
