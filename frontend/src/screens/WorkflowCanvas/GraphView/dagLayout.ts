export interface LayoutStep {
  id: string;
  dependsOn: string[];
  implicitDependsOn?: string[];
}

export type DependencyKind = 'explicit' | 'implicit';

export interface DagNode {
  stepId: string;
  x: number;
  y: number;
  row: number;
  col: number;
  columnGroup: number;
}

export interface DagEdge {
  fromId: string;
  toId: string;
  edgeType: 'sequential' | 'fanout' | 'merge';
  dependencyKind: DependencyKind;
}

export interface BranchBar {
  parentId: string;
  row: number;
  xStart: number;
  xEnd: number;
  xMid: number;
  y: number;
}

export interface GroupDivider {
  x: number;
  leftGroup: number;
  rightGroup: number;
}

export interface DagLayout {
  nodes: DagNode[];
  edges: DagEdge[];
  branchBars: BranchBar[];
  groupDividers: GroupDivider[];
  canvasWidth: number;
  canvasHeight: number;
  groupCount: number;
}

const CARD_W = 260;
const CARD_H = 110;
const V_GAP = 80;
const H_GAP = 60;
const GROUP_PADDING = 72;
const DIVIDER_WIDTH = 48;
const BRANCH_BAR_OFFSET_Y = 32;
const CANVAS_PADDING = 80;

function allDependsOn(step: LayoutStep): string[] {
  return [...step.dependsOn, ...(step.implicitDependsOn ?? [])];
}

function classifyEdgeType(
  parentChildCount: number,
  stepParentCount: number,
): DagEdge['edgeType'] {
  if (parentChildCount > 1 && stepParentCount === 1) {
    return 'fanout';
  }
  if (stepParentCount > 1 && parentChildCount === 1) {
    return 'merge';
  }
  if (parentChildCount > 1 && stepParentCount > 1) {
    return 'merge';
  }
  return 'sequential';
}

/**
 * Compute a deterministic, top-to-bottom DAG layout for a list of workflow steps.
 *
 * The layout is organised into independent "column groups" — one per root step
 * (a step with no dependencies). Steps reachable from multiple roots are merged
 * into the leftmost (lowest-index) group. Within each group, rows come from a
 * topological sort (Kahn's algorithm) where a step's row is one greater than the
 * maximum row of its parents, and columns are centred per row.
 *
 * The function is pure: it has no side effects, performs no I/O, and always
 * returns the same output for the same input. Determinism is anchored on the
 * original input ordering of `steps` (never on id string comparison).
 *
 * @param steps - The workflow steps, each with an id and its parent dependencies.
 * @returns A fully resolved {@link DagLayout} with pixel positions, classified
 *   edges, branch bars, group dividers and canvas dimensions.
 * @throws If a dependency references an unknown step id.
 * @throws If a cycle (including a self-reference) is detected.
 */
interface Cell { g: number; r: number; ids: string[] }

function buildGraph(steps: LayoutStep[]): {
  indexById: Map<string, number>;
  stepById: (id: string) => LayoutStep;
  byInputIndex: (a: string, b: string) => number;
  childrenOf: Map<string, string[]>;
} {
  const indexById = new Map<string, number>();
  steps.forEach((step, i) => indexById.set(step.id, i));

  for (const step of steps) {
    for (const depId of allDependsOn(step)) {
      if (!indexById.has(depId)) {
        throw new Error(`Step "${step.id}" depends on unknown step "${depId}"`);
      }
    }
  }

  const stepById = (id: string): LayoutStep => steps[indexById.get(id)!]!;
  const byInputIndex = (a: string, b: string): number => indexById.get(a)! - indexById.get(b)!;

  const childrenOf = new Map<string, string[]>();
  for (const step of steps) childrenOf.set(step.id, []);
  for (const step of steps) {
    for (const depId of allDependsOn(step)) {
      childrenOf.get(depId)!.push(step.id);
    }
  }

  // Cycle detection via DFS. A self-reference lands the node in the visiting set.
  const visited = new Set<string>();
  const visiting = new Set<string>();
  const detectCycle = (id: string): void => {
    if (visited.has(id)) return;
    if (visiting.has(id)) throw new Error(`Cycle detected involving step: ${id}`);
    visiting.add(id);
    for (const depId of allDependsOn(stepById(id))) detectCycle(depId);
    visiting.delete(id);
    visited.add(id);
  };
  for (const step of steps) detectCycle(step.id);

  return { indexById, stepById, byInputIndex, childrenOf };
}

