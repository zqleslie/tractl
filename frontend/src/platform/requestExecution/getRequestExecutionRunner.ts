import { detectSurface } from '@/platform'
import { desktopRequestExecutionRunner } from '@/platform/desktop/requestExecutionRunner'
import type { RequestExecutionRunner } from '@/platform/requestExecution/types'
import { webRequestExecutionRunner } from '@/platform/web/requestExecutionRunner'

export function getRequestExecutionRunner(): RequestExecutionRunner {
  return detectSurface() === 'desktop'
    ? desktopRequestExecutionRunner
    : webRequestExecutionRunner
}
