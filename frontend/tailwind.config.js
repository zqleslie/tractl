/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  darkMode: ['selector', '[data-theme="dark"]'],
  theme: {
    extend: {
      fontFamily: {
        sans: ['var(--font-sans)'],
        mono: ['var(--font-mono)'],
      },
      colors: {
        surface: 'var(--color-surface)',
        'surface-elevated': 'var(--color-surface-elevated)',
        border: 'var(--color-border)',
        text: 'var(--color-text)',
        'text-muted': 'var(--color-text-muted)',
        primary: {
          DEFAULT: 'var(--color-primary)',
          fg: 'var(--color-primary-fg)',
          bg: 'var(--color-primary-bg)',
        },
        success: {
          DEFAULT: 'var(--color-success)',
          fg: 'var(--color-success-fg)',
          bg: 'var(--color-success-bg)',
        },
        danger: {
          DEFAULT: 'var(--color-danger)',
          fg: 'var(--color-danger-fg)',
          bg: 'var(--color-danger-bg)',
        },
        warning: {
          DEFAULT: 'var(--color-warning)',
          fg: 'var(--color-warning-fg)',
          bg: 'var(--color-warning-bg)',
        },
        info: {
          DEFAULT: 'var(--color-info)',
          fg: 'var(--color-info-fg)',
          bg: 'var(--color-info-bg)',
        },
      },
      borderRadius: {
        ui: '7px',
        card: '10px',
      },
      fontSize: {
        'ui-xs': ['var(--font-size-ui-xs)', { lineHeight: 'var(--line-height-ui-xs)' }],
        'ui-sm': ['var(--font-size-ui-sm)', { lineHeight: 'var(--line-height-ui-sm)' }],
        'ui-base': [
          'var(--font-size-ui-base)',
          { lineHeight: 'var(--line-height-ui-base)' },
        ],
      },
    },
  },
  plugins: [],
}
