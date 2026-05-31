import type {
  StepAssertionResult,
  StepExtractResult,
  StepOutcome,
  StepResult,
} from '@/types/workflow'
import {
  isTractlWasmRunSuccess,
  type TractlWasmRunResult,
} from '@/platform/web/wasm/loadTractlWasmRuntime'

export type WorkflowRunOutcome = 'passed' | 'failed' | 'error'

export type WasmRunResult = TractlWasmRunResult

type WasmWorkflowOutcome = {
  WorkflowID?: string
  Passed?: boolean
  Skipped?: boolean
  Steps?: WasmStepOutcome[]
}

type WasmStepOutcome = {
  StepID?: string
  State?: string
  CausesFailure?: boolean
  AssertionResults?: WasmAssertionOutcome[]
  ResponseStatus?: number
  ResponseHeaders?: Record<string, string>
  ResponseBody?: string
  Error?: string
}

type WasmAssertionOutcome = {
  AssertionID?: string
  Kind?: string
  Outcome?: string
  Message?: string
  Severity?: string
}

type WasmDiagnostics = {
  Duration?: number
  Workflows?: Array<{
    WorkflowID?: string
    StartedAt?: string
    Duration?: number
    Steps?: Array<{
      StepID?: string
      StartedAt?: string
      Duration?: number
      Requests?: Array<{ URL?: string; Timeline?: { TotalMs?: number } }>
      Extracts?: Array<{ ID?: string; Source?: string; As?: string }>
    }>
  }>
}

export type CanvasRunOutcome = {
  workflowOutcome: WorkflowRunOutcome
  duration: number
  stepOutcomes: Record<string, StepResult>
  errorMessage?: string
}

function bridgeErrorMessage(result: TractlWasmRunResult): string | undefined {
  if (isTractlWasmRunSuccess(result)) return undefined
  const { code, message } = result.error
  return `${code}: ${message}`
}

function pipelineErrorMessage(result: Record<string, unknown>): string | undefined {
  const parseError = readString(result, 'ParseError')
  const validationError = readString(result, 'ValidationError')
  const planError = readString(result, 'PlanError')
  return parseError || validationError || planError || undefined
}

function readString(
  value: Record<string, unknown>,
  key: string,
): string | undefined {
  const entry = value[key]
  return typeof entry === 'string' && entry.trim().length > 0 ? entry : undefined
}

function durationMsFromDiagnostics(
  workflowId: string,
  stepId: string,
  diagnostics?: WasmDiagnostics,
): number {
  const workflow = diagnostics?.Workflows?.find((entry) => entry.WorkflowID === workflowId)
  const step = workflow?.Steps?.find((entry) => entry.StepID === stepId)
  if (!step) return 0

  if (typeof step.Duration === 'number' && step.Duration > 0) {
    return normalizeDurationMs(step.Duration)
  }

  const timeline = step.Requests?.[0]?.Timeline
  if (timeline && typeof timeline.TotalMs === 'number') {
    return timeline.TotalMs
  }

  return 0
}

function normalizeDurationMs(value: number): number {
  if (value > 1_000_000) {
    return Math.round(value / 1_000_000)
  }
  return Math.round(value)
}

function parseDiagnosticsTime(value: string | undefined): number | undefined {
  if (!value?.trim()) return undefined
  const parsed = Date.parse(value)
  return Number.isFinite(parsed) ? parsed : undefined
}

function startMsFromDiagnostics(
  workflowId: string,
  stepId: string,
  diagnostics?: WasmDiagnostics,
): number | undefined {
  const workflow = diagnostics?.Workflows?.find((entry) => entry.WorkflowID === workflowId)
  const step = workflow?.Steps?.find((entry) => entry.StepID === stepId)
  if (!workflow || !step) return undefined

  const workflowStart = parseDiagnosticsTime(workflow.StartedAt)
  const stepStart = parseDiagnosticsTime(step.StartedAt)
  if (workflowStart === undefined || stepStart === undefined) return undefined

  return Math.max(0, stepStart - workflowStart)
}

function mapStepOutcome(
  state: string | undefined,
  causesFailure: boolean | undefined,
  error: string | undefined,
): StepOutcome {
  if (error?.trim()) return 'failed'
  if (state === 'dependency-skipped' || state === 'conditional-skip') return 'skipped'
  if (state === 'failed' || causesFailure) return 'failed'
  if (state === 'succeeded') return 'passed'
  return 'skipped'
}

function mapAssertionSeverity(
  severity: string | undefined,
): StepAssertionResult['severity'] {
  if (severity === 'warning') return 'warn'
  if (severity === 'error') return 'error'
  return 'info'
}

function parseExpectedReceived(message: string): {
  expected?: string
  received?: string
} {
  const match = message.match(/expected\s+(.+?),\s+got\s+(.+)$/i)
  if (!match) return {}
  return { expected: match[1], received: match[2] }
}

function mapAssertion(assertion: WasmAssertionOutcome): StepAssertionResult {
  const description =
    assertion.Message?.trim() ||
    `${assertion.Kind ?? 'assertion'} ${assertion.AssertionID ?? ''}`.trim()

  return {
    description,
    passed: assertion.Outcome === 'pass',
    severity: mapAssertionSeverity(assertion.Severity),
    ...parseExpectedReceived(assertion.Message ?? ''),
  }
}

