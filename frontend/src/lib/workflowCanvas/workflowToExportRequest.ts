import type { WorkflowExportRequest, AssertionDef, ExtractDef } from '@/platform/types'
import type { Workflow } from '@/types/workflow'
import type { TraCtlAssertion, TraCtlExtract } from '@/components/request-editor/tractlSpecDocument'

function toAssertionDef(a: TraCtlAssertion): AssertionDef {
  return {
    id: a.id,
    kind: String(a.kind),
    op: String(a.op ?? 'equals'),
    expected: String(a.expected ?? ''),
    severity: String(a.severity ?? 'error'),
  }
}

function toExtractDef(e: TraCtlExtract): ExtractDef {
  return {
    id: e.id,
    source: String(e.source),
    path: String(e.path ?? ''),
    variableName: String(e.as ?? e.id),
    scope: String(e.scope ?? 'workflow'),
  }
}

export function workflowToExportRequest(workflow: Workflow): WorkflowExportRequest {
  return {
    id: workflow.id,
    name: workflow.name,
    variables: workflow.variables,
    concurrency: workflow.config.concurrency,
    failurePolicy: workflow.config.failurePolicy,
    steps: workflow.steps.map((step) => ({
      id: step.id,
      dependsOn: step.dependsOn.length > 0 ? step.dependsOn : undefined,
      url: step.url,
      method: step.method,
      assertions: step.assertions?.map(toAssertionDef),
      extracts: step.extracts?.map(toExtractDef),
    })),
  }
}
