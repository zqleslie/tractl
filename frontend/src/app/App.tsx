import { useEffect } from 'react'
import { AppShell } from '@/components/shell/AppShell'
import {
  pushRequestRoute,
  syncRequestRouteFromLocation,
} from '@/lib/requestEditor/syncRequestRoute'
import { EnvironmentManager } from '@/screens/EnvironmentManager'
import { FastStartScreen } from '@/screens/FastStartScreen'
import { WorkflowCanvasScreen } from '@/screens/WorkflowCanvas'
import { RequestEditor } from '@/screens/RequestEditor'
import { RunHistoryScreen } from '@/screens/RunHistoryScreen'
import { useUiStore } from '@/stores/uiStore'

export default function App() {
  const activeScreen = useUiStore((s) => s.activeScreen)
  const requestEditorKey = useUiStore((s) => s.requestEditorKey)
  const activeTabId = useUiStore((s) => s.activeTabId)

  useEffect(() => {
    syncRequestRouteFromLocation()
  }, [])

  useEffect(() => {
    if (activeScreen === 'request' && activeTabId) {
      pushRequestRoute(activeTabId)
    }
  }, [activeScreen, activeTabId])

  return (
    <AppShell>
      {activeScreen === 'fast-start' && <FastStartScreen />}
      {activeScreen === 'request' && <RequestEditor key={requestEditorKey} />}
      {activeScreen === 'workflow' && <WorkflowCanvasScreen />}
      {activeScreen === 'run-history' && <RunHistoryScreen />}
      {activeScreen === 'environments' && <EnvironmentManager />}
    </AppShell>
  )
}
