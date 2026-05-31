import { useState } from 'react'
import { cn } from '@/lib/cn'

const hints = {
  beforeStep:
    'Runs before the request · hooks.beforeStep · return { variables, cancel }',
  afterStep:
    'Runs after the response · hooks.afterStep · return { variables, extracts, assertions, cancel }',
} as const

type ScriptRunResult = {
  mutations: unknown
  logs: string[]
  error?: string
}

const chipSets = {
  beforeStep: [
    'ctx.spec',
    'ctx.workflow',
    'ctx.step',
    'ctx.environment',
    'ctx.runtime',
    'tractl.now()',
    'tractl.log()',
    'tractl.uuid()',
  ],
  afterStep: [
    'ctx.step.response',
    'ctx.step.timing',
    'ctx.step.extracts',
    'tractl.log()',
    'tractl.cancel()',
  ],
} as const

export type ScriptPanelProps = {
  hook?: 'beforeStep' | 'afterStep'
  value?: string
  code?: string
  hint?: string
  chips?: readonly string[]
  onChange: (value: string) => void
  readOnly?: boolean
}

export function ScriptPanel({
  hook = 'beforeStep',
  value,
  code,
  hint,
  chips,
  onChange,
  readOnly,
}: ScriptPanelProps) {
  const [copiedChip, setCopiedChip] = useState<string | null>(null)
  const [scriptResult, setScriptResult] = useState<ScriptRunResult | null>(null)
  const resolvedValue = value ?? code ?? ''
  const resolvedHint = hint ?? hints[hook]
  const resolvedChips = chips ?? chipSets[hook]

  const copyChip = async (token: string) => {
    try {
      await navigator.clipboard.writeText(token)
      setCopiedChip(token)
      window.setTimeout(() => setCopiedChip(null), 1200)
    } catch {
      setCopiedChip(null)
    }
  }

  const handleTestScript = () => {
    setScriptResult({
      error:
        'Script sandbox over HTTP is deferred (Phase 15.4). The local API only runs requests and workspace files.',
      mutations: null,
      logs: [],
    })
  }

  return (
    <div>
      <p
        className={cn(
          'mb-2 rounded-[7px] border-[0.5px] border-border bg-surface-elevated',
          'px-[var(--density-padding-sm)] py-[var(--density-padding-xs)]',
          'text-[length:var(--density-font-label)] text-text-muted',
        )}
      >
        {resolvedHint}
      </p>
      <textarea
        aria-label={resolvedHint}
        readOnly={readOnly}
        className={cn(
          'w-full resize-y rounded-[7px] border-[0.5px] border-border bg-surface-elevated',
          'p-[var(--density-padding-md)] font-mono text-[length:var(--density-font-mono)]',
          'leading-[var(--density-line-height)] text-text outline-none focus:border-primary',
          'min-h-[64px]',
        )}
        value={resolvedValue}
        onChange={(event) => onChange(event.currentTarget.value)}
      />
      <div className="relative mt-2 flex flex-wrap items-center gap-[var(--density-gap-sm)]">
        {resolvedChips.map((chip) => (
          <button
            key={chip}
            type="button"
            className="rounded-[4px] border-[0.5px] border-border bg-surface-elevated px-1.5 py-0.5 font-mono text-[10px] text-text-muted hover:text-text"
            onClick={() => void copyChip(chip)}
          >
            {chip}
          </button>
        ))}
        <button
          type="button"
          className="rounded-[4px] border-[0.5px] border-border bg-surface-elevated px-2 py-0.5 text-[10px] font-medium text-text hover:border-primary disabled:opacity-60"
          onClick={() => void handleTestScript()}
          disabled={readOnly}
        >
          Test
        </button>
        {copiedChip ? (
          <span className="absolute right-0 top-0 text-[10px] text-text-muted">
            Copied
          </span>
        ) : null}
      </div>
      {scriptResult ? (
        <pre
          className={cn(
            'mt-2 max-h-36 overflow-auto rounded-[7px] border-[0.5px] border-border bg-surface-elevated p-2 text-[10px]',
            scriptResult.error ? 'text-danger-fg' : 'text-text',
          )}
        >
          {scriptResult.error
            ? scriptResult.error
            : JSON.stringify(scriptResult.mutations, null, 2)}
        </pre>
      ) : null}
    </div>
  )
}
