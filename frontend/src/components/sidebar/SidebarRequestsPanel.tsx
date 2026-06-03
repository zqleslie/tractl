import { IconPlus, IconSearch } from '@tabler/icons-react'
import { CollectionFolderGroup } from '@/components/sidebar/CollectionFolderGroup'
import { RequestSidebarRow } from '@/components/sidebar/RequestSidebarRow'
import { SaveToCollectionDialog } from '@/components/sidebar/SaveToCollectionDialog'
import { SidebarSection } from '@/components/sidebar/SidebarSection'
import { useSidebarRequestsPanel } from '@/components/sidebar/useSidebarRequestsPanel'
import { historyEntryTitle } from '@/lib/history/historyEntryLabel'
import { cn } from '@/lib/cn'
import { RUN_HISTORY_SIDEBAR_PREVIEW } from '@/stores/runHistoryStore'

function formatTriggeredTime(iso: string): string {
  return new Date(iso).toLocaleTimeString(undefined, {
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function SidebarRequestsPanel() {
  const {
    openNewRequest,
    openHistoryEntry,
    openRequestItem,
    sidebarView,
    setSidebarView,
    showAllHistory,
    setShowAllHistory,
    entries,
    pinned,
    unpinned,
    hiddenUnpinnedCount,
    dateGroups,
    historyContextMenu,
    collectionTree,
    removeFolder,
    collectionItemContextMenu,
    newFolderName,
    setNewFolderName,
    handleCreateFolder,
    searchQuery,
    setSearchQuery,
    saveTarget,
    setSaveTarget,
    saveDialogState,
  } = useSidebarRequestsPanel()

  return (
    <div data-testid="sidebar-panel-requests" className="flex h-full min-h-0 flex-col">
      <div className="shrink-0 border-b-[0.5px] border-border px-3 pb-3 pt-2">
        <button
          type="button"
          className="flex h-8 w-full items-center justify-center gap-1.5 rounded-ui border-[0.5px] border-border bg-surface-elevated text-ui-sm font-medium text-text hover:border-primary"
          data-testid="sidebar-new-request"
          onClick={openNewRequest}
        >
          <IconPlus size={14} stroke={1.75} />
          New request
        </button>

        <label className="mt-3 flex items-center gap-2 rounded-ui border-[0.5px] border-border bg-surface px-2 focus-within:border-primary">
          <IconSearch size={14} stroke={1.75} className="pointer-events-none text-text-muted" />
          <input
            type="search"
            aria-label="Search requests"
            placeholder={sidebarView === 'history' ? 'Search history…' : 'Search collections…'}
            className="h-8 w-full bg-transparent text-ui-xs text-text outline-none"
            data-testid="sidebar-requests-search"
            value={searchQuery}
            onChange={(event) => setSearchQuery(event.currentTarget.value)}
          />
        </label>

        <div className="mt-3 flex border-b-[0.5px] border-border" role="tablist" aria-label="Requests sidebar">
          {(['history', 'collections'] as const).map((view) => (
            <button
              key={view}
              type="button"
              role="tab"
              aria-selected={sidebarView === view}
              data-testid={`sidebar-requests-tab-${view}`}
              className={cn(
                'flex-1 border-b-2 -mb-px border-transparent px-2 pb-1.5 pt-1 text-ui-sm capitalize transition-colors',
                sidebarView === view
                  ? 'border-primary text-primary font-medium'
                  : 'text-text-muted hover:text-text',
              )}
              onClick={() => setSidebarView(view)}
            >
              {view}
            </button>
          ))}
        </div>
      </div>

      <div className="min-h-0 flex-1 overflow-hidden">
        {sidebarView === 'history' ? (
          <div className="h-full overflow-auto">
            {pinned.length > 0 ? (
              <>
                <SidebarSection>Pinned</SidebarSection>
                {pinned.map((entry) => (
                  <RequestSidebarRow
                    key={entry.id}
                    testId={`sidebar-request-${entry.id}`}
                    method={entry.method}
                    title={historyEntryTitle(entry)}
                    statusCode={entry.statusCode}
                    meta={formatTriggeredTime(entry.timestamp)}
                    onClick={() => openHistoryEntry(entry)}
                    onRun={() => openHistoryEntry(entry)}
                    contextMenu={historyContextMenu(entry)}
                  />
                ))}
              </>
            ) : null}

            {dateGroups.map((group) => (
              <div key={group.label}>
                <SidebarSection>{group.label}</SidebarSection>
                {group.entries.map((entry) => (
                  <RequestSidebarRow
                    key={entry.id}
                    testId={`sidebar-request-${entry.id}`}
                    method={entry.method}
                    title={historyEntryTitle(entry)}
                    statusCode={entry.statusCode}
                    meta={formatTriggeredTime(entry.timestamp)}
                    onClick={() => openHistoryEntry(entry)}
                    onRun={() => openHistoryEntry(entry)}
                    contextMenu={historyContextMenu(entry)}
                  />
                ))}
              </div>
            ))}

            {entries.length === 0 ? (
              <p className="px-2 py-3 text-ui-xs text-text-muted">
                Run a request to build history. Up to 50 entries are kept.
              </p>
            ) : null}

            {hiddenUnpinnedCount > 0 ? (
              <button
                type="button"
                className="mx-2 mb-3 w-[calc(100%-1rem)] py-1.5 text-center text-ui-xs font-medium text-primary hover:underline"
                data-testid="sidebar-show-all-history"
                onClick={() => setShowAllHistory(true)}
              >
                Show recent ({unpinned.length})
              </button>
            ) : null}

            {showAllHistory && unpinned.length > RUN_HISTORY_SIDEBAR_PREVIEW ? (
              <button
                type="button"
                className="mx-2 mb-3 w-[calc(100%-1rem)] py-1.5 text-center text-ui-xs text-text-muted hover:text-text"
                onClick={() => setShowAllHistory(false)}
              >
                Show less
              </button>
            ) : null}
          </div>
        ) : (
          <div className="flex h-full flex-col overflow-hidden">
            <div className="flex shrink-0 gap-1.5 border-b-[0.5px] border-border px-2.5 py-2">
              <input
                aria-label="New collection folder name"
                placeholder="New collection folder…"
                className="min-w-0 flex-1 rounded-ui border-[0.5px] border-border bg-surface px-2.5 py-1.5 text-ui-xs text-text outline-none focus:border-primary"
                value={newFolderName}
                onChange={(event) => setNewFolderName(event.currentTarget.value)}
              />
              <button
                type="button"
                className="shrink-0 rounded-ui border-[0.5px] border-border bg-surface-elevated px-3 py-1.5 text-ui-xs font-medium text-text hover:border-primary"
                data-testid="sidebar-create-folder"
                onClick={handleCreateFolder}
              >
                Add
              </button>
            </div>

            <div className="flex-1 overflow-y-auto py-1">
              {collectionTree.length === 0 ? (
                <p className="px-3 py-4 text-ui-xs leading-relaxed text-text-muted">
                  Create a collection folder, then use Save in the request editor to store requests here.
                </p>
              ) : (
                collectionTree.map(({ folder, items: folderItems }) => (
                  <CollectionFolderGroup
                    key={folder.id}
                    title={folder.name}
                    suffix={String(folderItems.length)}
                    onDelete={() => removeFolder(folder.id)}
                  >
                    {folderItems.map((item) => (
                      <RequestSidebarRow
                        key={item.id}
                        testId={`sidebar-collection-${item.id}`}
                        method={item.state.method}
                        title={item.name}
                        meta={formatTriggeredTime(item.savedAt)}
                        onClick={() => openRequestItem(item)}
                        contextMenu={collectionItemContextMenu(item)}
                      />
                    ))}
                    {folderItems.length === 0 ? (
                      <p className="px-3 py-1 text-[10px] text-text-muted">No saved requests</p>
                    ) : null}
                  </CollectionFolderGroup>
                ))
              )}
            </div>
          </div>
        )}
      </div>

      {saveTarget && saveDialogState ? (
        <SaveToCollectionDialog
          open
          defaultName={saveDialogState.defaultName}
          getRequestState={() => saveDialogState.state}
          onClose={() => setSaveTarget(null)}
          onSaved={() => {
            setSaveTarget(null)
            setSidebarView('collections')
          }}
        />
      ) : null}
    </div>
  )
}
