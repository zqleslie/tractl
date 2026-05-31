import { IconBolt, IconPlayerPlay, IconTopologyStar3 } from '@tabler/icons-react'
import { Badge, Button } from '@/components/primitives'
import type { RecentItem, RecentItemType } from '@/types/recent'
import { cn } from '@/lib/cn'

const typeLabel: Record<RecentItemType, string> = {
  request: 'request',
  workflow: 'workflow',
}

function RecentIcon({ type }: { type: RecentItemType }) {
  const props = { size: 14, stroke: 1.75, className: 'shrink-0 text-text-muted' }
  if (type === 'request') return <IconBolt {...props} />
  return <IconTopologyStar3 {...props} />
}

export type RecentListProps = {
  items: RecentItem[]
  onViewAll: () => void
  onOpenItem?: (item: RecentItem) => void
  onRunItem?: (item: RecentItem) => void
}

export function RecentList({ items, onViewAll, onOpenItem, onRunItem }: RecentListProps) {
  return (
    <section data-testid="fast-start-recent">
      <div className="mb-2.5 flex items-center justify-between">
        <h2 className="text-ui-xs font-medium uppercase tracking-wider text-text-muted">
          Recent
        </h2>
        <button
          type="button"
          className="text-ui-sm text-primary hover:underline"
          data-testid="fast-start-view-all"
          onClick={onViewAll}
        >
          View all
        </button>
      </div>
      <ul className="space-y-0.5">
        {items.map((item) => (
          <li key={item.id}>
            <RecentRow
              item={item}
              onOpen={() => onOpenItem?.(item)}
              onRun={() => onRunItem?.(item)}
            />
          </li>
        ))}
      </ul>
    </section>
  )
}

type RecentRowProps = {
  item: RecentItem
  onOpen?: () => void
  onRun?: () => void
}

function RecentRow({ item, onOpen, onRun }: RecentRowProps) {
  return (
    <div
      className="group relative flex cursor-pointer items-center gap-2.5 rounded-ui px-2 py-1.5 hover:bg-surface-elevated"
      data-testid={`recent-item-${item.id}`}
      onClick={onOpen}
      onKeyDown={(event) => {
        if (event.key === 'Enter' || event.key === ' ') {
          event.preventDefault()
          onOpen?.()
        }
      }}
      role="button"
      tabIndex={0}
    >
      <RecentIcon type={item.type} />
      <span className="min-w-0 flex-1 truncate text-ui-sm text-text">{item.name}</span>
      <Badge tone="neutral" className="shrink-0 normal-case">
        {typeLabel[item.type]}
      </Badge>
      <span className="shrink-0 text-ui-sm text-text-muted">{item.timeAgo}</span>
      <Button
        variant="secondary"
        className={cn(
          'absolute right-2 h-auto py-0.5 px-1.5 text-[10px] opacity-0 transition-opacity group-hover:opacity-100',
        )}
        onClick={(event) => {
          event.stopPropagation()
          onRun?.()
        }}
      >
        <IconPlayerPlay size={10} stroke={2} />
        Run
      </Button>
    </div>
  )
}
