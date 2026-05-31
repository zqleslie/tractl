import { cn } from '@/lib/cn'

export type SeverityBadgeProps = {
  severity: 'error' | 'warning'
  onToggle?: () => void
}

export function SeverityBadge({ severity, onToggle }: SeverityBadgeProps) {
  const content = (
    <span
      className={cn(
        'inline-flex rounded px-1.5 py-0.5 text-[10px] font-medium capitalize',
        severity === 'error'
          ? 'bg-danger-bg text-danger-fg'
          : 'bg-warning-bg text-warning-fg',
      )}
    >
      {severity}
    </span>
  )

  if (!onToggle) return content

  return (
    <button type="button" className="shrink-0" onClick={onToggle}>
      {content}
    </button>
  )
}
