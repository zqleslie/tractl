import { stripAuthoringSystemFields } from '@/lib/workflowCanvas/stripAuthoringSystemFields'
import type { Workflow, WorkflowStep } from '@/types/workflow'

type SerializableValue =
  | string
  | number
  | boolean
  | SerializableValue[]
  | { [key: string]: SerializableValue }

type TraCtlWorkflowDocument = {
  schemaVersion: 1
  capabilities: ['protocol.http']
  metadata?: {
    name?: string
  }
  variables?: Record<string, string>
  workflows: TraCtlWorkflow[]
}

type TraCtlWorkflow = {
  id: string
  name?: string
  concurrency?: number
  failurePolicy?: 'resilient' | 'failFast'
  steps: TraCtlStep[]
}

type TraCtlStep = {
  id: string
  kind: 'request'
  dependsOn?: string[]
  request: {
    protocol: 'http'
    target: string
    operation: string
  }
  assertions?: {
    id: string
    kind: string
    op?: string
    expected?: string | number | boolean
    severity?: string
  }[]
  extracts?: {
    id: string
    source: string
    path?: string
    as?: string
  }[]
}

function compactObject<T extends Record<string, unknown>>(value: T): Partial<T> {
  return Object.fromEntries(
    Object.entries(value).filter(([, entry]) => {
      if (entry === null || entry === undefined) return false
      if (Array.isArray(entry)) return entry.length > 0
      if (typeof entry === 'string') return entry.trim().length > 0
      return true
    }),
  ) as Partial<T>
}

function serializeStep(step: WorkflowStep): TraCtlStep {
  return compactObject({
    id: step.id,
    kind: 'request' as const,
    dependsOn: step.dependsOn,
    request: {
      protocol: 'http' as const,
      target: step.url,
      operation: step.method.toUpperCase(),
    },
    assertions: step.assertions,
    extracts: step.extracts,
  }) as TraCtlStep
}

function buildDocument(workflow: Workflow): TraCtlWorkflowDocument {
  return {
    schemaVersion: 1,
    capabilities: ['protocol.http'],
    metadata: compactObject({
      name: workflow.name,
    }),
    variables: workflow.variables,
    workflows: [
      compactObject({
        id: workflow.id,
        name: workflow.name,
        concurrency: workflow.config.concurrency,
        failurePolicy: workflow.config.failurePolicy,
        steps: workflow.steps.map(serializeStep),
      }) as TraCtlWorkflow,
    ],
  }
}

// TODO: install js-yaml and replace this scoped serializer with yaml.dump().
function toYaml(value: SerializableValue, indent = 0): string {
  const pad = ' '.repeat(indent)

  if (Array.isArray(value)) {
    return value
      .map((entry) => {
        if (typeof entry === 'object' && entry !== null && !Array.isArray(entry)) {
          const serialized = toYaml(entry, indent + 2)
          return `${pad}- ${serialized.trimStart()}`
        }
        return `${pad}- ${formatScalar(entry)}`
      })
      .join('\n')
  }

  if (typeof value === 'object' && value !== null) {
    return Object.entries(value)
      .map(([key, entry]) => {
        if (Array.isArray(entry)) {
          return `${pad}${key}:\n${toYaml(entry, indent + 2)}`
        }
        if (typeof entry === 'object' && entry !== null) {
          return `${pad}${key}:\n${toYaml(entry, indent + 2)}`
        }
        return `${pad}${key}: ${formatScalar(entry)}`
      })
      .join('\n')
  }

  return `${pad}${formatScalar(value)}`
}

function formatScalar(value: SerializableValue): string {
  if (typeof value === 'number' || typeof value === 'boolean') return String(value)
  if (typeof value !== 'string') return JSON.stringify(value)
  if (/^[A-Za-z0-9_.:/-]+$/.test(value) && !value.includes(':')) return value
  return JSON.stringify(value)
}

export function serializeWorkflow(workflow: Workflow): string {
  const document = stripAuthoringSystemFields(buildDocument(workflow))
  return `${toYaml(document as unknown as SerializableValue)}\n`
}
