import { useCallback, useRef } from 'react'
import type { EditorLayout } from '@/stores/uiStore'

export type EditorDragHandleProps = {
  layout: EditorLayout
  onResize: (ratio: number) => void
}

export function EditorDragHandle({ layout, onResize }: EditorDragHandleProps) {
  const dragging = useRef(false)

  const onMouseDown = useCallback(
    (event: React.MouseEvent) => {
      event.preventDefault()
      dragging.current = true
      const container = (event.currentTarget as HTMLElement).parentElement
      if (!container) return

      const onMove = (moveEvent: MouseEvent) => {
        if (!dragging.current) return
        const rect = container.getBoundingClientRect()
        const ratio =
          layout === 'stacked'
            ? (moveEvent.clientY - rect.top) / rect.height
            : (moveEvent.clientX - rect.left) / rect.width
        onResize(Math.min(0.75, Math.max(0.25, ratio)))
      }

      const onUp = () => {
        dragging.current = false
        window.removeEventListener('mousemove', onMove)
        window.removeEventListener('mouseup', onUp)
      }

      window.addEventListener('mousemove', onMove)
      window.addEventListener('mouseup', onUp)
    },
    [layout, onResize],
  )

  return (
    <div
      role="separator"
      aria-orientation={layout === 'stacked' ? 'horizontal' : 'vertical'}
      data-testid="editor-drag-handle"
      className={
        layout === 'stacked'
          ? 'h-1 shrink-0 cursor-row-resize bg-border'
          : 'w-1 shrink-0 cursor-col-resize bg-border'
      }
      onMouseDown={onMouseDown}
    />
  )
}
