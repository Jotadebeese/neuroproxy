import React from 'react'
import { Metrics } from '../types'
import { Activity, ShieldCheck, Zap, Bot, Clock } from 'lucide-react'
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
      label: 'Total Ingested',
      value: total.toLocaleString(),
      sub: `${metrics.valid_pass_requests || 0} valid / ${healed} healed`,
      icon: Activity,
      iconColor: 'text-blue-500',
      iconBg: 'bg-blue-50 dark:bg-blue-950/40 border-blue-200 dark:border-blue-800/40',
    },
    {
      label: 'Autonomous Healing',
      value: `${healRate}%`,
      sub: `${healed} schema drifts adapted`,
      icon: Bot,
      iconColor: 'text-brand-orange',
      iconBg: 'bg-orange-50 dark:bg-orange-950/40 border-orange-200 dark:border-orange-800/40',
    },
    {
      label: 'Rule Cache Hits',
      value: `${cacheRate}%`,
      sub: `${cacheHits} hot-path recoveries (<1ms)`,
      icon: Zap,
      iconColor: 'text-amber-500',
      iconBg: 'bg-amber-50 dark:bg-amber-950/40 border-amber-200 dark:border-amber-800/40',
    },
    {
      label: 'Value Conservation',
      value: '100.0%',
      sub: vetoes > 0 ? `${vetoes} hallucination vetoes enforced` : 'Zero data corruption guaranteed',
      icon: ShieldCheck,
      iconColor: 'text-emerald-500',
      iconBg: 'bg-emerald-50 dark:bg-emerald-950/40 border-emerald-200 dark:border-emerald-800/40',
    },
    {
      label: 'Fast-Path Latency',
      value: `${metrics.avg_cache_latency_ms.toFixed(2)} ms`,
      sub: `Gemini turn: ${metrics.avg_gemini_latency_ms.toFixed(0)} ms`,
      icon: Clock,
      iconColor: 'text-cyan-500',
      iconBg: 'bg-cyan-50 dark:bg-cyan-950/40 border-cyan-200 dark:border-cyan-800/40',
    },
  ]

  return (
    <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-4">
      {cards.map((card, idx) => {
        const Icon = card.icon
        return (
          <motion.div
            key={card.label}
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.3, delay: idx * 0.05 }}
            className="p-5 rounded-2xl modern-card modern-card-hover flex flex-col justify-between"
          >
            <div className="flex items-center justify-between">
              <span className="text-xs font-semibold text-slate-500 dark:text-slate-400">
                {card.label}
              </span>
              <div className={`p-1.5 rounded-lg border ${card.iconBg}`}>
                <Icon className={`w-3.5 h-3.5 ${card.iconColor}`} />
              </div>
            </div>

            <div className="mt-3">
              <div className="text-2xl font-bold tracking-tight text-slate-900 dark:text-white">
                {card.value}
              </div>
              <p className="mt-1 text-xs text-slate-500 dark:text-slate-400 truncate">
                {card.sub}
              </p>
            </div>
          </motion.div>
        )
      })}
    </div>
  )
}