function assignRows(
  steps: LayoutStep[],
  stepById: (id: string) => LayoutStep,
  childrenOf: Map<string, string[]>,
  byInputIndex: (a: string, b: string) => number,
): Map<string, number> {
  const inDegree = new Map<string, number>();
  for (const step of steps) inDegree.set(step.id, allDependsOn(step).length);

  const row = new Map<string, number>();
  const queue: string[] = [];
  for (const step of steps) {
    if (allDependsOn(step).length === 0) { row.set(step.id, 0); queue.push(step.id); }
  }
  queue.sort(byInputIndex);

  while (queue.length > 0) {
    const id = queue.shift()!;
    for (const childId of childrenOf.get(id)!) {
      const remaining = inDegree.get(childId)! - 1;
      inDegree.set(childId, remaining);
      if (remaining === 0) {
        let maxParentRow = -1;
        for (const parentId of allDependsOn(stepById(childId))) {
          maxParentRow = Math.max(maxParentRow, row.get(parentId)!);
        }
        row.set(childId, maxParentRow + 1);
        queue.push(childId);
        queue.sort(byInputIndex);
      }
    }
  }
  return row;
}

function assignGroups(
  steps: LayoutStep[],
  childrenOf: Map<string, string[]>,
): { group: Map<string, number>; groupCount: number } {
  const roots = steps.filter((step) => allDependsOn(step).length === 0);
  const groupCount = roots.length;
  const rootGroup = new Map<string, number>();
  roots.forEach((rootStep, i) => rootGroup.set(rootStep.id, i));

  const group = new Map<string, number>();
  for (const rootStep of roots) {
    const g = rootGroup.get(rootStep.id)!;
    const stack: string[] = [rootStep.id];
    while (stack.length > 0) {
      const current = stack.pop()!;
      if (group.has(current)) continue;
      group.set(current, g);
      for (const childId of childrenOf.get(current)!) {
        if (!group.has(childId)) stack.push(childId);
      }
    }
  }
  return { group, groupCount };
}

function buildCells(
  steps: LayoutStep[],
  group: Map<string, number>,
  row: Map<string, number>,
  byInputIndex: (a: string, b: string) => number,
): { cells: Map<string, Cell>; col: Map<string, number>; cellKey: (g: number, r: number) => string } {
  const cellKey = (g: number, r: number): string => `${g}:${r}`;
  const cells = new Map<string, Cell>();
  for (const step of steps) {
    const g = group.get(step.id)!;
    const r = row.get(step.id)!;
    const key = cellKey(g, r);
    let cell = cells.get(key);
    if (!cell) { cell = { g, r, ids: [] }; cells.set(key, cell); }
    cell.ids.push(step.id);
  }
  const col = new Map<string, number>();
  for (const cell of cells.values()) {
    cell.ids.sort(byInputIndex);
    cell.ids.forEach((id, c) => col.set(id, c));
  }
  return { cells, col, cellKey };
}

function computeGroupOffsets(
  cells: Map<string, Cell>,
  groupCount: number,
): { groupWidth: number[]; groupXStart: number[] } {
  const rowWidthOf = (n: number) => n * CARD_W + (n - 1) * H_GAP;
  const groupWidth: number[] = new Array<number>(groupCount).fill(0);
  for (const cell of cells.values()) {
    const width = rowWidthOf(cell.ids.length);
    if (width > groupWidth[cell.g]!) groupWidth[cell.g] = width;
  }
  const groupXStart: number[] = new Array<number>(groupCount).fill(0);
  groupXStart[0] = GROUP_PADDING;
  for (let i = 1; i < groupCount; i++) {
    groupXStart[i] = groupXStart[i - 1]! + groupWidth[i - 1]! + DIVIDER_WIDTH + GROUP_PADDING;
  }
  return { groupWidth, groupXStart };
}

