import { useCallback, useEffect, useMemo, useState } from 'react'
import { createEmptyRequestFormState } from '@/components/request-editor/createEmptyRequestFormState'
import {
  addAssertionRow,
  addExtractRow,
  addKeyValueRow,
  countPopulatedRows,
  removeAssertionRow,
  removeExtractRow,
  removeKeyValueRow,
  updateAssertionRow,
  updateExtractRow,
  updateKeyValueRow,
} from '@/components/request-editor/requestFormStateMutations'
import type {
  AssertionRowModel,
  ExtractRowModel,
  KeyValueRow,
  RequestAuthDraft,
  RequestBodyDraft,
  RequestFormState,
  RequestScriptDraft,
} from '@/components/request-editor/types'
import type { HttpMethod } from '@/components/primitives'
import { useWorkflowCanvasStore } from '@/stores/workflowCanvasStore'
import type { WorkflowStep } from '@/types/workflow'

export interface StepDraftEditorSlice {
  stepId: string
  setStepId: (id: string) => void
  commitStepId: () => void

  method: HttpMethod
  url: string
  setMethod: (method: HttpMethod) => void
  setUrl: (url: string) => void

  params: {
    rows: KeyValueRow[]
    add: () => void
    update: (id: string, patch: Partial<Omit<KeyValueRow, 'id'>>) => void
    remove: (id: string) => void
  }

  headers: {
    rows: KeyValueRow[]
    add: () => void
    update: (id: string, patch: Partial<Omit<KeyValueRow, 'id'>>) => void
    remove: (id: string) => void
  }

  body: {
    value: RequestBodyDraft
    update: (patch: Partial<RequestBodyDraft>) => void
  }

  auth: {
    value: RequestAuthDraft
    update: (patch: Partial<RequestAuthDraft>) => void
  }

  scripts: {
    value: RequestScriptDraft
    update: (patch: Partial<RequestScriptDraft>) => void
  }

  assertions: {
    rows: AssertionRowModel[]
    add: () => void
    update: (id: string, patch: Partial<Omit<AssertionRowModel, 'id'>>) => void
    remove: (id: string) => void
  }

  extracts: {
    rows: ExtractRowModel[]
    add: () => void
    update: (id: string, patch: Partial<Omit<ExtractRowModel, 'id'>>) => void
    remove: (id: string) => void
  }

  tabCounts: {
    params: number
    headers: number
    assertions: number
    extracts: number
  }
}

function createInitialStepDraft(step: WorkflowStep): RequestFormState {
  const draft = createEmptyRequestFormState()

  return {
    ...draft,
    auth: {
      ...draft.auth,
      type: step.hasAuth ? 'Bearer token' : 'None',
    },
  }
}

function draftToStepPatch(draft: RequestFormState): Partial<WorkflowStep> {
  return {
    hasAuth: draft.auth.type !== 'None',
    hasPreScript: draft.scripts.pre.trim().length > 0,
    assertionCount: draft.assertions.length,
    extractCount: draft.extracts.length,
  }
}

export function useStepDraftEditor(step: WorkflowStep): StepDraftEditorSlice {
  const updateStep = useWorkflowCanvasStore((s) => s.updateStep)
  const renameStep = useWorkflowCanvasStore((s) => s.renameStep)

  const [stepId, setStepId] = useState(step.id)
  const [method, setMethodState] = useState<HttpMethod>(step.method)
  const [url, setUrlState] = useState<string>(step.url)
  const [draft, setDraft] = useState<RequestFormState>(() => createInitialStepDraft(step))

  useEffect(() => {
    setStepId(step.id)
    setMethodState(step.method)
    setUrlState(step.url)
    setDraft(createInitialStepDraft(step))
  }, [step.id, step.method, step.url, step.hasAuth])

  const persistDraft = useCallback(
    (nextDraft: RequestFormState) => {
      updateStep(step.id, draftToStepPatch(nextDraft))
    },
    [step.id, updateStep],
  )

  const updateDraft = useCallback(
    (updater: (current: RequestFormState) => RequestFormState) => {
      setDraft((current) => {
        const next = updater(current)
        persistDraft(next)
        return next
      })
    },
    [persistDraft],
  )

  const setMethod = useCallback(
    (nextMethod: HttpMethod) => {
      setMethodState(nextMethod)
      updateStep(step.id, { method: nextMethod })
    },
    [step.id, updateStep],
  )

  const setUrl = useCallback(
    (nextUrl: string) => {
      setUrlState(nextUrl)
      updateStep(step.id, { url: nextUrl })
    },
    [step.id, updateStep],
  )

  const commitStepId = useCallback(() => {
    const trimmed = stepId.trim()
    if (!trimmed || trimmed === step.id) {
      setStepId(step.id)
      return
    }
    renameStep(step.id, trimmed)
  }, [renameStep, step.id, stepId])

  const tabCounts = useMemo(
    () => ({
      params: countPopulatedRows(draft.params),
      headers: countPopulatedRows(draft.headers),
      assertions: draft.assertions.length,
      extracts: draft.extracts.length,
    }),
    [draft.params, draft.headers, draft.assertions.length, draft.extracts.length],
  )

  return {
    stepId,
    setStepId,
    commitStepId,
    method,
    url,
    setMethod,
    setUrl,
    params: {
      rows: draft.params,
      add: () =>
        updateDraft((current) => ({
          ...current,
          params: addKeyValueRow(current.params, 'param'),
        })),
      update: (id, patch) =>
        updateDraft((current) => ({
          ...current,
          params: updateKeyValueRow(current.params, id, patch),
        })),
      remove: (id) =>
        updateDraft((current) => ({
          ...current,
          params: removeKeyValueRow(current.params, id),
        })),
    },
    headers: {
      rows: draft.headers,
      add: () =>
        updateDraft((current) => ({
          ...current,
          headers: addKeyValueRow(current.headers, 'header'),
        })),
      update: (id, patch) =>
        updateDraft((current) => ({
          ...current,
          headers: updateKeyValueRow(current.headers, id, patch),
        })),
      remove: (id) =>
        updateDraft((current) => ({
          ...current,
          headers: removeKeyValueRow(current.headers, id),
        })),
    },
    body: {
      value: draft.body,
      update: (patch) =>
        updateDraft((current) => ({
          ...current,
          body: { ...current.body, ...patch },
        })),
    },
    auth: {
      value: draft.auth,
      update: (patch) =>
        updateDraft((current) => ({
          ...current,
          auth: { ...current.auth, ...patch },
        })),
    },
    scripts: {
      value: draft.scripts,
      update: (patch) =>
        updateDraft((current) => ({
          ...current,
          scripts: { ...current.scripts, ...patch },
        })),
    },
    assertions: {
      rows: draft.assertions,
      add: () =>
        updateDraft((current) => ({
          ...current,
          assertions: addAssertionRow(current.assertions),
        })),
      update: (id, patch) =>
        updateDraft((current) => ({
          ...current,
          assertions: updateAssertionRow(current.assertions, id, patch),
        })),
      remove: (id) =>
        updateDraft((current) => ({
          ...current,
          assertions: removeAssertionRow(current.assertions, id),
        })),
    },
    extracts: {
      rows: draft.extracts,
      add: () =>
        updateDraft((current) => ({
          ...current,
          extracts: addExtractRow(current.extracts),
        })),
      update: (id, patch) =>
        updateDraft((current) => ({
          ...current,
          extracts: updateExtractRow(current.extracts, id, patch),
        })),
      remove: (id) =>
        updateDraft((current) => ({
          ...current,
          extracts: removeExtractRow(current.extracts, id),
        })),
    },
    tabCounts,
  }
}
