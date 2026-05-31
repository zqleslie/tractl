import { useState, type ReactNode } from 'react'
import { IconChevronDown, IconChevronRight } from '@tabler/icons-react'
import { cn } from '@/lib/cn'

export type SidebarCollapsibleGroupProps = {
  title: string
  children?: ReactNode
  defaultOpen?: boolean
  disabled?: boolean
  suffix?: string
}

export function SidebarCollapsibleGroup({
  title,
  children,
  defaultOpen = true,
  disabled,
  suffix,
}: SidebarCollapsibleGroupProps) {
  const [open, setOpen] = useState(defaultOpen)

  return (
    <div className={cn(disabled && 'opacity-40')}>
      <button
        type="button"
        className="flex w-full items-center gap-1.5 px-3 py-2 text-ui-xs font-medium text-text hover:bg-surface-elevated"
        disabled={disabled}
        onClick={() => !disabled && setOpen((value) => !value)}
      >
        {open ? (
          <IconChevronDown size={11} stroke={2} />
        ) : (
          <IconChevronRight size={11} stroke={2} />
        )}
        {title}
        {suffix ? (
          <span className="ml-auto text-[10px] font-normal text-text-muted">
            {suffix}
          </span>
        ) : null}
      </button>
      {open && !disabled ? children : null}
    </div>
  )
}
