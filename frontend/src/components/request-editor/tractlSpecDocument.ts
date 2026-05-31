/** Minimal traCtlSpec document shape for UI serialization (ADR-016 / tractl_spec.md). */

export type TraCtlSpecDocument = {
  schemaVersion: 1
  capabilities: ['protocol.http']
  metadata?: {
    name?: string
    description?: string
  }
  variables?: Record<string, string>
  workflows: TraCtlWorkflow[]
}

export type TraCtlWorkflow = {
  id: string
  name?: string
  concurrency?: number
  failurePolicy?: string
  steps: TraCtlStep[]
}

export type TraCtlStep = {
  id: string
  kind: 'request'
  dependsOn?: string[]
  request: {
    protocol: 'http'
    target: string
    operation: string
    headers?: Record<string, string>
    body?: {
      encoding?: string
      content?: unknown
    }
  }
  assertions?: TraCtlAssertion[]
  extracts?: TraCtlExtract[]
  hooks?: {
    beforeStep?: TraCtlScript
    afterStep?: TraCtlScript
  }
  timeout?: string
  retry?: {
    maxAttempts?: number
    backoff?: string
    delay?: string
  }
}

export type TraCtlAssertion = {
  id: string
  kind: string
  op?: string
  target?: string
  expected?: string | number | boolean
  severity?: 'error' | 'warning'
}

export type TraCtlExtract = {
  id: string
  source: string
  path?: string
  as?: string
  scope?: string
}

export type TraCtlScript = {
  language: 'js'
  source: string
}
