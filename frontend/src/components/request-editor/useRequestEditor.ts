import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { createEmptyRequestDraft } from '@/components/request-editor/createEmptyRequestDraft'
import {
  addAssertionRow,
  addExtractRow,
  addKeyValueRow,
  countPopulatedRows,
  patchDraft,
  removeAssertionRow,
  removeExtractRow,
  removeKeyValueRow,
  updateAssertionRow,
  updateExtractRow,
  updateKeyValueRow,
} from '@/components/request-editor/requestDraftMutations'
import { requestDraftToTraCtlSpec } from '@/components/request-editor/requestDraftToTraCtlSpec'
import { validateRequestRunUrl } from '@/components/request-editor/normalizeRequestUrl'
import {
  formatMissingEnvironmentVariables,
  resolveTraCtlSpecEnvironmentVariables,
} from '@/lib/resolveEnvironmentVariables'
import type {
  AssertionRowModel,
  ConfigTab,
  ExtractRowModel,
  KeyValueRow,
  RequestAuthDraft,
  RequestBodyDraft,
  RequestDraft,
  RequestScriptDraft,
  RequestSettingsDraft,
  ResultTab,
} from '@/components/request-editor/types'
import type { HttpMethod } from '@/components/primitives'
import { getRequestExecutionRunner } from '@/platform/requestExecution/getRequestExecutionRunner'
import { LocalApiUnavailableError } from '@/platform/localApi/client'
import type { RequestRunResult } from '@/platform/localApi/types'
import { WasmRuntimeUnavailableError } from '@/platform/web/requestExecutionRunner'
import { useEnvironmentStore } from '@/stores/environmentStore'
import { saveRequestFile, runRequest as apiRunRequest } from '@/api/requests'
import {
  executionResultToRequestRunResult,
  requestRunResultToExecutionResult,
} from '@/lib/execution/mapRunResults'
import { buildRunRequestPayload } from '@/lib/requestEditor/buildRunPayload'
import { syncContentTypeHeader } from '@/lib/requestEditor/contentTypeHeader'
import {
  draftToRequestState,
  requestStateToDraft,
} from '@/lib/requestEditor/workspaceMapping'
import { recordRunHistoryEntry } from '@/lib/runHistory/recordRunHistoryEntry'
import { useExecutionStore } from '@/stores/executionStore'
import { useUiStore } from '@/stores/uiStore'
import { useWorkspaceStore } from '@/stores/workspaceStore'

const requestRunner = getRequestExecutionRunner()

function requestTabTitleFromUrl(url: string): string {
  const trimmed = url.trim()
  if (!trimmed) return 'Untitled request'

  try {
    const parsed = new URL(trimmed)
    const path = parsed.pathname === '/' ? '' : parsed.pathname
    return `${parsed.host}${path}`.slice(0, 48)
  } catch {
    return trimmed.replace(/^\{\{|\}\}$/g, '').slice(0, 48)
  }
}

function environmentResolutionErrorResult(missing: string[]): RequestRunResult {
  const message = formatMissingEnvironmentVariables(missing)

  return {
    passed: false,
    durationMs: 0,
    statusCode: 0,
    statusLabel: 'Environment error',
    contentType: 'text/plain',
    body: message,
    headers: [],
    assertionResults: [],
    extractResults: [],
    passedCount: 0,
    totalCount: 0,
    timeline: [],
    error: message,
  }
}

