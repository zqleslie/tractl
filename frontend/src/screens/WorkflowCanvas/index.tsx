import { useEffect, useMemo, useRef, useState } from 'react'
import { useShallow } from 'zustand/react/shallow'
import { detectSurface } from '@/platform'
import { engine } from '@/platform/engine'
import { mapRunResult } from '@/platform/web/workflowRunAdapter'
import { workflowToExportRequest } from '@/lib/workflowCanvas/workflowToExportRequest'
import { buildWorkflowRunPayload } from '@/lib/workflowCanvas/buildWorkflowRunPayload'
import { createEmptyWorkflow } from '@/lib/workflowCanvas/createEmptyWorkflow'
import { loadWorkflowIntoCanvas } from '@/lib/workflowCanvas/loadWorkflowIntoCanvas'
import { useUiStore } from '@/stores/uiStore'
import { useWorkflowCanvasStore } from '@/stores/workflowCanvasStore'
import { useWorkflowWorkspaceStore } from '@/stores/workflowWorkspaceStore'
import type { WorkflowRunSummary, WorkflowStep } from '@/types/workflow'
import { WorkflowBar } from './WorkflowBar'
import { FilterBar } from './FilterBar'
import { GraphView } from './GraphView'
import { CodeView } from './CodeView'
import { WorkflowConfigStrip } from './WorkflowConfigStrip'
import { StepDetailPopup } from './StepDetailPopup'
import { ResultBar } from './ResultBar'
import { FullResultsPopup } from './FullResultsPopup'

function resolveStepStartTimes(steps: WorkflowStep[]): Map<string, number> {
  const resolved = new Map<string, number>()
  for (const step of steps) {
    resolved.set(step.id, step.result?.startMs ?? 0)
  }
  return resolved
}

const WATERFALL_COLS = 24

function deriveRunSummary(
  steps: WorkflowStep[],
  engineOutcome?: WorkflowRunSummary['outcome'] | null,
  engineDurationMs?: number,
  errorMessage?: string | null,
): WorkflowRunSummary {
  const ran = steps.filter((step) => step.result !== undefined)
  const assertionsTotal = ran.reduce((sum, step) => sum + (step.result?.assertions.length ?? 0), 0)
  const assertionsPassed = ran.reduce(
    (sum, step) => sum + (step.result?.assertions.filter((a) => a.passed).length ?? 0),
    0,
  )
  const totalDurationMs =
    engineDurationMs ??
    ran.reduce((sum, step) => sum + (step.result?.durationMs ?? 0), 0)
  const anyFailed = ran.some((step) => step.result?.outcome === 'failed')

  const stepStartTimes = resolveStepStartTimes(steps)
  const timelineEndMs = Math.max(
    ...steps.map((step) => {
      const start = stepStartTimes.get(step.id) ?? 0
      return start + (step.result?.durationMs ?? 0)
    }),
    1,
  )

  const waterfallText = steps
    .map((step) => {
      const name = step.id.padEnd(14, ' ')
      const result = step.result

      if (!result) return `${name}  (not run)`

      const start = stepStartTimes.get(step.id) ?? 0
      const durationMs = result.durationMs ?? 0
      const offsetCols = Math.round((start / timelineEndMs) * WATERFALL_COLS)
      const barCols = Math.max(1, Math.round((durationMs / timelineEndMs) * WATERFALL_COLS))
      const mark = result.outcome === 'passed' ? '✓' : '✗'

      return `${name}${' '.repeat(offsetCols)}${'█'.repeat(barCols)} ${durationMs}ms  ${result.statusCode ?? '—'} ${mark}`
    })
    .join('\n')

  return {
    outcome: engineOutcome ?? (anyFailed ? 'failed' : 'passed'),
    totalDurationMs,
    stepsRan: ran.length,
    totalSteps: steps.length,
    assertionsPassed,
    assertionsTotal,
    waterfallText,
    errorMessage: errorMessage ?? undefined,
  }
}

