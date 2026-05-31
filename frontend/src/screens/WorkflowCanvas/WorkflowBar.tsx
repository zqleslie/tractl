import { useState } from 'react'
import type { KeyboardEvent } from 'react'
import { IconCode, IconGitBranch, IconLoader2, IconPlayerPlay } from '@tabler/icons-react'
import { cn } from '@/lib/cn'
import type { AutoSaveState, CanvasView } from '@/stores/workflowCanvasStore'
import type { RunState } from '@/types/workflow'

export interface WorkflowBarProps {
  workflowName: string
  stepCount: number
  topologySummary: string
  view: CanvasView
  onViewChange: (view: CanvasView) => void
  onRun: () => void
  runState: RunState
  autoSaveState: AutoSaveState
}

const toggleBase =
  'flex h-7 w-7 items-center justify-center rounded-[7px] border border-[0.5px] transition-colors'
const toggleInactive =
  'border-border text-text-muted hover:text-text'
const toggleActive =
  'border-primary bg-primary-bg text-primary'

export function WorkflowBar({
  workflowName,
  stepCount,
  topologySummary,
  view,
  onViewChange,
  onRun,
  runState,
  autoSaveState,
}: WorkflowBarProps) {
  const [isEditing, setIsEditing] = useState(false)
  const [draftName, setDraftName] = useState(workflowName)

  const commit = () => setIsEditing(false)
  const handleKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Enter') commit()
    if (event.key === 'Escape') {
      setDraftName(workflowName)
      setIsEditing(false)
    }
  }

  const isRunning = runState === 'running'

  return (
    <div className="flex h-10 shrink-0 items-center gap-3 border-b border-[0.5px] border-border bg-surface px-3">
      {isEditing ? (
        <input
          autoFocus
          value={draftName}
          onChange={(event) => setDraftName(event.target.value)}
          onBlur={commit}
          onKeyDown={handleKeyDown}
          className="border-0 border-b border-[0.5px] border-border bg-transparent px-0 text-[13px] font-[500] text-text outline-none focus:border-primary"
          aria-label="Workflow name"
        />
      ) : (
        <button
          type="button"
          onClick={() => {
            setDraftName(workflowName)
            setIsEditing(true)
          }}
          className="truncate text-[13px] font-[500] text-text"
        >
          {workflowName}
        </button>
      )}

      <span className="truncate text-[11px] font-[400] text-text-muted">
        {stepCount} {stepCount === 1 ? 'step' : 'steps'} · {topologySummary}
      </span>

      <div className="flex-1" />

      <div className="flex items-center gap-1">
        <button
          type="button"
          onClick={() => onViewChange('graph')}
          title="Graph view"
          aria-label="Graph view"
          aria-pressed={view === 'graph'}
          className={cn(toggleBase, view === 'graph' ? toggleActive : toggleInactive)}
        >
          <IconGitBranch size={15} stroke={1.5} />
        </button>
        <button
          type="button"
          onClick={() => onViewChange('code')}
          title="Code view"
          aria-label="Code view"
          aria-pressed={view === 'code'}
          className={cn(toggleBase, view === 'code' ? toggleActive : toggleInactive)}
        >
          <IconCode size={15} stroke={1.5} />
        </button>
      </div>

      <span className="text-[11px] font-[400] text-text-muted">
        {autoSaveState === 'saving' ? 'Saving…' : '✓ Saved'}
      </span>

      <button
        type="button"
        onClick={onRun}
        disabled={isRunning}
        className="flex h-7 items-center gap-1.5 rounded-[7px] bg-primary px-3 text-[12px] font-[500] text-primary-fg disabled:opacity-70"
      >
        {isRunning ? (
          <>
            <IconLoader2 size={14} stroke={1.5} className="animate-spin" />
            Running…
          </>
        ) : (
          <>
            <IconPlayerPlay size={14} stroke={1.5} />
            Run
          </>
        )}
      </button>
    </div>
  )
}
