import { getEngineDefaults } from '@/stores/engineDefaultsStore'
import type { Workflow } from '@/types/workflow'

export function createEmptyWorkflow(): Workflow {
  const defaults = getEngineDefaults()
  return {
    id: `wf-${crypto.randomUUID().slice(0, 8)}`,
    name: 'Untitled workflow',
    config: {
      timeoutMs: defaults.timeoutMs,
      failurePolicy: defaults.failurePolicy,
      hasHooks: false,
      hasOverlay: false,
    },
    steps: [],
  }
}
