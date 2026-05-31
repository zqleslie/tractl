import { IconTopologyStar3 } from '@tabler/icons-react'
import type { WorkflowStep } from '@/types/workflow'

export interface DependenciesTabProps {
  step: WorkflowStep
  allSteps: WorkflowStep[]
}

function SectionHeading({ children }: { children: string }) {
  return (
    <h3 className="mb-2 mt-4 text-ui-xs font-medium uppercase tracking-wide text-text-muted first:mt-0">
      {children}
    </h3>
  )
}

function StepChip({ stepId }: { stepId: string }) {
  return (
    <span className="inline-flex items-center gap-1 rounded-full border border-[0.5px] border-border bg-surface-elevated px-2.5 py-0.5 font-mono text-ui-xs text-text">
      <IconTopologyStar3 size={11} stroke={1.75} aria-hidden />
      {stepId}
    </span>
  )
}

function ChipList({ stepIds }: { stepIds: string[] }) {
  return (
    <div className="flex flex-wrap gap-2">
      {stepIds.map((id) => (
        <StepChip key={id} stepId={id} />
      ))}
    </div>
  )
}

function sameDependsOn(a: WorkflowStep, b: WorkflowStep): boolean {
  if (a.dependsOn.length !== b.dependsOn.length) return false
  return a.dependsOn.every((dep) => b.dependsOn.includes(dep))
}

export function DependenciesTab({ step, allSteps }: DependenciesTabProps) {
  const downstream = allSteps.filter((s) => s.dependsOn.includes(step.id))
  const siblings = allSteps.filter(
    (s) => s.id !== step.id && sameDependsOn(s, step),
  )

  return (
    <div className="space-y-1 text-ui-xs text-text">
      <SectionHeading>Depends on</SectionHeading>
      {step.dependsOn.length === 0 ? (
        <p className="text-text-muted">Root step — no dependencies.</p>
      ) : (
        <ChipList stepIds={step.dependsOn} />
      )}

      <SectionHeading>Required by</SectionHeading>
      {downstream.length === 0 ? (
        <p className="text-text-muted">No steps depend on this step.</p>
      ) : (
        <ChipList stepIds={downstream.map((s) => s.id)} />
      )}

      {siblings.length > 0 ? (
        <>
          <SectionHeading>Parallel eligibility</SectionHeading>
          <p className="text-text-muted">
            Runs in parallel with:{' '}
            <span className="inline-flex flex-wrap items-center gap-2 align-middle">
              {siblings.map((s) => (
                <StepChip key={s.id} stepId={s.id} />
              ))}
            </span>
          </p>
        </>
      ) : null}
    </div>
  )
}
