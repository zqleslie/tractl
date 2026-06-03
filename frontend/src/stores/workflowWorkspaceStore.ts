import { create } from 'zustand'
import type { RunHistorySourceFormat } from '@/stores/runHistoryStore'
import type { Workflow } from '@/types/workflow'

type OpenedWorkflowRecord = {
  workflow: Workflow
  sourceName: string
  sourceFormat: RunHistorySourceFormat
}

type WorkflowWorkspaceState = {
  entries: OpenedWorkflowRecord[]
  upsertWorkflow: (
    workflow: Workflow,
    meta: { sourceName: string; sourceFormat: RunHistorySourceFormat },
  ) => void
  syncWorkflow: (workflow: Workflow) => void
  removeWorkflow: (workflowId: string) => void
  getWorkflow: (workflowId: string) => Workflow | undefined
  getWorkflowEntry: (workflowId: string) => OpenedWorkflowRecord | undefined
  listWorkflows: () => Workflow[]
}

export const useWorkflowWorkspaceStore = create<WorkflowWorkspaceState>((set, get) => ({
  entries: [],

  upsertWorkflow: (workflow, meta) =>
    set((state) => {
      const rest = state.entries.filter((entry) => entry.workflow.id !== workflow.id)
      return {
        entries: [{ workflow, sourceName: meta.sourceName, sourceFormat: meta.sourceFormat }, ...rest],
      }
    }),

  syncWorkflow: (workflow) =>
    set((state) => {
      let changed = false
      const entries = state.entries.map((entry) => {
        if (entry.workflow.id !== workflow.id) return entry
        changed = true
        return { ...entry, workflow }
      })
      return changed ? { entries } : state
    }),

  removeWorkflow: (workflowId) =>
    set((state) => ({
      entries: state.entries.filter((entry) => entry.workflow.id !== workflowId),
    })),

  getWorkflow: (workflowId) =>
    get().entries.find((entry) => entry.workflow.id === workflowId)?.workflow,

  getWorkflowEntry: (workflowId) =>
    get().entries.find((entry) => entry.workflow.id === workflowId),

  listWorkflows: () => get().entries.map((entry) => entry.workflow),
}))
