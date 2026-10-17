import React, { useState } from 'react'
import { Zap, Download, RefreshCw, Layers, Sparkles, BookOpen, Terminal, Check } from 'lucide-react'
import confetti from 'canvas-confetti'
import { ThemeToggle } from './ThemeToggle'

interface HeaderProps {
  currentView: 'story' | 'console'
  onViewChange: (view: 'story' | 'console') => void
  theme: 'dark' | 'light'
  onToggleTheme: () => void
  onReset: () => void
  onBurstBenchmark: () => void
  onOpenContractModal: () => void
  isBenchmarking: boolean
}

export const Header: React.FC<HeaderProps> = ({
  currentView,
  onViewChange,
  theme,
  onToggleTheme,
  onReset,
  onBurstBenchmark,
  onOpenContractModal,
  isBenchmarking,
}) => {
  const [resetting, setResetting] = useState(false)

  const handleBurst = () => {
    confetti({
      particleCount: 50,
      spread: 60,
      origin: { y: 0.15 },
      colors: ['#f97316', '#fbbf24', '#10b981'],
    })
    onBurstBenchmark()
  }

  const handleReset = async () => {
    if (confirm('Reset gateway state? This will clear cache rules and recent incidents for fresh demonstration runs.')) {
      setResetting(true)
      await onReset()
      setResetting(false)
    }
  }

  return (
    <header className="border-b border-slate-200/80 dark:border-slate-800/80 bg-white/90 dark:bg-canvas-cardDark/90 backdrop-blur-md sticky top-0 z-40 transition-colors">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-3.5 flex flex-wrap items-center justify-between gap-4">
        {/* Brand & Identity */}
        <div className="flex items-center gap-3.5">
          <div className="relative group cursor-pointer" onClick={() => onViewChange('story')}>
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-brand-orange to-amber-600 flex items-center justify-center font-bold text-base text-white shadow-sm transition-transform duration-300 group-hover:scale-105">
              NP
            </div>
            <span className="absolute -bottom-0.5 -right-0.5 flex h-3 w-3">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
              <span className="relative inline-flex rounded-full h-3 w-3 bg-emerald-500"></span>
            </span>
          </div>

          <div>
            <div className="flex items-center gap-2 flex-wrap">
              <span className="text-lg font-bold tracking-tight text-slate-900 dark:text-white flex items-center gap-1.5">
                NeuroProxy
              </span>
              <span className="text-[11px] px-2.5 py-0.5 rounded-full font-semibold bg-brand-orange/10 text-brand-orange border border-brand-orange/20">
                Cloud Run Edge • v2.0
              </span>
              <span className="text-[11px] px-2.5 py-0.5 rounded-full font-semibold bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20 flex items-center gap-1">
                <Sparkles className="w-3 h-3" />
                Gemini 2.5 Flash
              </span>
            </div>
            <p className="text-xs text-slate-500 dark:text-slate-400 font-normal">
              Autonomous Self-Healing API Gateway • Value Conservation Guardrails
            </p>
          </div>
        </div>

        {/* View Switcher Tabs (Story vs Console) */}
        <div className="flex items-center p-1 rounded-xl bg-slate-100 dark:bg-canvas-subtleDark border border-slate-200 dark:border-slate-800">
          <button
            onClick={() => onViewChange('story')}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold transition-all cursor-pointer ${
              currentView === 'story'
                ? 'bg-white dark:bg-canvas-cardDark text-slate-900 dark:text-white shadow-sm'
                : 'text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-white'
            }`}
          >
            <BookOpen className="w-3.5 h-3.5 text-brand-orange" />
            <span>Interactive Story</span>
          </button>

          <button
            onClick={() => onViewChange('console')}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold transition-all cursor-pointer ${
              currentView === 'console'
                ? 'bg-white dark:bg-canvas-cardDark text-slate-900 dark:text-white shadow-sm'
                : 'text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-white'
            }`}
          >
            <Terminal className="w-3.5 h-3.5 text-blue-500" />
            <span>Mission Control</span>
          </button>
        </div>

        {/* Global Actions Toolbar */}
        <div className="flex items-center gap-2 flex-wrap">
          <button
            onClick={onOpenContractModal}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-medium bg-slate-50 dark:bg-canvas-subtleDark border border-slate-200 dark:border-slate-800 text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
            title="Edit microservice schemas & endpoints"
          >
            <Layers className="w-3.5 h-3.5 text-blue-500" />
            <span>Contracts</span>
          </button>

          <a
            href="/api/export/middleware.go"
            download="neuroproxy_middleware.go"
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-medium bg-slate-50 dark:bg-canvas-subtleDark border border-brand-orange/30 text-brand-orange hover:bg-brand-orange/10 transition-colors"
            title="Export learned adapters as zero-latency static Go middleware"
          >
            <Download className="w-3.5 h-3.5" />
            <span>Export Middleware</span>
          </a>

          <button
            onClick={handleBurst}
            disabled={isBenchmarking}
            className="flex items-center gap-1.5 px-3.5 py-1.5 rounded-xl text-xs font-bold bg-gradient-to-r from-brand-orange to-amber-500 text-slate-950 hover:brightness-105 active:scale-95 transition-all shadow-sm cursor-pointer disabled:opacity-50"
          >
            <Zap className={`w-3.5 h-3.5 ${isBenchmarking ? 'animate-spin' : ''}`} />
            <span>{isBenchmarking ? 'Benchmarking...' : 'Burst 10x'}</span>
          </button>

          <ThemeToggle theme={theme} onToggle={onToggleTheme} />

          <button
            onClick={handleReset}
            disabled={resetting}
            className="p-2 rounded-xl text-slate-400 hover:text-rose-500 hover:bg-rose-50 dark:hover:bg-rose-950/30 border border-transparent hover:border-rose-200 dark:hover:border-rose-900/30 transition-colors cursor-pointer"
            title="Reset Cache & Incidents"
          >
            <RefreshCw className={`w-4 h-4 ${resetting ? 'animate-spin' : ''}`} />
          </button>
        </div>
      </div>
    </header>
  )
}
