import type { TraCtlSpecDocument } from '@/components/request-editor/tractlSpecDocument'
import {
  isTractlWasmParseSuccess,
  isTractlWasmValidateFailure,
  isTractlWasmValidateSuccess,
  loadTractlWasmRuntime,
  type TractlWasmParseFormat,
  type TractlWasmValidationError,
} from '@/platform/web/wasm/loadTractlWasmRuntime'

export type DocumentValidationIssue = TractlWasmValidationError

export type ParseValidateSuccess = {
  ok: true
  spec: TraCtlSpecDocument
  format: TractlWasmParseFormat
  raw: string
}

export type ParseValidateFailure = {
  ok: false
  stage: 'runtime' | 'parse' | 'validate'
  message: string
  errors: DocumentValidationIssue[]
}

export type ParseValidateResult = ParseValidateSuccess | ParseValidateFailure

function failure(
  stage: ParseValidateFailure['stage'],
  message: string,
  errors: DocumentValidationIssue[] = [],
): ParseValidateFailure {
  return { ok: false, stage, message, errors }
}

export async function parseAndValidateDocument(
  raw: string,
  format: TractlWasmParseFormat,
): Promise<ParseValidateResult> {
  await loadTractlWasmRuntime()

  if (!window.tractl?.parse || !window.tractl?.validate) {
    return failure(
      'runtime',
      'WASM runtime is not ready — reload the page and try again.',
    )
  }

  const parsed = await window.tractl.parse(raw, format)
  if (!isTractlWasmParseSuccess(parsed)) {
    return failure('parse', parsed.error.message, [
      { message: parsed.error.message, code: parsed.error.code },
    ])
  }

  const validate = await window.tractl.validate(raw, format)
  if ('error' in validate && validate.error) {
    return failure('validate', validate.error.message, [
      { message: validate.error.message, code: validate.error.code },
    ])
  }

  if (isTractlWasmValidateFailure(validate)) {
    return failure(
      'validate',
      'Document failed validation',
      validate.errors,
    )
  }

  if (!isTractlWasmValidateSuccess(validate)) {
    return failure('validate', 'Unexpected validation response')
  }

  return {
    ok: true,
    spec: parsed.spec as TraCtlSpecDocument,
    format,
    raw,
  }
}
