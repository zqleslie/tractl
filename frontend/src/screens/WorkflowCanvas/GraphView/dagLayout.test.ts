import { describe, it, expect } from 'vitest';
import { computeDagLayout } from './dagLayout';
import type { LayoutStep } from './dagLayout';

const chain2: LayoutStep[] = [
  { id: 'A', dependsOn: [] },
  { id: 'B', dependsOn: ['A'] },
];
const chain3: LayoutStep[] = [
  { id: 'A', dependsOn: [] },
  { id: 'B', dependsOn: ['A'] },
  { id: 'C', dependsOn: ['B'] },
];
const fanout: LayoutStep[] = [
  { id: 'A', dependsOn: [] },
  { id: 'B', dependsOn: ['A'] },
  { id: 'C', dependsOn: ['A'] },
];
const fanoutMerge: LayoutStep[] = [
  { id: 'A', dependsOn: [] },
  { id: 'B', dependsOn: ['A'] },
  { id: 'C', dependsOn: ['A'] },
  { id: 'D', dependsOn: ['B', 'C'] },
];
const twoRoots: LayoutStep[] = [
  { id: 'X', dependsOn: [] },
  { id: 'Y', dependsOn: [] },
  { id: 'Z', dependsOn: ['Y'] },
];
const single: LayoutStep[] = [
  { id: 'A', dependsOn: [] },
];

describe('empty input', () => {
  it('returns empty layout with zero canvas dimensions', () => {
    const layout = computeDagLayout([]);
    expect(layout.nodes).toHaveLength(0);
    expect(layout.edges).toHaveLength(0);
    expect(layout.canvasWidth).toBe(0);
    expect(layout.canvasHeight).toBe(0);
    expect(layout.groupCount).toBe(0);
  });
});

describe('single step', () => {
  it('places the node at row 0, col 0, group 0', () => {
    const layout = computeDagLayout(single);
    expect(layout.nodes).toHaveLength(1);
    expect(layout.nodes[0]!.row).toBe(0);
    expect(layout.nodes[0]!.col).toBe(0);
    expect(layout.nodes[0]!.columnGroup).toBe(0);
  });
  it('produces no edges', () => {
    const layout = computeDagLayout(single);
    expect(layout.edges).toHaveLength(0);
  });
  it('produces no branch bars', () => {
    const layout = computeDagLayout(single);
    expect(layout.branchBars).toHaveLength(0);
  });
  it('produces no group dividers', () => {
    const layout = computeDagLayout(single);
    expect(layout.groupDividers).toHaveLength(0);
  });
  it('canvas dimensions are positive', () => {
    const layout = computeDagLayout(single);
    expect(layout.canvasWidth).toBeGreaterThan(0);
    expect(layout.canvasHeight).toBeGreaterThan(0);
  });
});

describe('sequential chain (A → B)', () => {
  it('assigns A to row 0 and B to row 1', () => {
    const layout = computeDagLayout(chain2);
    const a = layout.nodes.find(n => n.stepId === 'A')!;
    const b = layout.nodes.find(n => n.stepId === 'B')!;
    expect(a.row).toBe(0);
    expect(b.row).toBe(1);
  });
  it('A and B are in the same column group', () => {
    const layout = computeDagLayout(chain2);
    const a = layout.nodes.find(n => n.stepId === 'A')!;
    const b = layout.nodes.find(n => n.stepId === 'B')!;
    expect(a.columnGroup).toBe(b.columnGroup);
  });
  it('produces exactly one edge with type sequential', () => {
    const layout = computeDagLayout(chain2);
    expect(layout.edges).toHaveLength(1);
    expect(layout.edges[0]!.edgeType).toBe('sequential');
    expect(layout.edges[0]!.fromId).toBe('A');
    expect(layout.edges[0]!.toId).toBe('B');
  });
  it('A.y is less than B.y (top-to-bottom layout)', () => {
    const layout = computeDagLayout(chain2);
    const a = layout.nodes.find(n => n.stepId === 'A')!;
    const b = layout.nodes.find(n => n.stepId === 'B')!;
    expect(a.y).toBeLessThan(b.y);
  });
});

describe('sequential chain (A → B → C)', () => {
  it('assigns rows 0, 1, 2 in order', () => {
    const layout = computeDagLayout(chain3);
    const a = layout.nodes.find(n => n.stepId === 'A')!;
    const b = layout.nodes.find(n => n.stepId === 'B')!;
    const c = layout.nodes.find(n => n.stepId === 'C')!;
    expect(a.row).toBe(0);
    expect(b.row).toBe(1);
    expect(c.row).toBe(2);
  });
  it('produces 2 sequential edges', () => {
    const layout = computeDagLayout(chain3);
    expect(layout.edges).toHaveLength(2);
    expect(layout.edges.every(e => e.edgeType === 'sequential')).toBe(true);
  });
});

