import { desktopPlatform } from '@/platform/desktop/platform'
import type { PlatformCapabilities, SurfaceKind } from '@/platform/surface'
import { webPlatform } from '@/platform/web/platform'

export function detectSurface(): SurfaceKind {
  return __TRACTL_SURFACE__ === 'desktop' ? 'desktop' : 'web'
}

export function getPlatformCapabilities(): PlatformCapabilities {
  return detectSurface() === 'desktop' ? desktopPlatform : webPlatform
}

export type { PlatformCapabilities, SurfaceKind }
