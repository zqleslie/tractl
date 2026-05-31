import { useState, type ReactNode } from 'react'
import { IconChevronDown, IconChevronUp } from '@tabler/icons-react'
import { cn } from '@/lib/cn'
import {
  RUNTIME_VARIABLE_REFERENCES,
  listUpstreamStepExtracts,
  workflowVariableRows,
  type UpstreamStepExtractRef,
} from '@/lib/workflowCanvas/stepVariableContext'
import type { WorkflowStep } from '@/types/workflow'

export interface VariablesPanelProps {
  step: WorkflowStep
  allSteps: WorkflowStep[]
  workflowVariables?: Record<string, string>
}

function KeyValueRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="grid grid-cols-[minmax(0,1fr)_minmax(0,1fr)] gap-2 text-ui-xs">
      <span className="truncate font-mono text-info-fg">{label}</span>
      <span className="truncate font-mono text-text">{value}</span>
    </div>
  )
}

function ExtractRefRow({ entry }: { entry: UpstreamStepExtractRef }) {
  const detail = [entry.source, entry.path].filter(Boolean).join(' · ')

  return (
    <div className="space-y-0.5">
      <p className="font-mono text-ui-xs text-info-fg">{entry.reference}</p>
      {detail ? <p className="text-[10px] text-text-muted">{detail}</p> : null}
    </div>
  )
}

function SectionHeading({ children }: { children: ReactNode }) {
  return (
    <h4 className="text-[10px] font-medium uppercase tracking-wide text-text-muted">
      {children}
    </h4>
  )
}

export function VariablesPanel({
  step,
  allSteps,
  workflowVariables,
}: VariablesPanelProps) {
  const [isExpanded, setIsExpanded] = useState(true)
  const workflowVars = workflowVariableRows(workflowVariables)
  const upstreamExtracts = listUpstreamStepExtracts(step, allSteps)

  return (
    <section className="shrink-0 border-t border-[0.5px] border-border bg-surface">
      <button
        type="button"
        onClick={() => setIsExpanded((expanded) => !expanded)}
        aria-expanded={isExpanded}
        className="flex h-9 w-full items-center gap-2 px-4 text-left"
      >
        <span className="text-ui-xs font-medium text-text">Variables</span>
        <span className="text-[10px] text-text-muted">available in this step</span>
        <span className="flex-1" />
        {isExpanded ? (
          <IconChevronDown size={14} stroke={1.5} className="text-text-muted" />
        ) : (
          <IconChevronUp size={14} stroke={1.5} className="text-text-muted" />
        )}
      </button>

      {isExpanded ? (
        <div className="space-y-4 px-4 pb-4">
          <div className="space-y-2">
            <SectionHeading>Workflow vars (${'vars.*'})</SectionHeading>
            {workflowVars.length === 0 ? (
              <p className="text-ui-xs text-text-muted">No workflow variables defined</p>
            ) : (
              <div className="space-y-1.5 rounded-ui border-[0.5px] border-border bg-surface-elevated p-2">
                {workflowVars.map((row) => (
                  <KeyValueRow
                    key={row.key}
                    label={`\${${row.reference}}`}
                    value={row.value}
                  />
                ))}
              </div>
            )}
          </div>

          {upstreamExtracts.length > 0 ? (
            <div className="space-y-2">
              <SectionHeading>Step extracts (${'steps.*.extracts.*'})</SectionHeading>
              <div className="space-y-2 rounded-ui border-[0.5px] border-border bg-surface-elevated p-2">
                {upstreamExtracts.map((entry) => (
                  <ExtractRefRow key={`${entry.stepId}-${entry.extractId}`} entry={entry} />
                ))}
              </div>
            </div>
          ) : null}

          <div className="space-y-2">
            <SectionHeading>Runtime vars (${'runtime.*'})</SectionHeading>
            <div className="flex flex-wrap gap-1.5">
              {RUNTIME_VARIABLE_REFERENCES.map((reference) => (
                <span
                  key={reference}
                  className={cn(
                    'inline-flex items-center rounded-full border border-[0.5px] border-border bg-surface-elevated px-2 py-0.5 font-mono text-[10px] text-text-muted',
                  )}
                  title="Read-only at execution time"
                >
                  {reference}
                </span>
              ))}
            </div>
          </div>
        </div>
      ) : null}
    </section>
  )
}
