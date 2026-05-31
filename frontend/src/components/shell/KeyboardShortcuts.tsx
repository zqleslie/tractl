import { useEffect } from 'react'
import { IconX } from '@tabler/icons-react'
import { Button } from '@/components/primitives'
import { useUiStore } from '@/stores/uiStore'

const sections = [
  {
    title: 'General',
    items: [
      ['Cmd K', 'Search and commands'],
      ['Cmd /', 'Keyboard shortcuts'],
      ['Cmd N', 'New request'],
      ['Cmd Shift N', 'New workflow'],
      ['Cmd E', 'Switch environment'],
      ['Cmd ,', 'Settings'],
    ],
  },
  {
    title: 'Request',
    items: [
      ['Cmd Enter', 'Send request'],
      ['Cmd W', 'Close tab'],
      ['Cmd ] / Cmd [', 'Next / previous tab'],
      ['Opt G', 'Select GET method'],
      ['Opt P', 'Select POST method'],
      ['Opt X', 'Select DELETE method'],
    ],
  },
  {
    title: 'Response',
    items: [
      ['Cmd .', 'Copy response body'],
      ['Cmd Shift L', 'Toggle layout'],
    ],
  },
  {
    title: 'Navigation',
    items: [
      ['Cmd H', 'Open run history'],
      ['Opt W', 'Go to workflows'],
    ],
  },
]

export function KeyboardShortcuts() {
  const open = useUiStore((state) => state.keyboardShortcutsOpen)
  const setOpen = useUiStore((state) => state.setKeyboardShortcutsOpen)

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key === '/') {
        event.preventDefault()
        setOpen(true)
      }
      if (event.key === 'Escape') setOpen(false)
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [setOpen])

  if (!open) return null

  return (
    <div
      className="fixed inset-0 z-40 flex items-center justify-center bg-black/20"
      data-testid="keyboard-shortcuts"
      onMouseDown={() => setOpen(false)}
    >
      <div
        className="w-[560px] rounded-card border-[0.5px] border-border bg-surface"
        role="dialog"
        aria-modal="true"
        aria-label="Keyboard shortcuts"
        onMouseDown={(event) => event.stopPropagation()}
      >
        <div className="flex h-10 items-center justify-between border-b-[0.5px] border-border px-3">
          <h2 className="text-ui-sm font-medium text-text">Keyboard shortcuts</h2>
          <Button
            variant="ghost"
            className="h-7 px-1.5"
            aria-label="Close keyboard shortcuts"
            onClick={() => setOpen(false)}
          >
            <IconX size={14} stroke={1.75} />
          </Button>
        </div>
        <div className="grid grid-cols-2 gap-3 p-3">
          {sections.map((section) => (
            <section key={section.title}>
              <h3 className="mb-1 text-[10px] font-medium uppercase tracking-wider text-text-muted">
                {section.title}
              </h3>
              <div className="space-y-1">
                {section.items.map(([keys, label]) => (
                  <div
                    key={`${section.title}-${keys}`}
                    className="flex items-center justify-between gap-3 rounded-ui bg-surface-elevated px-2 py-1.5 text-ui-xs"
                  >
                    <span className="text-text">{label}</span>
                    <kbd className="font-mono text-[10px] text-text-muted">{keys}</kbd>
                  </div>
                ))}
              </div>
            </section>
          ))}
        </div>
      </div>
    </div>
  )
}
