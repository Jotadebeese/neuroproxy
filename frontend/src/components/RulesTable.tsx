import React from 'react'
import { LearnedRule } from '../types'
import { Database, Zap, ArrowRight } from 'lucide-react'

interface RulesTableProps {
  rules: LearnedRule[]
}

export const RulesTable: React.FC<RulesTableProps> = ({ rules }) => {
  return (
    <div className="rounded-2xl modern-card p-6 flex flex-col h-full">
      <div className="flex items-center justify-between pb-4 border-b border-slate-200 dark:border-slate-800 mb-4">
        <div className="flex items-center gap-2.5">
          <div className="p-2 rounded-xl bg-amber-50 dark:bg-amber-950/40 border border-amber-200 dark:border-amber-800/40 text-amber-600 dark:text-amber-400">
            <Database className="w-4 h-4" />
          </div>
          <div>
            <h2 className="text-base font-bold text-slate-900 dark:text-white tracking-tight">
              Active Rule Cache (Fast Path)
            </h2>
            <p className="text-xs text-slate-500 dark:text-slate-400">
              LRU memory &amp; GCP Memorystore Redis adapters
            </p>
          </div>
        </div>
        <span className="text-xs font-semibold px-2.5 py-1 rounded-full bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300">
          {rules.length} Compiled (5,000 Cap)
        </span>
      </div>

      <div className="flex-1 overflow-auto max-h-[380px] space-y-3 pr-1">
        {rules.length === 0 ? (
          <p className="text-xs text-slate-500 py-10 text-center">
            No rules in cache. Rules are synthesized and cached on the first encounter of schema drift.
          </p>
        ) : (
          rules.map((rule, idx) => (
            <div
              key={`${rule.service_id}-${rule.fingerprint}-${idx}`}
              className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-canvas-subtleDark/40 text-xs"
            >
              <div className="flex items-center justify-between mb-2.5">
                <div className="flex items-center gap-2">
                  <span className="px-2.5 py-0.5 rounded-md text-[11px] font-bold font-mono bg-amber-100 dark:bg-amber-950/40 text-amber-800 dark:text-amber-300 border border-amber-200 dark:border-amber-800/40">
                    {rule.service_id.toUpperCase()}
                  </span>
                  <span className="text-slate-500 dark:text-slate-400 font-mono text-[11px] truncate max-w-[200px]">
                    fp:{rule.fingerprint}
                  </span>
                </div>
                <div className="flex items-center gap-1.5 text-emerald-600 dark:text-emerald-400 font-bold text-xs">
                  <Zap className="w-3.5 h-3.5 text-amber-500" />
                  <span>{rule.hit_count || 0} Hits (<span className="text-slate-400 font-normal">sub-ms</span>)</span>
                </div>
              </div>

              {/* Mappings preview */}
              <div className="space-y-1.5 mb-2.5 bg-slate-50 dark:bg-slate-900/60 p-2.5 rounded-lg border border-slate-200 dark:border-slate-800 font-mono">
                {(rule.mappings || []).map((m, mIdx) => (
                  <div key={mIdx} className="flex items-center justify-between text-xs">
                    <div className="flex items-center gap-2 truncate">
                      <span className="text-rose-500 font-semibold">{m.source_path}</span>
                      <ArrowRight className="w-3.5 h-3.5 text-slate-400 shrink-0" />
                      <span className="text-emerald-500 font-semibold">{m.target_path}</span>
                    </div>
                    <span className="text-[11px] text-slate-400">
                      {m.transform}
                    </span>
                  </div>
                ))}
              </div>

              <div className="flex items-center justify-between text-[11px] text-slate-400">
                <span>Compiled: {new Date(rule.created_at).toLocaleTimeString()}</span>
                <span>Last active: {new Date(rule.last_used_at).toLocaleTimeString()}</span>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
