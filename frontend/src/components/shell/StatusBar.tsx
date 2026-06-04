import { StatusChip } from '@/components/primitives'
import { getPlatformCapabilities } from '@/platform'
import { useWorkflowCanvasStore } from '@/stores/workflowCanvasStore'

export function StatusBar() {
  const platform = getPlatformCapabilities()
  const lastRunLabel = useWorkflowCanvasStore((state) => state.lastRunLabel)
  const lastRunTone = useWorkflowCanvasStore((state) => state.lastRunTone)
  const autoSaveState = useWorkflowCanvasStore((state) => state.autoSaveState)

  return (
    <footer
      className="grid h-[22px] shrink-0 grid-cols-3 items-center border-t-[0.5px] border-border bg-surface-sidebar px-3 text-ui-xs"
      data-testid="app-statusbar"
    >
      <StatusChip tone={lastRunTone} label={lastRunLabel} />
      <span className="text-center text-text-muted" data-testid="autosave-state">
        {autoSaveState === 'saving' ? 'Saving…' : '✓ Saved'}
      </span>
      <div className="flex items-center justify-end gap-3 text-text-muted">
        <button
          type="button"
          className="text-ui-xs text-text-muted hover:text-text"
          data-testid="git-indicator"
        >
          git: untracked
        </button>
        <span data-testid="surface-label">{platform.statusLabel}</span>
      </div>
    </footer>
  )
}
