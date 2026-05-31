import { useEffect, useMemo, useState } from 'react'
import {
  IconBolt,
  IconClock,
  IconKeyboard,
  IconPlus,
  IconSettings,
  IconTopologyStar3,
} from '@tabler/icons-react'
import { useRunHistoryStore } from '@/stores/runHistoryStore'
import { useUiStore } from '@/stores/uiStore'

type PaletteItem = {
  id: string
  group: 'Recent' | 'Commands'
  name: string
  hint: string
  icon: JSX.Element
  run: () => void
}

export function CommandPalette() {
  const open = useUiStore((state) => state.commandPaletteOpen)
  const setOpen = useUiStore((state) => state.setCommandPaletteOpen)
  const openNewRequest = useUiStore((state) => state.openNewRequest)
  const openWorkflowPlaceholder = useUiStore(
    (state) => state.openWorkflowPlaceholder,
  )
  const setKeyboardShortcutsOpen = useUiStore(
    (state) => state.setKeyboardShortcutsOpen,
  )
  const setActiveScreen = useUiStore((state) => state.setActiveScreen)
  const openHistoryEntry = useUiStore((state) => state.openHistoryEntry)
  const entries = useRunHistoryStore((state) => state.entries)
  const [query, setQuery] = useState('')

  const items = useMemo<PaletteItem[]>(() => {
    const recent = entries.slice(0, 5).map((entry): PaletteItem => ({
      id: `recent-${entry.id}`,
      group: 'Recent',
      name: entry.requestName,
      hint: entry.method,
      icon: <IconClock size={15} stroke={1.75} />,
      run: () => openHistoryEntry(entry),
    }))

    return [
      ...recent,
      {
        id: 'new-request',
        group: 'Commands',
        name: 'New request',
        hint: 'Cmd N',
        icon: <IconPlus size={15} stroke={1.75} />,
        run: openNewRequest,
      },
      {
        id: 'new-workflow',
        group: 'Commands',
        name: 'New workflow',
        hint: 'Cmd Shift N',
        icon: <IconTopologyStar3 size={15} stroke={1.75} />,
        run: openWorkflowPlaceholder,
      },
      {
        id: 'run-history',
        group: 'Commands',
        name: 'Open run history',
        hint: 'Cmd H',
        icon: <IconBolt size={15} stroke={1.75} />,
        run: () => setActiveScreen('run-history'),
      },
      {
        id: 'shortcuts',
        group: 'Commands',
        name: 'Keyboard shortcuts',
        hint: 'Cmd /',
        icon: <IconKeyboard size={15} stroke={1.75} />,
        run: () => setKeyboardShortcutsOpen(true),
      },
      {
        id: 'settings',
        group: 'Commands',
        name: 'Settings',
        hint: 'Cmd ,',
        icon: <IconSettings size={15} stroke={1.75} />,
        run: () => setActiveScreen('environments'),
      },
    ]
  }, [
    entries,
    openHistoryEntry,
    openNewRequest,
    openWorkflowPlaceholder,
    setActiveScreen,
    setKeyboardShortcutsOpen,
  ])

  const filtered = items.filter((item) =>
    `${item.name} ${item.hint}`.toLowerCase().includes(query.toLowerCase()),
  )

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
        event.preventDefault()
        setOpen(true)
      }
      if (event.key === 'Escape') {
        setOpen(false)
      }
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [setOpen])

  if (!open) return null

  return (
    <div
      className="fixed inset-0 z-40 flex items-start justify-center bg-black/20 pt-[12vh]"
      data-testid="command-palette"
      onMouseDown={() => setOpen(false)}
    >
      <div
        className="w-[520px] rounded-card border-[0.5px] border-border bg-surface"
        role="dialog"
        aria-modal="true"
        aria-label="Command palette"
        onMouseDown={(event) => event.stopPropagation()}
      >
        <input
          autoFocus
          className="h-11 w-full border-b-[0.5px] border-border bg-surface px-3 text-ui-sm text-text outline-none"
          placeholder="Search and commands..."
          aria-label="Command search"
          value={query}
          onChange={(event) => setQuery(event.currentTarget.value)}
        />
        <div className="max-h-[320px] overflow-auto p-2">
          <PaletteGroup name="Recent" items={filtered.filter((i) => i.group === 'Recent')} />
          <PaletteGroup name="Commands" items={filtered.filter((i) => i.group === 'Commands')} />
          {filtered.length === 0 ? (
            <p className="px-2 py-4 text-ui-sm text-text-muted">No matches.</p>
          ) : null}
        </div>
      </div>
    </div>
  )
}

function PaletteGroup({ name, items }: { name: string; items: PaletteItem[] }) {
  const setOpen = useUiStore((state) => state.setCommandPaletteOpen)
  if (items.length === 0) return null
  return (
    <section className="mb-2" aria-label={name}>
      <div className="px-2 py-1 text-[10px] font-medium uppercase tracking-wider text-text-muted">
        {name}
      </div>
      {items.map((item) => (
        <button
          key={item.id}
          type="button"
          className="flex w-full items-center gap-2 rounded-ui px-2 py-2 text-left text-ui-sm text-text hover:bg-surface-elevated"
          onClick={() => {
            item.run()
            setOpen(false)
          }}
        >
          {item.icon}
          <span className="min-w-0 flex-1 truncate">{item.name}</span>
          <span className="font-mono text-[10px] text-text-muted">{item.hint}</span>
        </button>
      ))}
    </section>
  )
}
