export const colors = {
  success: {
    light: { bg: 'var(--color-success-bg)', text: 'var(--color-success-fg)' },
    dark: { bg: 'var(--color-success-bg)', text: 'var(--color-success-fg)' },
  },
  error: {
    light: { bg: 'var(--color-danger-bg)', text: 'var(--color-danger-fg)' },
    dark: { bg: 'var(--color-danger-bg)', text: 'var(--color-danger-fg)' },
  },
  warning: {
    light: { bg: 'var(--color-warning-bg)', text: 'var(--color-warning-fg)' },
    dark: { bg: 'var(--color-warning-bg)', text: 'var(--color-warning-fg)' },
  },
  info: {
    light: { bg: 'var(--color-info-bg)', text: 'var(--color-info-fg)' },
    dark: { bg: 'var(--color-info-bg)', text: 'var(--color-info-fg)' },
  },
  run: { bg: 'var(--color-primary)', text: 'var(--color-primary-fg)' },
} as const

export const methodColors: Record<string, { bg: string; text: string }> = {
  GET: { bg: 'var(--method-get-bg)', text: 'var(--method-get-fg)' },
  POST: { bg: 'var(--method-post-bg)', text: 'var(--method-post-fg)' },
  PUT: { bg: 'var(--method-put-bg)', text: 'var(--method-put-fg)' },
  PATCH: { bg: 'var(--method-patch-bg)', text: 'var(--method-patch-fg)' },
  DELETE: { bg: 'var(--method-delete-bg)', text: 'var(--method-delete-fg)' },
}

export const typography = {
  label: 'font-sans text-[12px] font-[400]',
  labelMedium: 'font-sans text-[12px] font-[500]',
  mono: 'font-mono text-[11px]',
} as const

export const border = { default: '0.5px', accent: '2px', card: '1.5px' } as const

export const radius = { input: '7px', button: '7px', card: '10px', pill: '999px' } as const
