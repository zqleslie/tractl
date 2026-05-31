import type { ButtonHTMLAttributes } from 'react'
import { cn } from '@/lib/cn'

export type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger'

export type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: ButtonVariant
}

const variantClass: Record<ButtonVariant, string> = {
  primary:
    'border-primary bg-primary text-primary-fg hover:opacity-90 disabled:opacity-50',
  secondary:
    'border-border bg-surface-elevated text-text hover:bg-surface disabled:opacity-50',
  ghost:
    'border-transparent bg-transparent text-text hover:bg-surface-elevated disabled:opacity-50',
  danger:
    'border-danger bg-danger-bg text-danger-fg hover:opacity-90 disabled:opacity-50',
}

export function Button({
  className,
  variant = 'secondary',
  type = 'button',
  ...props
}: ButtonProps) {
  return (
    <button
      type={type}
      className={cn(
        'inline-flex items-center justify-center gap-1.5 rounded-ui border-[0.5px] px-3 py-1.5 text-ui-sm font-medium transition-opacity',
        variantClass[variant],
        className,
      )}
      {...props}
    />
  )
}
