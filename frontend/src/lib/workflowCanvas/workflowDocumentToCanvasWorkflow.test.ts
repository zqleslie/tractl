import { describe, expect, it } from 'vitest'
import type { TraCtlSpecDocument } from '@/components/request-editor/tractlSpecDocument'
import { serializeWorkflow } from '@/platform/web/workflowSerializer'
import { workflowDocumentToCanvasWorkflow } from '@/lib/workflowCanvas/workflowDocumentToCanvasWorkflow'

const document: TraCtlSpecDocument = {
  schemaVersion: 1,
  capabilities: ['protocol.http'],
  metadata: { name: 'Auth flow' },
  variables: {
    baseUrl: 'https://httpbin.org',
  },
  workflows: [
    {
      id: 'wf-auth',
      name: 'Auth flow',
      concurrency: 2,
      failurePolicy: 'resilient',
      steps: [
        {
          id: 'step-auth',
          kind: 'request',
          dependsOn: [],
          request: {
            protocol: 'http',
            target: 'https://httpbin.org/get',
            operation: 'GET',
          },
          assertions: [{ id: 'a1', kind: 'status', op: 'eq', expected: 200 }],
        },
        {
          id: 'step-billing',
          kind: 'request',
          dependsOn: ['step-auth'],
          request: {
            protocol: 'http',
            target: 'https://httpbin.org/billing',
            operation: 'GET',
            headers: { Authorization: 'Bearer token' },
          },
          extracts: [{ id: 'billing-token', source: 'body', path: 'token' }],
        },
      ],
    },
  ],
}

describe('workflowDocumentToCanvasWorkflow', () => {
  it('maps spec workflows to canvas graph state without run results', () => {
    const workflow = workflowDocumentToCanvasWorkflow(document, {
      sourceName: 'auth-flow.yaml',
      raw: 'schemaVersion: 1',
    })

    expect(workflow.id).toBe('wf-auth')
    expect(workflow.name).toBe('Auth flow')
    expect(workflow.config.concurrency).toBe(2)
    expect(workflow.steps).toHaveLength(2)
    expect(workflow.steps[0]?.result).toBeUndefined()
    expect(workflow.steps[0]?.assertionCount).toBe(1)
    expect(workflow.steps[0]?.extractCount).toBe(0)
    expect(workflow.steps[1]?.hasAuth).toBe(true)
    expect(workflow.steps[1]?.dependsOn).toEqual(['step-auth'])
    expect(workflow.steps[1]?.extractCount).toBe(1)
    expect(workflow.steps[0]?.assertions).toHaveLength(1)
    expect(workflow.variables).toEqual({ baseUrl: 'https://httpbin.org' })
    expect(workflow.yaml).toBe('schemaVersion: 1')
  })

  it('round-trips spec variables and step assertions through the canvas serializer', () => {
    const workflow = workflowDocumentToCanvasWorkflow(document)
    const yaml = serializeWorkflow(workflow)

    expect(yaml).toContain('variables:')
    expect(yaml).toContain('baseUrl: "https://httpbin.org"')
    expect(yaml).toContain('assertions:')
    expect(yaml).toContain('step-auth')
  })

  it('leaves concurrency unset when the workflow omits it', () => {
    const withoutConcurrency: TraCtlSpecDocument = {
      schemaVersion: 1,
      capabilities: ['protocol.http'],
      workflows: [
        {
          id: 'wf-plain',
          steps: [
            {
              id: 'step-a',
              kind: 'request',
              request: {
                protocol: 'http',
                target: 'https://httpbin.org/get',
                operation: 'GET',
              },
            },
          ],
        },
      ],
    }

    const workflow = workflowDocumentToCanvasWorkflow(withoutConcurrency)
    expect(workflow.config.concurrency).toBeUndefined()
    expect(serializeWorkflow(workflow)).not.toContain('concurrency:')
  })

  it('infers implicit dependsOn from ${steps.<id>} references', () => {
    const diamond: TraCtlSpecDocument = {
      schemaVersion: 1,
      capabilities: ['protocol.http'],
      workflows: [
        {
          id: 'diamond-flow',
          steps: [
            {
              id: 'step-1',
              kind: 'request',
              request: {
                protocol: 'http',
                target: 'https://httpbin.org/status/200',
                operation: 'GET',
              },
            },
            {
              id: 'step-2',
              kind: 'request',
              dependsOn: ['step-1'],
              request: {
                protocol: 'http',
                target: 'https://httpbin.org/get',
                operation: 'GET',
              },
              extracts: [{ id: 'branch-url', source: 'body', path: 'url' }],
            },
            {
              id: 'step-3',
              kind: 'request',
              dependsOn: ['step-1'],
              request: {
                protocol: 'http',
                target: 'https://httpbin.org/headers',
                operation: 'GET',
              },
            },
            {
              id: 'step-4',
              kind: 'request',
              dependsOn: ['step-3'],
              request: {
                protocol: 'http',
                target: '${steps.step-2.extracts.branch-url}',
                operation: 'GET',
              },
            },
          ],
        },
      ],
    }

    const workflow = workflowDocumentToCanvasWorkflow(diamond)
    expect(workflow.steps.find((step) => step.id === 'step-4')?.dependsOn).toEqual(['step-3'])
    // implicit deps are now inferred by engine.inferDeps() — not the frontend
    expect(workflow.steps.find((step) => step.id === 'step-4')?.implicitDependsOn).toBeUndefined()
  })
})