describe('fan-out (A → B, A → C)', () => {
  it('A is at row 0; B and C are at row 1', () => {
    const layout = computeDagLayout(fanout);
    const a = layout.nodes.find(n => n.stepId === 'A')!;
    const b = layout.nodes.find(n => n.stepId === 'B')!;
    const c = layout.nodes.find(n => n.stepId === 'C')!;
    expect(a.row).toBe(0);
    expect(b.row).toBe(1);
    expect(c.row).toBe(1);
  });
  it('B and C are in different columns at the same row', () => {
    const layout = computeDagLayout(fanout);
    const b = layout.nodes.find(n => n.stepId === 'B')!;
    const c = layout.nodes.find(n => n.stepId === 'C')!;
    expect(b.col).not.toBe(c.col);
  });
  it('produces 2 fanout edges', () => {
    const layout = computeDagLayout(fanout);
    const edgeAB = layout.edges.find(e => e.fromId === 'A' && e.toId === 'B')!;
    const edgeAC = layout.edges.find(e => e.fromId === 'A' && e.toId === 'C')!;
    expect(edgeAB.edgeType).toBe('fanout');
    expect(edgeAC.edgeType).toBe('fanout');
  });
  it('produces exactly one branch bar for A', () => {
    const layout = computeDagLayout(fanout);
    expect(layout.branchBars).toHaveLength(1);
    expect(layout.branchBars[0]!.parentId).toBe('A');
  });
  it('branch bar xStart ≤ xEnd', () => {
    const layout = computeDagLayout(fanout);
    const bar = layout.branchBars[0]!;
    expect(bar.xStart).toBeLessThanOrEqual(bar.xEnd);
  });
  it('branch bar xMid is midpoint of xStart and xEnd', () => {
    const layout = computeDagLayout(fanout);
    const bar = layout.branchBars[0]!;
    expect(bar.xMid).toBeCloseTo((bar.xStart + bar.xEnd) / 2, 1);
  });
  it('B and C have different x positions (not stacked)', () => {
    const layout = computeDagLayout(fanout);
    const b = layout.nodes.find(n => n.stepId === 'B')!;
    const c = layout.nodes.find(n => n.stepId === 'C')!;
    expect(b.x).not.toBe(c.x);
  });
  it('produces no group dividers (single root)', () => {
    const layout = computeDagLayout(fanout);
    expect(layout.groupDividers).toHaveLength(0);
  });
});

describe('fan-out + merge (A → B, A → C, B+C → D)', () => {
  it('assigns A=row0, B=row1, C=row1, D=row2', () => {
    const layout = computeDagLayout(fanoutMerge);
    const a = layout.nodes.find(n => n.stepId === 'A')!;
    const b = layout.nodes.find(n => n.stepId === 'B')!;
    const c = layout.nodes.find(n => n.stepId === 'C')!;
    const d = layout.nodes.find(n => n.stepId === 'D')!;
    expect(a.row).toBe(0);
    expect(b.row).toBe(1);
    expect(c.row).toBe(1);
    expect(d.row).toBe(2);
  });
  it('D is in the same column group as A, B, C', () => {
    const layout = computeDagLayout(fanoutMerge);
    const groups = layout.nodes.map(n => n.columnGroup);
    expect(new Set(groups).size).toBe(1);
  });
  it('produces 4 edges total', () => {
    const layout = computeDagLayout(fanoutMerge);
    expect(layout.edges).toHaveLength(4);
  });
  it('A→B and A→C are fanout edges', () => {
    const layout = computeDagLayout(fanoutMerge);
    const ab = layout.edges.find(e => e.fromId === 'A' && e.toId === 'B')!;
    const ac = layout.edges.find(e => e.fromId === 'A' && e.toId === 'C')!;
    expect(ab.edgeType).toBe('fanout');
    expect(ac.edgeType).toBe('fanout');
  });
  it('B→D and C→D are merge edges', () => {
    const layout = computeDagLayout(fanoutMerge);
    const bd = layout.edges.find(e => e.fromId === 'B' && e.toId === 'D')!;
    const cd = layout.edges.find(e => e.fromId === 'C' && e.toId === 'D')!;
    expect(bd.edgeType).toBe('merge');
    expect(cd.edgeType).toBe('merge');
  });
  it('one branch bar for A at the fan-out', () => {
    const layout = computeDagLayout(fanoutMerge);
    expect(layout.branchBars).toHaveLength(1);
    expect(layout.branchBars[0]!.parentId).toBe('A');
  });
  it('D.y is greater than B.y and C.y', () => {
    const layout = computeDagLayout(fanoutMerge);
    const b = layout.nodes.find(n => n.stepId === 'B')!;
    const c = layout.nodes.find(n => n.stepId === 'C')!;
    const d = layout.nodes.find(n => n.stepId === 'D')!;
    expect(d.y).toBeGreaterThan(b.y);
    expect(d.y).toBeGreaterThan(c.y);
  });
  it('no group dividers (all steps in one root tree)', () => {
    const layout = computeDagLayout(fanoutMerge);
    expect(layout.groupDividers).toHaveLength(0);
  });
  it('canvas dimensions are positive and larger than single step', () => {
    const singleLayout = computeDagLayout(single);
    const layout = computeDagLayout(fanoutMerge);
    expect(layout.canvasWidth).toBeGreaterThanOrEqual(singleLayout.canvasWidth);
    expect(layout.canvasHeight).toBeGreaterThan(singleLayout.canvasHeight);
  });
});

