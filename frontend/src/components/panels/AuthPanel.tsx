import { useState } from 'react'
import { EmptyState } from '@/components/common/EmptyState'
import { EnvironmentVariableInput } from '@/components/environment/EnvironmentVariableInput'
import {
  LabeledInput,
  LabeledSelect,
} from '@/components/request-editor/shared/LabeledField'
import type { AuthType, RequestAuthDraft } from '@/components/request-editor/types'

export type AuthPanelProps = {
  value: RequestAuthDraft
  onChange: (value: RequestAuthDraft) => void
  readOnly?: boolean
}

const authTypes: AuthType[] = ['None', 'Bearer token', 'Basic auth', 'API key']

export function AuthPanel({ value, onChange, readOnly }: AuthPanelProps) {
  const [showPassword, setShowPassword] = useState(false)

  return (
    <div className="space-y-[var(--density-gap-md)]">
      <LabeledSelect
        label="Type"
        options={authTypes}
        value={value.type}
        onChange={(type) => onChange({ ...value, type: type as AuthType })}
      />

      {value.type === 'None' ? (
        <EmptyState icon="shield-off" text="No authentication" />
      ) : null}

      {value.type === 'Bearer token' ? (
        <>
          <EnvironmentVariableInput
            aria-label="Bearer token"
            className="w-full rounded-ui border-[0.5px] border-border bg-surface px-[var(--density-padding-sm)] py-[var(--density-padding-xs)] font-mono text-[length:var(--density-font-mono)] text-text"
            value={value.token || value.tokenRef}
            onChange={(token) => onChange({ ...value, token, tokenRef: token })}
          />
          <p className="text-[length:var(--density-font-label)] text-text-muted">
            Reference a stored secret:{' '}
            <span className="font-mono">{'${secrets.env.name}'}</span>
          </p>
          <p className="text-[length:var(--density-font-label)] text-text-muted">
            Add to header as: Authorization: Bearer &lt;token&gt;
          </p>
        </>
      ) : null}

      {value.type === 'Basic auth' ? (
        <>
          <LabeledInput
            label="Username"
            value={value.username}
            onChange={(username) => onChange({ ...value, username })}
          />
          <label className="block text-ui-sm text-text-muted">
            Password
            <div className="mt-1 flex gap-2">
              <input
                type={showPassword ? 'text' : 'password'}
                className="w-full rounded-ui border-[0.5px] border-border bg-surface px-2 py-1.5 text-text"
                value={value.password}
                readOnly={readOnly}
                onChange={(event) =>
                  onChange({ ...value, password: event.currentTarget.value })
                }
              />
              <button
                type="button"
                className="shrink-0 rounded-ui border-[0.5px] border-border px-2 text-ui-xs text-text-muted"
                onClick={() => setShowPassword((open) => !open)}
              >
                {showPassword ? 'Hide' : 'Show'}
              </button>
            </div>
          </label>
          <p className="text-[length:var(--density-font-label)] text-text-muted">
            Encoded as base64, sent as Authorization: Basic &lt;encoded&gt;
          </p>
        </>
      ) : null}

      {value.type === 'API key' ? (
        <>
          <LabeledInput
            label="Key name"
            value={value.keyName}
            onChange={(keyName) => onChange({ ...value, keyName })}
          />
          <EnvironmentVariableInput
            aria-label="API key value"
            className="w-full rounded-ui border-[0.5px] border-border bg-surface px-[var(--density-padding-sm)] py-[var(--density-padding-xs)] font-mono text-[length:var(--density-font-mono)] text-text"
            value={value.keyValue}
            onChange={(keyValue) => onChange({ ...value, keyValue })}
          />
          <LabeledSelect
            label="Placement"
            options={['header', 'query']}
            value={value.placement}
            onChange={(placement) =>
              onChange({
                ...value,
                placement: placement as RequestAuthDraft['placement'],
              })
            }
          />
        </>
      ) : null}
    </div>
  )
}
