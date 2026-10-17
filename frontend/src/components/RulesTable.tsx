import React from 'react'
import { LearnedRule } from '../types'
import { Database, Zap, Clock, ShieldCheck, ArrowRight } from 'lucide-react'

interface RulesTableProps {
  rules: LearnedRule[]
}

export const RulesTable: React.FC<RulesTableProps> = ({ rules }) => {
  return (
    <div className="rounded-xl border border-white/10 bg-canvas-cardDark/70 backdrop-blur-md p-5 flex flex-col h-full">
      <div className="flex items-center justify-between pb-3 border-b border-white/10 mb-3">
        <div className="flex items-center gap-2">
          <Database className="w-4 h-4 text-amber-400" />
          <h2 className="text-sm font-semibold text-white tracking-tight">
            Active LRU Rule Cache (Fast-Path Memory)
          </h2>
        </div>
        <span className="text-xs font-mono text-slate-400">
          {rules.length} Compiled Rules (5,000 Cap)
        </span>
      </div>

      <div className="flex-1 overflow-auto max-h-[380px] space-y-2.5 pr-1">
        {rules.length === 0 ? (
          <p className="text-xs text-slate-500 font-mono py-8 text-center">
            No rules in cache. Rules are synthesized and cached on the first encounter of schema drift.
          </p>
        ) : (
          rules.map((rule, idx) => (
            <div
              key={`${rule.service_id}-${rule.fingerprint}-${idx}`}
              className="p-3 rounded-lg border border-white/5 bg-canvas-subtleDark/40 text-xs font-mono"
            >
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-amber-500/20 text-amber-300 border border-amber-500/30">
                    {rule.service_id.toUpperCase()}
                  </span>
                  <span className="text-slate-400 text-[11px] truncate max-w-[180px]">
                    fp:{rule.fingerprint}
                  </span>
                </div>
                <div className="flex items-center gap-1.5 text-emerald-400 font-bold text-[11px]">
                  <Zap className="w-3 h-3 text-amber-400" />
                  <span>{rule.hit_count || 0} Hits (<span className="text-slate-400 font-normal">sub-ms</span>)</span>
                </div>
              </div>

              {/* Mappings preview */}
              <div className="space-y-1 mb-2 bg-black/40 p-2 rounded border border-white/5">
                {(rule.mappings || []).map((m, mIdx) => (
                  <div key={mIdx} className="flex items-center justify-between text-[11px]">
                    <div className="flex items-center gap-1.5 truncate">
                      <span className="text-rose-400">{m.source_path}</span>
                      <ArrowRight className="w-3 h-3 text-slate-600 shrink-0" />
                      <span className="text-emerald-400">{m.target_path}</span>
                    </div>
                    <span className="text-[10px] text-slate-400">
                      {m.transform}
                    </span>
                  </div>
                ))}
              </div>

              <div className="flex items-center justify-between text-[10px] text-slate-500">
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
