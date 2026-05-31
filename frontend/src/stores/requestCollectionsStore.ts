import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { RequestState } from '@/components/request-editor/requestState'

export type CollectionFolder = {
  id: string
  name: string
  createdAt: string
}

export type SavedRequestItem = {
  id: string
  folderId: string
  name: string
  savedAt: string
  state: RequestState
}

type RequestCollectionsState = {
  folders: CollectionFolder[]
  items: SavedRequestItem[]
  createFolder: (name: string) => string
  renameFolder: (id: string, name: string) => void
  removeFolder: (id: string) => void
  saveRequest: (folderId: string, state: RequestState, name?: string) => string
  removeItem: (id: string) => void
  getItemsByFolder: (folderId: string) => SavedRequestItem[]
}

export const useRequestCollectionsStore = create<RequestCollectionsState>()(
  persist(
    (set, get) => ({
      folders: [],
      items: [],
      createFolder: (name) => {
        const trimmed = name.trim()
        if (!trimmed) return ''
        const folder: CollectionFolder = {
          id: crypto.randomUUID(),
          name: trimmed,
          createdAt: new Date().toISOString(),
        }
        set((state) => ({ folders: [...state.folders, folder] }))
        return folder.id
      },
      renameFolder: (id, name) =>
        set((state) => ({
          folders: state.folders.map((folder) =>
            folder.id === id ? { ...folder, name: name.trim() } : folder,
          ),
        })),
      removeFolder: (id) =>
        set((state) => ({
          folders: state.folders.filter((folder) => folder.id !== id),
          items: state.items.filter((item) => item.folderId !== id),
        })),
      saveRequest: (folderId, state, name) => {
        const item: SavedRequestItem = {
          id: crypto.randomUUID(),
          folderId,
          name: name?.trim() || state.name || 'Untitled request',
          savedAt: new Date().toISOString(),
          state: { ...state },
        }
        set((s) => ({ items: [item, ...s.items] }))
        return item.id
      },
      removeItem: (id) =>
        set((state) => ({
          items: state.items.filter((item) => item.id !== id),
        })),
      getItemsByFolder: (folderId) =>
        get().items.filter((item) => item.folderId === folderId),
    }),
    { name: 'tractl-request-collections' },
  ),
)
