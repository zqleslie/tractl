import { serializeWorkflow } from '@/platform/web/workflowSerializer'
import { buildTopologySummary, useWorkflowCanvas } from './useWorkflowCanvas'
import { WorkflowBar } from './WorkflowBar'
import { FilterBar } from './FilterBar'
import { GraphView } from './GraphView'
import { CodeView } from './CodeView'
import { WorkflowConfigStrip } from './WorkflowConfigStrip'
import { StepDetailPopup } from './StepDetailPopup'
import { ResultBar } from './ResultBar'
import { FullResultsPopup } from './FullResultsPopup'

export function WorkflowCanvasScreen() {
  const {
    workflow,
    view,
    runState,
    autoSaveState,
    activeFilterId,
    openStepId,
    isConfigExpanded,
    isResultBarExpanded,
    isFullResultsOpen,
    steps,
    summary,
    sourceWorkflowNames,
    setView,
    openStep,
    toggleConfig,
    updateConfig,
    toggleResultBar,
    openFullResults,
    closeFullResults,
    handleRun,
    handleAddWorkflow,
    handleFilterChange,
  } = useWorkflowCanvas()

  if (workflow === null) {
    return <div className="h-full w-full bg-surface" />
  }

  return (
    <div className="relative flex h-full flex-col overflow-hidden bg-surface">
      <WorkflowBar
        workflowName={workflow.name}
        stepCount={steps.length}
        topologySummary={buildTopologySummary(steps)}
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
          <CodeView
            yaml={
              workflow.yaml && workflow.yaml.trim().length > 0
                ? workflow.yaml
                : serializeWorkflow(workflow)
            }
          />
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
