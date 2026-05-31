import { useState } from 'react'
import { IconChevronDown, IconChevronUp, IconQuestionMark } from '@tabler/icons-react'
import { cn } from '@/lib/cn'
import { detectSurface } from '@/platform'
import type { FailurePolicy, WorkflowConfig } from '@/types/workflow'

const ENGINE_DEFAULT_CONCURRENCY = 4
const MIN_CONCURRENCY = 1
const MAX_CONCURRENCY = 4

export interface WorkflowConfigStripProps {
  config: WorkflowConfig
  isExpanded: boolean
  onToggle: () => void
  onChange: (config: WorkflowConfig) => void
}

type TimeoutUnit = 's' | 'ms' | 'min'

const chipClass =
  'inline-flex h-5 shrink-0 items-center rounded-full border border-[0.5px] border-border px-2 text-[11px] font-[400] text-text-muted'
const inputClass =
  'h-7 rounded-[7px] border border-[0.5px] border-border bg-surface px-2 text-[12px] font-[400] text-text outline-none focus:border-[#185FA5]'
const policyBase =
  'inline-flex h-7 items-center rounded-[7px] border border-[0.5px] px-3 text-[12px] font-[400] transition-colors'
const policyInactive = 'border-border text-text-muted hover:text-text'
const policyActive =
  'border-[#185FA5] bg-[#E6F1FB] text-[#185FA5] dark:border-[#B5D4F4] dark:bg-[#042C53] dark:text-[#B5D4F4]'

function toUnitValue(timeoutMs: number, unit: TimeoutUnit): number {
  if (unit === 'ms') return timeoutMs
  if (unit === 'min') return timeoutMs / 60000
  return timeoutMs / 1000
}

function toMs(value: number, unit: TimeoutUnit): number {
  if (unit === 'ms') return value
  if (unit === 'min') return value * 60000
  return value * 1000
}

function clampConcurrency(value: number): number {
  return Math.min(MAX_CONCURRENCY, Math.max(MIN_CONCURRENCY, Math.round(value)))
}

