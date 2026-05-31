export type RecentItemType = 'request' | 'workflow'

export type RecentItem = {
  id: string
  name: string
  type: RecentItemType
  timeAgo: string
  historyEntryId?: string
}
