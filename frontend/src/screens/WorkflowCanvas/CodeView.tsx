import { useEffect, useRef, useState } from 'react'
import { IconCheck, IconCopy } from '@tabler/icons-react'

export interface CodeViewProps {
  yaml: string
}

export function CodeView({ yaml }: CodeViewProps) {
  const [copied, setCopied] = useState(false)
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(
    () => () => {
      if (timeoutRef.current !== null) clearTimeout(timeoutRef.current)
    },
    [],
  )

  const handleCopy = () => {
    void navigator.clipboard.writeText(yaml).then(() => {
      setCopied(true)
      if (timeoutRef.current !== null) clearTimeout(timeoutRef.current)
      timeoutRef.current = setTimeout(() => setCopied(false), 1500)
    })
  }

  const lines = yaml.split('\n')

  return (
    <div className="relative h-full w-full overflow-auto bg-surface-elevated">
      <button
        type="button"
        onClick={handleCopy}
        title="Copy to clipboard"
        aria-label="Copy to clipboard"
        className="absolute right-3 top-3 z-10 flex h-7 items-center gap-1.5 rounded-[7px] border border-[0.5px] border-border bg-surface px-2.5 text-[11px] font-[400] text-text-muted hover:text-text"
      >
        {copied ? (
          <>
            <IconCheck size={13} stroke={1.5} />
            Copied!
          </>
        ) : (
          <>
            <IconCopy size={13} stroke={1.5} />
            Copy
          </>
        )}
      </button>

      <pre className="m-0 min-h-full p-3 font-mono text-[12px] leading-relaxed text-text">
        <code>
          {lines.map((line, index) => (
            <div key={index} className="flex">
              <span className="mr-4 inline-block w-8 shrink-0 select-none text-right text-text-muted">
                {index + 1}
              </span>
              <span className="whitespace-pre">{line}</span>
            </div>
          ))}
        </code>
      </pre>
    </div>
  )
}
