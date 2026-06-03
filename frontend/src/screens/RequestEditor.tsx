import { useState } from 'react'
import { RequestConfigPanel } from '@/components/request-editor/config/RequestConfigPanel'
import { EditorDragHandle } from '@/components/request-editor/EditorDragHandle'
import { PanelCollapseDivider } from '@/components/request-editor/PanelCollapseDivider'
import { useRequestEditor } from '@/components/request-editor/useRequestEditor'
import { ResultsPanel } from '@/components/results/ResultsPanel'
import { SaveToCollectionDialog } from '@/components/sidebar/SaveToCollectionDialog'
import { UrlBar } from '@/components/shell/UrlBar'
import { useRequestEditorShortcuts } from '@/hooks/useRequestEditorShortcuts'
import { cn } from '@/lib/cn'
import { useRequestCollectionsStore } from '@/stores/requestCollectionsStore'
import { useUiStore } from '@/stores/uiStore'

export function RequestEditor() {
  const editor = useRequestEditor()
  const [saveDialogOpen, setSaveDialogOpen] = useState(false)
  const updateCollectionItem = useRequestCollectionsStore((s) => s.updateItem)
  const setRequestsSidebarView = useUiStore((s) => s.setRequestsSidebarView)

  const handleUpdate = () => {
    if (!editor.collectionItemId) return
    updateCollectionItem(editor.collectionItemId, editor.getRequestState())
    useUiStore.getState().updateActiveRequestTab({ isDirty: false })
  }
  const resultLayout = useUiStore((state) => state.editorLayout)
  const editorSplitRatio = useUiStore((state) => state.editorSplitRatio)
  const setEditorSplitRatio = useUiStore((state) => state.setEditorSplitRatio)

  useRequestEditorShortcuts(editor)

  const configBasis = `${editorSplitRatio * 100}%`

  return (
    <div className="flex h-full min-h-0 flex-col" data-testid="screen-request">
      {editor.apiAvailable === false ? (
        <div
          className="border-b-[0.5px] border-warning bg-warning-bg px-3 py-2 text-ui-xs text-warning-fg"
          data-testid="request-api-unavailable"
        >
          traCtl Desktop API is not running. Start the desktop app or run{' '}
          <code className="font-mono">go run ./cmd/localapi</code> for web dev, then
          reload.
        </div>
      ) : null}
      {editor.saveError ? (
        <div className="border-b-[0.5px] border-danger bg-danger-bg px-3 py-2 text-ui-xs text-danger-fg">
          {editor.saveError}
        </div>
      ) : null}
      {editor.runError ? (
        <div className="border-b-[0.5px] border-danger bg-danger-bg px-3 py-2 text-ui-xs text-danger-fg">
          {editor.runError}
        </div>
      ) : null}

      <UrlBar
        method={editor.method}
        url={editor.url}
        canRun={editor.canRun}
        isRunning={editor.isRunning}
        onMethodChange={editor.setMethod}
        onUrlChange={editor.setUrl}
        onRun={() => void editor.runRequest()}
        isSaved={editor.collectionItemId !== null}
        onSaveToCollection={() => setSaveDialogOpen(true)}
        onUpdate={handleUpdate}
      />

      <SaveToCollectionDialog
        open={saveDialogOpen}
        defaultName={editor.requestName}
        getRequestState={editor.getRequestState}
        onClose={() => setSaveDialogOpen(false)}
        onSaved={(itemId) => {
          setSaveDialogOpen(false)
          editor.setSavedCollectionItemId(itemId)
          useUiStore.getState().updateActiveRequestTab({ isDirty: false })
          setRequestsSidebarView('collections')
          useUiStore.getState().setSidebarTab('requests')
          useUiStore.getState().setSidebarCollapsed(false)
        }}
      />

      <div
        className={cn(
          'flex min-h-0 flex-1',
          resultLayout === 'stacked' ? 'flex-col' : 'flex-row',
        )}
        data-layout={resultLayout}
      >
        <RequestConfigPanel
          editor={editor}
          activeTab={editor.activeConfigTab}
          layout={resultLayout}
          resultsOpen={editor.resultsOpen}
          splitBasis={configBasis}
          onTabChange={editor.setActiveConfigTab}
        />

        <PanelCollapseDivider
          layout={resultLayout}
          resultsOpen={editor.resultsOpen}
          onToggle={editor.toggleResultsPanel}
        />

        {editor.resultsOpen ? (
          <>
            <EditorDragHandle
              layout={resultLayout}
              onResize={setEditorSplitRatio}
            />
            <ResultsPanel
              layout={resultLayout}
              activeTab={editor.activeResultTab}
              runState={editor.runState}
              runResult={editor.runResult}
              executionResult={editor.executionResult}
              isRunning={editor.isRunning}
              onTabChange={editor.setActiveResultTab}
            />
          </>
        ) : null}
      </div>
    </div>
  )
}
