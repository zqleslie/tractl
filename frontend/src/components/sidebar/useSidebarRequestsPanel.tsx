import { useMemo, useState } from 'react'
import {
  IconFolderPlus,
  IconPin,
  IconPinnedOff,
  IconTrash,
} from '@tabler/icons-react'
import { groupHistoryByDate } from '@/lib/history/groupHistoryByDate'
import { historyEntryTitle } from '@/lib/history/historyEntryLabel'
import { createEmptyRequestFormState } from '@/components/request-editor/createEmptyRequestFormState'
import { draftToRequestState } from '@/lib/requestEditor/workspaceMapping'
import type { RequestState } from '@/components/request-editor/requestState'
import {
  useRequestCollectionsStore,
  type SavedRequestItem,
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

export function useSidebarRequestsPanel() {
  const openNewRequest = useUiStore((s) => s.openNewRequest)
  const openHistoryEntry = useUiStore((s) => s.openHistoryEntry)
  const openRequestItem = useUiStore((s) => s.openRequestItem)
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

  const historyPreviewLimit = showAllHistory ? unpinned.length : RUN_HISTORY_SIDEBAR_PREVIEW
  const visibleUnpinned = unpinned.slice(0, historyPreviewLimit)
  const hiddenUnpinnedCount = Math.max(0, unpinned.length - visibleUnpinned.length)
  const dateGroups = useMemo(() => groupHistoryByDate(visibleUnpinned), [visibleUnpinned])

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
        ? { label: 'Unpin', icon: <IconPinnedOff size={12} stroke={1.75} />, onClick: () => unpinEntry(entry.id) }
        : { label: 'Pin', icon: <IconPin size={12} stroke={1.75} />, onClick: () => pinEntry(entry.id) },
      {
        label: 'Add to collection',
        icon: <IconFolderPlus size={12} stroke={1.75} />,
        onClick: () => setSaveTarget({ kind: 'history', entry }),
      },
      {
        label: 'Delete',
        icon: <IconTrash size={12} stroke={1.75} />,
        danger: true,
        onClick: () => removeEntry(entry.id),
      },
    ]
  }

  const collectionItemContextMenu = (item: SavedRequestItem) => [
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
        createEmptyRequestFormState(),
      )
      return { defaultName: historyEntryTitle(entry), state }
    }
    return { defaultName: saveTarget.name, state: saveTarget.state }
  })()

  return {
    // ui store
    openNewRequest,
    openHistoryEntry,
    openRequestItem,
    sidebarView,
    setSidebarView,
    showAllHistory,
    setShowAllHistory,
    // history
    entries,
    pinned,
    unpinned,
    hiddenUnpinnedCount,
    dateGroups,
    historyContextMenu,
    // collections
    collectionTree,
    folders,
    removeFolder,
    collectionItemContextMenu,
    newFolderName,
    setNewFolderName,
    handleCreateFolder,
    // search
    searchQuery,
    setSearchQuery,
    // save dialog
    saveTarget,
    setSaveTarget,
    saveDialogState,
  }
}
