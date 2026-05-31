import type { PlatformCapabilities } from '@/platform/types'

export const desktopPlatform: PlatformCapabilities = {
  kind: 'desktop',
  statusLabel: 'Desktop · v0.1.0-alpha',
  canUseNativeDialogs: true,
  canRevealInFileManager: true,
  canAccessKeychain: true,
  requiresDesktopForExecution: false,
}
