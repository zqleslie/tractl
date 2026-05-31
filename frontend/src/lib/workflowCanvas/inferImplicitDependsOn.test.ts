import { describe, expect, it } from 'vitest'
import type { TraCtlStep } from '@/components/request-editor/tractlSpecDocument'
import {
  collectReferencedStepIds,
  inferImplicitDependsOn,
} from '@/lib/workflowCanvas/inferImplicitDependsOn'

const workflowIds = new Set(['step-1', 'step-2', 'step-3', 'step-4'])

function requestStep(
  id: string,
  target: string,
  dependsOn?: string[],
): TraCtlStep {
  return {
    id,
    kind: 'request',
    dependsOn,
    request: { protocol: 'http', target, operation: 'GET' },
  }
}

describe('inferImplicitDependsOn', () => {
  it('detects ${steps.stepId.extracts.name} on request target', () => {
    const refs = collectReferencedStepIds(
      requestStep('step-4', '${steps.step-2.extracts.branch-url}', ['step-3']),
    )
    expect(refs).toEqual(['step-2'])
  })

  it('returns implicit deps not already in dependsOn', () => {
    const implicit = inferImplicitDependsOn(
      requestStep('step-4', '${steps.step-2.extracts.branch-url}', ['step-3']),
      workflowIds,
    )
    expect(implicit).toEqual(['step-2'])
  })

  it('omits refs already declared in dependsOn', () => {
    const implicit = inferImplicitDependsOn(
      requestStep('step-3', '${vars.baseUrl}/headers', ['step-1']),
      workflowIds,
    )
    expect(implicit).toEqual([])
  })

  it('ignores unknown step ids outside the workflow', () => {
    const implicit = inferImplicitDependsOn(
      requestStep('step-4', '${steps.ghost.extracts.token}', ['step-3']),
      workflowIds,
    )
    expect(implicit).toEqual([])
  })
})
