import { useEffect, useMemo, useRef, useState, type ReactNode } from 'react'
import {
  IconFolderPlus,
  IconPlayerPlay,
  IconPin,
  IconPinnedOff,
  IconPlus,
  IconSearch,
  IconTrash,
} from '@tabler/icons-react'
import { Button } from '@/components/primitives'
import { MethodBadge } from '@/components/primitives/MethodBadge'
import { SaveToCollectionDialog } from '@/components/sidebar/SaveToCollectionDialog'
import { SidebarCollapsibleGroup } from '@/components/sidebar/SidebarCollapsibleGroup'
import { SidebarSection } from '@/components/sidebar/SidebarSection'
import { groupHistoryByDate } from '@/lib/history/groupHistoryByDate'
import { historyEntryTitle } from '@/lib/history/historyEntryLabel'
import { createEmptyRequestDraft } from '@/components/request-editor/createEmptyRequestDraft'
import { draftToRequestState } from '@/lib/requestEditor/workspaceMapping'
import { cn } from '@/lib/cn'
import type { RequestState } from '@/components/request-editor/requestState'
import {
  useRequestCollectionsStore,
  type SavedCollectionRequest,
} from '@/stores/requestCollectionsStore'
import {
  RUN_HISTORY_SIDEBAR_PREVIEW,
  useRunHistoryStore,
  type RunHistoryEntry,
} from '@/stores/runHistoryStore'
import { useUiStore } from '@/stores/uiStore'

type SaveTarget =
  | { kind: 'history'; entry: RunHistoryEntry }
  | { kind: 'state'; state: RequestState; name: string }
  | null

type RowMenuItem = {
  label: string
  icon?: ReactNode
  danger?: boolean
  onClick: () => void
}

function RequestSidebarRow({
  method,
  title,
  meta,
  statusCode,
  testId,
  onClick,
  onRun,
  contextMenu,
}: {
  method: RunHistoryEntry['method']
  title: string
  meta?: string
  statusCode?: number
  testId: string
  onClick: () => void
  onRun?: () => void
  contextMenu?: RowMenuItem[]
}) {
  const [menuPos, setMenuPos] = useState<{ x: number; y: number } | null>(null)
  const menuRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!menuPos) return undefined
    const close = (event: MouseEvent) => {
      if (!menuRef.current?.contains(event.target as Node)) {
        setMenuPos(null)
      }
    }
    const closeKey = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setMenuPos(null)
    }
    document.addEventListener('mousedown', close)
    document.addEventListener('keydown', closeKey)
    return () => {
      document.removeEventListener('mousedown', close)
      document.removeEventListener('keydown', closeKey)
    }
  }, [menuPos])

  return (
    <>
      <div
        className="group relative flex h-9 cursor-pointer items-center gap-2 border-b-[0.5px] border-border px-2 text-left hover:bg-surface-elevated"
        data-testid={testId}
        role="button"
        tabIndex={0}
        onClick={onClick}
        onContextMenu={(event) => {
          if (!contextMenu?.length) return
          event.preventDefault()
          setMenuPos({ x: event.clientX, y: event.clientY })
        }}
        onKeyDown={(event) => {
          if (event.key !== 'Enter' && event.key !== ' ') return
          event.preventDefault()
          onClick()
        }}
      >
        <MethodBadge method={method} className="min-w-[42px] justify-center" />
        <span className="min-w-0 flex-1 truncate text-ui-sm font-medium text-text">
          {title}
        </span>
        {statusCode ? (
          <span
            className={cn(
              'h-1.5 w-1.5 shrink-0 rounded-full',
              statusCode >= 200 && statusCode < 400
                ? 'bg-success-fg'
                : 'bg-danger-fg',
            )}
            aria-hidden="true"
          />
        ) : null}
        {meta ? (
          <span className="shrink-0 text-[11px] text-text-muted">{meta}</span>
        ) : null}
        {onRun ? (
          <Button
            variant="ghost"
            className="absolute right-1 h-7 w-7 p-0 opacity-0 transition-opacity group-hover:opacity-100"
            aria-label="Run"
            onClick={(event) => {
              event.stopPropagation()
              onRun()
            }}
          >
            <IconPlayerPlay size={13} stroke={2} />
          </Button>
        ) : null}
      </div>

      {menuPos && contextMenu?.length ? (
        <div
          ref={menuRef}
          className="fixed z-50 min-w-[160px] rounded-ui border-[0.5px] border-border bg-surface py-1 shadow-lg"
          style={{ left: menuPos.x, top: menuPos.y }}
        >
          {contextMenu.map((item) => (
            <button
              key={item.label}
              type="button"
              className={cn(
                'flex w-full items-center gap-2 px-3 py-1.5 text-left text-ui-xs hover:bg-surface-elevated',
                item.danger ? 'text-danger-fg' : 'text-text',
              )}
              onClick={() => {
                item.onClick()
                setMenuPos(null)
              }}
            >
              {item.icon ? (
                <span className="shrink-0 text-current">{item.icon}</span>
              ) : null}
              {item.label}
            </button>
          ))}
        </div>
      ) : null}
    </>
  )
}

