import { useRef, useState, type ReactNode } from 'react'
import { useClickOutside } from '@/hooks/useClickOutside'
import { IconPlayerPlay } from '@tabler/icons-react'
import { Button } from '@/components/primitives'
import { MethodBadge } from '@/components/primitives/MethodBadge'
import { cn } from '@/lib/cn'
import type { RunHistoryEntry } from '@/stores/runHistoryStore'

type RowMenuItem = {
  label: string
  icon?: ReactNode
  danger?: boolean
  onClick: () => void
}

type RequestSidebarRowProps = {
  method: RunHistoryEntry['method']
  title: string
  meta?: string
  statusCode?: number
  testId: string
  onClick: () => void
  onRun?: () => void
  contextMenu?: RowMenuItem[]
}

export function RequestSidebarRow({
  method,
  title,
  meta,
  statusCode,
  testId,
  onClick,
  onRun,
  contextMenu,
}: RequestSidebarRowProps) {
  const [menuPos, setMenuPos] = useState<{ x: number; y: number } | null>(null)
  const menuRef = useRef<HTMLDivElement>(null)

  useClickOutside(menuRef, () => setMenuPos(null), menuPos !== null)

  return (
    <>
      <div
        className="group relative flex h-9 cursor-pointer items-center gap-2 border-b-[0.5px] border-border px-2 text-left hover:bg-surface-elevated"
        data-testid={testId}
        role="button"
        tabIndex={0}
        onClick={onClick}
        onContextMenu={(event) => {
          if (!contextMenu?.length) return
          event.preventDefault()
          setMenuPos({ x: event.clientX, y: event.clientY })
        }}
        onKeyDown={(event) => {
          if (event.key !== 'Enter' && event.key !== ' ') return
          event.preventDefault()
          onClick()
        }}
      >
        <MethodBadge method={method} className="min-w-[42px] justify-center" />
        <span className="min-w-0 flex-1 truncate text-ui-sm font-medium text-text">
          {title}
        </span>
        {statusCode ? (
          <span
            className={cn(
              'h-1.5 w-1.5 shrink-0 rounded-full',
              statusCode >= 200 && statusCode < 400
                ? 'bg-success-fg'
                : 'bg-danger-fg',
            )}
            aria-hidden="true"
          />
        ) : null}
        {meta ? (
          <span className="shrink-0 text-[11px] text-text-muted">{meta}</span>
        ) : null}
        {onRun ? (
          <Button
            variant="ghost"
            className="h-7 w-7 p-0 opacity-0 transition-opacity group-hover:opacity-100"
            aria-label="Run"
            onClick={(event) => {
              event.stopPropagation()
              onRun()
            }}
          >
            <IconPlayerPlay size={13} stroke={2} />
          </Button>
        ) : null}
      </div>

      {menuPos && contextMenu?.length ? (
        <div
          ref={menuRef}
          className="fixed z-50 min-w-[160px] rounded-ui border-[0.5px] border-border bg-surface py-1 shadow-lg"
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
              {item.icon ? (
                <span className="shrink-0 text-current">{item.icon}</span>
              ) : null}
              {item.label}
            </button>
          ))}
        </div>
      ) : null}
    </>
  )
}
