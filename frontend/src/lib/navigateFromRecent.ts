import type { RecentItem } from '@/types/recent'
import type { AppScreen } from '@/stores/uiStore'

export function screenForRecentItem(item: RecentItem): AppScreen {
  if (item.type === 'request') return 'request'
  return 'workflow'
}