export function useRequestEditor() {
  const pendingHistoryEntry = useUiStore((s) => s.pendingHistoryEntry)
  const clearPendingHistoryEntry = useUiStore((s) => s.clearPendingHistoryEntry)
  const pendingCollectionRequest = useUiStore((s) => s.pendingCollectionRequest)
  const clearPendingCollectionRequest = useUiStore(
    (s) => s.clearPendingCollectionRequest,
  )
  const updateActiveRequestTab = useUiStore((s) => s.updateActiveRequestTab)
  const activeTabId = useUiStore((s) => s.activeTabId)
  const requestEditorKey = useUiStore((s) => s.requestEditorKey)
  const requestId = activeTabId ?? `request-${requestEditorKey}`
  const setActiveRequestId = useWorkspaceStore((s) => s.setActiveRequestId)
  const upsertWorkspaceRequest = useWorkspaceStore((s) => s.upsertRequest)
  const getWorkspaceRequest = useWorkspaceStore((s) => s.getRequest)
  const createWorkspaceRequest = useWorkspaceStore((s) => s.createRequest)

  const [method, setMethod] = useState<HttpMethod>(
    () => pendingHistoryEntry?.method ?? 'GET',
  )
  const [url, setUrl] = useState(() => pendingHistoryEntry?.url ?? '')
  const [requestName] = useState(
    () => pendingHistoryEntry?.requestName ?? 'Untitled request',
  )
  const [fileId, setFileId] = useState<string | null>(null)
  const [draft, setDraft] = useState<RequestDraft>(createEmptyRequestDraft)
  const [activeConfigTab, setActiveConfigTab] = useState<ConfigTab>('params')
  const [activeResultTab, setActiveResultTab] = useState<ResultTab>('body')
  const [resultsOpen, setResultsOpen] = useState(
    () => pendingHistoryEntry != null || pendingCollectionRequest != null,
  )
  const [isSaving, setIsSaving] = useState(false)
  const [isRunning, setIsRunning] = useState(false)
  const [apiAvailable, setApiAvailable] = useState<boolean | null>(null)
  const [saveError, setSaveError] = useState<string | null>(null)
  const [runError, setRunError] = useState<string | null>(null)
  const [runResult, setRunResult] = useState<RequestRunResult | null>(
    () => pendingHistoryEntry?.result ?? null,
  )
  const activeEnvironment = useEnvironmentStore((state) =>
    state.environments.find(
      (environment) => environment.id === state.activeEnvironmentId,
    ),
  )

  const saveGeneration = useRef(0)
  const startRun = useExecutionStore((state) => state.startRun)
  const setExecutionResult = useExecutionStore((state) => state.setResult)
  const setExecutionError = useExecutionStore((state) => state.setError)
  const resetExecution = useExecutionStore((state) => state.reset)

  const syncWorkspace = useCallback(
    (nextDraft: RequestDraft, nextMethod: HttpMethod, nextUrl: string) => {
      upsertWorkspaceRequest(
        draftToRequestState(requestId, requestName, nextMethod, nextUrl, nextDraft),
      )
    },
    [requestId, requestName, upsertWorkspaceRequest],
  )

  useEffect(() => {
    const loadRequestState = () => {
      setActiveRequestId(requestId)
      const existing = getWorkspaceRequest(requestId)
      const request =
        existing ??
        createWorkspaceRequest(requestId, pendingHistoryEntry?.method ?? 'GET')

      if (pendingCollectionRequest) {
        const saved = {
          ...pendingCollectionRequest.state,
          id: requestId,
          name: pendingCollectionRequest.name,
        }
        setMethod(saved.method)
        setUrl(saved.url)
        setDraft(requestStateToDraft(saved))
        upsertWorkspaceRequest(saved)
        clearPendingCollectionRequest()
        return
      }

      if (pendingHistoryEntry) {
        setMethod(pendingHistoryEntry.method)
        setUrl(pendingHistoryEntry.url)
        if (pendingHistoryEntry.result) {
          setRunResult(pendingHistoryEntry.result)
          setExecutionResult(
            requestRunResultToExecutionResult(pendingHistoryEntry.result),
          )
        }
        clearPendingHistoryEntry()
        return
      }

      setMethod(request.method)
      setUrl(request.url)
      setDraft(requestStateToDraft(request))
    }

    if (useWorkspaceStore.persist.hasHydrated()) {
      loadRequestState()
      return undefined
    }

    return useWorkspaceStore.persist.onFinishHydration(() => {
      loadRequestState()
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [requestId])

  useEffect(() => {
    if (!requestRunner.supportsPersistence) return

    requestRunner
      .checkAvailable()
      .then(() => setApiAvailable(true))
      .catch(() => setApiAvailable(false))
  }, [])

  const persistDraft = useCallback(async () => {
    const generation = ++saveGeneration.current
    const document = requestDraftToTraCtlSpec(method, url, draft, requestName)

    if (!requestRunner.supportsPersistence) {
      setIsSaving(true)
      setSaveError(null)
      try {
        const saved = await saveRequestFile({
          id: fileId,
          name: requestName,
          yaml: JSON.stringify(document, null, 2),
        })
        if (generation !== saveGeneration.current) return saved
        setFileId(saved.id)
        return saved
      } catch (error) {
        if (generation !== saveGeneration.current) return null
        const message = error instanceof Error ? error.message : 'Save failed'
        setSaveError(message)
        return null
      } finally {
        if (generation === saveGeneration.current) {
          setIsSaving(false)
        }
      }
    }

    setIsSaving(true)
    setSaveError(null)

    try {
      await requestRunner.checkAvailable()
      const saved = await requestRunner.saveRequest({
        id: fileId,
        name: requestName,
        document,
      })

      if (generation !== saveGeneration.current) return saved

      if (saved) {
        setFileId(saved.id)
      }
      return saved
    } catch (error) {
      if (generation !== saveGeneration.current) return null

      const message =
        error instanceof LocalApiUnavailableError
          ? 'Desktop API unavailable — start traCtl Desktop'
          : error instanceof Error
            ? error.message
            : 'Save failed'
      setSaveError(message)
      return null
    } finally {
      if (generation === saveGeneration.current) {
        setIsSaving(false)
      }
    }
  }, [draft, fileId, method, requestName, url])

  const markDirty = useCallback(() => {
    resetExecution()
    setRunResult(null)
    setRunError(null)
  }, [resetExecution])

  const updateDraft = useCallback(
    (updater: (current: RequestDraft) => RequestDraft) => {
      setDraft((current) => {
        const next = updater(current)
        syncWorkspace(next, method, url)
        return next
      })
      markDirty()
    },
    [markDirty, method, syncWorkspace, url],
  )

  const configTabCounts = useMemo(
    () => ({
      headers: countPopulatedRows(draft.headers),
      assertions: draft.assertions.length,
      extracts: draft.extracts.length,
    }),
    [draft.assertions.length, draft.extracts.length, draft.headers],
  )

  const canRun = url.trim().length > 0 && !isRunning

  const getRequestState = useCallback(
    () => draftToRequestState(requestId, requestName, method, url, draft),
    [draft, method, requestId, requestName, url],
  )

  const recordHistoryResult = useCallback(
    (result: RequestRunResult, document: ReturnType<typeof requestDraftToTraCtlSpec>) => {
      recordRunHistoryEntry({
        requestName,
        method,
        url,
        result,
        sourceType: 'request',
        document,
      })
    },
    [method, requestName, url],
  )

  const runState = useExecutionStore((state) => state.runState)
  const executionResult = useExecutionStore((state) => state.result)
  const runRequest = useCallback(async () => {
    if (!url.trim()) return

    updateActiveRequestTab({
      method,
      title: requestTabTitleFromUrl(url),
    })
    startRun()
    setIsRunning(true)
    setRunError(null)
    setResultsOpen(true)

    try {
      const document = requestDraftToTraCtlSpec(method, url, draft, requestName)
      const resolved = resolveTraCtlSpecEnvironmentVariables(
        document,
        activeEnvironment?.variables ?? {},
      )
      if (resolved.missing.length > 0) {
        console.warn('[tractl:request-run] Missing environment variables', {
          missing: resolved.missing,
          activeEnvironmentName: activeEnvironment?.name ?? null,
        })
        const result = environmentResolutionErrorResult(resolved.missing)
        setRunResult(result)
        setExecutionResult(requestRunResultToExecutionResult(result))
        setRunError(result.error ?? null)
        recordHistoryResult(result, document)
        return
      }

      const resolvedTarget =
        resolved.value.workflows[0]?.steps[0]?.request.target ?? url
      const urlError = validateRequestRunUrl(resolvedTarget)
      if (urlError) {
        setRunError(urlError)
        setRunResult(null)
        return
      }

      let runFileId = fileId

      if (requestRunner.supportsPersistence) {
        await requestRunner.checkAvailable()
        const saved = await requestRunner.saveRequest({
          id: fileId,
          name: requestName,
          document: resolved.value,
        })
        if (!saved?.id) {
          throw new Error('Request must be saved before running')
        }
        runFileId = saved.id
        setFileId(saved.id)
      }
      const result = await requestRunner.runRequest({
        document: resolved.value,
        fileId: runFileId,
        name: requestName,
      })
      setRunResult(result)
      setExecutionResult(requestRunResultToExecutionResult(result))
      if (result.error) {
        setRunError(result.error)
        setExecutionError(result.error)
      }
      recordHistoryResult(result, resolved.value)
    } catch (error) {
      const useApiStub =
        error instanceof LocalApiUnavailableError ||
        error instanceof WasmRuntimeUnavailableError

      if (useApiStub) {
        try {
          const payload = buildRunRequestPayload({
            method,
            url,
            draft,
            environmentId: activeEnvironment?.id ?? null,
          })
          const apiResult = await apiRunRequest(payload)
          const mapped = executionResultToRequestRunResult(apiResult)
          setRunResult(mapped)
          setExecutionResult(apiResult)
          recordHistoryResult(mapped, requestDraftToTraCtlSpec(method, url, draft, requestName))
          return
        } catch (stubError) {
          console.error('[tractl:request-run] API stub run failed', stubError)
        }
      }

      const message =
        error instanceof LocalApiUnavailableError
          ? 'Desktop API unavailable — start traCtl Desktop on port 7428'
          : error instanceof WasmRuntimeUnavailableError
            ? 'WASM runtime is not ready — reload the page'
            : error instanceof Error
              ? error.message
              : 'Run failed'
      console.error('[tractl:request-run] Run failed', error)
      setRunError(message)
      setExecutionError(message)
      setRunResult(null)
    } finally {
      setIsRunning(false)
    }
  }, [
    activeEnvironment,
    draft,
    fileId,
    method,
    recordHistoryResult,
    requestName,
    setExecutionError,
    setExecutionResult,
    startRun,
    updateActiveRequestTab,
    url,
  ])

  const toggleResultsPanel = useCallback(() => {
    setResultsOpen((open) => !open)
  }, [])

  return {
    method,
    url,
    draft,
    fileId,
    requestName,
    activeConfigTab,
    activeResultTab,
    resultsOpen,
    isSaving,
    getRequestState,
    isRunning,
    apiAvailable,
    saveError,
    runError,
    runResult,
    runState,
    executionResult,
    canRun,
    configTabCounts,
    requestId,
    setMethod: (value: HttpMethod) => {
      setMethod(value)
      updateActiveRequestTab({ method: value })
      setDraft((current) => {
        syncWorkspace(current, value, url)
        return current
      })
      markDirty()
    },
    setUrl: (value: string) => {
      setUrl(value)
      setDraft((current) => {
        syncWorkspace(current, method, value)
        return current
      })
      markDirty()
    },
    setActiveConfigTab,
    setActiveResultTab,
    toggleResultsPanel,
    runRequest,
    params: {
      rows: draft.params,
      add: () =>
        updateDraft((current) =>
          patchDraft(current, { params: addKeyValueRow(current.params, 'param') }),
        ),
      update: (id: string, patch: Partial<Omit<KeyValueRow, 'id'>>) =>
        updateDraft((current) =>
          patchDraft(current, {
            params: updateKeyValueRow(current.params, id, patch),
          }),
        ),
      remove: (id: string) =>
        updateDraft((current) =>
          patchDraft(current, {
            params: removeKeyValueRow(current.params, id),
          }),
        ),
    },
    headers: {
      rows: draft.headers,
      replace: (rows: KeyValueRow[]) =>
        updateDraft((current) => patchDraft(current, { headers: rows })),
      add: () =>
        updateDraft((current) =>
          patchDraft(current, {
            headers: addKeyValueRow(current.headers, 'header'),
          }),
        ),
      update: (id: string, patch: Partial<Omit<KeyValueRow, 'id'>>) =>
        updateDraft((current) =>
          patchDraft(current, {
            headers: updateKeyValueRow(current.headers, id, patch),
          }),
        ),
      remove: (id: string) =>
        updateDraft((current) =>
          patchDraft(current, {
            headers: removeKeyValueRow(current.headers, id),
          }),
        ),
    },
    body: {
      value: draft.body,
      update: (patch: Partial<RequestBodyDraft>) =>
        updateDraft((current) =>
          patchDraft(current, {
            body: { ...current.body, ...patch },
            headers: syncContentTypeHeader(current.headers, patch, current.body),
          }),
        ),
    },
    auth: {
      value: draft.auth,
      update: (patch: Partial<RequestAuthDraft>) =>
        updateDraft((current) =>
          patchDraft(current, { auth: { ...current.auth, ...patch } }),
        ),
    },
    scripts: {
      value: draft.scripts,
      update: (patch: Partial<RequestScriptDraft>) =>
        updateDraft((current) =>
          patchDraft(current, { scripts: { ...current.scripts, ...patch } }),
        ),
    },
    assertions: {
      rows: draft.assertions,
      add: () =>
        updateDraft((current) =>
          patchDraft(current, {
            assertions: addAssertionRow(current.assertions),
          }),
        ),
      update: (id: string, patch: Partial<Omit<AssertionRowModel, 'id'>>) =>
        updateDraft((current) =>
          patchDraft(current, {
            assertions: updateAssertionRow(current.assertions, id, patch),
          }),
        ),
      remove: (id: string) =>
        updateDraft((current) =>
          patchDraft(current, {
            assertions: removeAssertionRow(current.assertions, id),
          }),
        ),
    },
    extracts: {
      rows: draft.extracts,
      add: () =>
        updateDraft((current) =>
          patchDraft(current, {
            extracts: addExtractRow(current.extracts),
          }),
        ),
      update: (id: string, patch: Partial<Omit<ExtractRowModel, 'id'>>) =>
        updateDraft((current) =>
          patchDraft(current, {
            extracts: updateExtractRow(current.extracts, id, patch),
          }),
        ),
      remove: (id: string) =>
        updateDraft((current) =>
          patchDraft(current, {
            extracts: removeExtractRow(current.extracts, id),
          }),
        ),
    },
    settings: {
      value: draft.settings,
      update: (patch: Partial<RequestSettingsDraft>) =>
        updateDraft((current) =>
          patchDraft(current, { settings: { ...current.settings, ...patch } }),
        ),
    },
  }
}

export type RequestEditorController = ReturnType<typeof useRequestEditor>
