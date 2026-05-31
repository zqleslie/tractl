import { useId, useMemo, type ChangeEvent } from 'react'
import { useEnvironmentStore } from '@/stores/environmentStore'

export type EnvironmentVariableInputProps = {
  value: string
  onChange: (value: string) => void
  placeholder?: string
  className?: string
  'aria-label'?: string
  multiline?: boolean
  rows?: number
}

export function EnvironmentVariableInput({
  value,
  onChange,
  placeholder,
  className,
  'aria-label': ariaLabel,
  multiline = false,
  rows = 4,
}: EnvironmentVariableInputProps) {
  const listId = useId()
  const activeEnvironmentId = useEnvironmentStore((state) => state.activeEnvironmentId)
  const environments = useEnvironmentStore((state) => state.environments)

  const variableNames = useMemo(() => {
    const active =
      environments.find((environment) => environment.id === activeEnvironmentId) ??
      null
    return Object.keys(active?.variables ?? {}).sort((left, right) =>
      left.localeCompare(right),
    )
  }, [activeEnvironmentId, environments])

  const sharedProps = {
    'aria-label': ariaLabel,
    placeholder,
    className,
    value,
    onChange: (event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) =>
      onChange(event.currentTarget.value),
    list: variableNames.length > 0 ? listId : undefined,
  }

  return (
    <>
      {multiline ? (
        <textarea {...sharedProps} rows={rows} />
      ) : (
        <input type="text" {...sharedProps} />
      )}
      {variableNames.length > 0 ? (
        <datalist id={listId}>
          {variableNames.map((name) => (
            <option key={name} value={`{{${name}}}`} label={name} />
          ))}
        </datalist>
      ) : null}
    </>
  )
}
