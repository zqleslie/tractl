import type { KeyValueRow, RequestBodyDraft } from '@/components/request-editor/types'

const CONTENT_TYPE_KEY = 'content-type'

function findContentTypeIndex(headers: KeyValueRow[]): number {
  return headers.findIndex(
    (row) => row.key.trim().toLowerCase() === CONTENT_TYPE_KEY,
  )
}

function contentTypeForBody(body: RequestBodyDraft): string | null {
  switch (body.encoding) {
    case 'JSON':
      return 'application/json'
    case 'Form data':
      return 'application/x-www-form-urlencoded'
    case 'Multipart':
      return 'multipart/form-data'
    case 'Raw':
      return body.rawContentType || 'text/plain'
    case 'Binary': {
      const name = body.binaryFile.trim().toLowerCase()
      if (name.endsWith('.json')) return 'application/json'
      if (name.endsWith('.xml')) return 'application/xml'
      if (name.endsWith('.html') || name.endsWith('.htm')) return 'text/html'
      if (name.endsWith('.png')) return 'image/png'
      if (name.endsWith('.jpg') || name.endsWith('.jpeg')) return 'image/jpeg'
      if (name.endsWith('.pdf')) return 'application/pdf'
      return 'application/octet-stream'
    }
    case 'None':
      return null
    default:
      return null
  }
}

function upsertContentTypeHeader(
  headers: KeyValueRow[],
  value: string,
): KeyValueRow[] {
  const index = findContentTypeIndex(headers)
  if (index === -1) {
    return [
      ...headers,
      {
        id: `header-content-type-${Date.now()}`,
        enabled: true,
        key: 'Content-Type',
        value,
      },
    ]
  }

  return headers.map((row, rowIndex) =>
    rowIndex === index ? { ...row, enabled: true, value } : row,
  )
}

function removeContentTypeHeader(headers: KeyValueRow[]): KeyValueRow[] {
  const index = findContentTypeIndex(headers)
  if (index === -1) return headers
  return headers.filter((_, rowIndex) => rowIndex !== index)
}

export function syncContentTypeHeader(
  headers: KeyValueRow[],
  bodyPatch: Partial<RequestBodyDraft>,
  currentBody: RequestBodyDraft,
): KeyValueRow[] {
  const nextBody: RequestBodyDraft = { ...currentBody, ...bodyPatch }
  const encodingChanged =
    bodyPatch.encoding != null && bodyPatch.encoding !== currentBody.encoding
  const rawTypeChanged =
    bodyPatch.rawContentType != null &&
    bodyPatch.rawContentType !== currentBody.rawContentType
  const binaryChanged =
    bodyPatch.binaryFile != null && bodyPatch.binaryFile !== currentBody.binaryFile

  if (
    !encodingChanged &&
    !rawTypeChanged &&
    !binaryChanged &&
    bodyPatch.encoding == null
  ) {
    return headers
  }

  const contentType = contentTypeForBody(nextBody)
  if (!contentType) return removeContentTypeHeader(headers)
  return upsertContentTypeHeader(headers, contentType)
}

