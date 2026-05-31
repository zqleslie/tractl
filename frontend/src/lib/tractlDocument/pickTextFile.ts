const DEFAULT_ACCEPT = '.yaml,.yml,.json,.toon,application/json,text/yaml,text/x-yaml'

type FilePickerWindow = Window & {
  showOpenFilePicker?: (options?: {
    multiple?: boolean
    types?: Array<{
      description?: string
      accept: Record<string, string[]>
    }>
  }) => Promise<Array<{ getFile: () => Promise<File> }>>
}

async function pickWithNativeFilePicker(): Promise<File | null> {
  const picker = (window as FilePickerWindow).showOpenFilePicker
  if (!picker) return null

  const [handle] = await picker({
    multiple: false,
    types: [
      {
        description: 'traCtl documents',
        accept: {
          'application/json': ['.json'],
          'text/yaml': ['.yaml', '.yml'],
          'text/plain': ['.toon'],
        },
      },
    ],
  })

  return handle ? handle.getFile() : null
}

export function pickTextFile(accept = DEFAULT_ACCEPT): Promise<File | null> {
  const nativePicker = (window as FilePickerWindow).showOpenFilePicker
  if (nativePicker) {
    return pickWithNativeFilePicker().catch((error) => {
      if (error instanceof DOMException && error.name === 'AbortError') {
        return null
      }
      throw error
    })
  }

  return new Promise((resolve) => {
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = accept
    input.style.position = 'fixed'
    input.style.left = '-9999px'
    input.style.top = '0'
    input.style.width = '1px'
    input.style.height = '1px'
    input.style.opacity = '0'

    const cleanup = () => {
      input.remove()
    }

    input.addEventListener('change', () => {
      const file = input.files?.[0] ?? null
      cleanup()
      resolve(file)
    })

    input.addEventListener('cancel', () => {
      cleanup()
      resolve(null)
    })

    document.body.appendChild(input)
    input.click()
  })
}

export async function readFileAsText(file: File): Promise<string> {
  return file.text()
}
