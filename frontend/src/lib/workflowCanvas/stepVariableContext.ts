import type { TraCtlExtract } from '@/components/request-editor/tractlSpecDocument'
import type { WorkflowStep } from '@/types/workflow'

const EXPRESSION_REF_RE = /\$\{([^}]+)\}/g

/** Read-only runtime keys documented for `${runtime.*}` (tractl_spec §13). */
export const RUNTIME_VARIABLE_REFERENCES = [
  'runtime.runId',
  'runtime.surface',
  'runtime.timestamp',
] as const

export type UpstreamStepExtractRef = {
  stepId: string
  extractId: string
  reference: string
  source?: string
  path?: string
}

function collectAncestorIds(step: WorkflowStep, allSteps: WorkflowStep[]): Set<string> {
  const byId = new Map(allSteps.map((entry) => [entry.id, entry]))
  const ancestors = new Set<string>()
  const pending = [...step.dependsOn, ...(step.implicitDependsOn ?? [])]

  while (pending.length > 0) {
    const stepId = pending.pop()
    if (!stepId || ancestors.has(stepId)) continue
    const ancestor = byId.get(stepId)
    if (!ancestor) continue

    ancestors.add(stepId)
    pending.push(...ancestor.dependsOn, ...(ancestor.implicitDependsOn ?? []))
  }

  return ancestors
}

/** Upstream steps in workflow document order (transitive dependsOn + implicit deps). */
export function listUpstreamSteps(
  step: WorkflowStep,
  allSteps: WorkflowStep[],
): WorkflowStep[] {
  const ancestorIds = collectAncestorIds(step, allSteps)
  return allSteps.filter((entry) => ancestorIds.has(entry.id))
}

export function listUpstreamStepExtracts(
  step: WorkflowStep,
  allSteps: WorkflowStep[],
): UpstreamStepExtractRef[] {
  const refs: UpstreamStepExtractRef[] = []

  for (const upstream of listUpstreamSteps(step, allSteps)) {
    for (const extract of upstream.extracts ?? []) {
      refs.push({
        stepId: upstream.id,
        extractId: extract.id,
        reference: `steps.${upstream.id}.extracts.${extract.id}`,
        source: extract.source,
        path: extract.path,
      })
    }
  }

  return refs
}

/** Unique `${...}` inner expressions found in a string (trimmed, stable order). */
export function parseVariableReferences(value: string): string[] {
  if (!value.includes('${')) return []

  const seen = new Set<string>()
  const ordered: string[] = []
  const pattern = new RegExp(EXPRESSION_REF_RE.source, 'g')

  for (const match of value.matchAll(pattern)) {
    const ref = match[1]?.trim()
    if (!ref || seen.has(ref)) continue
    seen.add(ref)
    ordered.push(ref)
  }

  return ordered
}

export function formatWorkflowVariableReference(key: string): string {
  return `vars.${key}`
}

export function workflowVariableRows(
  variables: Record<string, string> | undefined,
): Array<{ key: string; value: string; reference: string }> {
  if (!variables) return []

  return Object.entries(variables)
    .sort(([left], [right]) => left.localeCompare(right))
    .map(([key, value]) => ({
      key,
      value,
      reference: formatWorkflowVariableReference(key),
    }))
}

export function extractReferenceLabel(extract: TraCtlExtract, stepId: string): string {
  return `steps.${stepId}.extracts.${extract.id}`
}
