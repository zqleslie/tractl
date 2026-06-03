import {
  addAssertionRow,
  addExtractRow,
  addKeyValueRow,
  removeAssertionRow,
  removeExtractRow,
  removeKeyValueRow,
  patchDraft,
  updateAssertionRow,
  updateExtractRow,
  updateKeyValueRow,
} from '@/components/request-editor/requestFormStateMutations'
import { syncContentTypeHeader } from '@/lib/requestEditor/contentTypeHeader'
import type {
  AssertionRowModel,
  ExtractRowModel,
  KeyValueRow,
  RequestAuthDraft,
  RequestBodyDraft,
  RequestFormState,
  RequestScriptDraft,
  RequestSettingsDraft,
} from '@/components/request-editor/types'

type UpdateDraft = (updater: (current: RequestFormState) => RequestFormState) => void

export function buildFieldControllers(draft: RequestFormState, updateDraft: UpdateDraft) {
  return {
    params: {
      rows: draft.params,
      add: () =>
        updateDraft((current) =>
          patchDraft(current, { params: addKeyValueRow(current.params, 'param') }),
        ),
      update: (id: string, patch: Partial<Omit<KeyValueRow, 'id'>>) =>
        updateDraft((current) =>
          patchDraft(current, { params: updateKeyValueRow(current.params, id, patch) }),
        ),
      remove: (id: string) =>
        updateDraft((current) =>
          patchDraft(current, { params: removeKeyValueRow(current.params, id) }),
        ),
    },
    headers: {
      rows: draft.headers,
      replace: (rows: KeyValueRow[]) =>
        updateDraft((current) => patchDraft(current, { headers: rows })),
      add: () =>
        updateDraft((current) =>
          patchDraft(current, { headers: addKeyValueRow(current.headers, 'header') }),
        ),
      update: (id: string, patch: Partial<Omit<KeyValueRow, 'id'>>) =>
        updateDraft((current) =>
          patchDraft(current, { headers: updateKeyValueRow(current.headers, id, patch) }),
        ),
      remove: (id: string) =>
        updateDraft((current) =>
          patchDraft(current, { headers: removeKeyValueRow(current.headers, id) }),
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
          patchDraft(current, { assertions: addAssertionRow(current.assertions) }),
        ),
      update: (id: string, patch: Partial<Omit<AssertionRowModel, 'id'>>) =>
        updateDraft((current) =>
          patchDraft(current, { assertions: updateAssertionRow(current.assertions, id, patch) }),
        ),
      remove: (id: string) =>
        updateDraft((current) =>
          patchDraft(current, { assertions: removeAssertionRow(current.assertions, id) }),
        ),
    },
    extracts: {
      rows: draft.extracts,
      add: () =>
        updateDraft((current) =>
          patchDraft(current, { extracts: addExtractRow(current.extracts) }),
        ),
      update: (id: string, patch: Partial<Omit<ExtractRowModel, 'id'>>) =>
        updateDraft((current) =>
          patchDraft(current, { extracts: updateExtractRow(current.extracts, id, patch) }),
        ),
      remove: (id: string) =>
        updateDraft((current) =>
          patchDraft(current, { extracts: removeExtractRow(current.extracts, id) }),
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
