export type AddRowProps = {
  label: string
  onClick?: () => void
}

export function AddRow({ label, onClick }: AddRowProps) {
  return (
    <button
      type="button"
      className="mt-2 w-full rounded-ui border-[0.5px] border-dashed border-border px-[var(--density-padding-sm)] py-[var(--density-padding-xs)] text-left text-[length:var(--density-font-label)] text-text-muted hover:border-primary hover:text-text"
      onClick={onClick}
    >
      + {label}
    </button>
  )
}
