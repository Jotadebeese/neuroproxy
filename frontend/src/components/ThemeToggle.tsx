import React from 'react'
import { Sun, Moon } from 'lucide-react'
import { motion } from 'framer-motion'

interface ThemeToggleProps {
  theme: 'dark' | 'light'
  onToggle: () => void
}

export const ThemeToggle: React.FC<ThemeToggleProps> = ({ theme, onToggle }) => {
  const isDark = theme === 'dark'

  return (
    <button
      type="button"
      onClick={onToggle}
      title={isDark ? 'Switch to Light Mode' : 'Switch to Dark Mode'}
      aria-label="Toggle theme"
      className="relative p-2 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-canvas-cardDark text-slate-600 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white hover:bg-slate-100 dark:hover:bg-slate-800/80 transition-all shadow-sm focus:outline-none focus:ring-2 focus:ring-brand-orange/40 cursor-pointer"
    >
      <div className="relative w-4 h-4 flex items-center justify-center">
        <motion.div
          initial={false}
          animate={{
            scale: isDark ? 0 : 1,
            rotate: isDark ? -90 : 0,
            opacity: isDark ? 0 : 1,
          }}
          transition={{ duration: 0.2 }}
          className="absolute"
        >
          <Sun className="w-4 h-4 text-amber-500" />
        </motion.div>

        <motion.div
          initial={false}
          animate={{
            scale: isDark ? 1 : 0,
            rotate: isDark ? 0 : 90,
            opacity: isDark ? 1 : 0,
          }}
          transition={{ duration: 0.2 }}
          className="absolute"
        >
          <Moon className="w-4 h-4 text-indigo-400" />
        </motion.div>
      </div>
    </button>
  )
}
