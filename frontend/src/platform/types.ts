export type SurfaceKind = 'web' | 'desktop'

export type PlatformCapabilities = {
  kind: SurfaceKind
  statusLabel: string
  canUseNativeDialogs: boolean
  canRevealInFileManager: boolean
  canAccessKeychain: boolean
  requiresDesktopForExecution: boolean
}
