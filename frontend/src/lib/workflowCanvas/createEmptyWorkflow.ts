import type { Workflow } from '@/types/workflow'

export function createEmptyWorkflow(): Workflow {
  return {
    id: `wf-${crypto.randomUUID().slice(0, 8)}`,
    name: 'Untitled workflow',
    config: {
      timeoutMs: 30_000,
      failurePolicy: 'resilient',
      hasHooks: false,
      hasOverlay: false,
    },
    steps: [],
  }
}
