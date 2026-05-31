import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { HttpMethod } from '@/components/primitives'
import type { RunHistoryEntry } from '@/stores/runHistoryStore'
import type { SavedRequestItem } from '@/stores/requestCollectionsStore'

export type RequestsSidebarView = 'history' | 'collections'

export type Theme = 'light' | 'dark' | 'system'
export type UiScale = 'compact' | 'default' | 'comfortable'
/** Spec alias — same values as {@link UiScale} (comfortable | default | compact). */
export type Density = UiScale
export type SidebarTab =
  | 'requests'
  | 'workflows'
  | 'overlays'
  | 'history'
  | 'envs'
  | 'settings'
export type AppScreen =
  | 'fast-start'
  | 'request'
  | 'workflow'
  | 'run-history'
  | 'environments'
export type EditorLayout = 'stacked' | 'side-by-side'
/** Spec alias for {@link EditorLayout}. */
export type ResultLayout = EditorLayout
export type WorkspaceTabType = 'request' | 'workflow'

export type WorkspaceTab = {
  id: string
  type: WorkspaceTabType
  title: string
  method: HttpMethod | 'WF'
  isDirty: boolean
  workflowId?: string | null
}

type UiState = {
  theme: Theme
  uiScale: UiScale
  sidebarTab: SidebarTab
  sidebarCollapsed: boolean
  activityBarExpanded: boolean
  editorLayout: EditorLayout
  editorSplitRatio: number
  wordWrap: boolean
  commandPaletteOpen: boolean
  keyboardShortcutsOpen: boolean
  tabs: WorkspaceTab[]
  activeTabId: string | null
  activeScreen: AppScreen
  activeWorkflowId: string | null
  pendingHistoryEntry: RunHistoryEntry | null
  pendingRequestItem: SavedRequestItem | null
  requestsSidebarView: RequestsSidebarView
  requestsHistoryShowAll: boolean
  requestEditorKey: number
  setTheme: (theme: Theme) => void
  toggleTheme: () => void
  setUiScale: (scale: UiScale) => void
  /** Spec alias for {@link setUiScale} — drives `data-density` on the document root. */
  setDensity: (density: Density) => void
  cycleUiScale: () => void
  setSidebarTab: (tab: SidebarTab) => void
  setSidebarCollapsed: (collapsed: boolean) => void
  toggleActivityBarExpanded: () => void
  toggleSidebarTab: (tab: SidebarTab) => void
  setEditorLayout: (layout: EditorLayout) => void
  /** Spec alias for {@link setEditorLayout}. */
  setResultLayout: (layout: ResultLayout) => void
  setEditorSplitRatio: (ratio: number) => void
  toggleEditorLayout: () => void
  /** Spec alias for {@link toggleEditorLayout}. */
  toggleResultLayout: () => void
  toggleWordWrap: () => void
  setCommandPaletteOpen: (open: boolean) => void
  setKeyboardShortcutsOpen: (open: boolean) => void
  activateTab: (tabId: string) => void
  closeTab: (tabId: string) => void
  updateActiveRequestTab: (
    patch: Partial<Pick<WorkspaceTab, 'method' | 'title' | 'isDirty'>>,
  ) => void
  setActiveScreen: (screen: AppScreen) => void
  openNewRequest: () => void
  openWorkflow: (id: string | null) => void
  openWorkflowPlaceholder: () => void
  openHistoryEntry: (entry: RunHistoryEntry) => void
  clearPendingHistoryEntry: () => void
  openRequestItem: (item: SavedRequestItem) => void
  clearPendingRequestItem: () => void
  setRequestsSidebarView: (view: RequestsSidebarView) => void
  setRequestsHistoryShowAll: (showAll: boolean) => void
}

