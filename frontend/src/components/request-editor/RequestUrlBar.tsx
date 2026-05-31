import { useEffect, useRef, useState } from 'react'
import {
  IconChevronDown,
  IconCopy,
  IconLoader2,
  IconPlayerPlay,
} from '@tabler/icons-react'
import { EnvironmentVariableInput } from '@/components/environment/EnvironmentVariableInput'
import { Button, type HttpMethod } from '@/components/primitives'
import { cn } from '@/lib/cn'

const methods = ['POST', 'GET', 'PUT', 'PATCH', 'DELETE'] satisfies HttpMethod[]

export type RequestUrlBarProps = {
  method: HttpMethod
  url: string
  canRun: boolean
  isRunning: boolean
  onMethodChange: (method: HttpMethod) => void
  onUrlChange: (url: string) => void
  onRun: () => void
  onSaveToCollection: () => void
}

export function RequestUrlBar({
  method,
  url,
  canRun,
  isRunning,
  onMethodChange,
  onUrlChange,
  onRun,
  onSaveToCollection,
}: RequestUrlBarProps) {
  const [sendMenuOpen, setSendMenuOpen] = useState(false)
  const sendMenuRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    if (!sendMenuOpen) return undefined
    const closeOnOutside = (event: MouseEvent) => {
      if (!sendMenuRef.current?.contains(event.target as Node)) {
        setSendMenuOpen(false)
      }
    }
    document.addEventListener('mousedown', closeOnOutside)
    return () => document.removeEventListener('mousedown', closeOnOutside)
  }, [sendMenuOpen])

  const copyUrl = async () => {
    if (!url.trim()) return
    try {
      await navigator.clipboard.writeText(url)
    } catch {
      // clipboard may be unavailable
    }
    setSendMenuOpen(false)
  }

  return (
    <div className="flex shrink-0 items-center gap-2.5 border-b-[0.5px] border-border bg-surface px-4 py-2.5">
      <select
        aria-label="HTTP method"
        className="h-9 min-w-[80px] rounded-ui border-[0.5px] border-border bg-surface-elevated px-2 text-ui-sm font-semibold text-text"
        value={method}
        onChange={(event) =>
          onMethodChange(event.currentTarget.value as HttpMethod)
        }
      >
        {methods.map((option) => (
          <option key={option} value={option}>
            {option}
          </option>
        ))}
      </select>
      <EnvironmentVariableInput
        aria-label="Request URL"
        placeholder="https://api.example.com/resource or {{baseUrl}}/path"
        className="h-9 min-w-0 flex-1 rounded-ui border-[0.5px] border-border bg-surface px-3 font-mono text-ui-sm text-text outline-none focus:border-primary"
        value={url}
        onChange={onUrlChange}
      />

      <div ref={sendMenuRef} className="relative flex shrink-0">
        <Button
          variant="primary"
          className="h-9 rounded-r-none px-4"
          data-testid="request-run"
          disabled={!canRun}
          onClick={onRun}
        >
          {isRunning ? (
            <IconLoader2 size={14} stroke={2} className="animate-spin" />
          ) : (
            <IconPlayerPlay size={14} stroke={2} />
          )}
          {isRunning ? 'Sending…' : 'Send'}
        </Button>
        <Button
          variant="primary"
          className={cn('h-9 rounded-l-none border-l border-primary/30 px-1.5')}
          aria-label="Send options"
          aria-expanded={sendMenuOpen}
          aria-haspopup="menu"
          data-testid="request-send-menu"
          disabled={!canRun && !url.trim()}
          onClick={() => setSendMenuOpen((open) => !open)}
        >
          <IconChevronDown size={14} stroke={2} />
        </Button>
        {sendMenuOpen ? (
          <div
            role="menu"
            className="absolute right-0 top-10 z-20 min-w-[160px] rounded-ui border-[0.5px] border-border bg-surface py-1 shadow-lg"
          >
            <button
              type="button"
              role="menuitem"
              className="flex w-full items-center gap-2 px-3 py-1.5 text-left text-ui-xs text-text hover:bg-surface-elevated"
              data-testid="request-copy-url"
              onClick={() => void copyUrl()}
            >
              <IconCopy size={14} stroke={1.75} />
              Copy URL
            </button>
          </div>
        ) : null}
      </div>

      <Button
        variant="secondary"
        className="h-9 shrink-0 px-4"
        data-testid="request-save-collection"
        onClick={onSaveToCollection}
      >
        Save
      </Button>
    </div>
  )
}
