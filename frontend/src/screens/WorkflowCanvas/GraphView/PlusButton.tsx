import { IconPlus } from '@tabler/icons-react'
import { cn } from '@/lib/cn'

export interface PlusButtonProps {
  x: number
  y: number
  onClick: () => void
  alwaysVisible?: boolean
  /** Tailwind class enabling reveal via a parent named group, e.g. group-hover/connector:opacity-100. */
  revealClassName?: string
  title?: string
}

export function PlusButton({
  x,
  y,
  onClick,
  alwaysVisible = false,
  revealClassName,
  title = 'Add step',
}: PlusButtonProps) {
  return (
    <button
      type="button"
      title={title}
      aria-label={title}
      onClick={onClick}
      style={{ left: x, top: y }}
      className={cn(
        'absolute z-10 flex h-4 w-4 -translate-x-1/2 -translate-y-1/2 items-center justify-center',
        'rounded-full border border-[0.5px] border-slate-300 bg-white text-slate-400 transition-opacity',
        'hover:border-[#185FA5] hover:text-[#185FA5]',
        'dark:border-slate-600 dark:bg-[#16171d] dark:text-slate-500 dark:hover:border-[#B5D4F4] dark:hover:text-[#B5D4F4]',
        alwaysVisible ? 'opacity-100' : 'opacity-0 hover:opacity-100',
        !alwaysVisible && revealClassName,
      )}
    >
      <IconPlus size={10} stroke={1.5} />
    </button>
  )
}
