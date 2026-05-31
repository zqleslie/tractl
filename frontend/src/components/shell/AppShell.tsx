import type { ReactNode } from 'react'
import { ActivityBar } from '@/components/shell/ActivityBar'
import { BottomToolbar } from '@/components/shell/BottomToolbar'
import { CommandPalette } from '@/components/shell/CommandPalette'
import { KeyboardShortcuts } from '@/components/shell/KeyboardShortcuts'
import { SidePanel } from '@/components/shell/SidePanel'
import { StatusBar } from '@/components/shell/StatusBar'
import { TabBar } from '@/components/shell/TabBar'
import { Topbar } from '@/components/shell/Topbar'
import { getPlatformCapabilities } from '@/platform'
import { WebDesktopBanner } from '@/platform/web/WebDesktopRequired'

type AppShellProps = {
  children: ReactNode
}

export function AppShell({ children }: AppShellProps) {
  const platform = getPlatformCapabilities()

  return (
    <div
      className="flex h-full min-h-screen flex-col bg-surface text-text"
      data-testid="app-shell"
    >
      {platform.requiresDesktopForExecution ? <WebDesktopBanner /> : null}
      <div className="flex min-h-0 flex-1">
        <ActivityBar />
        <SidePanel />
        <div className="flex min-w-0 flex-1 flex-col">
          <Topbar />
          <TabBar />
          <main className="min-h-0 flex-1 overflow-auto">{children}</main>
          <BottomToolbar />
          <StatusBar />
        </div>
      </div>
      <CommandPalette />
      <KeyboardShortcuts />
    </div>
  )
}
