import type { StatusTone } from '@/components/primitives'
import type { WorkflowRunOutcome } from '@/platform/web/workflowRunAdapter'

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

export function buildLastRunStatusLabel(input: {
  runState: 'idle' | 'running' | 'complete' | 'error'
  workflowOutcome?: WorkflowRunOutcome | null
  durationMs?: number
  assertionsPassed?: number
  assertionsTotal?: number
  failedSteps?: number
  errorMessage?: string
}): { tone: StatusTone; label: string } {
  if (input.runState === 'running') {
    return { tone: 'running', label: 'Running workflow…' }
  }

  if (input.runState === 'idle' || input.workflowOutcome == null) {
    return { tone: 'ready', label: 'Engine ready' }
  }

  const duration = formatDuration(input.durationMs ?? 0)

  if (input.workflowOutcome === 'error') {
    const short = input.errorMessage?.split(':').slice(1).join(':').trim()
    return {
      tone: 'error',
      label: short
        ? `Error: ${short.slice(0, 48)}${short.length > 48 ? '…' : ''}`
        : 'Error: workflow run failed',
    }
  }

  if (input.workflowOutcome === 'failed') {
    const failed = input.failedSteps ?? 0
    return {
      tone: 'error',
      label: `Last run: failed · ${duration} · ${failed} failed`,
    }
  }

  const assertions =
    input.assertionsTotal != null
      ? `${input.assertionsPassed ?? 0}/${input.assertionsTotal} assertions`
      : undefined

  return {
    tone: 'success',
    label: assertions
      ? `Last run: passed · ${duration} · ${assertions}`
      : `Last run: passed · ${duration}`,
  }
}