export function WorkflowConfigStrip({
  config,
  isExpanded,
  onToggle,
  onChange,
}: WorkflowConfigStripProps) {
  const [timeoutUnit, setTimeoutUnit] = useState<TimeoutUnit>('s')

  const setPolicy = (failurePolicy: FailurePolicy) => onChange({ ...config, failurePolicy })

  const setConcurrency = (raw: number) => {
    if (Number.isNaN(raw)) return
    onChange({ ...config, concurrency: clampConcurrency(raw) })
  }

  const setTimeoutValue = (raw: number) => {
    if (Number.isNaN(raw)) return
    onChange({ ...config, timeoutMs: Math.max(0, Math.round(toMs(raw, timeoutUnit))) })
  }

  return (
    <div className="shrink-0 border-t border-b border-[0.5px] border-border bg-surface">
      <button
        type="button"
        onClick={onToggle}
        className="flex h-8 w-full items-center gap-2 px-3 text-left"
        aria-expanded={isExpanded}
      >
        <span className={chipClass}>
          Concurrency: {config.concurrency ?? `default (${ENGINE_DEFAULT_CONCURRENCY})`}
        </span>
        <span className={chipClass}>Timeout: {config.timeoutMs / 1000}s</span>
        <span className={chipClass}>{config.failurePolicy}</span>
        {config.hasHooks && <span className={chipClass}>Hooks</span>}
        {config.hasOverlay && (
          <span className={chipClass}>Overlay: {config.overlayName ?? 'on'}</span>
        )}
        <span className="flex-1" />
        {isExpanded ? (
          <IconChevronUp size={15} stroke={1.5} className="text-text-muted" />
        ) : (
          <IconChevronDown size={15} stroke={1.5} className="text-text-muted" />
        )}
      </button>

      {isExpanded && (
        <div className="flex flex-wrap items-end gap-6 px-3 pb-3 pt-1">
          <label className="flex flex-col gap-1">
            <span className="flex items-center gap-1 text-[11px] font-[400] text-text-muted">
              Max concurrency
              {detectSurface() === 'web' ? (
                <span className="group relative inline-flex">
                  <span
                    className="inline-flex h-4 w-4 items-center justify-center rounded-full border border-[0.5px] border-border text-text-muted"
                    aria-label="Browser concurrency help"
                    tabIndex={0}
                  >
                    <IconQuestionMark size={10} stroke={1.8} />
                  </span>
                  <span className="pointer-events-none absolute left-1/2 top-5 z-20 hidden w-64 -translate-x-1/2 rounded-card border border-[0.5px] border-border bg-surface-elevated p-2 text-[11px] font-[400] leading-relaxed text-text shadow-none group-hover:block group-focus-within:block">
                    Browser runs execute steps one at a time today. Max concurrency applies on CLI
                    and desktop runs; the waterfall reflects actual browser timing until parallel
                    WASM scheduling ships.
                  </span>
                </span>
              ) : null}
            </span>
            <input
              type="number"
              min={MIN_CONCURRENCY}
              max={MAX_CONCURRENCY}
              step={1}
              placeholder={String(ENGINE_DEFAULT_CONCURRENCY)}
              value={config.concurrency ?? ENGINE_DEFAULT_CONCURRENCY}
              onChange={(event) => setConcurrency(event.target.valueAsNumber)}
              onBlur={(event) => setConcurrency(event.target.valueAsNumber)}
              className={cn(inputClass, 'w-20')}
            />
          </label>

          <div className="flex flex-col gap-1">
            <span className="text-[11px] font-[400] text-text-muted">Timeout</span>
            <div className="flex items-center gap-1.5">
              <input
                type="number"
                min={0}
                value={toUnitValue(config.timeoutMs, timeoutUnit)}
                onChange={(event) => setTimeoutValue(event.target.valueAsNumber)}
                className={cn(inputClass, 'w-24')}
              />
              <select
                value={timeoutUnit}
                onChange={(event) => setTimeoutUnit(event.target.value as TimeoutUnit)}
                aria-label="Timeout unit"
                className={cn(inputClass, 'w-16')}
              >
                <option value="s">s</option>
                <option value="ms">ms</option>
                <option value="min">min</option>
              </select>
            </div>
          </div>

          <div className="flex flex-col gap-1">
            <span className="text-[11px] font-[400] text-text-muted">Failure policy</span>
            <div className="flex items-center gap-1.5">
              <button
                type="button"
                onClick={() => setPolicy('resilient')}
                className={cn(
                  policyBase,
                  config.failurePolicy === 'resilient' ? policyActive : policyInactive,
                )}
              >
                resilient
              </button>
              <button
                type="button"
                onClick={() => setPolicy('failFast')}
                className={cn(
                  policyBase,
                  config.failurePolicy === 'failFast' ? policyActive : policyInactive,
                )}
              >
                failFast
              </button>
            </div>
          </div>

          <div className="flex flex-col gap-1">
            <span className="text-[11px] font-[400] text-text-muted">Hooks</span>
            <div className="flex items-center gap-2">
              <span className={chipClass}>{config.hasHooks ? 'Enabled' : 'None'}</span>
              <button
                type="button"
                className="text-[11px] font-[400] text-[#185FA5] hover:underline dark:text-[#B5D4F4]"
              >
                Configure
              </button>
            </div>
          </div>

          <label className="flex flex-col gap-1">
            <span className="text-[11px] font-[400] text-text-muted">Overlay</span>
            <select
              value={config.hasOverlay ? (config.overlayName ?? 'on') : ''}
              onChange={(event) => {
                const value = event.target.value
                if (value === '') {
                  onChange({ ...config, hasOverlay: false, overlayName: undefined })
                } else {
                  onChange({ ...config, hasOverlay: true, overlayName: value })
                }
              }}
              className={cn(inputClass, 'w-40')}
            >
              <option value="">None</option>
              {config.hasOverlay && config.overlayName && (
                <option value={config.overlayName}>{config.overlayName}</option>
              )}
            </select>
          </label>
        </div>
      )}
    </div>
  )
}
