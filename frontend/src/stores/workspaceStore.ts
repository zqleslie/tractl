import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { createEmptyRequestFormState } from '@/components/request-editor/createEmptyRequestFormState'
import type { HttpMethod } from '@/components/primitives'
import type { RequestState } from '@/components/request-editor/requestState'
import { getEngineDefaults } from '@/stores/engineDefaultsStore'

export type WorkspaceTab = {
  id: string
  type: 'request' | 'workflow'
  isDirty: boolean
  request?: RequestState
}

type WorkspaceStore = {
  requests: Record<string, RequestState>
  activeRequestId: string | null
  setActiveRequestId: (id: string | null) => void
  upsertRequest: (request: RequestState) => void
  patchRequest: (id: string, patch: Partial<RequestState>) => void
  getRequest: (id: string) => RequestState | undefined
  createRequest: (id: string, method?: HttpMethod) => RequestState
}

export const useWorkspaceStore = create<WorkspaceStore>()(
  persist(
    (set, get) => ({
      requests: {},
      activeRequestId: null,
      setActiveRequestId: (activeRequestId) => set({ activeRequestId }),
      upsertRequest: (request) =>
        set((state) => ({
          requests: { ...state.requests, [request.id]: request },
        })),
      patchRequest: (id, patch) =>
        set((state) => {
          const current = state.requests[id]
          if (!current) return state
          return {
            requests: {
              ...state.requests,
              [id]: { ...current, ...patch },
            },
          }
        }),
      getRequest: (id) => get().requests[id],
      createRequest: (id, method = 'GET') => {
        const draft = createEmptyRequestFormState()
        const request: RequestState = {
          id,
          name: 'Untitled request',
          method,
          url: '',
          params: draft.params,
          headers: draft.headers,
          body: {
            encoding: 'json',
            content: draft.body.value,
            formRows: [],
            rawContentType: 'text/plain',
          },
          auth: {
            type: 'none',
          },
          preScript: '',
          postScript: '',
          assertions: [],
          extracts: [],
          settings: {
            timeoutMs: getEngineDefaults().timeoutMs,
            retry: null,
            failurePolicy: getEngineDefaults().failurePolicy,
          },
        }
        set((state) => ({
          requests: { ...state.requests, [id]: request },
        }))
        return request
      },
    }),
    {
      name: 'tractl-workspace',
      partialize: (state) => ({
        requests: state.requests,
      }),
    },
  ),
)
