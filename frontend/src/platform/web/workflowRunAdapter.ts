import type { StepAssertionResult, StepResult } from '@/types/workflow'
import { isTractlWasmRunSuccess, type TractlWasmRunResult } from '@/platform/web/wasm/loadTractlWasmRuntime'

export type WorkflowRunOutcome = 'passed' | 'failed' | 'error'
export type WasmRunResult = TractlWasmRunResult

export type CanvasRunOutcome = {
  workflowOutcome: WorkflowRunOutcome
  duration: number
  stepOutcomes: Record<string, StepResult>
  errorMessage?: string
}

type RunRecord = Record<string, unknown>

function str(v: unknown): string { return typeof v === 'string' ? v : '' }
function num(v: unknown): number { return typeof v === 'number' ? Math.round(v) : 0 }
function optNum(v: unknown): number | undefined { return typeof v === 'number' ? Math.round(v) : undefined }
function bool(v: unknown): boolean { return v === true }
function arr<T>(v: unknown): T[] { return Array.isArray(v) ? (v as T[]) : [] }
function obj(v: unknown): RunRecord { return (v && typeof v === 'object' && !Array.isArray(v)) ? v as RunRecord : {} }

function stepOutcome(step: RunRecord): StepResult['outcome'] {
  const state = str(step['State'])
  if (str(step['Error'])) return 'failed'
  if (state === 'dependency-skipped' || state === 'conditional-skip') return 'skipped'
  if (state === 'failed' || bool(step['CausesFailure'])) return 'failed'
  if (state === 'succeeded') return 'passed'
  return 'skipped'
}

function stepDuration(workflowId: string, stepId: string, diagnostics: RunRecord): number {
  const wf = arr<RunRecord>(obj(diagnostics)['Workflows']).find((w) => str(w['WorkflowID']) === workflowId)
  const st = arr<RunRecord>(wf?.['Steps']).find((s) => str(s['StepID']) === stepId)
  if (!st) return 0
  if (num(st['Duration']) > 0) return num(st['Duration'])
  const req = arr<RunRecord>(st['Requests'])[0]
  return num(obj(req?.['Timeline'])['TotalMs'])
}

function mapStep(step: RunRecord, workflowId: string, diagnostics: RunRecord): StepResult {
  const assertions: StepAssertionResult[] = arr<RunRecord>(step['AssertionResults']).map((a) => ({
    description: str(a['Message']) || `${str(a['Kind'])} ${str(a['AssertionID'])}`.trim(),
    passed: str(a['Outcome']) === 'pass',
    severity: str(a['Severity']) === 'warning' ? 'warn' : 'error',
    // Go's assertion.AssertionResult has no Expected/Received fields on this path;
    // the full message is in `description` above.
  }))
  const stepId = str(step['StepID'])
  const req = arr<RunRecord>(
    arr<RunRecord>(
      arr<RunRecord>(obj(diagnostics)['Workflows']).find((w) => str(w['WorkflowID']) === workflowId)?.['Steps']
    ).find((s) => str(s['StepID']) === stepId)?.['Requests']
  )[0]
  return {
    outcome: stepOutcome(step),
    statusCode: optNum(step['ResponseStatus']),
    durationMs: stepDuration(workflowId, stepId, diagnostics),
    assertions,
    extracts: [],
    responseBody: str(step['ResponseBody']) || undefined,
    responseHeaders: obj(step['ResponseHeaders']) as Record<string, string>,
    requestUrl: str(obj(req)['URL']) || undefined,
  }
}

export function mapRunResult(wasmResult: WasmRunResult, workflowId: string): CanvasRunOutcome {
  if (!isTractlWasmRunSuccess(wasmResult)) {
    const { code, message } = (wasmResult as { error: { code: string; message: string } }).error
    return { workflowOutcome: 'error', duration: 0, stepOutcomes: {}, errorMessage: `${code}: ${message}` }
  }

  const run = wasmResult as RunRecord
  const parseErr = str(run['ParseError']) || str(run['ValidationError']) || str(run['PlanError'])
  if (parseErr) return { workflowOutcome: 'error', duration: 0, stepOutcomes: {}, errorMessage: parseErr }

  const diagnostics = obj(run['diagnostics'] ?? run['Diagnostics'])
  const workflows = arr<RunRecord>(run['Workflows'])
  const wf = workflows.find((w) => str(w['WorkflowID']) === workflowId) ?? workflows[0]
  if (!wf) return { workflowOutcome: 'error', duration: 0, stepOutcomes: {}, errorMessage: 'Run produced no workflow result' }

  const resolvedId = str(wf['WorkflowID']) || workflowId
  const stepOutcomes: Record<string, StepResult> = {}
  for (const step of arr<RunRecord>(wf['Steps'])) {
    const id = str(step['StepID'])
    if (id) stepOutcomes[id] = mapStep(step, resolvedId, diagnostics)
  }

  const wfDiag = arr<RunRecord>(diagnostics['Workflows']).find((w) => str(w['WorkflowID']) === resolvedId)
  const duration = num(wfDiag?.['Duration']) || num(diagnostics['Duration']) ||
    Object.values(stepOutcomes).reduce((s, r) => s + (r.durationMs ?? 0), 0)

  const passed = bool(wf['Passed'])
  return {
    workflowOutcome: bool(wf['Skipped']) ? 'failed' : passed ? 'passed' : 'failed',
    duration,
    stepOutcomes,
    errorMessage: !passed ? Object.values(stepOutcomes).some((s) => s.outcome === 'failed') ? 'One or more steps failed' : undefined : undefined,
  }
}
