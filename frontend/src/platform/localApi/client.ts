import type {
  ApiErrorBody,
  ApiStatusResponse,
  EngineWorkflowRunResult,
  FileWriteResponse,
  RequestRunResult,
  RunRequestInput,
  SaveRequestFileInput,
  SaveRequestFileResponse,
  WorkflowRunDocumentInput,
} from '@/platform/localApi/types'
import { type GoRunResult, mapFlatRunResult } from '@/lib/execution/mapRunResults'
import { useSettingsStore } from '@/stores/settingsStore'

const DEFAULT_BASE_URL = 'http://127.0.0.1:7428'

function baseUrl(): string {
  return (
    useSettingsStore.getState().serverUrl ??
    import.meta.env.VITE_TRACTL_API_BASE_URL ??
    DEFAULT_BASE_URL
  )
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
  const path = input.path ?? `requests/${input.request.id}.yaml`
  const content = `${JSON.stringify(input.request, null, 2)}\n`
  const saved = await requestJson<FileWriteResponse>(
    `/api/v1/files/${encodeFilePath(path)}`,
    {
      method: 'POST',
      body: JSON.stringify({ content }),
    },
  )
  return {
    path: saved.path,
    updatedAt: saved.modifiedAt,
  }
}

export async function writeWorkspaceFile(
  path: string,
  content: string,
): Promise<FileWriteResponse> {
  return requestJson<FileWriteResponse>(`/api/v1/files/${encodeFilePath(path)}`, {
    method: 'POST',
    body: JSON.stringify({ content }),
  })
}

export async function runRequestFile(
  input: RunRequestInput,
): Promise<RequestRunResult> {
  const raw = await requestJson<GoRunResult>('/api/v1/run', {
    method: 'POST',
    body: JSON.stringify(input.request),
  })
  return mapFlatRunResult(raw)
}

export async function runWorkflowDocument(
  input: WorkflowRunDocumentInput,
): Promise<EngineWorkflowRunResult> {
  return requestJson<EngineWorkflowRunResult>('/api/v1/workflows/run', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

function encodeFilePath(path: string): string {
  return path
    .split('/')
    .map((part) => encodeURIComponent(part))
    .join('/')
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