function placeNodes(
  steps: LayoutStep[],
  group: Map<string, number>,
  row: Map<string, number>,
  col: Map<string, number>,
  cells: Map<string, Cell>,
  cellKey: (g: number, r: number) => string,
  groupWidth: number[],
  groupXStart: number[],
): { nodes: DagNode[]; nodeById: Map<string, DagNode> } {
  const rowWidthOf = (n: number) => n * CARD_W + (n - 1) * H_GAP;
  const nodes: DagNode[] = steps.map((step) => {
    const g = group.get(step.id)!;
    const r = row.get(step.id)!;
    const c = col.get(step.id)!;
    const cell = cells.get(cellKey(g, r))!;
    const xOffset = (groupWidth[g]! - rowWidthOf(cell.ids.length)) / 2;
    const x = groupXStart[g]! + xOffset + c * (CARD_W + H_GAP);
    const y = r * (CARD_H + V_GAP);
    return { stepId: step.id, x, y, row: r, col: c, columnGroup: g };
  });
  const nodeById = new Map<string, DagNode>();
  for (const node of nodes) nodeById.set(node.stepId, node);
  return { nodes, nodeById };
}

function buildEdges(
  steps: LayoutStep[],
  childrenOf: Map<string, string[]>,
): DagEdge[] {
  const edges: DagEdge[] = [];
  for (const step of steps) {
    for (const parentId of step.dependsOn) {
      edges.push({
        fromId: parentId, toId: step.id,
        edgeType: classifyEdgeType(childrenOf.get(parentId)!.length, allDependsOn(step).length),
        dependencyKind: 'explicit',
      });
    }
    const explicit = new Set(step.dependsOn);
    for (const parentId of step.implicitDependsOn ?? []) {
      if (explicit.has(parentId)) continue;
      edges.push({
        fromId: parentId, toId: step.id,
        edgeType: classifyEdgeType(childrenOf.get(parentId)!.length, allDependsOn(step).length),
        dependencyKind: 'implicit',
      });
    }
  }
  return edges;
}

function buildBranchBars(steps: LayoutStep[], childrenOf: Map<string, string[]>, nodeById: Map<string, DagNode>): BranchBar[] {
  const branchBars: BranchBar[] = [];
  for (const step of steps) {
    const children = childrenOf.get(step.id)!;
    if (children.length <= 1) continue;
    const parent = nodeById.get(step.id)!;
    const centres = children.map((childId) => nodeById.get(childId)!.x + CARD_W / 2);
    const xStart = Math.min(...centres);
    const xEnd = Math.max(...centres);
    branchBars.push({ parentId: step.id, row: parent.row, xStart, xEnd, xMid: (xStart + xEnd) / 2, y: parent.y + CARD_H + BRANCH_BAR_OFFSET_Y });
  }
  return branchBars;
}

export function computeDagLayout(steps: LayoutStep[]): DagLayout {
  if (steps.length === 0) {
    return { nodes: [], edges: [], branchBars: [], groupDividers: [], canvasWidth: 0, canvasHeight: 0, groupCount: 0 };
  }

  const { byInputIndex, stepById, childrenOf } = buildGraph(steps);
  const row = assignRows(steps, stepById, childrenOf, byInputIndex);
  const { group, groupCount } = assignGroups(steps, childrenOf);
  const { cells, col, cellKey } = buildCells(steps, group, row, byInputIndex);
  const { groupWidth, groupXStart } = computeGroupOffsets(cells, groupCount);
  const { nodes, nodeById } = placeNodes(steps, group, row, col, cells, cellKey, groupWidth, groupXStart);
  const edges = buildEdges(steps, childrenOf);
  const branchBars = buildBranchBars(steps, childrenOf, nodeById);

  const groupDividers: GroupDivider[] = [];
  for (let i = 0; i < groupCount - 1; i++) {
    groupDividers.push({ x: groupXStart[i + 1]! - DIVIDER_WIDTH / 2, leftGroup: i, rightGroup: i + 1 });
  }

  let maxRight = 0;
  let maxBottom = 0;
  for (const node of nodes) {
    maxRight = Math.max(maxRight, node.x + CARD_W);
    maxBottom = Math.max(maxBottom, node.y + CARD_H);
  }

  return { nodes, edges, branchBars, groupDividers, canvasWidth: maxRight + CANVAS_PADDING, canvasHeight: maxBottom + CANVAS_PADDING, groupCount };
}
