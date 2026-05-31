import { PanelTabBar } from '@/components/common/PanelTabBar'
import { configTabs } from '@/components/request-editor/constants'
import type { ConfigTab } from '@/components/request-editor/types'

export type RequestConfigTabsProps = {
  activeTab: ConfigTab
  tabCounts: {
    headers: number
    assertions: number
    extracts: number
  }
  onTabChange: (tab: ConfigTab) => void
}

export function RequestConfigTabs({
  activeTab,
  tabCounts,
  onTabChange,
}: RequestConfigTabsProps) {
  const counts: Partial<Record<ConfigTab, number>> = {
    headers: tabCounts.headers,
    assertions: tabCounts.assertions,
    extracts: tabCounts.extracts,
  }

  return (
    <PanelTabBar
      tabs={configTabs.map((tab) => ({
        id: tab.id,
        label: tab.label,
        count: counts[tab.id],
        variant: tab.script ? 'script' : 'default',
      }))}
      activeTab={activeTab}
      testIdPrefix="request-config-tab"
      onTabChange={(id) => onTabChange(id as ConfigTab)}
    />
  )
}
