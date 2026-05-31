export const colors = {
  success: { light: { bg: '#EAF3DE', text: '#3B6D11' }, dark: { bg: '#173404', text: '#C0DD97' } },
  error: { light: { bg: '#FCEBEB', text: '#A32D2D' }, dark: { bg: '#501313', text: '#F7C1C1' } },
  warning: { light: { bg: '#FAEEDA', text: '#854F0B' }, dark: { bg: '#412402', text: '#FAC775' } },
  info: { light: { bg: '#E6F1FB', text: '#185FA5' }, dark: { bg: '#042C53', text: '#B5D4F4' } },
  run: { bg: '#185FA5', text: '#E6F1FB' },
} as const

export const methodColors: Record<string, { bg: string; text: string }> = {
  GET: { bg: '#E6F1FB', text: '#185FA5' },
  POST: { bg: '#EAF3DE', text: '#3B6D11' },
  PUT: { bg: '#FAEEDA', text: '#854F0B' },
  PATCH: { bg: '#FAEEDA', text: '#854F0B' },
  DELETE: { bg: '#FCEBEB', text: '#A32D2D' },
}

export const typography = {
  label: 'font-sans text-[12px] font-[400]',
  labelMedium: 'font-sans text-[12px] font-[500]',
  mono: 'font-mono text-[11px]',
} as const

export const border = { default: '0.5px', accent: '2px', card: '1.5px' } as const

export const radius = { input: '7px', button: '7px', card: '10px', pill: '999px' } as const
