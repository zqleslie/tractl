import { describe, it, expect } from 'vitest'
import { computeDagLayout } from './dagLayout'
import type { WorkflowLayoutResponse } from '@/platform/types'

function makeLayout(overrides: Partial<WorkflowLayoutResponse> = {}): WorkflowLayoutResponse {
  return {
    rows: [],
    edges: [],
    groups: [],
    topologySummary: '',
    ...overrides,
  }
}

describe('computeDagLayout', () => {
  it('returns empty layout for empty rows', () => {
    const result = computeDagLayout(makeLayout())
    expect(result.nodes).toHaveLength(0)
    expect(result.canvasWidth).toBe(0)
    expect(result.groupCount).toBe(0)
  })

  it('places a single node at group-padded x and row 0 y', () => {
    const layout = makeLayout({
      rows: [{ rowIndex: 0, stepIds: ['A'] }],
      groups: [{ groupIndex: 0, rootStepId: 'A', stepIds: ['A'] }],
    })
    const result = computeDagLayout(layout)
    expect(result.nodes).toHaveLength(1)
    expect(result.nodes[0]!.stepId).toBe('A')
    expect(result.nodes[0]!.row).toBe(0)
    expect(result.nodes[0]!.col).toBe(0)
    expect(result.nodes[0]!.columnGroup).toBe(0)
  })

  it('places two sequential nodes at different rows', () => {
    const layout = makeLayout({
      rows: [
        { rowIndex: 0, stepIds: ['A'] },
        { rowIndex: 1, stepIds: ['B'] },
      ],
      edges: [{ from: 'A', to: 'B', kind: 'sequential', implicit: false }],
      groups: [{ groupIndex: 0, rootStepId: 'A', stepIds: ['A', 'B'] }],
    })
    const result = computeDagLayout(layout)
    const nodeA = result.nodes.find((n) => n.stepId === 'A')!
    const nodeB = result.nodes.find((n) => n.stepId === 'B')!
    expect(nodeA.y).toBeLessThan(nodeB.y)
    expect(nodeA.x).toBe(nodeB.x)
  })

  it('classifies fanout edges correctly', () => {
    const layout = makeLayout({
      rows: [
        { rowIndex: 0, stepIds: ['A'] },
        { rowIndex: 1, stepIds: ['B', 'C'] },
      ],
      edges: [
        { from: 'A', to: 'B', kind: 'fanout', implicit: false },
        { from: 'A', to: 'C', kind: 'fanout', implicit: false },
      ],
      groups: [{ groupIndex: 0, rootStepId: 'A', stepIds: ['A', 'B', 'C'] }],
    })
    const result = computeDagLayout(layout)
    expect(result.edges.every((e) => e.edgeType === 'fanout')).toBe(true)
    expect(result.branchBars).toHaveLength(1)
    expect(result.branchBars[0]!.parentId).toBe('A')
  })

  it('creates a group divider between two independent groups', () => {
    const layout = makeLayout({
      rows: [
        { rowIndex: 0, stepIds: ['A', 'X'] },
      ],
      edges: [],
      groups: [
        { groupIndex: 0, rootStepId: 'A', stepIds: ['A'] },
        { groupIndex: 1, rootStepId: 'X', stepIds: ['X'] },
      ],
    })
    const result = computeDagLayout(layout)
    expect(result.groupCount).toBe(2)
    expect(result.groupDividers).toHaveLength(1)
  })
})
