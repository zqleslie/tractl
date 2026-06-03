import { useRef, useState, type ReactNode } from 'react'
import { useClickOutside } from '@/hooks/useClickOutside'
import { IconDotsVertical, IconTrash } from '@tabler/icons-react'

type CollectionFolderGroupProps = {
  title: string
  suffix: string
  onDelete: () => void
  children?: ReactNode
}

export function CollectionFolderGroup({
  title,
  suffix,
  onDelete,
  children,
}: CollectionFolderGroupProps) {
  const [open, setOpen] = useState(true)
  const [menuOpen, setMenuOpen] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)

  useClickOutside(menuRef, () => setMenuOpen(false), menuOpen)

  return (
    <div>
      <div className="group/folder flex items-center gap-1.5 px-3 py-2 text-ui-xs font-medium text-text hover:bg-surface-elevated">
        <button
          type="button"
          className="min-w-0 flex-1 truncate text-left"
          onClick={() => setOpen((value) => !value)}
        >
          {open ? '▾' : '▸'} {title}
        </button>
        <span className="text-[10px] font-normal text-text-muted">{suffix}</span>
        <div className="relative" ref={menuRef}>
          <button
            type="button"
            aria-label={`Collection folder menu for ${title}`}
            className="flex h-6 w-6 items-center justify-center rounded-ui text-text-muted opacity-0 hover:bg-surface group-hover/folder:opacity-100"
            onClick={(event) => {
              event.stopPropagation()
              setMenuOpen((value) => !value)
            }}
          >
            <IconDotsVertical size={13} stroke={1.75} />
          </button>
          {menuOpen ? (
            <div className="fixed z-20 min-w-[140px] rounded-ui border-[0.5px] border-border bg-surface py-1 shadow-lg">
              <button
                type="button"
                className="flex w-full items-center gap-2 px-3 py-1.5 text-left text-ui-xs text-danger-fg hover:bg-surface-elevated"
                onClick={() => {
                  onDelete()
                  setMenuOpen(false)
                }}
              >
                <IconTrash size={12} stroke={1.75} />
                Delete folder
              </button>
            </div>
          ) : null}
        </div>
      </div>
      {open ? children : null}
    </div>
  )
}
