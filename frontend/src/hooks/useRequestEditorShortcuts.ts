import { useEffect } from 'react'
import type { HttpMethod } from '@/components/primitives'
import type { RequestEditorController } from '@/components/request-editor/useRequestEditor'
import { useUiStore } from '@/stores/uiStore'

const methodShortcuts: Record<string, HttpMethod> = {
  g: 'GET',
  p: 'POST',
  x: 'DELETE',
  u: 'PUT',
  a: 'PATCH',
}

function isEditableTarget(target: EventTarget | null) {
  if (!(target instanceof HTMLElement)) return false
  const tag = target.tagName
  return tag === 'INPUT' || tag === 'TEXTAREA' || target.isContentEditable
}

export function useRequestEditorShortcuts(editor: RequestEditorController) {
  const toggleResultLayout = useUiStore((state) => state.toggleResultLayout)
  const closeTab = useUiStore((state) => state.closeTab)
  const activeTabId = useUiStore((state) => state.activeTabId)
  const tabs = useUiStore((state) => state.tabs)
  const activateTab = useUiStore((state) => state.activateTab)

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      const meta = event.metaKey || event.ctrlKey

      if (meta && event.key === 'Enter') {
        event.preventDefault()
        void editor.runRequest()
        return
      }

      if (isEditableTarget(event.target) && !(meta && event.key === 'Enter')) {
        return
      }

      if (meta && event.shiftKey && event.key.toLowerCase() === 'l') {
        event.preventDefault()
        toggleResultLayout()
        return
      }

      if (event.altKey && !meta) {
        const method = methodShortcuts[event.key.toLowerCase()]
        if (method) {
          event.preventDefault()
          editor.setMethod(method)
        }
        return
      }

      if (meta && event.key === '.') {
        event.preventDefault()
        const body = editor.runResult?.body
        if (body) void navigator.clipboard.writeText(body)
        return
      }

      if (meta && event.key.toLowerCase() === 'w') {
        event.preventDefault()
        if (activeTabId) closeTab(activeTabId)
        return
      }

      if (meta && event.key === ']') {
        event.preventDefault()
        const index = tabs.findIndex((tab) => tab.id === activeTabId)
        const next = tabs[index + 1] ?? tabs[0]
        if (next) activateTab(next.id)
        return
      }

      if (meta && event.key === '[') {
        event.preventDefault()
        const index = tabs.findIndex((tab) => tab.id === activeTabId)
        const prev = tabs[index - 1] ?? tabs[tabs.length - 1]
        if (prev) activateTab(prev.id)
      }
    }

    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [
    activateTab,
    activeTabId,
    closeTab,
    editor,
    tabs,
    toggleResultLayout,
  ])
}
