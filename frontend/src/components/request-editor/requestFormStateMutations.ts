import type {
  AssertionRowModel,
  ExtractRowModel,
  KeyValueRow,
  RequestFormState,
} from '@/components/request-editor/types'

let rowCounter = 0

function createRowId(prefix: string): string {
  rowCounter += 1
  return `${prefix}-${rowCounter}`
}

export function updateKeyValueRow(
  rows: KeyValueRow[],
  id: string,
  patch: Partial<Omit<KeyValueRow, 'id'>>,
): KeyValueRow[] {
  return rows.map((row) => (row.id === id ? { ...row, ...patch } : row))
}

export function addKeyValueRow(rows: KeyValueRow[], prefix: string): KeyValueRow[] {
  return [
    ...rows,
    { id: createRowId(prefix), enabled: true, key: '', value: '' },
  ]
}

export function removeKeyValueRow(rows: KeyValueRow[], id: string): KeyValueRow[] {
  return rows.filter((row) => row.id !== id)
}

export function updateAssertionRow(
  rows: AssertionRowModel[],
  id: string,
  patch: Partial<Omit<AssertionRowModel, 'id'>>,
): AssertionRowModel[] {
  return rows.map((row) => (row.id === id ? { ...row, ...patch } : row))
}

export function addAssertionRow(rows: AssertionRowModel[]): AssertionRowModel[] {
  return [
    ...rows,
    {
      id: createRowId('assertion'),
      kind: 'status',
      operator: 'equals',
      expected: '200',
      severity: 'error',
    },
  ]
}

export function removeAssertionRow(
  rows: AssertionRowModel[],
  id: string,
): AssertionRowModel[] {
  return rows.filter((row) => row.id !== id)
}

export function updateExtractRow(
  rows: ExtractRowModel[],
  id: string,
  patch: Partial<Omit<ExtractRowModel, 'id'>>,
): ExtractRowModel[] {
  return rows.map((row) => (row.id === id ? { ...row, ...patch } : row))
}

export function addExtractRow(rows: ExtractRowModel[]): ExtractRowModel[] {
  return [
    ...rows,
    {
      id: createRowId('extract'),
      source: 'body',
      path: '$.id',
      variable: '',
      scope: 'workflow',
    },
  ]
}

export function removeExtractRow(rows: ExtractRowModel[], id: string): ExtractRowModel[] {
  return rows.filter((row) => row.id !== id)
}

export function countPopulatedRows(rows: KeyValueRow[]): number {
  return rows.filter((row) => row.key.trim().length > 0).length
}

export function patchDraft(
  draft: RequestFormState,
  patch: Partial<RequestFormState>,
): RequestFormState {
  return { ...draft, ...patch }
}
