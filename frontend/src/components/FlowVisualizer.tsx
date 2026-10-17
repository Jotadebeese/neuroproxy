import React from 'react'
import { motion } from 'framer-motion'
import { Globe, Database, Cpu, ShieldCheck, CheckCircle2, AlertOctagon, GitPullRequest, ArrowRight } from 'lucide-react'
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
      color: 'border-slate-500 text-slate-300 bg-slate-900/60',
    },
    {
      id: 'cache',
      title: '2. Rule Cache',
      subtitle: isCacheHit ? 'Hit: 0.12ms (LRU)' : 'Miss: Evaluating',
      icon: Database,
      active: isCacheHit || isProcessing,
      highlight: isCacheHit,
      color: isCacheHit
        ? 'border-amber-400 text-amber-300 bg-amber-950/40 shadow-glow-orange'
        : 'border-slate-700 text-slate-400 bg-slate-900/40',
    },
    {
      id: 'agent',
      title: '3. Gemini Agent',
      subtitle: isGemini ? `Turn ${lastIncident?.agent_turns || 1} Synthesizer` : 'Bypassed / Idle',
      icon: Cpu,
      active: isGemini || (isProcessing && !isCacheHit),
      highlight: isGemini,
      color: isGemini
        ? 'border-brand-orange text-brand-orange bg-orange-950/40 shadow-glow-orange'
        : 'border-slate-700 text-slate-400 bg-slate-900/40',
    },
    {
      id: 'guardrail',
      title: '4. Value Guardrail',
      subtitle: isVeto ? 'SAFETY VETO Blocked' : '100% Value Conserved',
      icon: isVeto ? AlertOctagon : ShieldCheck,
      active: Boolean(lastIncident),
      highlight: !isVeto && Boolean(lastIncident),
      color: isVeto
        ? 'border-rose-500 text-rose-400 bg-rose-950/40'
        : 'border-emerald-500 text-emerald-400 bg-emerald-950/40 shadow-glow-emerald',
    },
    {
      id: 'upstream',
      title: '5. Upstream Delivery',
      subtitle: isVeto || isBypass ? '400 / 422 Rejected' : '200 OK Upstream Ack',
      icon: CheckCircle2,
      active: Boolean(lastIncident),
      highlight: !isVeto && !isBypass && Boolean(lastIncident),
      color: !isVeto && !isBypass && lastIncident
        ? 'border-emerald-400 text-emerald-300 bg-emerald-950/40'
        : isBypass || isVeto
        ? 'border-rose-500 text-rose-400 bg-rose-950/30'
        : 'border-slate-700 text-slate-400 bg-slate-900/40',
    },
  ]

  return (
    <div className="p-4 rounded-xl border border-white/10 bg-canvas-cardDark/70 backdrop-blur-md">
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <div className="w-2.5 h-2.5 rounded-full bg-brand-orange animate-pulse" />
          <h2 className="text-xs font-mono font-bold uppercase tracking-wider text-slate-300">
            NeuroProxy Autonomous Execution Pipeline
          </h2>
        </div>
        {lastIncident && (
          <span className="text-[11px] font-mono text-slate-400">
            Mode: <strong className="text-white">{lastIncident.resolution_mode}</strong> • Latency: <strong className="text-emerald-400">{(lastIncident.total_latency_ms || 0).toFixed(1)}ms</strong>
          </span>
        )}
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-5 gap-2 relative">
        {stages.map((stage, idx) => {
          const Icon = stage.icon
          return (
            <motion.div
              key={stage.id}
              initial={false}
              animate={{
                scale: stage.highlight ? [1, 1.03, 1] : 1,
              }}
              transition={{ repeat: stage.highlight ? 1 : 0, duration: 0.4 }}
              className={`p-3 rounded-lg border ${stage.color} flex flex-col justify-between transition-all duration-300 relative overflow-hidden`}
            >
              {stage.highlight && (
                <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/5 to-transparent animate-flow" />
              )}
              <div className="flex items-center justify-between mb-1.5">
                <span className="text-[11px] font-mono font-semibold">{stage.title}</span>
                <Icon className="w-4 h-4 opacity-80" />
              </div>
              <p className="text-[10px] font-mono opacity-80 truncate">{stage.subtitle}</p>
            </motion.div>
          )
        })}
      </div>

      {lastIncident?.github_pr_title && (
        <div className="mt-3 pt-2.5 border-t border-white/5 flex items-center justify-between text-xs font-mono text-slate-400">
          <div className="flex items-center gap-1.5 truncate">
            <GitPullRequest className="w-3.5 h-3.5 text-purple-400 shrink-0" />
            <span className="text-purple-300">Autonomous PR Synthesized:</span>
            <span className="text-slate-300 truncate">{lastIncident.github_pr_title}</span>
          </div>
          <span className="px-1.5 py-0.5 rounded text-[10px] bg-purple-500/10 text-purple-300 border border-purple-500/20 shrink-0">
            Zero Human Intervention
          </span>
        </div>
      )}
    </div>
  )
}
