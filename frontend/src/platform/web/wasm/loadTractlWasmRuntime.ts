/**
 * UI-0.1B–0.1E: browser WASM bootstrap.
 * Loads /wasm_exec.js + /tractl.wasm; exposes metadata, echo, parse, validate, and run APIs.
 */

type WasmRuntimeState = 'idle' | 'loading' | 'ready' | 'error'

interface WasmRuntimeStatus {
  state: WasmRuntimeState
  error?: string
}

interface TractlWasmVersion {
  surface: string
  runtime: string
  version: string
}

interface TractlWasmCapabilities {
  surface: string
  networkProvider: string
  supports: string[]
  unsupported: string[]
}

export interface TractlWasmError {
  code: string
  message: string
}

/** Successful echo payload from the WASM bridge (UI-0.1C). */
interface TractlWasmEchoSuccess {
  surface: string
  received: true
  length: number
  preview: string
}

/** Structured bridge failure (invalid args, panic recovery, etc.). */
interface TractlWasmEchoFailure {
  error: TractlWasmError
}

type TractlWasmEchoResult = TractlWasmEchoSuccess | TractlWasmEchoFailure


/** Canonical spec document returned by the parser pipeline (UI-0.1D). */
export type TractlWasmParsedSpec = Record<string, unknown>

/** Successful parse payload from the WASM bridge (UI-0.1D). */
export interface TractlWasmParseSuccess {
  surface: string
  parsed: true
  spec: TractlWasmParsedSpec
}

/** Structured parser or bridge failure (UI-0.1D). */
export interface TractlWasmParseFailure {
  error: TractlWasmError
}

export type TractlWasmParseResult = TractlWasmParseSuccess | TractlWasmParseFailure

export type TractlWasmParseFormat = 'yaml' | 'yml' | 'json' | 'toon'

export function isTractlWasmParseSuccess(
  result: TractlWasmParseResult,
): result is TractlWasmParseSuccess {
  return 'parsed' in result && result.parsed === true
}

/** Single validation error from the canonical validation pipeline (UI-0.1E). */
export interface TractlWasmValidationError {
  message: string
  field?: string
  code?: string
}

/** Successful validation payload from the WASM bridge (UI-0.1E). */
export interface TractlWasmValidateSuccess {
  valid: true
  surface: string
}

/** Failed validation payload from the WASM bridge (UI-0.1E). */
export interface TractlWasmValidateFailure {
  valid: false
  surface: string
  errors: TractlWasmValidationError[]
}

/** Bridge-level failure for invalid args or panic recovery (UI-0.1E). */
export interface TractlWasmValidateBridgeFailure {
  error: TractlWasmError
}

export type TractlWasmValidateResult =
  | TractlWasmValidateSuccess
  | TractlWasmValidateFailure
  | TractlWasmValidateBridgeFailure

/**
 * Engine run success — WASM metadata fields present in all successful run responses (UI-0.1F).
 *
 * handleRun returns a flat localapi.RunResult (lowercase keys: passed, statusCode, …).
 * The workflow canvas adapter reads the engine format (Workflows, Steps, PascalCase) via
 * run.Workflows. Both shapes satisfy this type because the base is Record<string, unknown>.
 */
export type TractlWasmRunSuccess = Record<string, unknown> & {
  surface: string
  executionMode: string
  networkProvider: string
}

/** Engine or bridge failure for run (UI-0.1F). */
export interface TractlWasmRunFailure {
  error: TractlWasmError
}

export type TractlWasmRunResult = TractlWasmRunSuccess | TractlWasmRunFailure

export function isTractlWasmRunSuccess(
  result: TractlWasmRunResult,
): result is TractlWasmRunSuccess {
  return !('error' in result && result.error !== undefined)
}

export function isTractlWasmValidateSuccess(
  result: TractlWasmValidateResult,
): result is TractlWasmValidateSuccess {
  return 'valid' in result && result.valid === true
}

export function isTractlWasmValidateFailure(
  result: TractlWasmValidateResult,
): result is TractlWasmValidateFailure {
  return 'valid' in result && result.valid === false && 'errors' in result
}

