/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        canvas: {
          light: '#f8fafc',
          dark: '#090d16',
          cardLight: '#ffffff',
          cardDark: '#0f1524',
          subtleLight: '#f1f5f9',
          subtleDark: '#161c2e',
          borderDark: '#1e293b',
          borderLight: '#e2e8f0',
        },
        brand: {
          orange: '#f97316',
          amber: '#f59e0b',
          emerald: '#10b981',
          rose: '#ef4444',
          blue: '#3b82f6',
          indigo: '#6366f1',
          purple: '#8b5cf6',
          cyan: '#06b6d4',
          slate: '#64748b',
        }
      },
      fontFamily: {
        sans: ['"Plus Jakarta Sans"', 'Inter', 'system-ui', '-apple-system', 'sans-serif'],
        display: ['"Plus Jakarta Sans"', 'Inter', 'system-ui', '-apple-system', 'sans-serif'],
        mono: ['"JetBrains Mono"', 'ui-monospace', 'monospace'],
      },
      boxShadow: {
        'subtle-light': '0 1px 3px 0 rgba(0, 0, 0, 0.05), 0 1px 2px 0 rgba(0, 0, 0, 0.03)',
        'card-light': '0 4px 6px -1px rgba(0, 0, 0, 0.04), 0 2px 4px -1px rgba(0, 0, 0, 0.03)',
        'card-dark': '0 10px 15px -3px rgba(0, 0, 0, 0.4), 0 4px 6px -2px rgba(0, 0, 0, 0.3)',
        'glow-orange': '0 0 25px -5px rgba(249, 115, 22, 0.35)',
        'glow-emerald': '0 0 25px -5px rgba(16, 185, 129, 0.35)',
        'glow-blue': '0 0 25px -5px rgba(59, 130, 246, 0.35)',
      },
      animation: {
        'pulse-subtle': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        'flow': 'flow 1.5s ease-in-out infinite',
      },
      keyframes: {
        flow: {
          '0%, 100%': { transform: 'translateX(0)', opacity: '0.6' },
          '50%': { transform: 'translateX(6px)', opacity: '1' },
        }
      }
    },
  },
  plugins: [],
}
