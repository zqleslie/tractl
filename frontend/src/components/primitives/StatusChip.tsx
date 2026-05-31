import type { HTMLAttributes } from 'react'
import { cn } from '@/lib/cn'

export type StatusTone = 'ready' | 'running' | 'success' | 'error' | 'muted'

export type StatusChipProps = HTMLAttributes<HTMLSpanElement> & {
  tone?: StatusTone
  label: string
}

const dotClass: Record<StatusTone, string> = {
  ready: 'bg-success',
  running: 'bg-warning',
  success: 'bg-success',
  error: 'bg-danger',
  muted: 'bg-text-muted',
}

export function StatusChip({
  className,
  tone = 'muted',
  label,
  ...props
}: StatusChipProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 text-ui-xs text-text-muted',
        className,
      )}
      {...props}
    >
      <span
        className={cn('h-1.5 w-1.5 shrink-0 rounded-full', dotClass[tone])}
        aria-hidden
      />
      <span>{label}</span>
    </span>
  )
}
