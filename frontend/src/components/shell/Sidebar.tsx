import { IconHome, IconSettings } from '@tabler/icons-react'
import { Button } from '@/components/primitives'
import { SidebarEnvsPanel } from '@/components/sidebar/SidebarEnvsPanel'
import { SidebarRequestsPanel } from '@/components/sidebar/SidebarRequestsPanel'
import { SidebarWorkflowsPanel } from '@/components/sidebar/SidebarWorkflowsPanel'
import { cn } from '@/lib/cn'
import { type SidebarTab, useUiStore } from '@/stores/uiStore'

const tabs: Array<{ id: SidebarTab; label: string }> = [
  { id: 'requests', label: 'Requests' },
  { id: 'workflows', label: 'Workflows' },
  { id: 'envs', label: 'Envs' },
]

type SidebarProps = {
  onNavigateHome: () => void
}

export function Sidebar({ onNavigateHome }: SidebarProps) {
  const sidebarTab = useUiStore((s) => s.sidebarTab)
  const setSidebarTab = useUiStore((s) => s.setSidebarTab)
  const collapsed = useUiStore((s) => s.sidebarCollapsed)

  if (collapsed) {
    return null
  }

  return (
    <aside
      className="flex w-[216px] shrink-0 flex-col border-r-[0.5px] border-border bg-surface"
      data-testid="app-sidebar"
    >
      <div className="flex items-center gap-2 border-b-[0.5px] border-border px-3 py-2">
        <Button
          variant="ghost"
          className="px-2"
          aria-label="Home"
          onClick={onNavigateHome}
        >
          <IconHome size={16} stroke={1.75} />
        </Button>
        <span className="text-ui-sm font-medium text-text">traCtl</span>
      </div>

      <div
        className="flex border-b-[0.5px] border-border"
        role="tablist"
        aria-label="Sidebar sections"
      >
        {tabs.map((tab) => (
          <button
            key={tab.id}
            type="button"
            role="tab"
            aria-selected={sidebarTab === tab.id}
            data-testid={`sidebar-tab-${tab.id}`}
            className={cn(
              'flex-1 border-b-2 px-2 py-2 text-ui-xs font-medium',
              sidebarTab === tab.id
                ? 'border-primary text-primary'
                : 'border-transparent text-text-muted hover:text-text',
            )}
            onClick={() => setSidebarTab(tab.id)}
          >
            {tab.label}
          </button>
        ))}
      </div>

      <div className="flex-1 overflow-auto px-1 py-2 text-ui-xs">
        <SidebarTabPanel tab={sidebarTab} />
      </div>

      <div className="border-t-[0.5px] border-border p-2">
        <Button variant="ghost" className="w-full justify-start gap-2">
          <IconSettings size={16} stroke={1.75} />
          Settings
        </Button>
      </div>
    </aside>
  )
}

function SidebarTabPanel({ tab }: { tab: SidebarTab }) {
  if (tab === 'requests') return <SidebarRequestsPanel />
  if (tab === 'workflows') return <SidebarWorkflowsPanel />
  return <SidebarEnvsPanel />
}
