import type { ReactNode } from 'react'
import { cn } from '@/lib/cn'

export type ActionCardIconTone = 'request' | 'workflow' | 'file' | 'openapi' | 'history'

const iconToneClass: Record<ActionCardIconTone, string> = {
  request: 'bg-primary-bg text-primary',
  workflow: 'bg-success-bg text-success-fg',
  file: 'bg-[#E1F5EE] text-[#0F6E56] dark:bg-[#0F6E56]/20 dark:text-[#7EC8B8]',
  openapi: 'bg-warning-bg text-warning-fg',
  history: 'bg-surface text-text-muted',
}

export type ActionCardProps = {
  title: string
  description: string
  icon: ReactNode
  iconTone: ActionCardIconTone
  featured?: boolean
  disabled?: boolean
  onClick?: () => void
  testId?: string
  /** Render as a <label> bound to a control so native file dialogs open reliably. */
  htmlFor?: string
}

export function ActionCard({
  title,
  description,
  icon,
  iconTone,
  featured,
  disabled,
  onClick,
  testId,
  htmlFor,
}: ActionCardProps) {
  const className = cn(
    'flex w-full items-start gap-3.5 rounded-card border-[0.5px] border-border bg-surface-elevated p-4 text-left transition-colors hover:border-text-muted',
    featured && 'border-[2px] border-primary',
    disabled && 'cursor-not-allowed opacity-60',
    htmlFor && !disabled && 'cursor-pointer',
  )

  const body = (
    <>
      <span
        className={cn(
          'flex h-[34px] w-[34px] shrink-0 items-center justify-center rounded-ui text-[17px]',
          iconToneClass[iconTone],
        )}
        aria-hidden
      >
        {icon}
      </span>
      <span>
        <span className="block text-ui-sm font-medium text-text">{title}</span>
        <span className="mt-0.5 block text-ui-xs leading-snug text-text-muted">
          {description}
        </span>
      </span>
    </>
  )

  if (htmlFor) {
    return (
      <label htmlFor={htmlFor} data-testid={testId} className={className}>
        {body}
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
      {body}
    </button>
  )
}
