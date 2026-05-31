import type { ReactNode } from 'react'

export function SidebarSection({ children }: { children: ReactNode }) {
  return (
    <div className="px-3 pb-1 pt-3 text-[10px] font-semibold uppercase tracking-[0.08em] text-text-muted">
      {children}
    </div>
  )
}
