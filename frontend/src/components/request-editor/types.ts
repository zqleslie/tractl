export type ConfigTab =
  | 'params'
  | 'headers'
  | 'body'
  | 'auth'
  | 'pre-script'
  | 'post-script'
  | 'assertions'
  | 'extracts'
  | 'settings'

export type ResultTab = 'body' | 'headers' | 'assertions' | 'extracts'

export type BodyEncoding =
  | 'JSON'
  | 'Form data'
  | 'Multipart'
  | 'Raw'
  | 'Binary'
  | 'None'

export type AuthType = 'None' | 'Bearer token' | 'Basic auth' | 'API key'

export type RetryStrategy = 'None' | 'Fixed' | 'Linear' | 'Exponential'

export type FailurePolicy = 'resilient' | 'failFast'

export type ConfigTabDefinition = {
  id: ConfigTab
  label: string
  count?: number
  script?: boolean
}

export type ResultTabDefinition = {
  id: ResultTab
  label: string
}

export type KVRow = {
  id: string
  enabled: boolean
  key: string
  value: string
}

export type KeyValueRow = KVRow

export type FormRow = {
  id: string
  enabled: boolean
  key: string
  value: string
  type: 'text' | 'file'
  filename?: string
}

export type BodyEncodingId =
  | 'json'
  | 'form'
  | 'multipart'
  | 'raw'
  | 'binary'
  | 'none'

export type RequestBodyState = {
  encoding: BodyEncodingId
  content: string
  formRows?: FormRow[]
  rawContentType?: string
  binaryFile?: string
}

export type RequestAuthState = {
  type: 'none' | 'bearer' | 'basic' | 'apikey'
  token?: string
  username?: string
  password?: string
  keyName?: string
  keyValue?: string
  placement?: 'header' | 'query'
}

export type AssertionDef = {
  id: string
  kind: string
  op: string
  expected: string
  severity: 'error' | 'warning'
}

export type ExtractDef = {
  id: string
  source: string
  path: string
  variableName: string
  scope: string
}

export type RetryConfig = {
  strategy: 'fixed' | 'linear' | 'exponential'
  maxAttempts: number
  delayMs: number
  backoffFactor?: number
}

export type AssertionSeverity = 'error' | 'warning'

export type AssertionRowModel = {
  id: string
  kind: string
  operator: string
  expected: string
  severity: AssertionSeverity
}

export type ExtractRowModel = {
  id: string
  source: string
  path: string
  variable: string
  scope: string
}

export type RequestBodyDraft = {
  encoding: BodyEncoding
  value: string
  formRows: FormRow[]
  rawContentType: string
  binaryFile: string
}

export type RequestAuthDraft = {
  type: AuthType
  tokenRef: string
  token: string
  username: string
  password: string
  keyName: string
  keyValue: string
  placement: 'header' | 'query'
}

export type RequestScriptDraft = {
  pre: string
  post: string
}

export type RequestSettingsDraft = {
  timeoutValue: string
  timeoutUnit: 'milliseconds' | 'seconds' | 'minutes'
  retry: RetryStrategy
  retryConfig: RetryConfig | null
  failurePolicy: FailurePolicy
}

export type RequestDraft = {
  params: KeyValueRow[]
  headers: KeyValueRow[]
  body: RequestBodyDraft
  auth: RequestAuthDraft
  scripts: RequestScriptDraft
  assertions: AssertionRowModel[]
  extracts: ExtractRowModel[]
  settings: RequestSettingsDraft
}
