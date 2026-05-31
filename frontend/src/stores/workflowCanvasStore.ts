import { create } from 'zustand'
import type { StatusTone } from '@/components/primitives'
import { toHistorySourceFormat } from '@/lib/tractlDocument/inferDocumentFormat'
import { parseAndValidateDocument } from '@/lib/tractlDocument/parseValidateDocument'
import { buildLastRunStatusLabel } from '@/lib/workflowCanvas/lastRunStatusLabel'
import { workflowDocumentToCanvasWorkflow } from '@/lib/workflowCanvas/workflowDocumentToCanvasWorkflow'
import { serializeWorkflow } from '@/platform/web/workflowSerializer'
import type { CanvasRunOutcome } from '@/platform/web/workflowRunAdapter'
import type { TractlWasmParseFormat } from '@/platform/web/wasm/loadTractlWasmRuntime'
import { useUiStore } from '@/stores/uiStore'
import { useWorkflowWorkspaceStore } from '@/stores/workflowWorkspaceStore'
import type {
  RunState,
  StepResult,
  Workflow,
  WorkflowConfig,
  WorkflowRunSummary,
  WorkflowStep,
} from '@/types/workflow'

const SKIPPED_STEP_RESULT: StepResult = {
  outcome: 'skipped',
  durationMs: 0,
  assertions: [],
  extracts: [],
}

export type CanvasView = 'graph' | 'code'
export type AutoSaveState = 'saved' | 'saving'
export type OpenStepDefaultTab =
  | 'request'
  | 'dependencies'
  | 'pre-script'
  | 'post-script'
  | 'assertions'
  | 'extracts'
  | 'result'

export interface OpenStepOptions {
  defaultTab?: OpenStepDefaultTab
}

const READY_STATUS = {
  runState: 'idle' as RunState,
  runOutcome: null,
  runDuration: 0,
  runErrorMessage: null,
  lastRunLabel: 'Engine ready',
  lastRunTone: 'ready' as StatusTone,
  openStepId: null,
  openStepDefaultTab: 'request' as OpenStepDefaultTab,
  isFullResultsOpen: false,
  isResultBarExpanded: false,
}

function createBlankStep(id: string, dependsOn: string[] = []): WorkflowStep {
  return {
    id,
    method: 'GET',
    url: '',
    dependsOn,
    hasAuth: false,
    assertionCount: 0,
    hasPreScript: false,
    extractCount: 0,
  }
}

/** Canvas edits must re-serialize on run; stale opened-file yaml would otherwise win. */
function withoutStaleYaml(workflow: Workflow): Workflow {
  if (workflow.yaml === undefined) return workflow
  const { yaml: _yaml, ...rest } = workflow
  return rest
}

function persistCanvasWorkflow(workflow: Workflow): void {
  const workspace = useWorkflowWorkspaceStore.getState()
  const entry = workspace.getWorkflowEntry(workflow.id)
  if (entry) {
    workspace.syncWorkflow(workflow)
  } else {
    workspace.upsertWorkflow(workflow, {
      sourceName: workflow.name,
      sourceFormat: 'yaml',
    })
  }
  if (useUiStore.getState().activeWorkflowId === null) {
    useUiStore.getState().openWorkflow(workflow.id)
  }
}

/** Refresh stored YAML from canvas state so config edits apply on the next run. */
function withSerializedYaml(workflow: Workflow): Workflow {
  const yaml = serializeWorkflow({ ...workflow, yaml: undefined })
  return { ...workflow, yaml }
}

