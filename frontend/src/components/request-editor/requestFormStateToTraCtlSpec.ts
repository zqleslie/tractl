import type { RequestFormState } from '@/components/request-editor/types'
import type { TraCtlSpecDocument } from '@/components/request-editor/tractlSpecDocument'
import { normalizeRequestTargetUrl } from '@/components/request-editor/normalizeRequestUrl'
import type { HttpMethod } from '@/components/primitives'
import { draftToRequestState } from '@/lib/requestEditor/workspaceMapping'
import { requestDefToDocumentObject } from '@/platform/mappers/requestDefToDocument'

export function requestFormStateToTraCtlSpec(
  method: HttpMethod,
  url: string,
  draft: RequestFormState,
  requestName = 'Untitled request',
): TraCtlSpecDocument {
  const target = normalizeRequestTargetUrl(url)
  const request = draftToRequestState('request-step', requestName, method, target, draft)
  return requestDefToDocumentObject(request)
}
