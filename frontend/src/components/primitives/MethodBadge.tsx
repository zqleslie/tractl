import type { HTMLAttributes } from 'react'
import { cn } from '@/lib/cn'

export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
export type MethodBadgeValue = HttpMethod | 'WF'

export type MethodBadgeProps = HTMLAttributes<HTMLSpanElement> & {
  method: MethodBadgeValue
}

const methodClass: Record<MethodBadgeValue, string> = {
  GET: 'method-badge-get',
  POST: 'method-badge-post',
  PUT: 'method-badge-put',
  PATCH: 'method-badge-patch',
  DELETE: 'method-badge-delete',
  WF: 'method-badge-wf',
}

export function MethodBadge({ method, className, ...props }: MethodBadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex shrink-0 rounded-[4px] px-1.5 py-0.5 text-[10px] font-semibold leading-none',
        methodClass[method],
        className,
      )}
      {...props}
    >
      {method}
    </span>
  )
}
