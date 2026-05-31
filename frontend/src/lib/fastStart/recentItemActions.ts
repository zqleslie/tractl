import { screenForRecentItem } from '@/lib/navigateFromRecent'
import type { RecentItem } from '@/types/recent'
import { openWorkflowFromHistoryDocument } from '@/lib/workflowCanvas/openWorkflowFromHistoryDocument'
import { recordRunHistoryEntry } from '@/lib/runHistory/recordRunHistoryEntry'
import { runTraCtlDocument } from '@/lib/tractlDocument/runDocument'
import { useEnvironmentStore } from '@/stores/environmentStore'
import { useRunHistoryStore } from '@/stores/runHistoryStore'
import { useUiStore } from '@/stores/uiStore'

function activeEnvironmentVariables(): Record<string, string> {
  const { environments, activeEnvironmentId } = useEnvironmentStore.getState()
  const active = environments.find(
    (environment) => environment.id === activeEnvironmentId,
  )
  return active?.variables ?? {}
}

export function openRecentItem(item: RecentItem): void {
  const ui = useUiStore.getState()
  const historyId = item.historyEntryId ?? item.id
  const entry = useRunHistoryStore
    .getState()
    .entries.find((candidate) => candidate.id === historyId)

  if (!entry) return

  const screen = screenForRecentItem({
    ...item,
    type: entry.sourceType === 'workflow' ? 'workflow' : 'request',
  })

  if (screen === 'request') {
    ui.openHistoryEntry(entry)
    return
  }

  if (entry.document && entry.sourceType === 'workflow') {
    openWorkflowFromHistoryDocument({
      document: entry.document,
      sourceName: entry.sourceName,
      sourceFormat: entry.sourceFormat,
    })
    return
  }

  ui.setSidebarTab('workflows')
  ui.openWorkflowPlaceholder()
}

export async function runRecentItem(item: RecentItem): Promise<string | null> {
  const historyId = item.historyEntryId ?? item.id
  const entry = useRunHistoryStore
    .getState()
    .entries.find((candidate) => candidate.id === historyId)

  if (!entry?.document) {
    openRecentItem(item)
    return null
  }

  const sourceType = entry.sourceType ?? 'request'
  const format = entry.sourceFormat ?? 'json'
  const raw = JSON.stringify(entry.document)

  const result = await runTraCtlDocument({
    raw,
    format,
    sourceName: entry.sourceName ?? entry.requestName,
    sourceType,
    environmentVariables: activeEnvironmentVariables(),
  })

  if (!result.ok) {
    return result.message
  }

  const newId = recordRunHistoryEntry({
    requestName: result.summary.requestName,
    method: result.summary.method,
    url: result.summary.url,
    result: result.result,
    sourceType: entry.sourceType,
    sourceName: entry.sourceName,
    sourceFormat: result.sourceFormat,
    document: result.spec,
  })

  const updated = useRunHistoryStore.getState().entries.find((e) => e.id === newId)
  if (updated && (entry.sourceType === 'request' || !entry.sourceType)) {
    useUiStore.getState().openHistoryEntry(updated)
  } else {
    useUiStore.getState().setActiveScreen('run-history')
  }

  return null
}
