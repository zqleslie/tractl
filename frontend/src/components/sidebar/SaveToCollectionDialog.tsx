import { useEffect, useState } from 'react'
import { Button } from '@/components/primitives'
import type { RequestState } from '@/components/request-editor/requestState'
import { useRequestCollectionsStore } from '@/stores/requestCollectionsStore'

export type SaveToCollectionDialogProps = {
  open: boolean
  defaultName: string
  getRequestState: () => RequestState
  onClose: () => void
  onSaved?: (itemId: string) => void
}

export function SaveToCollectionDialog({
  open,
  defaultName,
  getRequestState,
  onClose,
  onSaved,
}: SaveToCollectionDialogProps) {
  const folders = useRequestCollectionsStore((s) => s.folders)
  const createFolder = useRequestCollectionsStore((s) => s.createFolder)
  const saveRequest = useRequestCollectionsStore((s) => s.saveRequest)

  const [selectedFolderId, setSelectedFolderId] = useState('')
  const [newFolderName, setNewFolderName] = useState('')
  const [requestName, setRequestName] = useState(defaultName)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!open) return
    setRequestName(defaultName)
    setError(null)
    setNewFolderName('')
    setSelectedFolderId(folders[0]?.id ?? '')
  }, [open, defaultName, folders])

  if (!open) return null

  const handleSave = () => {
    let folderId = selectedFolderId

    if (newFolderName.trim()) {
      folderId = createFolder(newFolderName)
      if (!folderId) {
        setError('Enter a collection folder name.')
        return
      }
    }

    if (!folderId) {
      setError('Select or create a collection folder.')
      return
    }

    const state = getRequestState()
    const itemId = saveRequest(folderId, state, requestName)
    onSaved?.(itemId)
    onClose()
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
      role="presentation"
      data-testid="save-collection-dialog-backdrop"
      onClick={onClose}
    >
      <div
        role="dialog"
        aria-labelledby="save-collection-title"
        className="w-full max-w-md rounded-ui border-[0.5px] border-border bg-surface p-4 shadow-lg"
        data-testid="save-collection-dialog"
        onClick={(event) => event.stopPropagation()}
      >
        <h2 id="save-collection-title" className="text-ui-sm font-medium text-text">
          Save to collection
        </h2>
        <p className="mt-1 text-ui-xs text-text-muted">
          Choose a parent folder. Saved requests appear under Collections only.
        </p>

        <div className="mt-4 space-y-3">
          <label className="block">
            <span className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-text-muted">
              Request name
            </span>
            <input
              className="h-8 w-full rounded-ui border-[0.5px] border-border bg-surface-elevated px-2 text-ui-sm text-text outline-none focus:border-primary"
              value={requestName}
              onChange={(event) => setRequestName(event.currentTarget.value)}
            />
          </label>

          {folders.length > 0 ? (
            <label className="block">
              <span className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-text-muted">
                Collection folder
              </span>
              <select
                className="h-8 w-full rounded-ui border-[0.5px] border-border bg-surface-elevated px-2 text-ui-sm text-text outline-none focus:border-primary"
                value={selectedFolderId}
                onChange={(event) => {
                  setSelectedFolderId(event.currentTarget.value)
                  setNewFolderName('')
                }}
              >
                <option value="">Select folder…</option>
                {folders.map((folder) => (
                  <option key={folder.id} value={folder.id}>
                    {folder.name}
                  </option>
                ))}
              </select>
            </label>
          ) : null}

          <label className="block">
            <span className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-text-muted">
              {folders.length > 0 ? 'Or create new folder' : 'Collection folder (required)'}
            </span>
            <input
              placeholder="e.g. GitHub REST API"
              className="h-8 w-full rounded-ui border-[0.5px] border-border bg-surface-elevated px-2 text-ui-sm text-text outline-none focus:border-primary"
              value={newFolderName}
              onChange={(event) => {
                setNewFolderName(event.currentTarget.value)
                if (event.currentTarget.value.trim()) {
                  setSelectedFolderId('')
                }
              }}
            />
          </label>
        </div>

        {error ? (
          <p className="mt-2 text-ui-xs text-danger-fg" role="alert">
            {error}
          </p>
        ) : null}

        <div className="mt-4 flex justify-end gap-2">
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button variant="primary" data-testid="save-collection-confirm" onClick={handleSave}>
            Save
          </Button>
        </div>
      </div>
    </div>
  )
}
