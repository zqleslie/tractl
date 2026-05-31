import { cn } from '@/lib/cn'

export type ResultRowProps = {
  passed: boolean
  label: string
  detail: string
}

export function ResultRow({ passed, label, detail }: ResultRowProps) {
  return (
    <div
      className={cn(
        'rounded-ui p-2',
        passed ? 'bg-success-bg text-success-fg' : 'bg-danger-bg text-danger-fg',
      )}
    >
      <div className="font-medium">{label}</div>
      <div className="text-ui-xs text-text-muted">{detail}</div>
    </div>
  )
}
