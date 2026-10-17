import React from 'react'
import { motion } from 'framer-motion'
import { Globe, Database, Cpu, ShieldCheck, CheckCircle2, AlertOctagon, GitPullRequest } from 'lucide-react'
import { IncidentRecord } from '../types'

interface FlowVisualizerProps {
  lastIncident: IncidentRecord | null
  isProcessing: boolean
}

export const FlowVisualizer: React.FC<FlowVisualizerProps> = ({ lastIncident, isProcessing }) => {
  const isCacheHit = lastIncident?.resolution_mode === 'CACHE_HIT'
  const isGemini = lastIncident?.resolution_mode === 'GEMINI_SYNTHESIS'
  const isVeto = lastIncident?.resolution_mode === 'SAFETY_VETO' || lastIncident?.guardrail?.hallucination_veto
  const isBypass = lastIncident?.resolution_mode === 'BYPASS_FAILED'

  const stages = [
    {
      id: 'client',
      title: '1. Ingestion',
      subtitle: 'Drifted Webhook / Client',
      icon: Globe,
      active: true,
      color: 'border-slate-200 dark:border-slate-800 text-slate-700 dark:text-slate-300 bg-slate-50 dark:bg-canvas-subtleDark',
    },
    {
      id: 'cache',
      title: '2. Rule Cache',
      subtitle: isCacheHit ? 'Hit: 0.12ms (LRU / Redis)' : 'Miss: Evaluating',
      icon: Database,
      active: isCacheHit || isProcessing,
      highlight: isCacheHit,
      color: isCacheHit
        ? 'border-amber-400 text-amber-600 dark:text-amber-400 bg-amber-50 dark:bg-amber-950/30 ring-2 ring-amber-400/20'
        : 'border-slate-200 dark:border-slate-800 text-slate-500 dark:text-slate-400 bg-slate-50 dark:bg-canvas-subtleDark',
    },
    {
      id: 'agent',
      title: '3. Gemini Agent',
      subtitle: isGemini ? `Turn ${lastIncident?.agent_turns || 1} Synthesizer` : 'Bypassed / Idle',
      icon: Cpu,
      active: isGemini || (isProcessing && !isCacheHit),
      highlight: isGemini,
      color: isGemini
        ? 'border-brand-orange text-brand-orange bg-orange-50 dark:bg-orange-950/30 ring-2 ring-brand-orange/20'
        : 'border-slate-200 dark:border-slate-800 text-slate-500 dark:text-slate-400 bg-slate-50 dark:bg-canvas-subtleDark',
    },
    {
      id: 'guardrail',
      title: '4. Value Guardrail',
      subtitle: isVeto ? 'SAFETY VETO Blocked' : '100% Value Conserved',
      icon: isVeto ? AlertOctagon : ShieldCheck,
      active: Boolean(lastIncident),
      highlight: !isVeto && Boolean(lastIncident),
      color: isVeto
        ? 'border-rose-400 text-rose-600 dark:text-rose-400 bg-rose-50 dark:bg-rose-950/30 ring-2 ring-rose-400/20'
        : 'border-emerald-400 text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-950/30 ring-2 ring-emerald-400/20',
    },
    {
      id: 'upstream',
      title: '5. Upstream Delivery',
      subtitle: isVeto || isBypass ? '400 / 422 Rejected' : '200 OK Upstream Ack',
      icon: CheckCircle2,
      active: Boolean(lastIncident),
      highlight: !isVeto && !isBypass && Boolean(lastIncident),
      color: !isVeto && !isBypass && lastIncident
        ? 'border-emerald-400 text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-950/30 ring-2 ring-emerald-400/20'
        : isBypass || isVeto
        ? 'border-rose-400 text-rose-600 dark:text-rose-400 bg-rose-50 dark:bg-rose-950/30'
        : 'border-slate-200 dark:border-slate-800 text-slate-500 dark:text-slate-400 bg-slate-50 dark:bg-canvas-subtleDark',
    },
  ]

  return (
    <div className="p-6 rounded-2xl modern-card">
      <div className="flex items-center justify-between mb-4 flex-wrap gap-2">
        <div className="flex items-center gap-2.5">
          <span className="w-2.5 h-2.5 rounded-full bg-brand-orange animate-pulse" />
          <h2 className="text-xs font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300">
            Autonomous In-Flight Execution Pipeline
          </h2>
        </div>
        {lastIncident && (
          <div className="flex items-center gap-2 text-xs text-slate-500 dark:text-slate-400">
            <span>Mode: <strong className="text-slate-800 dark:text-white font-semibold">{lastIncident.resolution_mode}</strong></span>
            <span>• Latency: <strong className="text-emerald-600 dark:text-emerald-400 font-semibold font-mono">{(lastIncident.total_latency_ms || 0).toFixed(1)}ms</strong></span>
          </div>
        )}
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-5 gap-3 relative">
        {stages.map((stage) => {
          const Icon = stage.icon
          return (
            <motion.div
              key={stage.id}
              initial={false}
              animate={{
                scale: stage.highlight ? [1, 1.02, 1] : 1,
              }}
              transition={{ repeat: stage.highlight ? 1 : 0, duration: 0.4 }}
              className={`p-4 rounded-xl border ${stage.color} flex flex-col justify-between transition-all duration-300 relative overflow-hidden`}
            >
              <div className="flex items-center justify-between mb-2">
                <span className="text-xs font-bold tracking-tight">{stage.title}</span>
                <Icon className="w-4 h-4 opacity-80" />
              </div>
              <p className="text-xs opacity-85 truncate font-normal">{stage.subtitle}</p>
            </motion.div>
          )
        })}
      </div>

      {lastIncident?.github_pr_title && (
        <div className="mt-4 pt-3.5 border-t border-slate-200 dark:border-slate-800 flex items-center justify-between text-xs text-slate-600 dark:text-slate-400 flex-wrap gap-2">
          <div className="flex items-center gap-2 truncate">
            <GitPullRequest className="w-4 h-4 text-purple-500 shrink-0" />
            <span className="font-semibold text-purple-700 dark:text-purple-300">Automated PR Synthesized:</span>
            <span className="text-slate-800 dark:text-slate-300 truncate font-mono text-xs">{lastIncident.github_pr_title}</span>
          </div>
          <span className="px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-purple-100 dark:bg-purple-950/40 text-purple-700 dark:text-purple-300 border border-purple-200 dark:border-purple-800/40 shrink-0">
            Zero Human Intervention
          </span>
        </div>
      )}
    </div>
  )
}