export const useUiStore = create<UiState>()(
  persist(
    (set, get) => ({
      theme: 'light',
      uiScale: 'default',
      sidebarTab: 'requests',
      sidebarCollapsed: true,
      activityBarExpanded: false,
      editorLayout: 'stacked',
      editorSplitRatio: 0.54,
      wordWrap: true,
      commandPaletteOpen: false,
      keyboardShortcutsOpen: false,
      tabs: [],
      activeTabId: null,
      activeScreen: 'fast-start',
      activeWorkflowId: null,
      pendingHistoryEntry: null,
      pendingRequestItem: null,
      requestsSidebarView: 'history',
      requestsHistoryShowAll: false,
      requestEditorKey: 0,
      setTheme: (theme) => set({ theme }),
      toggleTheme: () =>
        set({ theme: get().theme === 'dark' ? 'light' : 'dark' }),
      setUiScale: (uiScale) => set({ uiScale }),
      setDensity: (density) => set({ uiScale: density }),
      cycleUiScale: () =>
        set((state) => ({
          uiScale:
            state.uiScale === 'compact'
              ? 'default'
              : state.uiScale === 'default'
                ? 'comfortable'
                : 'compact',
        })),
      setSidebarTab: (sidebarTab) => set({ sidebarTab }),
      setSidebarCollapsed: (sidebarCollapsed) => set({ sidebarCollapsed }),
      toggleActivityBarExpanded: () =>
        set((state) => ({ activityBarExpanded: !state.activityBarExpanded })),
      toggleSidebarTab: (sidebarTab) =>
        set((state) => {
          const togglingSame = state.sidebarTab === sidebarTab
          const nextCollapsed = togglingSame ? !state.sidebarCollapsed : false

          if (
            sidebarTab === 'workflows' &&
            state.activeScreen === 'request' &&
            !togglingSame
          ) {
            const tabId = `workflow-${state.activeWorkflowId ?? 'draft'}`
            const existing = state.tabs.some((tab) => tab.id === tabId)
            return {
              sidebarTab,
              sidebarCollapsed: true,
              activeScreen: 'workflow',
              activeTabId: tabId,
              tabs: existing
                ? state.tabs
                : [
                    ...state.tabs,
                    {
                      id: tabId,
                      type: 'workflow',
                      title: state.activeWorkflowId ?? 'Untitled workflow',
                      method: 'WF',
                      isDirty: false,
                      workflowId: state.activeWorkflowId,
                    },
                  ],
            }
          }

          return {
            sidebarTab,
            sidebarCollapsed: nextCollapsed,
          }
        }),
      setEditorLayout: (editorLayout) => set({ editorLayout }),
      setResultLayout: (editorLayout) => set({ editorLayout }),
      setEditorSplitRatio: (editorSplitRatio) => set({ editorSplitRatio }),
      toggleEditorLayout: () =>
        set((state) => ({
          editorLayout:
            state.editorLayout === 'stacked' ? 'side-by-side' : 'stacked',
        })),
      toggleResultLayout: () =>
        set((state) => ({
          editorLayout:
            state.editorLayout === 'stacked' ? 'side-by-side' : 'stacked',
        })),
      toggleWordWrap: () => set((state) => ({ wordWrap: !state.wordWrap })),
      setCommandPaletteOpen: (commandPaletteOpen) => set({ commandPaletteOpen }),
      setKeyboardShortcutsOpen: (keyboardShortcutsOpen) =>
        set({ keyboardShortcutsOpen }),
      activateTab: (tabId) =>
        set((state) => {
          const tab = state.tabs.find((candidate) => candidate.id === tabId)
          if (!tab) return state
          return {
            activeTabId: tab.id,
            activeScreen: tab.type === 'request' ? 'request' : 'workflow',
            activeWorkflowId:
              tab.type === 'workflow' ? (tab.workflowId ?? null) : state.activeWorkflowId,
          }
        }),
      closeTab: (tabId) =>
        set((state) => {
          const tabIndex = state.tabs.findIndex((tab) => tab.id === tabId)
          if (tabIndex === -1) return state

          const tabs = state.tabs.filter((tab) => tab.id !== tabId)
          if (state.activeTabId !== tabId) return { tabs }

          const nextTab = tabs[Math.min(tabIndex, tabs.length - 1)] ?? null
          return {
            tabs,
            activeTabId: nextTab?.id ?? null,
            activeScreen:
              nextTab?.type === 'request'
                ? 'request'
                : nextTab?.type === 'workflow'
                  ? 'workflow'
                  : 'fast-start',
            activeWorkflowId:
              nextTab?.type === 'workflow'
                ? (nextTab.workflowId ?? null)
                : state.activeWorkflowId,
          }
        }),
      updateActiveRequestTab: (patch) =>
        set((state) => {
          if (!state.activeTabId) return state

          return {
            tabs: state.tabs.map((tab) =>
              tab.id === state.activeTabId && tab.type === 'request'
                ? { ...tab, ...patch }
                : tab,
            ),
          }
        }),
      setActiveScreen: (activeScreen) =>
        set({
          activeScreen,
          activeTabId: null,
        }),
      openNewRequest: () =>
        set((s) => ({
          tabs: [
            ...s.tabs,
            {
              id: `request-${s.requestEditorKey + 1}`,
              type: 'request',
              title: 'Untitled request',
              method: 'GET',
              isDirty: false,
            },
          ],
          activeTabId: `request-${s.requestEditorKey + 1}`,
          activeScreen: 'request',
          pendingHistoryEntry: null,
          requestEditorKey: s.requestEditorKey + 1,
        })),
      openWorkflow: (id) =>
        set((state) => {
          const tabId = `workflow-${id ?? 'draft'}`
          const existing = state.tabs.some((tab) => tab.id === tabId)
          return {
            tabs: existing
              ? state.tabs
              : [
                  ...state.tabs,
                  {
                    id: tabId,
                    type: 'workflow',
                    title: id ?? 'Untitled workflow',
                    method: 'WF',
                    isDirty: false,
                    workflowId: id,
                  },
                ],
            activeTabId: tabId,
            activeScreen: 'workflow',
            activeWorkflowId: id,
          }
        }),
      openWorkflowPlaceholder: () =>
        get().openWorkflow(null),
      openHistoryEntry: (entry) =>
        set((s) => ({
          tabs: [
            ...s.tabs,
            {
              id: `history-${entry.id}`,
              type: 'request',
              title: entry.requestName,
              method: entry.method,
              isDirty: false,
            },
          ],
          activeTabId: `history-${entry.id}`,
          activeScreen: 'request',
          pendingHistoryEntry: entry,
          requestEditorKey: s.requestEditorKey + 1,
        })),
      clearPendingHistoryEntry: () => set({ pendingHistoryEntry: null }),
      openRequestItem: (item) =>
        set((s) => {
          const tabId = `collection-${item.id}`
          const existing = s.tabs.some((tab) => tab.id === tabId)
          return {
            tabs: existing
              ? s.tabs
              : [
                  ...s.tabs,
                  {
                    id: tabId,
                    type: 'request',
                    title: item.name,
                    method: item.state.method,
                    isDirty: false,
                  },
                ],
            activeTabId: tabId,
            activeScreen: 'request',
            pendingRequestItem: item,
            pendingHistoryEntry: null,
            requestEditorKey: s.requestEditorKey + 1,
          }
        }),
      clearPendingRequestItem: () => set({ pendingRequestItem: null }),
      setRequestsSidebarView: (requestsSidebarView) => set({ requestsSidebarView }),
      setRequestsHistoryShowAll: (requestsHistoryShowAll) =>
        set({ requestsHistoryShowAll }),
    }),
    {
      name: 'tractl-ui',
      partialize: (state) => ({
        theme: state.theme,
        uiScale: state.uiScale,
        sidebarTab: state.sidebarTab,
        sidebarCollapsed: state.sidebarCollapsed,
        activityBarExpanded: state.activityBarExpanded,
        editorLayout: state.editorLayout,
        editorSplitRatio: state.editorSplitRatio,
        wordWrap: state.wordWrap,
        tabs: state.tabs,
        activeTabId: state.activeTabId,
        activeScreen: state.activeScreen,
        requestsSidebarView: state.requestsSidebarView,
        requestsHistoryShowAll: state.requestsHistoryShowAll,
      }),
      onRehydrateStorage: () => (state) => {
        if (!state?.activeTabId) return
        const tab = state.tabs.find((candidate) => candidate.id === state.activeTabId)
        if (!tab) return
        if (state.activeScreen === 'fast-start') {
          state.activeScreen = tab.type === 'request' ? 'request' : 'workflow'
        }
      },
    },
  ),
)
