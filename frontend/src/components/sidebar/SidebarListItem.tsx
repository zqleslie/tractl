import type { ReactNode } from 'react'
import { useEffect, useRef, useState } from 'react'
import { IconPlayerPlay } from '@tabler/icons-react'
import { Button } from '@/components/primitives'
import { cn } from '@/lib/cn'

export type ContextMenuItem = {
  label: string
  icon?: ReactNode
  danger?: boolean
  onClick: () => void
}

export type SidebarListItemProps = {
  icon: ReactNode
  label: ReactNode
  active?: boolean
  muted?: boolean
  onClick?: () => void
  onRun?: () => void
  contextMenu?: ContextMenuItem[]
  testId?: string
}

export function SidebarListItem({
  icon,
  label,
  active,
  muted,
  onClick,
  onRun,
  contextMenu,
  testId,
}: SidebarListItemProps) {
  const [menuPos, setMenuPos] = useState<{ x: number; y: number } | null>(null)
  const menuRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!menuPos) return
    const close = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuPos(null)
      }
    }
    const closeKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setMenuPos(null)
    }
    document.addEventListener('mousedown', close)
    document.addEventListener('keydown', closeKey)
    return () => {
      document.removeEventListener('mousedown', close)
      document.removeEventListener('keydown', closeKey)
    }
  }, [menuPos])

  const handleContextMenu = (e: React.MouseEvent) => {
    if (!contextMenu?.length) return
    e.preventDefault()
    setMenuPos({ x: e.clientX, y: e.clientY })
  }

  return (
    <>
      <div
        className={cn(
          'group relative flex cursor-pointer items-center gap-2 rounded-ui px-2 py-1.5 text-ui-sm',
          active ? 'bg-surface-elevated text-text' : 'text-text hover:bg-surface-elevated',
        )}
        data-testid={testId}
        onClick={onClick}
        onContextMenu={handleContextMenu}
        onKeyDown={(event) => {
          if (event.key === 'Enter' || event.key === ' ') {
            event.preventDefault()
            onClick?.()
          }
        }}
        role="button"
        tabIndex={0}
      >
        <span className="flex w-4 shrink-0 justify-center text-text-muted">{icon}</span>
        <span
          className={cn(
            'min-w-0 flex-1 truncate',
            muted && 'text-text-muted',
          )}
        >
          {label}
        </span>
        {onRun ? (
          <Button
            variant="ghost"
            className="absolute right-1 h-6 w-6 p-0 opacity-0 group-hover:opacity-100"
            aria-label="Run"
            onClick={(event) => {
              event.stopPropagation()
              onRun()
            }}
          >
            <IconPlayerPlay size={12} stroke={2} />
          </Button>
        ) : null}
      </div>

      {menuPos && contextMenu?.length ? (
        <div
          ref={menuRef}
          className="fixed z-50 min-w-[140px] rounded-ui border-[0.5px] border-border bg-surface py-1 shadow-lg"
          style={{ left: menuPos.x, top: menuPos.y }}
        >
          {contextMenu.map((item) => (
            <button
              key={item.label}
              type="button"
              className={cn(
                'flex w-full items-center gap-2 px-3 py-1.5 text-left text-ui-xs hover:bg-surface-elevated',
                item.danger ? 'text-danger-fg' : 'text-text',
              )}
              onClick={() => {
                item.onClick()
                setMenuPos(null)
              }}
            >
              {item.icon && (
                <span className="shrink-0 text-current">{item.icon}</span>
              )}
              {item.label}
            </button>
          ))}
        </div>
      ) : null}
    </>
  )
}
