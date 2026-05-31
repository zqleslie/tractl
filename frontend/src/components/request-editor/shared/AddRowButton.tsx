import { IconPlus } from '@tabler/icons-react'

export type AddRowButtonProps = {
  label: string
  onClick?: () => void
}

export function AddRowButton({ label, onClick }: AddRowButtonProps) {
  return (
    <button
      type="button"
      className="mt-2 flex items-center gap-1.5 rounded-ui px-1 py-1 text-ui-xs text-text-muted hover:bg-surface-elevated hover:text-text"
      onClick={onClick}
    >
      <IconPlus size={13} stroke={1.75} />
      {label}
    </button>
  )
}
