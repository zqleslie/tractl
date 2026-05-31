import type { TablerIcon } from '@tabler/icons-react'
import * as TablerIcons from '@tabler/icons-react'

export type EmptyStateProps = {
  icon: string
  text: string
  sub?: string
  shortcut?: string
}

function resolveIcon(name: string): TablerIcon | null {
  const key = `Icon${name
    .split(/[-_]/)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join('')}` as keyof typeof TablerIcons
  const icon = TablerIcons[key]
  return typeof icon === 'function' ? (icon as TablerIcon) : null
}

export function EmptyState({ icon, text, sub, shortcut }: EmptyStateProps) {
  const Icon = resolveIcon(icon)

  return (
    <div
      className="flex flex-col items-center justify-center opacity-40"
      data-testid="empty-state"
    >
      {Icon ? (
        <Icon
          size={22}
          stroke={1.75}
          className="text-text-muted [data-density=compact]_&:size-4 [data-density=comfortable]_&:size-[22px]"
          aria-hidden
        />
      ) : null}
      <p className="mt-2 text-[length:var(--density-font-ui)] text-text-muted">{text}</p>
      {sub ? (
        <p className="mt-1 text-[11px] text-text-muted">{sub}</p>
      ) : null}
      {shortcut ? (
        <p className="mt-0.5 text-[11px] text-text-muted">{shortcut}</p>
      ) : null}
    </div>
  )
}
