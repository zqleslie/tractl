import type { RequestFormState } from '@/components/request-editor/types'

/** Blank request editor state — no prefilled rows or sample URL. */
export function createEmptyRequestFormState(): RequestFormState {
  return {
    params: [],
    headers: [],
    body: {
      encoding: 'JSON',
      value: '',
      formRows: [],
      rawContentType: 'text/plain',
      binaryFile: '',
    },
    auth: {
      type: 'None',
      tokenRef: '',
      token: '',
      username: '',
      password: '',
      keyName: '',
      keyValue: '',
      placement: 'header',
    },
    scripts: {
      pre: '',
      post: '',
    },
    assertions: [],
    extracts: [],
    settings: {
      timeoutValue: '30',
      timeoutUnit: 'seconds',
      retry: 'None',
      retryConfig: null,
      failurePolicy: 'resilient',
    },
  }
}
