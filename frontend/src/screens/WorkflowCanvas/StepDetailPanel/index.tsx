import { useState } from 'react'
import { IconX } from '@tabler/icons-react'
import { MethodBadge } from '@/components/primitives'
import type { WorkflowStep } from '@/types/workflow'
import { StepDetailTabs } from './StepDetailTabs'
import type { StepDetailTab } from './types'
import { useStepDraftEditor } from './useStepDraftEditor'
import { AssertionsTab } from './tabs/AssertionsTab'
import { DependenciesTab } from './tabs/DependenciesTab'
import { ExtractsTab } from './tabs/ExtractsTab'
import { PostScriptTab } from './tabs/PostScriptTab'
import { PreScriptTab } from './tabs/PreScriptTab'
import { RequestTab } from './tabs/RequestTab'
import { ResultTab } from './tabs/ResultTab'
import { VariablesPanel } from './VariablesPanel'

export type { StepDetailTab } from './types'

export interface StepDetailPanelProps {
  step: WorkflowStep
  allSteps: WorkflowStep[]
  workflowVariables?: Record<string, string>
  defaultTab?: StepDetailTab
  onClose: () => void
}

export function StepDetailPanel({
  step,
  allSteps,
  workflowVariables,
  defaultTab,
  onClose,
}: StepDetailPanelProps) {
  const [activeTab, setActiveTab] = useState<StepDetailTab>(defaultTab ?? 'request')
  const editor = useStepDraftEditor(step)

  const hasResult = step.result !== undefined
  const failedAssertionCount =
    step.result?.assertions.filter((a) => !a.passed).length ?? 0

  return (
    <aside className="absolute right-0 top-0 z-20 flex h-full w-[min(640px,55vw)] flex-col border-l border-[0.5px] border-border bg-surface">
      <header className="flex h-10 shrink-0 items-center gap-2 border-b border-[0.5px] border-border px-3">
        <MethodBadge method={editor.method} />
        <input
          aria-label="Step id"
          value={editor.stepId}
          onChange={(event) => editor.setStepId(event.currentTarget.value)}
          onBlur={editor.commitStepId}
          onKeyDown={(event) => {
            if (event.key === 'Enter') event.currentTarget.blur()
          }}
          className="max-w-[140px] shrink-0 rounded-ui border-[0.5px] border-border bg-surface px-1.5 py-0.5 font-mono text-ui-xs text-text outline-none focus:border-primary"
        />
        <span className="min-w-0 flex-1 truncate font-mono text-ui-xs text-text-muted">
          {editor.url || '(no url)'}
        </span>
        <button
          type="button"
          onClick={onClose}
          title="Close"
          aria-label="Close step detail"
          className="ml-auto flex h-7 w-7 shrink-0 items-center justify-center rounded-ui text-text-muted hover:bg-surface-elevated hover:text-text"
        >
          <IconX size={14} stroke={1.75} />
        </button>
      </header>

      <StepDetailTabs
        activeTab={activeTab}
        onTabChange={setActiveTab}
        tabCounts={{
          assertions: editor.tabCounts.assertions,
          extracts: editor.tabCounts.extracts,
        }}
        hasResult={hasResult}
        resultPassed={step.result?.outcome === 'passed'}
        failedAssertionCount={failedAssertionCount}
      />

      <div className="flex-1 overflow-auto p-4">
        {activeTab === 'request' && <RequestTab editor={editor} />}
        {activeTab === 'dependencies' && (
          <DependenciesTab step={step} allSteps={allSteps} />
        )}
        {activeTab === 'pre-script' && <PreScriptTab editor={editor} />}
        {activeTab === 'post-script' && <PostScriptTab editor={editor} />}
        {activeTab === 'assertions' && <AssertionsTab editor={editor} />}
        {activeTab === 'extracts' && <ExtractsTab editor={editor} />}
        {activeTab === 'result' && <ResultTab step={step} />}
      </div>

      <VariablesPanel
        step={step}
        allSteps={allSteps}
        workflowVariables={workflowVariables}
      />
    </aside>
  )
}