describe('two independent roots (X, Y→Z)', () => {
  it('X and Y are in different column groups', () => {
    const layout = computeDagLayout(twoRoots);
    const x = layout.nodes.find(n => n.stepId === 'X')!;
    const y = layout.nodes.find(n => n.stepId === 'Y')!;
    expect(x.columnGroup).not.toBe(y.columnGroup);
  });
  it('Z is in the same group as Y (its root)', () => {
    const layout = computeDagLayout(twoRoots);
    const y = layout.nodes.find(n => n.stepId === 'Y')!;
    const z = layout.nodes.find(n => n.stepId === 'Z')!;
    expect(z.columnGroup).toBe(y.columnGroup);
  });
  it('produces exactly one group divider', () => {
    const layout = computeDagLayout(twoRoots);
    expect(layout.groupDividers).toHaveLength(1);
    expect(layout.groupDividers[0]!.leftGroup).toBe(0);
    expect(layout.groupDividers[0]!.rightGroup).toBe(1);
  });
  it('groupCount is 2', () => {
    const layout = computeDagLayout(twoRoots);
    expect(layout.groupCount).toBe(2);
  });
  it('X and Y have different x positions (side by side, not stacked)', () => {
    const layout = computeDagLayout(twoRoots);
    const x = layout.nodes.find(n => n.stepId === 'X')!;
    const y = layout.nodes.find(n => n.stepId === 'Y')!;
    expect(x.x).not.toBe(y.x);
  });
  it('Y and Z are in the same x column (Z is below Y)', () => {
    const layout = computeDagLayout(twoRoots);
    const y = layout.nodes.find(n => n.stepId === 'Y')!;
    const z = layout.nodes.find(n => n.stepId === 'Z')!;
    expect(y.x).toBe(z.x);
    expect(z.y).toBeGreaterThan(y.y);
  });
  it('the divider x is between X group and Y group', () => {
    const layout = computeDagLayout(twoRoots);
    const x = layout.nodes.find(n => n.stepId === 'X')!;
    const y = layout.nodes.find(n => n.stepId === 'Y')!;
    const divider = layout.groupDividers[0]!;
    expect(divider.x).toBeGreaterThan(x.x + 260);
    expect(divider.x).toBeLessThan(y.x);
  });
});

describe('error cases', () => {
  it('throws on cycle (A → B → A)', () => {
    const cyclic: LayoutStep[] = [
      { id: 'A', dependsOn: ['B'] },
      { id: 'B', dependsOn: ['A'] },
    ];
    expect(() => computeDagLayout(cyclic)).toThrow(/cycle/i);
  });
  it('throws on unknown dependency', () => {
    const unknown: LayoutStep[] = [
      { id: 'A', dependsOn: ['Z'] },
    ];
    expect(() => computeDagLayout(unknown)).toThrow(/unknown/i);
  });
  it('throws on self-referential step (A → A)', () => {
    const selfRef: LayoutStep[] = [
      { id: 'A', dependsOn: ['A'] },
    ];
    expect(() => computeDagLayout(selfRef)).toThrow(/cycle/i);
  });
});

describe('implicit dependencies', () => {
  const diamondImplicit: LayoutStep[] = [
    { id: 'step-1', dependsOn: [] },
    { id: 'step-2', dependsOn: ['step-1'] },
    { id: 'step-3', dependsOn: ['step-1'] },
    { id: 'step-4', dependsOn: ['step-3'], implicitDependsOn: ['step-2'] },
  ];

  it('adds a dashed implicit edge without duplicating explicit edges', () => {
    const layout = computeDagLayout(diamondImplicit);
    const explicit = layout.edges.find(
      (edge) => edge.fromId === 'step-3' && edge.toId === 'step-4',
    );
    const implicit = layout.edges.find(
      (edge) => edge.fromId === 'step-2' && edge.toId === 'step-4',
    );
    expect(explicit?.dependencyKind).toBe('explicit');
    expect(implicit?.dependencyKind).toBe('implicit');
    expect(
      layout.edges.filter((edge) => edge.fromId === 'step-2' && edge.toId === 'step-4'),
    ).toHaveLength(1);
  });

  it('uses implicit deps for row assignment', () => {
    const layout = computeDagLayout(diamondImplicit);
    const step4 = layout.nodes.find((node) => node.stepId === 'step-4')!;
    expect(step4.row).toBe(2);
  });
});

describe('determinism', () => {
  it('same input always produces same output', () => {
    const a = computeDagLayout(fanoutMerge);
    const b = computeDagLayout(fanoutMerge);
    expect(JSON.stringify(a)).toBe(JSON.stringify(b));
  });
  it('input order determines left-right ordering at same row', () => {
    const layout = computeDagLayout(fanoutMerge);
    const b = layout.nodes.find(n => n.stepId === 'B')!;
    const c = layout.nodes.find(n => n.stepId === 'C')!;
    expect(b.col).toBeLessThan(c.col);
  });
});