function requestUrlFromDiagnostics(
  workflowId: string,
  stepId: string,
  diagnostics?: WasmDiagnostics,
): string | undefined {
  const step = diagnostics?.Workflows?.find((entry) => entry.WorkflowID === workflowId)
    ?.Steps?.find((entry) => entry.StepID === stepId)
  const url = step?.Requests?.[0]?.URL?.trim()
  return url || undefined
}

function mapExtracts(
  workflowId: string,
  stepId: string,
  diagnostics?: WasmDiagnostics,
): StepExtractResult[] {
  const step = diagnostics?.Workflows?.find((entry) => entry.WorkflowID === workflowId)
    ?.Steps?.find((entry) => entry.StepID === stepId)

  return (step?.Extracts ?? []).map((extract) => {
    const variable = extract.As || extract.ID || 'extract'
    return {
      variable,
      expression: extract.Source || '',
      value: extract.Source || '',
      scope: 'workflow',
    }
  })
}

function mapStep(
  step: WasmStepOutcome,
  workflowId: string,
  diagnostics?: WasmDiagnostics,
): StepResult {
  const stepId = step.StepID ?? ''
  const assertions = (step.AssertionResults ?? []).map(mapAssertion)

  const startMs = startMsFromDiagnostics(workflowId, stepId, diagnostics)

  return {
    outcome: mapStepOutcome(step.State, step.CausesFailure, step.Error),
    statusCode: step.ResponseStatus,
    durationMs: durationMsFromDiagnostics(workflowId, stepId, diagnostics),
    startMs,
    assertions,
    extracts: mapExtracts(workflowId, stepId, diagnostics),
    responseBody: step.ResponseBody,
    responseHeaders: step.ResponseHeaders,
    requestUrl: requestUrlFromDiagnostics(workflowId, stepId, diagnostics),
  }
}

function workflowDurationMs(
  workflow: WasmWorkflowOutcome | undefined,
  diagnostics: WasmDiagnostics | undefined,
  stepResults: StepResult[],
): number {
  const diagnosticWorkflow = diagnostics?.Workflows?.find(
    (entry) => entry.WorkflowID === workflow?.WorkflowID,
  )
  if (diagnosticWorkflow && typeof diagnosticWorkflow.Duration === 'number') {
    return normalizeDurationMs(diagnosticWorkflow.Duration)
  }
  if (diagnostics && typeof diagnostics.Duration === 'number') {
    return normalizeDurationMs(diagnostics.Duration)
  }
  return stepResults.reduce((sum, step) => sum + (step.durationMs ?? 0), 0)
}

function mapWorkflowOutcome(
  workflow: WasmWorkflowOutcome | undefined,
  runPassed: boolean | undefined,
): WorkflowRunOutcome {
  if (workflow?.Skipped) return 'failed'
  if (workflow?.Passed === true) return 'passed'
  if (workflow?.Passed === false) return 'failed'
  if (runPassed === true) return 'passed'
  return 'failed'
}

export function mapRunResult(
  wasmResult: WasmRunResult,
  workflowId: string,
): CanvasRunOutcome {
  const bridgeError = bridgeErrorMessage(wasmResult)
  if (!isTractlWasmRunSuccess(wasmResult)) {
    return {
      workflowOutcome: 'error',
      duration: 0,
      stepOutcomes: {},
      errorMessage: bridgeError ?? bridgeErrorMessage(wasmResult),
    }
  }

  const run = wasmResult as Record<string, unknown>
  const pipelineError = pipelineErrorMessage(run)
  if (pipelineError) {
    return {
      workflowOutcome: 'error',
      duration: 0,
      stepOutcomes: {},
      errorMessage: pipelineError,
    }
  }

  const diagnostics = (run.diagnostics ?? run.Diagnostics) as WasmDiagnostics | undefined
  const workflows = (run.Workflows as WasmWorkflowOutcome[] | undefined) ?? []
  const workflow =
    workflows.find((entry) => entry.WorkflowID === workflowId) ?? workflows[0]

  if (!workflow) {
    return {
      workflowOutcome: 'error',
      duration: 0,
      stepOutcomes: {},
      errorMessage: 'Run produced no workflow result',
    }
  }

  const resolvedWorkflowId = workflow.WorkflowID ?? workflowId
  const stepOutcomes: Record<string, StepResult> = {}
  for (const step of workflow.Steps ?? []) {
    if (!step.StepID) continue
    stepOutcomes[step.StepID] = mapStep(step, resolvedWorkflowId, diagnostics)
  }

  const results = Object.values(stepOutcomes)

  return {
    workflowOutcome: mapWorkflowOutcome(workflow, run.Passed as boolean | undefined),
    duration: workflowDurationMs(workflow, diagnostics, results),
    stepOutcomes,
    errorMessage: results.some((step) => step.outcome === 'failed')
      ? 'One or more steps failed'
      : undefined,
  }
}
