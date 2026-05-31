import { detectSurface } from '@/platform'
import { desktopWorkflowRunRunner } from '@/platform/desktop/workflowRunRunner'
import { webWorkflowRunRunner } from '@/platform/web/workflowRunRunner'
import type { WorkflowRunRunner } from '@/platform/workflowRun/types'

export function getWorkflowRunRunner(): WorkflowRunRunner {
  return detectSurface() === 'desktop'
    ? desktopWorkflowRunRunner
    : webWorkflowRunRunner
}
