import { useCallback, useEffect, useMemo, useState } from 'react'
import { createEmptyRequestFormState } from '@/components/request-editor/createEmptyRequestFormState'
import { countPopulatedRows } from '@/components/request-editor/requestFormStateMutations'
import { buildFieldControllers } from '@/components/request-editor/useRequestEditorFields'
import type { RequestFormState } from '@/components/request-editor/types'
import type { HttpMethod } from '@/components/primitives'
import { useWorkflowCanvasStore } from '@/stores/workflowCanvasStore'
import type { WorkflowStep } from '@/types/workflow'

export type StepDraftEditorSlice = {
  stepId: string
  setStepId: (id: string) => void
  commitStepId: () => void
  method: HttpMethod
  url: string
  setMethod: (method: HttpMethod) => void
  setUrl: (url: string) => void
  tabCounts: { params: number; headers: number; assertions: number; extracts: number }
} & ReturnType<typeof buildFieldControllers>

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
    tabCounts,
    ...buildFieldControllers(draft, updateDraft),
  }
}
