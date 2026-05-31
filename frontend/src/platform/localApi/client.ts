import type {
  ApiErrorBody,
  ApiStatusResponse,
  EngineWorkflowRunResult,
  RequestRunResult,
  RunRequestInput,
  SaveRequestFileInput,
  SaveRequestFileResponse,
  WorkflowRunDocumentInput,
} from '@/platform/localApi/types'

const DEFAULT_BASE_URL = 'http://127.0.0.1:7428'

function baseUrl(): string {
  return import.meta.env.VITE_TRACTL_API_BASE_URL ?? DEFAULT_BASE_URL
}

async function parseError(response: Response): Promise<Error> {
  let message = `Request failed (${response.status})`
  try {
    const body = (await response.json()) as ApiErrorBody
    if (body.error) {
      message = body.details ? `${body.error}: ${body.details}` : body.error
    }
  } catch {
    // ignore JSON parse errors
  }
  return new Error(message)
}

async function requestJson<T>(
  path: string,
  init?: RequestInit,
): Promise<T> {
  const response = await fetch(`${baseUrl()}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers ?? {}),
    },
  })

  if (!response.ok) {
    throw await parseError(response)
  }

  return (await response.json()) as T
}

export async function getApiStatus(): Promise<ApiStatusResponse> {
  return requestJson<ApiStatusResponse>('/api/v1/status')
}

export async function saveRequestFile(
  input: SaveRequestFileInput,
): Promise<SaveRequestFileResponse> {
  return requestJson<SaveRequestFileResponse>('/api/v1/files', {
    method: 'POST',
    body: JSON.stringify({
      id: input.id ?? undefined,
      kind: 'request',
      name: input.name,
      document: input.document,
    }),
  })
}

export async function runRequestFile(
  input: RunRequestInput,
): Promise<RequestRunResult> {
  return requestJson<RequestRunResult>('/api/v1/run', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export async function runWorkflowDocument(
  input: WorkflowRunDocumentInput,
): Promise<EngineWorkflowRunResult> {
  return requestJson<EngineWorkflowRunResult>('/api/v1/workflows/run', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export class LocalApiUnavailableError extends Error {
  constructor(message = 'traCtl Desktop API is not reachable at localhost:7428') {
    super(message)
    this.name = 'LocalApiUnavailableError'
  }
}

export async function ensureApiAvailable(): Promise<void> {
  try {
    const status = await getApiStatus()
    if (!status.ok) {
      throw new LocalApiUnavailableError()
    }
  } catch (error) {
    if (error instanceof LocalApiUnavailableError) throw error
    throw new LocalApiUnavailableError()
  }
}