function applyLoadCanvasWorkflow(
  previous: WorkflowCanvasState,
  workflow: Workflow,
): Partial<WorkflowCanvasState> {
  const workspace = useWorkflowWorkspaceStore.getState()
  if (workspace.getWorkflowEntry(workflow.id)) {
    workspace.syncWorkflow(workflow)
  } else {
    workspace.upsertWorkflow(workflow, {
      sourceName: workflow.name,
      sourceFormat: 'yaml',
    })
  }

  const switchingWorkflow = previous.workflow?.id !== workflow.id
  const activeFilterId = switchingWorkflow
    ? workflow.name
    : (previous.activeFilterId ?? workflow.name)

  if (!switchingWorkflow && previous.workflow?.id === workflow.id) {
    return { workflow, activeFilterId }
  }

  return {
    workflow,
    view: 'graph',
    activeFilterId,
    ...READY_STATUS,
  }
}

interface WorkflowCanvasState {
  workflow: Workflow | null
  view: CanvasView
  activeFilterId: string | null
  runState: RunState
  runOutcome: WorkflowRunSummary['outcome'] | null
  runDuration: number
  runErrorMessage: string | null
  lastRunLabel: string
  lastRunTone: StatusTone
  autoSaveState: AutoSaveState
  openStepId: string | null
  openStepDefaultTab: OpenStepDefaultTab
  isConfigExpanded: boolean
  isResultBarExpanded: boolean
  isFullResultsOpen: boolean

  setWorkflow: (workflow: Workflow) => void
  loadWorkflow: (document: string, id?: string, format?: TractlWasmParseFormat) => void
  loadCanvasWorkflow: (workflow: Workflow) => void
  clearWorkflow: () => void
  setView: (view: CanvasView) => void
  setActiveFilter: (id: string | null) => void
  setRunState: (runState: RunState) => void
  setAutoSaveState: (autoSaveState: AutoSaveState) => void
  openStep: (id: string, options?: OpenStepOptions) => void
  closeStep: () => void
  setConfigExpanded: (expanded: boolean) => void
  toggleConfig: () => void
  setResultBarExpanded: (expanded: boolean) => void
  toggleResultBar: () => void
  openFullResults: () => void
  closeFullResults: () => void
  updateConfig: (config: WorkflowConfig) => void
  applyRunResult: (outcome: CanvasRunOutcome) => void

  updateStep: (stepId: string, patch: Partial<WorkflowStep>) => void
  renameStep: (oldId: string, newId: string) => void
  insertStepAfter: (afterStepId: string) => void
  insertParallelStep: (parentStepId: string) => void
  insertStepBetween: (fromStepId: string, toStepId: string) => void
}

