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
export function computeDagLayout(steps: LayoutStep[]): DagLayout {
  if (steps.length === 0) {
    return {
      nodes: [],
      edges: [],
      branchBars: [],
      groupDividers: [],
      canvasWidth: 0,
      canvasHeight: 0,
      groupCount: 0,
    };
  }

  // --- Step 1: validate input ---------------------------------------------
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
  const byInputIndex = (a: string, b: string): number =>
    indexById.get(a)! - indexById.get(b)!;

  // Children of each step, recorded in input order for determinism.
  const childrenOf = new Map<string, string[]>();
  for (const step of steps) childrenOf.set(step.id, []);
  for (const step of steps) {
    for (const depId of allDependsOn(step)) {
      childrenOf.get(depId)!.push(step.id);
    }
  }

  // Cycle detection via DFS over the "depends on" direction. A self-reference
  // (A -> A) lands the node in the visiting set during its own traversal.
  const visited = new Set<string>();
  const visiting = new Set<string>();
  const detectCycle = (id: string): void => {
    if (visited.has(id)) return;
    if (visiting.has(id)) {
      throw new Error(`Cycle detected involving step: ${id}`);
    }
    visiting.add(id);
    for (const depId of allDependsOn(stepById(id))) {
      detectCycle(depId);
    }
    visiting.delete(id);
    visited.add(id);
  };
  for (const step of steps) detectCycle(step.id);

  // --- Step 2: topological sort (Kahn) -> row assignment ------------------
  const inDegree = new Map<string, number>();
  for (const step of steps) inDegree.set(step.id, allDependsOn(step).length);

  const row = new Map<string, number>();
  const queue: string[] = [];
  for (const step of steps) {
    if (allDependsOn(step).length === 0) {
      row.set(step.id, 0);
      queue.push(step.id);
    }
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

  // --- Step 3: column group assignment (independent root detection) -------
  const roots = steps.filter((step) => allDependsOn(step).length === 0);
  const groupCount = roots.length;
  const rootGroup = new Map<string, number>();
  roots.forEach((rootStep, i) => rootGroup.set(rootStep.id, i));

  // Roots are processed in ascending group order, so the lowest (leftmost)
  // group claims any node reachable from multiple roots.
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

  // --- Step 4: column index within (group, row) ---------------------------
  interface Cell {
    g: number;
    r: number;
    ids: string[];
  }
  const cells = new Map<string, Cell>();
  const cellKey = (g: number, r: number): string => `${g}:${r}`;
  for (const step of steps) {
    const g = group.get(step.id)!;
    const r = row.get(step.id)!;
    const key = cellKey(g, r);
    let cell = cells.get(key);
    if (!cell) {
      cell = { g, r, ids: [] };
      cells.set(key, cell);
    }
    cell.ids.push(step.id);
  }

  const col = new Map<string, number>();
  for (const cell of cells.values()) {
    cell.ids.sort(byInputIndex);
    cell.ids.forEach((id, c) => col.set(id, c));
  }

  const rowWidthOf = (n: number): number => n * CARD_W + (n - 1) * H_GAP;

  // --- Step 5: group x offsets --------------------------------------------
  const groupWidth: number[] = new Array<number>(groupCount).fill(0);
  for (const cell of cells.values()) {
    const width = rowWidthOf(cell.ids.length);
    if (width > groupWidth[cell.g]!) groupWidth[cell.g] = width;
  }

  const groupXStart: number[] = new Array<number>(groupCount).fill(0);
  groupXStart[0] = GROUP_PADDING;
  for (let i = 1; i < groupCount; i++) {
    groupXStart[i] =
      groupXStart[i - 1]! + groupWidth[i - 1]! + DIVIDER_WIDTH + GROUP_PADDING;
  }

  // --- Step 6: pixel positions --------------------------------------------
  const nodes: DagNode[] = steps.map((step) => {
    const g = group.get(step.id)!;
    const r = row.get(step.id)!;
    const c = col.get(step.id)!;
    const cell = cells.get(cellKey(g, r))!;
    const rowWidth = rowWidthOf(cell.ids.length);
    const xOffset = (groupWidth[g]! - rowWidth) / 2;
    const x = groupXStart[g]! + xOffset + c * (CARD_W + H_GAP);
    const y = r * (CARD_H + V_GAP);
    return { stepId: step.id, x, y, row: r, col: c, columnGroup: g };
  });
  const nodeById = new Map<string, DagNode>();
  for (const node of nodes) nodeById.set(node.stepId, node);

  // --- Step 7: classify edges ---------------------------------------------
  const edges: DagEdge[] = [];
  const pushEdge = (
    parentId: string,
    childStep: LayoutStep,
    dependencyKind: DependencyKind,
  ) => {
    const stepParentCount = allDependsOn(childStep).length;
    const parentChildCount = childrenOf.get(parentId)!.length;
    edges.push({
      fromId: parentId,
      toId: childStep.id,
      edgeType: classifyEdgeType(parentChildCount, stepParentCount),
      dependencyKind,
    });
  };

  for (const step of steps) {
    for (const parentId of step.dependsOn) {
      pushEdge(parentId, step, 'explicit');
    }
    const explicit = new Set(step.dependsOn);
    for (const parentId of step.implicitDependsOn ?? []) {
      if (explicit.has(parentId)) continue;
      pushEdge(parentId, step, 'implicit');
    }
  }

  // --- Step 8: branch bars (one per parent with >1 child) -----------------
  const branchBars: BranchBar[] = [];
  for (const step of steps) {
    const children = childrenOf.get(step.id)!;
    if (children.length > 1) {
      const parent = nodeById.get(step.id)!;
      const centres = children.map((childId) => nodeById.get(childId)!.x + CARD_W / 2);
      const xStart = Math.min(...centres);
      const xEnd = Math.max(...centres);
      branchBars.push({
        parentId: step.id,
        row: parent.row,
        xStart,
        xEnd,
        xMid: (xStart + xEnd) / 2,
        y: parent.y + CARD_H + BRANCH_BAR_OFFSET_Y,
      });
    }
  }

  // --- Step 9: group dividers ---------------------------------------------
  const groupDividers: GroupDivider[] = [];
  for (let i = 0; i < groupCount - 1; i++) {
    groupDividers.push({
      x: groupXStart[i + 1]! - DIVIDER_WIDTH / 2,
      leftGroup: i,
      rightGroup: i + 1,
    });
  }

  // --- Step 10: canvas dimensions -----------------------------------------
  let maxRight = 0;
  let maxBottom = 0;
  for (const node of nodes) {
    maxRight = Math.max(maxRight, node.x + CARD_W);
    maxBottom = Math.max(maxBottom, node.y + CARD_H);
  }

  return {
    nodes,
    edges,
    branchBars,
    groupDividers,
    canvasWidth: maxRight + CANVAS_PADDING,
    canvasHeight: maxBottom + CANVAS_PADDING,
    groupCount,
  };
}
