import { desktopPlatform } from '@/platform/desktop/platform'
import { getRequestExecutionRunner } from '@/platform/requestExecution/getRequestExecutionRunner'
import type { PlatformCapabilities, SurfaceKind } from '@/platform/types'
import { webPlatform } from '@/platform/web/platform'

export function detectSurface(): SurfaceKind {
  return __TRACTL_SURFACE__ === 'desktop' ? 'desktop' : 'web'
}

export function getPlatformCapabilities(): PlatformCapabilities {
  return detectSurface() === 'desktop' ? desktopPlatform : webPlatform
}

export { getRequestExecutionRunner }
export type { PlatformCapabilities, SurfaceKind }
