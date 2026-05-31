import type { HttpMethod } from '@/components/primitives'
import type { TraCtlSpecDocument } from '@/components/request-editor/tractlSpecDocument'

export type DocumentRunSummary = {
  requestName: string
  method: HttpMethod
  url: string
  sourceType: 'request' | 'workflow'
}

const HTTP_METHODS = new Set<HttpMethod>([
  'GET',
  'POST',
  'PUT',
  'PATCH',
  'DELETE',
])

function asHttpMethod(operation: string | undefined): HttpMethod {
  const upper = (operation ?? 'GET').toUpperCase()
  if (HTTP_METHODS.has(upper as HttpMethod)) {
    return upper as HttpMethod
  }
  return 'GET'
}

export function extractRunSummaryFromDocument(
  spec: TraCtlSpecDocument,
  fallbackName: string,
): DocumentRunSummary | null {
  const workflow = spec.workflows?.[0]
  if (!workflow?.steps?.length) return null

  const requestStep = workflow.steps.find(
    (step) => step.kind === 'request' && step.request?.target,
  )
  if (!requestStep?.request) return null

  const stepCount = workflow.steps.length
  const requestName =
    spec.metadata?.name?.trim() ||
    workflow.name?.trim() ||
    fallbackName

  return {
    requestName,
    method: asHttpMethod(requestStep.request.operation),
    url: requestStep.request.target,
    sourceType: stepCount > 1 ? 'workflow' : 'request',
  }
}
