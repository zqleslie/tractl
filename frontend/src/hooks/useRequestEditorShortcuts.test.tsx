import { render } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { useRequestEditorShortcuts } from '@/hooks/useRequestEditorShortcuts'
import { useUiStore } from '@/stores/uiStore'

function ShortcutHarness({
  editor,
}: {
  editor: Parameters<typeof useRequestEditorShortcuts>[0]
}) {
  useRequestEditorShortcuts(editor)
  return null
}

describe('useRequestEditorShortcuts', () => {
  it('triggers run on meta+Enter', () => {
    const runRequest = vi.fn()
    render(
      <ShortcutHarness
        editor={
          {
            runRequest,
            setMethod: vi.fn(),
            runResult: null,
          } as never
        }
      />,
    )

    window.dispatchEvent(
      new KeyboardEvent('keydown', { key: 'Enter', metaKey: true, bubbles: true }),
    )

    expect(runRequest).toHaveBeenCalled()
  })

  it('toggles layout on meta+shift+l', () => {
    useUiStore.setState({ editorLayout: 'stacked' })
    const runRequest = vi.fn()
    render(
      <ShortcutHarness
        editor={
          {
            runRequest,
            setMethod: vi.fn(),
            runResult: null,
          } as never
        }
      />,
    )

    window.dispatchEvent(
      new KeyboardEvent('keydown', {
        key: 'l',
        metaKey: true,
        shiftKey: true,
        bubbles: true,
      }),
    )

    expect(useUiStore.getState().editorLayout).toBe('side-by-side')
  })
})
