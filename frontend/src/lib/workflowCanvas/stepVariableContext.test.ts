import { describe, expect, it } from 'vitest'
import type { WorkflowStep } from '@/types/workflow'
import {
  listUpstreamStepExtracts,
  listUpstreamSteps,
  parseVariableReferences,
  workflowVariableRows,
} from '@/lib/workflowCanvas/stepVariableContext'

const steps: WorkflowStep[] = [
  {
    id: 'step-1',
    method: 'GET',
    url: 'https://example.com/1',
    dependsOn: [],
    hasAuth: false,
    assertionCount: 0,
    hasPreScript: false,
    extractCount: 0,
  },
  {
    id: 'step-2',
    method: 'GET',
    url: 'https://example.com/2',
    dependsOn: ['step-1'],
    hasAuth: false,
    assertionCount: 0,
    hasPreScript: false,
    extractCount: 1,
    extracts: [{ id: 'branch-url', source: 'body', path: 'url' }],
  },
  {
    id: 'step-3',
    method: 'GET',
    url: 'https://example.com/3',
    dependsOn: ['step-1'],
    hasAuth: false,
    assertionCount: 0,
    hasPreScript: false,
    extractCount: 0,
  },
  {
    id: 'step-4',
    method: 'GET',
    url: '${steps.step-2.extracts.branch-url}',
    dependsOn: ['step-3'],
    implicitDependsOn: ['step-2'],
    hasAuth: false,
    assertionCount: 0,
    hasPreScript: false,
    extractCount: 0,
  },
]

describe('stepVariableContext', () => {
  it('parses unique variable references from a string', () => {
    expect(
      parseVariableReferences(
        '${vars.baseUrl}/x and ${steps.step-2.extracts.token} ${vars.baseUrl}',
      ),
    ).toEqual(['vars.baseUrl', 'steps.step-2.extracts.token'])
  })

  it('lists upstream steps in workflow order including implicit deps', () => {
    const current = steps.find((step) => step.id === 'step-4')
    expect(current).toBeDefined()
    expect(listUpstreamSteps(current!, steps).map((step) => step.id)).toEqual([
      'step-1',
      'step-2',
      'step-3',
    ])
  })

  it('lists upstream extract references for the variables panel', () => {
    const current = steps.find((step) => step.id === 'step-4')
    expect(current).toBeDefined()
    expect(listUpstreamStepExtracts(current!, steps)).toEqual([
      {
        stepId: 'step-2',
        extractId: 'branch-url',
        reference: 'steps.step-2.extracts.branch-url',
        source: 'body',
        path: 'url',
      },
    ])
  })

  it('builds workflow variable rows with vars.* references', () => {
    expect(workflowVariableRows({ baseUrl: 'https://httpbin.org' })).toEqual([
      {
        key: 'baseUrl',
        value: 'https://httpbin.org',
        reference: 'vars.baseUrl',
      },
    ])
  })
})