interface TractlWasmGlobal {
  version: () => TractlWasmVersion
  capabilities: () => TractlWasmCapabilities
  /** Promise-based document echo; no parser/validator/engine calls. */
  echo: (document: string) => Promise<TractlWasmEchoResult>
  /** Promise-based parser pipeline; no SpecValidator or engine calls. */
  parse: (document: string, format: TractlWasmParseFormat) => Promise<TractlWasmParseResult>
  /** Promise-based validation pipeline; no engine execution. */
  validate: (
    document: string,
    format: TractlWasmParseFormat,
  ) => Promise<TractlWasmValidateResult>
  /**
   * Full engine pipeline (parse → validate → plan → execute).
   * Async — runs on a Go goroutine via asyncPromiseHandler to avoid
   * blocking the syscall/js callback thread during net/http fetch.
   */
  run: (document: string, format: TractlWasmParseFormat) => Promise<TractlWasmRunResult>
  /**
   * Full engine pipeline for a single request (ADR-017 §3 amendment 2026-06-02 B).
   * Accepts RequestDef JSON. Go-side RunRequestDef() handles all spec construction.
   * Use this instead of run() for single-request execution from the editor.
   */
  runRequest: (requestDefJson: string) => Promise<TractlWasmRunResult>
}

declare global {
  interface Window {
    tractl?: TractlWasmGlobal
    Go?: new () => GoWasmRuntime
  }
}

interface GoWasmRuntime {
  importObject: WebAssembly.Imports
  run(instance: WebAssembly.Instance): Promise<void>
}

const WASM_EXEC_SCRIPT = '/wasm_exec.js'
const WASM_BINARY = '/tractl.wasm'
const TRACTL_WAIT_MS = 10_000

let runtimeStatus: WasmRuntimeStatus = { state: 'idle' }


export function isTractlWasmRuntimeReady(): boolean {
  return runtimeStatus.state === 'ready' && typeof window.tractl !== 'undefined'
}

let loadPromise: Promise<void> | null = null

/** Idempotent WASM bootstrap for the web surface. */
export function loadTractlWasmRuntime(): Promise<void> {
  if (loadPromise) {
    return loadPromise
  }

  loadPromise = bootstrap()
  return loadPromise
}

async function bootstrap(): Promise<void> {
  runtimeStatus = { state: 'loading' }

  try {
    await loadScript(WASM_EXEC_SCRIPT)

    const GoCtor = window.Go
    if (!GoCtor) {
      throw new Error('Go runtime constructor missing after wasm_exec.js load')
    }

    const go = new GoCtor()
    const instance = await instantiateWasm(go.importObject)
    void go.run(instance)

    await waitForTractlGlobal()
    runtimeStatus = { state: 'ready' }

    if (import.meta.env.DEV) {
      console.info('[tractl:wasm] runtime ready', {
        version: window.tractl?.version(),
        capabilities: window.tractl?.capabilities(),
      })
    }
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err)
    runtimeStatus = { state: 'error', error: message }
    console.error('[tractl:wasm] bootstrap failed:', message)
    throw err
  }
}

function loadScript(src: string): Promise<void> {
  return new Promise((resolve, reject) => {
    const marker = `script[data-tractl-src="${src}"]`
    if (document.querySelector(marker)) {
      resolve()
      return
    }

    const script = document.createElement('script')
    script.src = src
    script.dataset.tractlSrc = src
    script.async = true
    script.onload = () => resolve()
    script.onerror = () => reject(new Error(`failed to load script: ${src}`))
    document.head.appendChild(script)
  })
}

async function instantiateWasm(
  importObject: WebAssembly.Imports,
): Promise<WebAssembly.Instance> {
  if (typeof WebAssembly.instantiateStreaming === 'function') {
    try {
      const result = await WebAssembly.instantiateStreaming(
        fetch(WASM_BINARY),
        importObject,
      )
      return result.instance
    } catch {
      // Vite dev server may not set application/wasm; fall back below.
    }
  }

  const response = await fetch(WASM_BINARY)
  if (!response.ok) {
    throw new Error(`failed to fetch ${WASM_BINARY}: ${response.status}`)
  }

  const bytes = await response.arrayBuffer()
  const result = await WebAssembly.instantiate(bytes, importObject)
  return result.instance
}

function waitForTractlGlobal(): Promise<void> {
  return new Promise((resolve, reject) => {
    const deadline = Date.now() + TRACTL_WAIT_MS

    const poll = () => {
      if (typeof window.tractl !== 'undefined') {
        resolve()
        return
      }
      if (Date.now() >= deadline) {
        reject(new Error('window.tractl was not registered before timeout'))
        return
      }
      requestAnimationFrame(poll)
    }

    poll()
  })
}