function formatTriggeredTime(iso: string): string {
  return new Date(iso).toLocaleTimeString(undefined, {
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function SidebarRequestsPanel() {
  const openNewRequest = useUiStore((s) => s.openNewRequest)
  const openHistoryEntry = useUiStore((s) => s.openHistoryEntry)
  const openCollectionRequest = useUiStore((s) => s.openCollectionRequest)
  const sidebarView = useUiStore((s) => s.requestsSidebarView)
  const setSidebarView = useUiStore((s) => s.setRequestsSidebarView)
  const showAllHistory = useUiStore((s) => s.requestsHistoryShowAll)
  const setShowAllHistory = useUiStore((s) => s.setRequestsHistoryShowAll)

  const entries = useRunHistoryStore((s) => s.entries)
  const pinnedIds = useRunHistoryStore((s) => s.pinnedIds)
  const pinEntry = useRunHistoryStore((s) => s.pinEntry)
  const unpinEntry = useRunHistoryStore((s) => s.unpinEntry)
  const removeEntry = useRunHistoryStore((s) => s.removeEntry)

  const folders = useRequestCollectionsStore((s) => s.folders)
  const items = useRequestCollectionsStore((s) => s.items)
  const createFolder = useRequestCollectionsStore((s) => s.createFolder)
  const removeItem = useRequestCollectionsStore((s) => s.removeItem)
  const removeFolder = useRequestCollectionsStore((s) => s.removeFolder)

  const [searchQuery, setSearchQuery] = useState('')
  const [saveTarget, setSaveTarget] = useState<SaveTarget>(null)
  const [newFolderName, setNewFolderName] = useState('')

  const normalizedSearch = searchQuery.trim().toLowerCase()

  const matchesHistory = (entry: RunHistoryEntry) => {
    if (!normalizedSearch) return true
    return (
      historyEntryTitle(entry).toLowerCase().includes(normalizedSearch) ||
      entry.url.toLowerCase().includes(normalizedSearch) ||
      entry.method.toLowerCase().includes(normalizedSearch)
    )
  }

  const pinned = useMemo(
    () =>
      pinnedIds
        .map((id) => entries.find((e) => e.id === id))
        .filter((e): e is RunHistoryEntry => e !== undefined)
        .filter(matchesHistory),
    [entries, pinnedIds, normalizedSearch],
  )

  const unpinned = useMemo(
    () => entries.filter((e) => !pinnedIds.includes(e.id)).filter(matchesHistory),
    [entries, pinnedIds, normalizedSearch],
  )

  const historyPreviewLimit = showAllHistory
    ? unpinned.length
    : RUN_HISTORY_SIDEBAR_PREVIEW

  const visibleUnpinned = unpinned.slice(0, historyPreviewLimit)
  const hiddenUnpinnedCount = Math.max(0, unpinned.length - visibleUnpinned.length)
  const dateGroups = useMemo(
    () => groupHistoryByDate(visibleUnpinned),
    [visibleUnpinned],
  )

  const collectionTree = useMemo(() => {
    const filteredFolders = folders.filter((folder) => {
      if (!normalizedSearch) return true
      const folderItems = items.filter((item) => item.folderId === folder.id)
      return (
        folder.name.toLowerCase().includes(normalizedSearch) ||
        folderItems.some((item) => item.name.toLowerCase().includes(normalizedSearch))
      )
    })

    return filteredFolders.map((folder) => ({
      folder,
      items: items.filter((item) => {
        if (item.folderId !== folder.id) return false
        if (!normalizedSearch) return true
        return item.name.toLowerCase().includes(normalizedSearch)
      }),
    }))
  }, [folders, items, normalizedSearch])

  const historyContextMenu = (entry: RunHistoryEntry) => {
    const isPinned = pinnedIds.includes(entry.id)
    return [
      isPinned
        ? {
            label: 'Unpin',
            icon: <IconPinnedOff size={12} stroke={1.75} />,
            onClick: () => unpinEntry(entry.id),
          }
        : {
            label: 'Pin',
            icon: <IconPin size={12} stroke={1.75} />,
            onClick: () => pinEntry(entry.id),
          },
      {
        label: 'Add to collection',
        icon: <IconFolderPlus size={12} stroke={1.75} />,
        onClick: () =>
          setSaveTarget({
            kind: 'history',
            entry,
          }),
      },
      {
        label: 'Delete',
        icon: <IconTrash size={12} stroke={1.75} />,
        danger: true,
        onClick: () => removeEntry(entry.id),
      },
    ]
  }

  const collectionItemContextMenu = (item: SavedCollectionRequest) => [
    {
      label: 'Delete',
      icon: <IconTrash size={12} stroke={1.75} />,
      danger: true,
      onClick: () => removeItem(item.id),
    },
  ]

  const handleCreateFolder = () => {
    const id = createFolder(newFolderName)
    if (id) setNewFolderName('')
  }

  const saveDialogState = (() => {
    if (!saveTarget) return null
    if (saveTarget.kind === 'history') {
      const entry = saveTarget.entry
      const state = draftToRequestState(
        `history-save-${entry.id}`,
        historyEntryTitle(entry),
        entry.method,
        entry.url,
        createEmptyRequestDraft(),
      )
      return { defaultName: historyEntryTitle(entry), state }
    }
    return { defaultName: saveTarget.name, state: saveTarget.state }
  })()

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

        <label className="relative mt-3 flex items-center">
          <IconSearch
            size={14}
            stroke={1.75}
            className="pointer-events-none absolute left-2.5 text-text-muted"
          />
          <input
            type="search"
            aria-label="Search requests"
            placeholder={sidebarView === 'history' ? 'Search history…' : 'Search collections…'}
            className="h-8 w-full rounded-ui border-[0.5px] border-border bg-surface pl-8 pr-2 text-ui-xs text-text outline-none focus:border-primary"
            data-testid="sidebar-requests-search"
            value={searchQuery}
            onChange={(event) => setSearchQuery(event.currentTarget.value)}
          />
        </label>

        <div
          className="mt-3 flex border-b-[0.5px] border-border"
          role="tablist"
          aria-label="Requests sidebar"
        >
          {(['history', 'collections'] as const).map((view) => (
            <button
              key={view}
              type="button"
              role="tab"
              aria-selected={sidebarView === view}
              data-testid={`sidebar-requests-tab-${view}`}
              className={cn(
                'flex-1 border-b-2 border-transparent px-2 pb-1.5 pt-1 text-ui-sm font-semibold capitalize transition-colors',
                sidebarView === view
                  ? '-mb-[0.5px] border-primary text-text'
                  : 'text-text-muted hover:text-text',
              )}
              onClick={() => setSidebarView(view)}
            >
              {view}
            </button>
          ))}
        </div>
      </div>

      <div className="min-h-0 flex-1 overflow-auto">
        {sidebarView === 'history' ? (
          <>
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
          </>
        ) : (
          <>
            <div className="mx-2 mb-2 flex gap-1.5">
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

            {collectionTree.length === 0 ? (
              <p className="px-3 py-4 text-ui-xs leading-relaxed text-text-muted">
                Create a collection folder, then use Save in the request editor to
                store requests here.
              </p>
            ) : (
              collectionTree.map(({ folder, items: folderItems }) => (
                <SidebarCollapsibleGroup
                  key={folder.id}
                  title={folder.name}
                  suffix={String(folderItems.length)}
                  defaultOpen
                >
                  {folderItems.map((item) => (
                    <RequestSidebarRow
                      key={item.id}
                      testId={`sidebar-collection-${item.id}`}
                      method={item.state.method}
                      title={item.name}
                      meta={formatTriggeredTime(item.savedAt)}
                      onClick={() => openCollectionRequest(item)}
                      contextMenu={collectionItemContextMenu(item)}
                    />
                  ))}
                  {folderItems.length === 0 ? (
                    <p className="px-2 py-1 text-[10px] text-text-muted">No saved requests</p>
                  ) : null}
                  <button
                    type="button"
                    className="mx-3 mb-2 text-[10px] text-text-muted hover:text-danger-fg"
                    onClick={() => removeFolder(folder.id)}
                  >
                    Delete folder
                  </button>
                </SidebarCollapsibleGroup>
              ))
            )}
          </>
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
