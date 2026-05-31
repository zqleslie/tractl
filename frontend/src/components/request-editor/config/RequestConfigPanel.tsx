import { RequestConfigTabContent } from '@/components/request-editor/config/RequestConfigTabContent'
import { RequestConfigTabs } from '@/components/request-editor/config/RequestConfigTabs'
import type { RequestEditorController } from '@/components/request-editor/useRequestEditor'
import type { ConfigTab } from '@/components/request-editor/types'
import { cn } from '@/lib/cn'
import type { EditorLayout } from '@/stores/uiStore'

export type RequestConfigPanelProps = {
  editor: RequestEditorController
  activeTab: ConfigTab
  layout: EditorLayout
  resultsOpen: boolean
  splitBasis?: string
  onTabChange: (tab: ConfigTab) => void
}

export function RequestConfigPanel({
  editor,
  activeTab,
  layout,
  resultsOpen,
  splitBasis = '54%',
  onTabChange,
}: RequestConfigPanelProps) {
  return (
    <section
      className={cn(
        'flex min-w-0 flex-col bg-surface',
        layout === 'side-by-side' && resultsOpen && 'shrink-0',
        layout === 'stacked' && resultsOpen && 'shrink-0 border-b-[0.5px] border-border',
        !(layout === 'side-by-side' && resultsOpen) && 'flex-1',
      )}
      style={
        resultsOpen
          ? layout === 'stacked'
            ? { flex: `0 0 ${splitBasis}`, minHeight: 0 }
            : { flex: `0 0 ${splitBasis}`, minWidth: 0 }
          : undefined
      }
      aria-label="Request configuration"
    >
      <RequestConfigTabs
        activeTab={activeTab}
        tabCounts={editor.configTabCounts}
        onTabChange={onTabChange}
      />
      <div className="min-h-0 flex-1 overflow-auto p-3">
        <RequestConfigTabContent tab={activeTab} editor={editor} />
      </div>
    </section>
  )
}
