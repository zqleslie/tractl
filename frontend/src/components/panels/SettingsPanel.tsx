import {
  LabeledSelect,
} from '@/components/request-editor/shared/LabeledField'
import { getEngineDefaults } from '@/stores/engineDefaultsStore'
import type {
  FailurePolicy,
  RequestSettingsDraft,
  RetryStrategy,
} from '@/components/request-editor/types'

export type SettingsPanelProps = {
  value: RequestSettingsDraft
  onChange: (value: RequestSettingsDraft) => void
  readOnly?: boolean
}

const retryStrategies: RetryStrategy[] = [
  'None',
  'Fixed',
  'Linear',
  'Exponential',
]

export function SettingsPanel({ value, onChange, readOnly }: SettingsPanelProps) {
  const showRetryFields = value.retry !== 'None'

  return (
    <div className="space-y-[var(--density-gap-md)] text-[length:var(--density-font-ui)]">
      <label className="flex items-center gap-[var(--density-gap-sm)]">
        <span className="w-24 shrink-0 text-text-muted">Timeout</span>
        <input
          className="w-[52px] rounded-ui border-[0.5px] border-border bg-surface px-2 py-1 text-text"
          value={value.timeoutValue}
          readOnly={readOnly}
          onChange={(event) =>
            onChange({ ...value, timeoutValue: event.currentTarget.value })
          }
        />
        <select
          className="min-w-0 flex-1 rounded-ui border-[0.5px] border-border bg-surface px-2 py-1 text-text"
          value={value.timeoutUnit}
          disabled={readOnly}
          onChange={(event) =>
            onChange({
              ...value,
              timeoutUnit: event.currentTarget.value as RequestSettingsDraft['timeoutUnit'],
            })
          }
        >
          <option value="seconds">seconds</option>
          <option value="milliseconds">milliseconds</option>
          <option value="minutes">minutes</option>
        </select>
      </label>

      <LabeledSelect
        label="Retry"
        options={retryStrategies}
        value={value.retry}
        onChange={(retry) =>
          onChange({
            ...value,
            retry: retry as RetryStrategy,
            retryConfig:
              retry === 'None'
                ? null
                : value.retryConfig ?? {
                    strategy: retry.toLowerCase() as 'fixed',
                    maxAttempts: getEngineDefaults().retry.maxAttempts,
                    delayMs: 1000,
                    backoffFactor: getEngineDefaults().retry.backoffFactor,
                  },
          })
        }
      />

      {showRetryFields && value.retryConfig ? (
        <div className="space-y-2 rounded-ui border-[0.5px] border-border bg-surface-elevated p-[var(--density-padding-md)]">
          <label className="flex items-center gap-2">
            <span className="w-24 text-text-muted">Max attempts</span>
            <input
              type="number"
              min={1}
              max={10}
              className="w-[52px] rounded-ui border-[0.5px] border-border bg-surface px-2 py-1"
              value={value.retryConfig.maxAttempts}
              readOnly={readOnly}
              onChange={(event) =>
                onChange({
                  ...value,
                  retryConfig: {
                    ...value.retryConfig!,
                    maxAttempts: Number(event.currentTarget.value),
                  },
                })
              }
            />
          </label>
          <label className="flex items-center gap-2">
            <span className="w-24 text-text-muted">Delay</span>
            <input
              type="number"
              className="w-[52px] rounded-ui border-[0.5px] border-border bg-surface px-2 py-1"
              value={value.retryConfig.delayMs}
              readOnly={readOnly}
              onChange={(event) =>
                onChange({
                  ...value,
                  retryConfig: {
                    ...value.retryConfig!,
                    delayMs: Number(event.currentTarget.value),
                  },
                })
              }
            />
            <span className="text-text-muted">ms</span>
          </label>
          {value.retry === 'Linear' || value.retry === 'Exponential' ? (
            <label className="flex items-center gap-2">
              <span className="w-24 text-text-muted">Backoff factor</span>
              <input
                type="number"
                className="w-[52px] rounded-ui border-[0.5px] border-border bg-surface px-2 py-1"
                value={value.retryConfig.backoffFactor ?? getEngineDefaults().retry.backoffFactor}
                readOnly={readOnly}
                onChange={(event) =>
                  onChange({
                    ...value,
                    retryConfig: {
                      ...value.retryConfig!,
                      backoffFactor: Number(event.currentTarget.value),
                    },
                  })
                }
              />
            </label>
          ) : null}
        </div>
      ) : null}

      <LabeledSelect
        label="Failure policy"
        options={['resilient', 'failFast']}
        value={value.failurePolicy}
        onChange={(failurePolicy) =>
          onChange({
            ...value,
            failurePolicy: failurePolicy as FailurePolicy,
          })
        }
      />
      <p className="text-[length:var(--density-font-label)] text-text-muted">
        {value.failurePolicy === getEngineDefaults().failurePolicy
          ? 'Dependent steps are skipped on failure, independent steps continue'
          : 'All pending steps are cancelled immediately on first failure'}
      </p>
    </div>
  )
}
