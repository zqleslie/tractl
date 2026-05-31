import type { ReactNode } from 'react'
import { cn } from '@/lib/cn'

export type SidebarActionProps = {
  children: ReactNode
  primary?: boolean
  disabled?: boolean
  onClick?: () => void
  testId?: string
  /** Render as <label htmlFor="…"> so file inputs open via native label activation. */
  htmlFor?: string
}

export function SidebarAction({
  children,
  primary,
  disabled,
  onClick,
  testId,
  htmlFor,
}: SidebarActionProps) {
  const className = cn(
    'mb-px flex w-full items-center gap-1.5 rounded-ui px-2 py-1.5 text-ui-sm text-text-muted hover:bg-surface-elevated hover:text-text',
    primary && 'font-medium text-text',
    disabled && 'cursor-not-allowed opacity-60',
    htmlFor && !disabled && 'cursor-pointer',
  )

  if (htmlFor) {
    return (
      <label htmlFor={htmlFor} data-testid={testId} className={className}>
        {children}
      </label>
    )
  }

  return (
    <button
      type="button"
      data-testid={testId}
      disabled={disabled}
      className={className}
      onClick={onClick}
    >
      {children}
    </button>
  )
}
