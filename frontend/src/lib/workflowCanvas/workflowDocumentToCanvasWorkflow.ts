import type {
  TraCtlSpecDocument,
  TraCtlStep,
  TraCtlWorkflow,
} from '@/components/request-editor/tractlSpecDocument'
import { stripAuthoringSystemFields } from '@/lib/workflowCanvas/stripAuthoringSystemFields'
import { getEngineDefaults } from '@/stores/engineDefaultsStore'
import type {
  FailurePolicy,
  HttpMethod,
  Workflow,
  WorkflowConfig,
  WorkflowStep,
} from '@/types/workflow'

const FALLBACK_TIMEOUT_MS = 30000
const HTTP_METHODS: HttpMethod[] = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE']

type WorkflowDocumentSource = {
  sourceName?: string
  raw?: string
  lastModified?: string
  workflowId?: string
}

type WorkflowWithConfig = TraCtlWorkflow & {
  concurrency?: number
  timeout?: string
}

function isHttpMethod(value: string): value is HttpMethod {
  return HTTP_METHODS.includes(value as HttpMethod)
}

function normalizeMethod(value: string): HttpMethod {
  const upper = value.toUpperCase()
  return isHttpMethod(upper) ? upper : 'GET'
}

function parseTimeoutMs(value: string | undefined): number {
  if (!value) return FALLBACK_TIMEOUT_MS
  const match = value.match(/^(\d+)(ms|s)?$/)
  if (!match) return FALLBACK_TIMEOUT_MS
  const amount = Number(match[1])
  return match[2] === 's' ? amount * 1000 : amount
}

function hasAuth(step: TraCtlStep): boolean {
  const headers = step.request.headers ?? {}
  return Object.keys(headers).some((key) => key.toLowerCase() === 'authorization')
}

function hasPreScript(step: TraCtlStep): boolean {
  return Boolean(step.hooks?.beforeStep?.source.trim())
}

function toWorkflowStep(step: TraCtlStep, _workflowStepIds: Set<string>): WorkflowStep {
  const dependsOn = step.dependsOn ?? []

  return {
    id: step.id,
    method: normalizeMethod(step.request.operation),
    url: step.request.target,
    dependsOn,
    hasAuth: hasAuth(step),
    assertionCount: step.assertions?.length ?? 0,
    hasPreScript: hasPreScript(step),
    extractCount: step.extracts?.length ?? 0,
    assertions: step.assertions
      ? stripAuthoringSystemFields(step.assertions)
      : undefined,
    extracts: step.extracts ? stripAuthoringSystemFields(step.extracts) : undefined,
  }
}

function workflowConfig(workflow: WorkflowWithConfig): WorkflowConfig {
  return {
    concurrency: workflow.concurrency,
    timeoutMs: parseTimeoutMs(workflow.timeout),
    failurePolicy: (workflow.failurePolicy ?? getEngineDefaults().failurePolicy) as FailurePolicy,
    hasHooks: workflow.steps.some((step) => Boolean(step.hooks)),
    hasOverlay: false,
  }
}

function workflowName(
  workflow: TraCtlWorkflow,
  document: TraCtlSpecDocument,
  source?: WorkflowDocumentSource,
): string {
  return (
    workflow.name?.trim() ||
    document.metadata?.name?.trim() ||
    source?.sourceName?.trim() ||
    workflow.id
  )
}

export function workflowDocumentToCanvasWorkflow(
  document: TraCtlSpecDocument,
  source?: WorkflowDocumentSource,
): Workflow {
  const workflow =
    document.workflows.find((entry) => entry.id === source?.workflowId) ??
    document.workflows[0]

  if (!workflow) {
    throw new Error('Document does not contain a workflow')
  }

  const configWorkflow = workflow as WorkflowWithConfig
  const workflowStepIds = new Set(workflow.steps.map((entry) => entry.id))

  const variables =
    document.variables && Object.keys(document.variables).length > 0
      ? document.variables
      : undefined

  return {
    id: workflow.id,
    name: workflowName(workflow, document, source),
    config: workflowConfig(configWorkflow),
    steps: workflow.steps.map((entry) => toWorkflowStep(entry, workflowStepIds)),
    variables,
    yaml: source?.raw,
    lastModified: source?.lastModified,
  }
}
