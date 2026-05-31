import type { PlatformCapabilities } from '@/platform/types'

export const webPlatform: PlatformCapabilities = {
  kind: 'web',
  statusLabel: 'Web · v0.1.0-alpha',
  canUseNativeDialogs: false,
  canRevealInFileManager: false,
  canAccessKeychain: false,
  requiresDesktopForExecution: false,
}
