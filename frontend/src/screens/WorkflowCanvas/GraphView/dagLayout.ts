import type { WorkflowLayoutResponse } from '@/platform/types'

export type DependencyKind = 'explicit' | 'implicit'

export interface DagNode {
  stepId: string
  x: number
  y: number
  row: number
  col: number
  columnGroup: number
}

export interface DagEdge {
  fromId: string
  toId: string
  edgeType: 'sequential' | 'fanout' | 'merge'
  dependencyKind: DependencyKind
}

export interface BranchBar {
  parentId: string
  row: number
  xStart: number
  xEnd: number
  xMid: number
  y: number
}

export interface GroupDivider {
  x: number
  leftGroup: number
  rightGroup: number
}

export interface DagLayout {
  nodes: DagNode[]
  edges: DagEdge[]
  branchBars: BranchBar[]
  groupDividers: GroupDivider[]
  canvasWidth: number
  canvasHeight: number
  groupCount: number
}

const CARD_W = 260
const CARD_H = 110
const V_GAP = 80
const H_GAP = 60
const GROUP_PADDING = 72
const DIVIDER_WIDTH = 48
const BRANCH_BAR_OFFSET_Y = 32
const CANVAS_PADDING = 80

// computeDagLayout converts engine layout data to pixel positions.
// No graph theory. Pure coordinate math.
export function computeDagLayout(layout: WorkflowLayoutResponse): DagLayout {
  if (layout.rows.length === 0) {
    return { nodes: [], edges: [], branchBars: [], groupDividers: [], canvasWidth: 0, canvasHeight: 0, groupCount: 0 }
  }

  const groupCount = Math.max(layout.groups.length, 1)

  // Build step → row/group lookups
  const stepRow = new Map<string, number>()
  for (const row of layout.rows) {
    for (const id of row.stepIds) {
      stepRow.set(id, row.rowIndex)
    }
  }

  const stepGroup = new Map<string, number>()
  for (const group of layout.groups) {
    for (const id of group.stepIds) {
      stepGroup.set(id, group.groupIndex)
    }
  }

  // For each (groupIndex, rowIndex) bucket, collect stepIds in order
  const bucketKey = (g: number, r: number) => `${g}:${r}`
  const buckets = new Map<string, string[]>()
  for (const row of layout.rows) {
    for (const id of row.stepIds) {
      const g = stepGroup.get(id) ?? 0
      const key = bucketKey(g, row.rowIndex)
      if (!buckets.has(key)) buckets.set(key, [])
      buckets.get(key)!.push(id)
    }
  }

  // Column index within each bucket
  const stepBucketCol = new Map<string, number>()
  for (const [, ids] of buckets) {
    ids.forEach((id, idx) => stepBucketCol.set(id, idx))
  }

  const rowWidthOf = (n: number): number => n * CARD_W + Math.max(0, n - 1) * H_GAP

  // Group max widths
  const groupWidth: number[] = new Array<number>(groupCount).fill(0)
  for (const [key, ids] of buckets) {
    const g = Number(key.split(':')[0])
    const w = rowWidthOf(ids.length)
    if (w > (groupWidth[g] ?? 0)) groupWidth[g] = w
  }

  // Group x start positions
  const groupXStart: number[] = new Array<number>(groupCount).fill(0)
  groupXStart[0] = GROUP_PADDING
  for (let i = 1; i < groupCount; i++) {
    groupXStart[i] = (groupXStart[i - 1] ?? 0) + (groupWidth[i - 1] ?? 0) + DIVIDER_WIDTH + GROUP_PADDING
  }

  // Pixel positions
  const nodes: DagNode[] = []
  const nodeById = new Map<string, DagNode>()
  const allIds = new Set(layout.rows.flatMap((r) => r.stepIds))

  for (const id of allIds) {
    const g = stepGroup.get(id) ?? 0
    const r = stepRow.get(id) ?? 0
    const c = stepBucketCol.get(id) ?? 0
    const key = bucketKey(g, r)
    const bucketSize = buckets.get(key)?.length ?? 1
    const rowW = rowWidthOf(bucketSize)
    const xOffset = ((groupWidth[g] ?? 0) - rowW) / 2
    const x = (groupXStart[g] ?? 0) + xOffset + c * (CARD_W + H_GAP)
    const y = r * (CARD_H + V_GAP)
    const node: DagNode = { stepId: id, x, y, row: r, col: c, columnGroup: g }
    nodes.push(node)
    nodeById.set(id, node)
  }

  // Edges
  const edges: DagEdge[] = layout.edges.map((e) => {
    let edgeType: DagEdge['edgeType'] = 'sequential'
    if (e.kind === 'fanout') edgeType = 'fanout'
    else if (e.kind === 'merge') edgeType = 'merge'
    return { fromId: e.from, toId: e.to, edgeType, dependencyKind: e.implicit ? 'implicit' : 'explicit' }
  })

  // Branch bars — one per fan-out parent
  const fanoutChildren = new Map<string, string[]>()
  for (const e of layout.edges) {
    if (e.kind === 'fanout') {
      if (!fanoutChildren.has(e.from)) fanoutChildren.set(e.from, [])
      fanoutChildren.get(e.from)!.push(e.to)
    }
  }

  const branchBars: BranchBar[] = []
  for (const [parentId, children] of fanoutChildren) {
    if (children.length < 2) continue
    const parent = nodeById.get(parentId)
    if (!parent) continue
    const centres = children.map((id) => (nodeById.get(id)?.x ?? 0) + CARD_W / 2)
    const xStart = Math.min(...centres)
    const xEnd = Math.max(...centres)
    branchBars.push({
      parentId,
      row: parent.row,
      xStart,
      xEnd,
      xMid: (xStart + xEnd) / 2,
      y: parent.y + CARD_H + BRANCH_BAR_OFFSET_Y,
    })
  }

  // Group dividers
  const groupDividers: GroupDivider[] = []
  for (let i = 0; i < groupCount - 1; i++) {
    groupDividers.push({
      x: (groupXStart[i + 1] ?? 0) - DIVIDER_WIDTH / 2,
      leftGroup: i,
      rightGroup: i + 1,
    })
  }

  // Canvas size
  let maxRight = 0
  let maxBottom = 0
  for (const n of nodes) {
    maxRight = Math.max(maxRight, n.x + CARD_W)
    maxBottom = Math.max(maxBottom, n.y + CARD_H)
  }

  return {
    nodes,
    edges,
    branchBars,
    groupDividers,
    canvasWidth: maxRight + CANVAS_PADDING,
    canvasHeight: maxBottom + CANVAS_PADDING,
    groupCount,
  }
}
