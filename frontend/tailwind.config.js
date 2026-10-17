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
          light: '#f5f5f7',
          dark: '#0c0d0e',
          cardLight: '#ffffff',
          cardDark: '#141618',
          subtleDark: '#1c1f24',
          borderDark: '#2a2f38',
          borderLight: '#e2e8f0',
        },
        brand: {
          orange: '#f97316',
          amber: '#fbbf24',
          emerald: '#10b981',
          rose: '#f43f5e',
          blue: '#3b82f6',
          purple: '#8b5cf6',
          cyan: '#06b6d4',
          slate: '#64748b',
        }
      },
      fontFamily: {
        sans: ['"Inter"', 'system-ui', '-apple-system', 'sans-serif'],
        mono: ['"JetBrains Mono"', 'monospace'],
      },
      boxShadow: {
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
