import React, { useState } from 'react'
import { Zap, Download, RefreshCw, Cpu, Layers, Sparkles } from 'lucide-react'
import confetti from 'canvas-confetti'

interface HeaderProps {
  onReset: () => void
  onBurstBenchmark: () => void
  onOpenContractModal: () => void
  isBenchmarking: boolean
}

export const Header: React.FC<HeaderProps> = ({
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
    <header className="border-b border-white/10 bg-canvas-cardDark/80 backdrop-blur-md sticky top-0 z-40">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-3.5 flex flex-wrap items-center justify-between gap-4">
        {/* Brand & Mission Statement */}
        <div className="flex items-center gap-3.5">
          <div className="relative group">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-brand-orange to-amber-600 flex items-center justify-center font-mono font-bold text-lg text-white shadow-glow-orange transition-transform duration-300 group-hover:scale-105">
              NP
            </div>
            <span className="absolute -bottom-1 -right-1 flex h-3 w-3">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
              <span className="relative inline-flex rounded-full h-3 w-3 bg-emerald-500"></span>
            </span>
          </div>

          <div>
            <div className="flex items-center gap-2.5 flex-wrap">
              <span className="text-xl font-bold tracking-tight text-white flex items-center gap-1.5">
                NeuroProxy
                <span className="text-xs px-2 py-0.5 rounded-full font-mono bg-brand-orange/20 text-brand-orange border border-brand-orange/30">
                  Cloud Run Edge • v2.0
                </span>
              </span>
              <span className="text-xs px-2 py-0.5 rounded-full font-mono bg-emerald-500/15 text-emerald-400 border border-emerald-500/30 flex items-center gap-1">
                <Sparkles className="w-3 h-3" />
                Gemini 2.5 Flash Autonomous Agent
              </span>
            </div>
            <p className="text-xs text-slate-400 mt-0.5">
              Self-Healing API Gateway with Zero-Hallucination Value Conservation Guardrails
            </p>
          </div>
        </div>

        {/* Global Action Toolbar */}
        <div className="flex items-center gap-2.5 flex-wrap">
          <button
            onClick={onOpenContractModal}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium bg-canvas-subtleDark border border-white/10 text-slate-300 hover:text-white hover:bg-white/10 transition-colors"
            title="Edit microservice schemas & endpoints"
          >
            <Layers className="w-3.5 h-3.5 text-blue-400" />
            Contracts & Upstreams
          </button>

          <a
            href="/api/export/middleware.go"
            download="neuroproxy_middleware.go"
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium bg-canvas-subtleDark border border-brand-orange/30 text-brand-orange hover:bg-brand-orange/15 transition-colors"
            title="Export learned adapters as zero-latency static Go middleware"
          >
            <Download className="w-3.5 h-3.5" />
            Export Go Middleware
          </a>

          <button
            onClick={handleBurst}
            disabled={isBenchmarking}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold bg-gradient-to-r from-brand-orange to-amber-500 text-slate-950 hover:brightness-110 active:scale-95 transition shadow-glow-orange disabled:opacity-50"
          >
            <Zap className={`w-3.5 h-3.5 ${isBenchmarking ? 'animate-spin' : ''}`} />
            {isBenchmarking ? 'Benchmarking...' : 'Burst 10x Benchmark'}
          </button>

          <button
            onClick={handleReset}
            disabled={resetting}
            className="p-1.5 rounded-lg text-slate-400 hover:text-rose-400 hover:bg-rose-500/10 border border-transparent hover:border-rose-500/20 transition-colors"
            title="Reset Cache & Incidents"
          >
            <RefreshCw className={`w-4 h-4 ${resetting ? 'animate-spin' : ''}`} />
          </button>
        </div>
      </div>
    </header>
  )
}