export function WorkflowCanvasScreen() {
  const activeWorkflowId = useUiStore((s) => s.activeWorkflowId)
  const workflow = useWorkflowCanvasStore((s) => s.workflow)
  const view = useWorkflowCanvasStore((s) => s.view)
  const runState = useWorkflowCanvasStore((s) => s.runState)
  const autoSaveState = useWorkflowCanvasStore((s) => s.autoSaveState)
  const activeFilterId = useWorkflowCanvasStore((s) => s.activeFilterId)
  const openStepId = useWorkflowCanvasStore((s) => s.openStepId)
  const isConfigExpanded = useWorkflowCanvasStore((s) => s.isConfigExpanded)
  const isResultBarExpanded = useWorkflowCanvasStore((s) => s.isResultBarExpanded)
  const isFullResultsOpen = useWorkflowCanvasStore((s) => s.isFullResultsOpen)

  const [codeViewYaml, setCodeViewYaml] = useState('')
  const engineLayout = useWorkflowCanvasStore((s) => s.layout)
  const setLayout = useWorkflowCanvasStore((s) => s.setLayout)
  const setView = useWorkflowCanvasStore((s) => s.setView)
  const setActiveFilter = useWorkflowCanvasStore((s) => s.setActiveFilter)
  const runOutcome = useWorkflowCanvasStore((s) => s.runOutcome)
  const runDuration = useWorkflowCanvasStore((s) => s.runDuration)
  const runErrorMessage = useWorkflowCanvasStore((s) => s.runErrorMessage)
  const setRunState = useWorkflowCanvasStore((s) => s.setRunState)
  const applyRunResult = useWorkflowCanvasStore((s) => s.applyRunResult)
  const setResultBarExpanded = useWorkflowCanvasStore((s) => s.setResultBarExpanded)
  const openStep = useWorkflowCanvasStore((s) => s.openStep)
  const toggleConfig = useWorkflowCanvasStore((s) => s.toggleConfig)
  const updateConfig = useWorkflowCanvasStore((s) => s.updateConfig)
  const toggleResultBar = useWorkflowCanvasStore((s) => s.toggleResultBar)
  const openFullResults = useWorkflowCanvasStore((s) => s.openFullResults)
  const closeFullResults = useWorkflowCanvasStore((s) => s.closeFullResults)

  const sourceWorkflowEntries = useWorkflowWorkspaceStore(
    useShallow((s) => {
      const activeEntry = activeWorkflowId
        ? s.entries.find((entry) => entry.workflow.id === activeWorkflowId)
        : undefined

      if (!activeEntry) {
        if (!workflow) return []
        const canvasEntry = s.entries.find((entry) => entry.workflow.id === workflow.id)
        return canvasEntry ? [canvasEntry] : []
      }

      return s.entries.filter(
        (entry) =>
          entry.sourceName === activeEntry.sourceName &&
          entry.sourceFormat === activeEntry.sourceFormat,
      )
    }),
  )
  const sourceWorkflowNames = useMemo(
    () => sourceWorkflowEntries.map((entry) => entry.workflow.name),
    [sourceWorkflowEntries],
  )

  const handleAddWorkflow = () => {
    const nextWorkflow = createEmptyWorkflow()
    useWorkflowWorkspaceStore.getState().upsertWorkflow(nextWorkflow, {
      sourceName: nextWorkflow.name,
      sourceFormat: 'yaml',
    })
    useUiStore.getState().openWorkflow(nextWorkflow.id)
    loadWorkflowIntoCanvas(nextWorkflow)
    setActiveFilter(nextWorkflow.name)
  }

  const handleFilterChange = (filterId: string | null) => {
    setActiveFilter(filterId)
    if (filterId === null) return

    const entry = sourceWorkflowEntries.find(
      (item) => item.workflow.name === filterId || item.workflow.id === filterId,
    )
    if (entry) {
      useUiStore.getState().openWorkflow(entry.workflow.id)
    }
  }

  useEffect(() => {
    if (activeWorkflowId === null) {
      const existing = useWorkflowCanvasStore.getState().workflow
      if (existing) return
      const workflow = createEmptyWorkflow()
      useWorkflowWorkspaceStore.getState().upsertWorkflow(workflow, {
        sourceName: workflow.name,
        sourceFormat: 'yaml',
      })
      useUiStore.getState().openWorkflow(workflow.id)
      loadWorkflowIntoCanvas(workflow)
      return
    }

    const canvasWorkflow = useWorkflowCanvasStore.getState().workflow
    if (canvasWorkflow?.id === activeWorkflowId) {
      return
    }

    const workspace = useWorkflowWorkspaceStore.getState()
    const entry = workspace.getWorkflowEntry(activeWorkflowId)
    if (entry) {
      loadWorkflowIntoCanvas(entry.workflow)
      return
    }

    if (detectSurface() === 'desktop') {
      // TODO: load workflow via GET /api/v1/files/:id when the local API exposes file read.
      console.warn(
        '[tractl:workflow-load] workflow not in workspace; desktop file read API is not implemented',
        { activeWorkflowId },
      )
    }
  }, [activeWorkflowId])

  useEffect(() => {
    // TODO: wire Cmd+Enter to run the workflow.
  }, [])

  const steps = useMemo(() => workflow?.steps ?? [], [workflow])

  const stepStructureKey = useMemo(
    () => JSON.stringify(steps.map((s) => ({ id: s.id, d: s.dependsOn ?? [] }))),
    [steps],
  )
  useEffect(() => {
    setLayout(null)
    if (steps.length === 0) return
    const refs = steps.map((s) => ({ id: s.id, dependsOn: s.dependsOn ?? [] }))
    engine.computeLayout(refs)
      .then(setLayout)
      .catch(console.error)
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [stepStructureKey])

  const codeViewGenRef = useRef(0)
  useEffect(() => {
    if (view !== 'code' || !workflow) return
    if (workflow.yaml?.trim()) { setCodeViewYaml(workflow.yaml); return }
    const gen = ++codeViewGenRef.current
    engine.exportWorkflow(workflowToExportRequest(workflow))
      .then((yaml) => { if (codeViewGenRef.current === gen) setCodeViewYaml(yaml) })
      .catch(console.error)
  }, [view, workflow])

  const summary = useMemo(
    () => deriveRunSummary(steps, runOutcome, runDuration, runErrorMessage),
    [steps, runOutcome, runDuration, runErrorMessage],
  )
  const handleRun = async () => {
    if (runState === 'running' || workflow === null) return

    setRunState('running')
    setResultBarExpanded(true)

    try {
      const entry = useWorkflowWorkspaceStore.getState().getWorkflowEntry(workflow.id)
      const { document, format, source } = await buildWorkflowRunPayload(workflow, {
        sourceFormat: entry?.sourceFormat,
      })
      console.info('[tractl:workflow-run] payload sent to Go', {
        format,
        workflowId: workflow.id,
        source,
        surface: detectSurface(),
        document,
      })
      const engineResult = await engine.runWorkflow(document, format)
      const outcome = mapRunResult(engineResult as Parameters<typeof mapRunResult>[0], workflow.id)
      applyRunResult(outcome)
      setRunState(outcome.workflowOutcome === 'error' ? 'error' : 'complete')
    } catch (error) {
      applyRunResult({
        workflowOutcome: 'error',
        duration: 0,
        stepOutcomes: {},
        errorMessage: error instanceof Error ? error.message : 'Workflow run failed',
      })
      setRunState('error')
    }
  }

  if (workflow === null) {
    return <div className="h-full w-full bg-surface" />
  }

  return (
    <div className="relative flex h-full flex-col overflow-hidden bg-surface">
      <WorkflowBar
        workflowName={workflow.name}
        stepCount={steps.length}
        topologySummary={engineLayout?.topologySummary ?? ''}
        view={view}
        onViewChange={setView}
        onRun={handleRun}
        runState={runState}
        autoSaveState={autoSaveState}
      />

      <FilterBar
        workflowNames={sourceWorkflowNames}
        activeFilterId={activeFilterId}
        onFilterChange={handleFilterChange}
        onAddWorkflow={handleAddWorkflow}
      />

      <div className="relative flex-1 overflow-hidden">
        {view === 'graph' ? (
          <GraphView />
        ) : (
          <CodeView yaml={codeViewYaml} />
        )}

        {openStepId !== null ? (
          <div className="pointer-events-none absolute inset-0 z-10 bg-black/20" />
        ) : null}

        <StepDetailPopup />
      </div>

      <WorkflowConfigStrip
        config={workflow.config}
        isExpanded={isConfigExpanded}
        onToggle={toggleConfig}
        onChange={updateConfig}
      />

      <ResultBar
        isExpanded={isResultBarExpanded}
        onToggle={toggleResultBar}
        runState={runState}
        summary={summary}
        steps={steps}
        onStepChipClick={openStep}
        onViewFullResults={openFullResults}
      />

      {isFullResultsOpen && (
        <FullResultsPopup summary={summary} steps={steps} onClose={closeFullResults} />
      )}
    </div>
  )
}