export const useWorkflowCanvasStore = create<WorkflowCanvasState>((set) => ({
  workflow: null,
  view: 'graph',
  activeFilterId: null,
  runState: 'idle',
  runOutcome: null,
  runDuration: 0,
  runErrorMessage: null,
  lastRunLabel: 'Engine ready',
  lastRunTone: 'ready',
  autoSaveState: 'saved',
  openStepId: null,
  openStepDefaultTab: 'request',
  isConfigExpanded: false,
  isResultBarExpanded: false,
  isFullResultsOpen: false,

  setWorkflow: (workflow) => {
    persistCanvasWorkflow(workflow)
    set({ workflow })
  },

  loadWorkflow: (document, id, format = 'yaml') => {
    void parseAndValidateDocument(document, format)
      .then((result) => {
        if (!result.ok) {
          console.warn('[tractl:workflow-load] could not parse workflow document', result.message)
          return
        }

        const workflow = workflowDocumentToCanvasWorkflow(result.spec, {
          raw: document,
        })
        const clientWorkflowId = id ?? `wf-${crypto.randomUUID().slice(0, 8)}`
        const loadedWorkflow = { ...workflow, id: clientWorkflowId, yaml: document }

        useWorkflowWorkspaceStore.getState().upsertWorkflow(loadedWorkflow, {
          sourceName: loadedWorkflow.name,
          sourceFormat: toHistorySourceFormat(format),
        })
        set((previous) => applyLoadCanvasWorkflow(previous, loadedWorkflow))
      })
      .catch((error) => {
        console.warn(
          '[tractl:workflow-load] could not load workflow document',
          error instanceof Error ? error.message : error,
        )
      })
  },
  loadCanvasWorkflow: (workflow) =>
    set((previous) => applyLoadCanvasWorkflow(previous, workflow)),
  clearWorkflow: () =>
    set({
      workflow: null,
      ...READY_STATUS,
    }),
  setView: (view) => set({ view }),
  setActiveFilter: (activeFilterId) => set({ activeFilterId }),
  setRunState: (runState) =>
    set((state) => {
      const status = buildLastRunStatusLabel({
        runState,
        workflowOutcome: state.runOutcome,
        durationMs: state.runDuration,
        errorMessage: state.runErrorMessage ?? undefined,
      })
      return {
        runState,
        lastRunLabel: status.label,
        lastRunTone: status.tone,
      }
    }),
  setAutoSaveState: (autoSaveState) => set({ autoSaveState }),
  openStep: (openStepId, options) =>
    set({ openStepId, openStepDefaultTab: options?.defaultTab ?? 'request' }),
  closeStep: () => set({ openStepId: null }),
  setConfigExpanded: (isConfigExpanded) => set({ isConfigExpanded }),
  toggleConfig: () => set((s) => ({ isConfigExpanded: !s.isConfigExpanded })),
  setResultBarExpanded: (isResultBarExpanded) => set({ isResultBarExpanded }),
  toggleResultBar: () => set((s) => ({ isResultBarExpanded: !s.isResultBarExpanded })),
  openFullResults: () => set({ isFullResultsOpen: true }),
  closeFullResults: () => set({ isFullResultsOpen: false }),
  updateConfig: (config) =>
    set((s) => {
      if (!s.workflow) return {}
      const workflow = withSerializedYaml({ ...s.workflow, config })
      persistCanvasWorkflow(workflow)
      return { workflow }
    }),

  updateStep: (stepId, patch) =>
    set((state) => {
      if (!state.workflow) return {}
      const steps = state.workflow.steps.map((step) =>
        step.id === stepId ? { ...step, ...patch } : step,
      )
      const workflow = withoutStaleYaml({ ...state.workflow, steps })
      persistCanvasWorkflow(workflow)
      return { workflow }
    }),

  renameStep: (oldId, newId) =>
    set((state) => {
      if (!state.workflow) return {}

      const trimmed = newId.trim()
      if (!trimmed || trimmed === oldId) return {}
      if (state.workflow.steps.some((step) => step.id === trimmed)) return {}

      const steps = state.workflow.steps.map((step) => {
        if (step.id === oldId) {
          return { ...step, id: trimmed }
        }
        return {
          ...step,
          dependsOn: step.dependsOn.map((dependency) =>
            dependency === oldId ? trimmed : dependency,
          ),
          implicitDependsOn: step.implicitDependsOn?.map((dependency) =>
            dependency === oldId ? trimmed : dependency,
          ),
        }
      })
      const workflow = withoutStaleYaml({ ...state.workflow, steps })
      persistCanvasWorkflow(workflow)

      return {
        workflow,
        openStepId: state.openStepId === oldId ? trimmed : state.openStepId,
      }
    }),

  applyRunResult: (outcome) =>
    set((state) => {
      if (!state.workflow) return {}

      const hasStepResults = Object.keys(outcome.stepOutcomes).length > 0
      const steps: WorkflowStep[] = state.workflow.steps.map((step): WorkflowStep => {
        const mapped = outcome.stepOutcomes[step.id]
        if (mapped) {
          return { ...step, result: mapped }
        }
        if (!hasStepResults) {
          const { result: _removed, ...stepWithoutResult } = step
          return stepWithoutResult
        }
        return {
          ...step,
          result: SKIPPED_STEP_RESULT,
        }
      })

      const assertionsTotal = steps.reduce(
        (sum, step) => sum + (step.result?.assertions.length ?? 0),
        0,
      )
      const assertionsPassed = steps.reduce(
        (sum, step) =>
          sum + (step.result?.assertions.filter((assertion) => assertion.passed).length ?? 0),
        0,
      )
      const failedSteps = steps.filter((step) => step.result?.outcome === 'failed').length

      const status = buildLastRunStatusLabel({
        runState: outcome.workflowOutcome === 'error' ? 'error' : 'complete',
        workflowOutcome: outcome.workflowOutcome,
        durationMs: outcome.duration,
        assertionsPassed,
        assertionsTotal,
        failedSteps,
        errorMessage: outcome.errorMessage,
      })

      const nextWorkflow = { ...state.workflow, steps }
      persistCanvasWorkflow(nextWorkflow)

      return {
        workflow: nextWorkflow,
        runOutcome: outcome.workflowOutcome,
        runDuration: outcome.duration,
        runErrorMessage: outcome.errorMessage ?? null,
        lastRunLabel: status.label,
        lastRunTone: status.tone,
      }
    }),

  insertStepAfter: (afterStepId) =>
    set((state) => {
      if (!state.workflow) return {}

      const afterStepIndex = state.workflow.steps.findIndex((step) => step.id === afterStepId)
      if (afterStepIndex === -1 && state.workflow.steps.length > 0) return {}

      const newStep = createBlankStep(
        `step-${Date.now()}`,
        afterStepId === '' ? [] : [afterStepId],
      )
      const insertIndex = afterStepIndex === -1 ? 0 : afterStepIndex + 1
      const steps = [
        ...state.workflow.steps.slice(0, insertIndex),
        newStep,
        ...state.workflow.steps.slice(insertIndex),
      ]
      const workflow = withoutStaleYaml({ ...state.workflow, steps })
      persistCanvasWorkflow(workflow)

      return {
        workflow,
        openStepId: newStep.id,
        openStepDefaultTab: 'request',
      }
    }),
  insertParallelStep: (parentStepId) =>
    set((state) => {
      if (!state.workflow) return {}

      const parentStepIndex = state.workflow.steps.findIndex((step) => step.id === parentStepId)
      if (parentStepIndex === -1) return {}

      const newStep = createBlankStep(`step-${Date.now()}`, [parentStepId])
      const lastChildIndex = state.workflow.steps.reduce(
        (lastIndex, step, index) =>
          step.dependsOn.includes(parentStepId) ? index : lastIndex,
        parentStepIndex,
      )
      const insertIndex = lastChildIndex + 1
      const steps = [
        ...state.workflow.steps.slice(0, insertIndex),
        newStep,
        ...state.workflow.steps.slice(insertIndex),
      ]
      const workflow = withoutStaleYaml({ ...state.workflow, steps })
      persistCanvasWorkflow(workflow)

      return {
        workflow,
        openStepId: newStep.id,
        openStepDefaultTab: 'request',
      }
    }),
  insertStepBetween: (fromStepId, toStepId) =>
    set((state) => {
      if (!state.workflow) return {}

      const fromStepIndex = state.workflow.steps.findIndex((step) => step.id === fromStepId)
      const toStepIndex = state.workflow.steps.findIndex((step) => step.id === toStepId)
      if (fromStepIndex === -1 || toStepIndex === -1) return {}

      const newStep = createBlankStep(`step-${Date.now()}`, [fromStepId])
      const rewiredSteps = state.workflow.steps.map((step) => {
        if (step.id !== toStepId) return step
        return {
          ...step,
          dependsOn: step.dependsOn.map((dependency) =>
            dependency === fromStepId ? newStep.id : dependency,
          ),
        }
      })
      const insertIndex = fromStepIndex < toStepIndex ? toStepIndex : fromStepIndex + 1
      const steps = [
        ...rewiredSteps.slice(0, insertIndex),
        newStep,
        ...rewiredSteps.slice(insertIndex),
      ]
      const workflow = withoutStaleYaml({ ...state.workflow, steps })
      persistCanvasWorkflow(workflow)

      return {
        workflow,
        openStepId: newStep.id,
        openStepDefaultTab: 'request',
      }
    }),
}))
