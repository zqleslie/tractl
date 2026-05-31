import {
  parseRequestRoute,
  requestPathForId,
} from '@/lib/requestEditor/resolveRequestRoute'
import { useUiStore } from '@/stores/uiStore'
import { useWorkspaceStore } from '@/stores/workspaceStore'

export function syncRequestRouteFromLocation(pathname = window.location.pathname) {
  const route = parseRequestRoute(pathname)
  if (!route) return

  const ui = useUiStore.getState()
  const workspace = useWorkspaceStore.getState()
  const requestId = route.isNew ? `request-${ui.requestEditorKey + 1}` : route.id

  if (!workspace.getRequest(requestId)) {
    workspace.createRequest(requestId)
  }

  const existingTab = ui.tabs.find((tab) => tab.id === requestId)
  if (!existingTab) {
    useUiStore.setState((state) => ({
      tabs: [
        ...state.tabs,
        {
          id: requestId,
          type: 'request',
          title: route.isNew ? 'Untitled request' : route.id,
          method: 'GET',
        },
      ],
      activeTabId: requestId,
      activeScreen: 'request',
      requestEditorKey: route.isNew
        ? state.requestEditorKey + 1
        : state.requestEditorKey,
      pendingHistoryEntry: null,
    }))
    return
  }

  useUiStore.setState({
    activeTabId: requestId,
    activeScreen: 'request',
  })
}

export function pushRequestRoute(requestId: string) {
  const nextPath = requestPathForId(requestId)
  if (window.location.pathname !== nextPath) {
    window.history.replaceState(null, '', nextPath)
  }
}
