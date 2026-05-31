export type RequestRoute = {
  id: string
  isNew: boolean
}

export function parseRequestRoute(pathname: string): RequestRoute | null {
  const match = pathname.match(/^\/request\/([^/]+)\/?$/)
  if (!match?.[1]) return null
  const id = decodeURIComponent(match[1])
  return { id, isNew: id === 'new' }
}

export function requestPathForId(id: string): string {
  return `/request/${encodeURIComponent(id)}`
}
