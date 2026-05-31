import type { HTMLAttributes } from 'react'
import { cn } from '@/lib/cn'

export type BadgeTone = 'neutral' | 'info' | 'success' | 'warning' | 'danger'

export type BadgeProps = HTMLAttributes<HTMLSpanElement> & {
  tone?: BadgeTone
}

const toneClass: Record<BadgeTone, string> = {
  neutral: 'border-border bg-surface-elevated text-text-muted',
  info: 'border-primary bg-primary-bg text-primary',
  success: 'border-success bg-success-bg text-success-fg',
  warning: 'border-warning bg-warning-bg text-warning-fg',
  danger: 'border-danger bg-danger-bg text-danger-fg',
}

export function Badge({
  className,
  tone = 'neutral',
  ...props
}: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-ui border-[0.5px] px-1.5 py-0.5 text-ui-xs font-medium',
        toneClass[tone],
        className,
      )}
      {...props}
    />
  )
}
