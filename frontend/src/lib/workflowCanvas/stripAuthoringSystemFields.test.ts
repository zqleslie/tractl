import { describe, expect, it } from 'vitest'
import { serializeWorkflow } from '@/platform/web/workflowSerializer'
import { workflowDocumentToCanvasWorkflow } from '@/lib/workflowCanvas/workflowDocumentToCanvasWorkflow'
import { stripAuthoringSystemFields } from '@/lib/workflowCanvas/stripAuthoringSystemFields'
import type { TraCtlSpecDocument } from '@/components/request-editor/tractlSpecDocument'

describe('stripAuthoringSystemFields', () => {
  it('removes _ulid and reserved-prefix keys recursively', () => {
    const input = {
      id: 'step-1',
      _ulid: '01ARZ3NDEKTSV4RRFFQ69G5FAV',
      assertions: [{ id: 'a1', _ulid: '01ARZ3NDEKTSV4RRFFQ69G5FAV', kind: 'status' }],
      '_traCtl.debug': true,
    }

    expect(stripAuthoringSystemFields(input)).toEqual({
      id: 'step-1',
      assertions: [{ id: 'a1', kind: 'status' }],
    })
  })
})

describe('serializeWorkflow', () => {
  it('does not emit _ulid fields from parsed step metadata', () => {
    const document = {
      schemaVersion: 1,
      capabilities: ['protocol.http'],
      workflows: [
        {
          id: 'wf-1',
          steps: [
            {
              id: 'step-1',
              kind: 'request',
              request: {
                protocol: 'http',
                target: 'https://example.com',
                operation: 'GET',
              },
              assertions: [
                {
                  id: 'a1',
                  kind: 'status',
                  op: 'eq',
                  expected: 200,
                  _ulid: '01ARZ3NDEKTSV4RRFFQ69G5FAV',
                },
              ],
            },
          ],
        },
      ],
    } as unknown as TraCtlSpecDocument

    const workflow = workflowDocumentToCanvasWorkflow(document)
    const yaml = serializeWorkflow(workflow)

    expect(yaml).not.toContain('_ulid')
    expect(yaml).toContain('assertions:')
    expect(yaml).toContain('kind: status')
  })
})
