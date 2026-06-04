import { IconX } from '@tabler/icons-react'
import { Button } from '@/components/primitives'
import { SidebarEnvsPanel } from '@/components/sidebar/SidebarEnvsPanel'
import { SidebarRequestsPanel } from '@/components/sidebar/SidebarRequestsPanel'
import { SidebarWorkflowsPanel } from '@/components/sidebar/SidebarWorkflowsPanel'
import { type SidebarTab, useUiStore } from '@/stores/uiStore'

const titles: Record<SidebarTab, string> = {
  requests: 'Requests',
  workflows: 'Workflows',
  overlays: 'Overlays',
  history: 'History',
  envs: 'Environments',
  settings: 'Settings',
}

export function SidePanel() {
  const sidebarTab = useUiStore((state) => state.sidebarTab)
  const collapsed = useUiStore((state) => state.sidebarCollapsed)
  const setSidebarCollapsed = useUiStore((state) => state.setSidebarCollapsed)
  const setActiveScreen = useUiStore((state) => state.setActiveScreen)

  if (collapsed) return null

  const isRequestsPanel = sidebarTab === 'requests'

  return (
    <aside
      className="flex w-[260px] shrink-0 flex-col border-r-[0.5px] border-border bg-surface-sidebar"
      data-testid="side-panel"
    >
      {!isRequestsPanel ? (
        <div className="flex h-10 shrink-0 items-center justify-between border-b-[0.5px] border-border px-3">
          <span className="text-[10px] font-medium uppercase tracking-wider text-text-muted">
            {titles[sidebarTab]}
          </span>
          <Button
            variant="ghost"
            className="h-7 px-1.5"
            aria-label="Close side panel"
            data-testid="side-panel-close"
            onClick={() => setSidebarCollapsed(true)}
          >
            <IconX size={14} stroke={1.75} />
          </Button>
        </div>
      ) : (
        <div className="flex h-9 shrink-0 items-center justify-end border-b-[0.5px] border-border px-2">
          <Button
            variant="ghost"
            className="h-7 px-1.5"
            aria-label="Close side panel"
            data-testid="side-panel-close"
            onClick={() => setSidebarCollapsed(true)}
          >
            <IconX size={14} stroke={1.75} />
          </Button>
        </div>
      )}
      <div
        className={
          isRequestsPanel
            ? 'flex min-h-0 flex-1 flex-col overflow-hidden'
            : 'min-h-0 flex-1 overflow-auto px-1 py-2 text-ui-xs'
        }
      >
        <SidePanelContent
          tab={sidebarTab}
          openScreen={(screen) => setActiveScreen(screen)}
        />
      </div>
    </aside>
  )
}

function SidePanelContent({
  tab,
  openScreen,
}: {
  tab: SidebarTab
  openScreen: (screen: 'run-history' | 'environments') => void
}) {
  if (tab === 'requests') return <SidebarRequestsPanel />
  if (tab === 'workflows') return <SidebarWorkflowsPanel />
  if (tab === 'envs') return <SidebarEnvsPanel />
  if (tab === 'history') {
    return (
      <div className="space-y-2 px-2" data-testid="side-panel-history">
        <button
          type="button"
          className="w-full rounded-ui border-[0.5px] border-border bg-surface-elevated px-2 py-2 text-left text-ui-sm text-text hover:border-primary"
          onClick={() => openScreen('run-history')}
        >
          Open run history
        </button>
      </div>
    )
  }
  if (tab === 'overlays') {
    return (
      <p className="px-2 py-2 text-ui-xs text-text-muted" data-testid="side-panel-overlays">
        Overlay files will appear here.
      </p>
    )
  }
  return (
    <div className="space-y-2 px-2" data-testid="side-panel-settings">
      <button
        type="button"
        className="w-full rounded-ui border-[0.5px] border-border bg-surface-elevated px-2 py-2 text-left text-ui-sm text-text hover:border-primary"
        onClick={() => openScreen('environments')}
      >
        Environment settings
      </button>
    </div>
  )
}
