import type { ButtonHTMLAttributes } from 'react'
import { cn } from '@/lib/cn'
import type { TractlEnvironment } from '@/stores/environmentStore'
import { isDangerousEnvironment } from '@/components/primitives/envPillModel'

type EnvPillProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  environment: TractlEnvironment | null
}

export function EnvPill({ environment, className, ...props }: EnvPillProps) {
  const dangerous = isDangerousEnvironment(environment)

  return (
    <button
      type="button"
      className={cn(
        'inline-flex h-7 items-center gap-1.5 rounded-ui border-[1.5px] px-2 text-ui-xs font-medium',
        environment
          ? dangerous
            ? 'env-pill-danger'
            : 'env-pill-standard'
          : 'border-border bg-surface-elevated text-text-muted',
        className,
      )}
      data-dangerous={dangerous ? 'true' : 'false'}
      data-testid="env-pill"
      {...props}
    >
      <span
        className={cn(
          'h-1.5 w-1.5 rounded-full',
          environment
            ? dangerous
              ? 'bg-danger-fg'
              : 'bg-primary'
            : 'bg-text-muted',
        )}
      />
      {environment?.name ?? 'No environment'}
    </button>
  )
}
