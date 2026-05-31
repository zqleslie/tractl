import { useState } from 'react'
import {
  IconFolderOpen,
  IconPlus,
  IconTopologyStar3,
} from '@tabler/icons-react'
import { SidebarAction } from '@/components/sidebar/SidebarAction'
import { SidebarCollapsibleGroup } from '@/components/sidebar/SidebarCollapsibleGroup'
import { SidebarListItem } from '@/components/sidebar/SidebarListItem'
import { SidebarSection } from '@/components/sidebar/SidebarSection'
import { createEmptyWorkflow } from '@/lib/workflowCanvas/createEmptyWorkflow'
import { loadWorkflowIntoCanvas } from '@/lib/workflowCanvas/loadWorkflowIntoCanvas'
import { TRACTL_FILE_INPUT_ACCEPT } from '@/lib/tractlDocument/openTraCtlFileFromDisk'
import type { TraCtlFileActionError } from '@/lib/tractlDocument/openTraCtlFileFromDisk'
import { openTraCtlWorkflowFromDisk } from '@/lib/workflowCanvas/openTraCtlWorkflowFromDisk'
import { useWorkflowWorkspaceStore } from '@/stores/workflowWorkspaceStore'
import { useUiStore } from '@/stores/uiStore'

const SIDEBAR_OPEN_FILE_INPUT_ID = 'sidebar-workflows-open-file-input'

export function SidebarWorkflowsPanel() {
  const openWorkflowAction = useUiStore((s) => s.openWorkflow)
  const activeWorkflowId = useUiStore((s) => s.activeWorkflowId)
  const openedWorkflowEntries = useWorkflowWorkspaceStore((s) => s.entries)
  const upsertWorkflow = useWorkflowWorkspaceStore((s) => s.upsertWorkflow)

  const [isOpeningFile, setIsOpeningFile] = useState(false)
  const [fileError, setFileError] = useState<TraCtlFileActionError | null>(null)

  const openWorkflowById = (workflowId: string) => {
    const workflow = useWorkflowWorkspaceStore.getState().getWorkflow(workflowId)
    if (!workflow) return
    openWorkflowAction(workflowId)
    loadWorkflowIntoCanvas(workflow)
  }

  const handleNewWorkflow = () => {
    const workflow = createEmptyWorkflow()
    upsertWorkflow(workflow, {
      sourceName: workflow.name,
      sourceFormat: 'yaml',
    })
    openWorkflowAction(workflow.id)
    loadWorkflowIntoCanvas(workflow)
  }

  const handleSelectedFile = async (file: File) => {
    setFileError(null)
    setIsOpeningFile(true)
    try {
      const result = await openTraCtlWorkflowFromDisk(file)
      if (!result.ok) {
        setFileError(result.error)
      }
    } finally {
      setIsOpeningFile(false)
    }
  }

  return (
    <div data-testid="sidebar-panel-workflows">
      <SidebarAction primary testId="sidebar-new-workflow" onClick={handleNewWorkflow}>
        <IconPlus size={14} stroke={1.75} />
        New workflow
      </SidebarAction>
      <SidebarAction
        testId="sidebar-open-file"
        htmlFor={SIDEBAR_OPEN_FILE_INPUT_ID}
        disabled={isOpeningFile}
      >
        <IconFolderOpen size={14} stroke={1.75} />
        {isOpeningFile ? 'Opening file…' : 'Open file'}
      </SidebarAction>
      <input
        id={SIDEBAR_OPEN_FILE_INPUT_ID}
        type="file"
        accept={TRACTL_FILE_INPUT_ACCEPT}
        className="sr-only"
        tabIndex={-1}
        data-testid="sidebar-open-file-input"
        onChange={(event) => {
          const file = event.currentTarget.files?.[0]
          event.currentTarget.value = ''
          if (file) {
            void handleSelectedFile(file)
          }
        }}
      />
      {fileError ? (
        <p
          className="mb-2 px-2 text-ui-xs text-danger-fg"
          data-testid="sidebar-open-file-error"
          role="alert"
        >
          {fileError.title}: {fileError.message}
        </p>
      ) : null}

      {openedWorkflowEntries.length > 0 ? (
        <>
          <SidebarSection>Open</SidebarSection>
          {openedWorkflowEntries.map(({ workflow }) => (
            <SidebarListItem
              key={workflow.id}
              testId={`sidebar-workflow-${workflow.id}`}
              active={activeWorkflowId === workflow.id}
              icon={<IconTopologyStar3 size={13} stroke={1.75} />}
              label={workflow.name}
              onClick={() => openWorkflowById(workflow.id)}
            />
          ))}
        </>
      ) : (
        <p className="px-2 py-3 text-ui-xs text-text-muted">
          Open a traCtl workflow file to display it on the canvas.
        </p>
      )}

      <SidebarCollapsibleGroup title="OpenAPI" disabled suffix="v1" />
      <SidebarCollapsibleGroup title="WSDL" disabled suffix="v1" />
      <SidebarCollapsibleGroup title="GraphQL SDL" disabled suffix="v1" />
    </div>
  )
}
