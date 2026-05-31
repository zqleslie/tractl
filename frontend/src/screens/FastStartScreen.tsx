import { useCallback, useMemo, useState } from 'react'
import {
  IconBolt,
  IconFolderOpen,
  IconHistory,
  IconTopologyStar3,
} from '@tabler/icons-react'
import { ActionCard, RecentList } from '@/components/fast-start'
import type { RecentItem } from '@/types/recent'
import { buildRecentItems } from '@/lib/fastStart/buildRecentItems'
import { openRecentItem, runRecentItem } from '@/lib/fastStart/recentItemActions'
import {
  openTraCtlFileFromDisk,
  TRACTL_FILE_INPUT_ACCEPT,
  type TraCtlFileActionError,
} from '@/lib/tractlDocument/openTraCtlFileFromDisk'
import type { DocumentValidationIssue } from '@/lib/tractlDocument/parseValidateDocument'
import { useEnvironmentStore } from '@/stores/environmentStore'
import { useRunHistoryStore } from '@/stores/runHistoryStore'
import { useUiStore } from '@/stores/uiStore'

function greeting(): string {
  const hour = new Date().getHours()
  if (hour < 12) return 'Good morning'
  if (hour < 18) return 'Good afternoon'
  return 'Good evening'
}

function ValidationErrorPanel({
  title,
  message,
  errors,
}: {
  title: string
  message: string
  errors: DocumentValidationIssue[]
}) {
  return (
    <div
      className="mt-4 rounded-ui border-[0.5px] border-danger bg-danger-bg px-3 py-2"
      data-testid="fast-start-action-error"
      role="alert"
    >
      <p className="text-ui-sm font-medium text-danger-fg">{title}</p>
      <p className="mt-1 text-ui-xs text-danger-fg">{message}</p>
      {errors.length > 0 ? (
        <ul className="mt-2 list-inside list-disc space-y-0.5 text-ui-xs text-danger-fg">
          {errors.map((error, index) => (
            <li key={`${error.code ?? 'err'}-${index}`}>
              {error.field ? `${error.field}: ` : ''}
              {error.message}
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  )
}

export function FastStartScreen() {
  const setActiveScreen = useUiStore((s) => s.setActiveScreen)
  const setSidebarTab = useUiStore((s) => s.setSidebarTab)
  const openNewRequest = useUiStore((s) => s.openNewRequest)
  const openWorkflowPlaceholder = useUiStore((s) => s.openWorkflowPlaceholder)

  const historyEntries = useRunHistoryStore((s) => s.entries)

  const activeEnvironment = useEnvironmentStore((state) =>
    state.environments.find(
      (environment) => environment.id === state.activeEnvironmentId,
    ),
  )

  const [actionError, setActionError] = useState<TraCtlFileActionError | null>(
    null,
  )
  const [isRunningFile, setIsRunningFile] = useState(false)
  const fileInputId = 'fast-start-open-file-input'

  const recentItems = useMemo(
    () => buildRecentItems(historyEntries),
    [historyEntries],
  )

  const environmentVariables = activeEnvironment?.variables ?? {}

  const clearActionError = useCallback(() => setActionError(null), [])

  const goRequest = () => {
    clearActionError()
    setSidebarTab('requests')
    openNewRequest()
  }

  const goWorkflow = () => {
    clearActionError()
    setSidebarTab('workflows')
    openWorkflowPlaceholder()
  }

  const handleSelectedFile = async (file: File) => {
    clearActionError()
    setIsRunningFile(true)
    try {
      const result = await openTraCtlFileFromDisk(file, environmentVariables)
      if (!result.ok) {
        setActionError(result.error)
      }
    } finally {
      setIsRunningFile(false)
    }
  }

  const handleOpenRecent = (item: RecentItem) => {
    clearActionError()
    openRecentItem(item)
  }

  const handleRunRecent = async (item: RecentItem) => {
    clearActionError()
    const error = await runRecentItem(item)
    if (error) {
      setActionError({
        title: 'Could not run item',
        message: error,
        errors: [{ message: error }],
      })
    }
  }

  return (
    <div
      className="mx-auto max-w-3xl px-8 pb-6 pt-9"
      data-testid="screen-fast-start"
    >
      <h1 className="text-xl font-medium text-text">{greeting()}</h1>
      <p className="mt-1 text-ui-sm text-text-muted">
        What would you like to validate today?
      </p>

      <div className="mt-8 grid grid-cols-2 gap-2.5">
        <ActionCard
          testId="fast-start-new-request"
          title="New request"
          description="Single HTTP step with assertions and extracts"
          icon={<IconBolt size={17} stroke={1.75} />}
          iconTone="request"
          featured
          onClick={goRequest}
        />
        <ActionCard
          testId="fast-start-run-history"
          title="Run history"
          description="Revisit and inspect previous executions"
          icon={<IconHistory size={17} stroke={1.75} />}
          iconTone="history"
          onClick={() => setActiveScreen('run-history')}
        />
        <ActionCard
          testId="fast-start-new-workflow"
          title="New workflow"
          description="Multi-step DAG with dependencies and overlays"
          icon={<IconTopologyStar3 size={17} stroke={1.75} />}
          iconTone="workflow"
          onClick={goWorkflow}
        />
        <ActionCard
          testId="fast-start-open-file"
          title={isRunningFile ? 'Running file…' : 'Open file'}
          description="Run YAML, JSON, or TOON from the repo examples/ folder"
          icon={<IconFolderOpen size={17} stroke={1.75} />}
          iconTone="file"
          htmlFor={fileInputId}
          disabled={isRunningFile}
        />
      </div>
      <input
        id={fileInputId}
        type="file"
        accept={TRACTL_FILE_INPUT_ACCEPT}
        className="sr-only"
        tabIndex={-1}
        data-testid="fast-start-file-input"
        onChange={(event) => {
          const file = event.currentTarget.files?.[0]
          event.currentTarget.value = ''
          if (file) {
            void handleSelectedFile(file)
          }
        }}
      />

      {actionError ? (
        <ValidationErrorPanel
          title={actionError.title}
          message={actionError.message}
          errors={actionError.errors}
        />
      ) : null}

      <div className="mt-9">
        {recentItems.length > 0 ? (
          <RecentList
            items={recentItems}
            onViewAll={() => setActiveScreen('run-history')}
            onOpenItem={handleOpenRecent}
            onRunItem={(item) => void handleRunRecent(item)}
          />
        ) : (
          <section data-testid="fast-start-recent-empty">
            <h2 className="text-ui-xs font-medium uppercase tracking-wider text-text-muted">
              Recent
            </h2>
            <p className="mt-2 text-ui-sm text-text-muted">
              No history yet. Run a request or open a file from{' '}
              <code className="font-mono text-ui-xs">examples/</code> to see it
              here.
            </p>
          </section>
        )}
      </div>
    </div>
  )
}
