import { useEffect, type RefObject } from 'react'

export function useClickOutside(
  ref: RefObject<HTMLElement | null>,
  onClose: () => void,
  enabled = true,
): void {
  useEffect(() => {
    if (!enabled) return undefined

    const handleMouseDown = (event: MouseEvent) => {
      if (!ref.current?.contains(event.target as Node)) {
        onClose()
      }
    }
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose()
    }

    document.addEventListener('mousedown', handleMouseDown)
    document.addEventListener('keydown', handleKeyDown)
    return () => {
      document.removeEventListener('mousedown', handleMouseDown)
      document.removeEventListener('keydown', handleKeyDown)
    }
  }, [ref, onClose, enabled])
}
