import type { TraCtlStep } from '@/components/request-editor/tractlSpecDocument'

/** Matches `${steps.<stepId>...}` per internal/normalizer/scan.go and tractl_spec §13.5 */
const STEP_REF_RE = /\$\{steps\.([a-zA-Z][a-zA-Z0-9_-]*)\b/g

function scanStepRefs(...values: string[]): Set<string> {
  const refs = new Set<string>()
  for (const value of values) {
    if (!value) continue
    const pattern = new RegExp(STEP_REF_RE.source, 'g')
    for (const match of value.matchAll(pattern)) {
      const stepId = match[1]
      if (stepId) refs.add(stepId)
    }
  }
  return refs
}

function scanRefsInValue(value: unknown, into: Set<string>): void {
  if (typeof value === 'string') {
    for (const ref of scanStepRefs(value)) into.add(ref)
    return
  }
  if (Array.isArray(value)) {
    for (const entry of value) scanRefsInValue(entry, into)
    return
  }
  if (value && typeof value === 'object') {
    for (const entry of Object.values(value as Record<string, unknown>)) {
      scanRefsInValue(entry, into)
    }
  }
}

/** Collect step IDs referenced in expression-bearing fields (mirrors Go collectStepRefs). */
export function collectReferencedStepIds(step: TraCtlStep): string[] {
  const refs = new Set<string>()

  const when = (step as TraCtlStep & { when?: string }).when
  if (when) {
    for (const ref of scanStepRefs(when)) refs.add(ref)
  }

  for (const ref of scanStepRefs(step.request.target)) refs.add(ref)

  const headers = step.request.headers ?? {}
  for (const value of Object.values(headers)) {
    for (const ref of scanStepRefs(value)) refs.add(ref)
  }

  if (step.request.body?.content !== undefined) {
    scanRefsInValue(step.request.body.content, refs)
  }

  return [...refs].sort()
}

/**
 * Returns same-workflow step IDs referenced in expressions but not listed in dependsOn.
 * Cross-workflow references are ignored (engine resolves at runtime; tractl_spec §12.2).
 */
export function inferImplicitDependsOn(
  step: TraCtlStep,
  workflowStepIds: Set<string>,
): string[] {
  const explicit = new Set(step.dependsOn ?? [])
  const implicit: string[] = []

  for (const ref of collectReferencedStepIds(step)) {
    if (ref === step.id) continue
    if (!workflowStepIds.has(ref)) continue
    if (explicit.has(ref)) continue
    implicit.push(ref)
  }

  return implicit
}
