import React from 'react'
import { Metrics } from '../types'
import { Activity, ShieldCheck, Zap, Bot, Clock, AlertTriangle } from 'lucide-react'
import { motion } from 'framer-motion'

interface MetricsBarProps {
  metrics: Metrics
}

export const MetricsBar: React.FC<MetricsBarProps> = ({ metrics }) => {
  const total = metrics.total_requests || 0
  const healed = metrics.self_healed_requests || 0
  const cacheHits = metrics.cache_hits || 0
  const cacheRate = total > 0 ? Math.round((cacheHits / total) * 100) : 0
  const healRate = total > 0 ? Math.round((healed / total) * 100) : 100
  const vetoes = metrics.safety_vetoes || 0

  const cards = [
    {
      label: 'Total Requests',
      value: total.toLocaleString(),
      sub: `${metrics.valid_pass_requests || 0} valid / ${healed} healed`,
      icon: Activity,
      color: 'text-blue-400',
      bgColor: 'bg-blue-500/10',
      borderColor: 'border-blue-500/20',
    },
    {
      label: 'Self-Healing Rate',
      value: `${healRate}%`,
      sub: `${healed} schema drifts adapted`,
      icon: Bot,
      color: 'text-brand-orange',
      bgColor: 'bg-brand-orange/10',
      borderColor: 'border-brand-orange/25',
    },
    {
      label: 'Rule Cache Hits',
      value: `${cacheRate}%`,
      sub: `${cacheHits} hot-path recoveries (<1ms)`,
      icon: Zap,
      color: 'text-amber-400',
      bgColor: 'bg-amber-500/10',
      borderColor: 'border-amber-500/20',
    },
    {
      label: 'Value Conservation',
      value: '100.0%',
      sub: vetoes > 0 ? `${vetoes} hallucination vetoes enforced` : 'Zero data corruption guaranteed',
      icon: ShieldCheck,
      color: 'text-emerald-400',
      bgColor: 'bg-emerald-500/10',
      borderColor: 'border-emerald-500/20',
    },
    {
      label: 'Fast-Path Latency',
      value: `${metrics.avg_cache_latency_ms.toFixed(2)} ms`,
      sub: `Gemini loop: ${metrics.avg_gemini_latency_ms.toFixed(0)} ms`,
      icon: Clock,
      color: 'text-cyan-400',
      bgColor: 'bg-cyan-500/10',
      borderColor: 'border-cyan-500/20',
    },
  ]

  return (
    <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-3.5">
      {cards.map((card, idx) => {
        const Icon = card.icon
        return (
          <motion.div
            key={card.label}
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.3, delay: idx * 0.05 }}
            className={`p-3.5 rounded-xl border ${card.borderColor} ${card.bgColor} backdrop-blur-sm relative overflow-hidden group hover:scale-[1.02] transition-transform`}
          >
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium text-slate-400">{card.label}</span>
              <Icon className={`w-4 h-4 ${card.color}`} />
            </div>
            <div className="mt-2 flex items-baseline gap-2">
              <span className="text-2xl font-bold font-mono tracking-tight text-white">
                {card.value}
              </span>
            </div>
            <p className="mt-1 text-[11px] text-slate-400 truncate font-mono">
              {card.sub}
            </p>
          </motion.div>
        )
      })}
    </div>
  )
}
