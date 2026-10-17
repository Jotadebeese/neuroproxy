import React from 'react'
import { IncidentRecord } from '../types'
import { Activity, CheckCircle2, AlertOctagon, Zap, ShieldAlert, ChevronRight } from 'lucide-react'

interface IncidentsTimelineProps {
  incidents: IncidentRecord[]
  selectedIncident: IncidentRecord | null
  onSelectIncident: (inc: IncidentRecord) => void
}

export const IncidentsTimeline: React.FC<IncidentsTimelineProps> = ({
  incidents,
  selectedIncident,
  onSelectIncident,
}) => {
  return (
    <div className="rounded-2xl modern-card p-6 flex flex-col h-full">
      <div className="flex items-center justify-between pb-4 border-b border-slate-200 dark:border-slate-800 mb-4">
        <div className="flex items-center gap-2.5">
          <div className="p-2 rounded-xl bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-200 dark:border-emerald-800/40 text-emerald-600 dark:text-emerald-400">
            <Activity className="w-4 h-4" />
          </div>
          <div>
            <h2 className="text-base font-bold text-slate-900 dark:text-white tracking-tight">
              Live Incident Audit Trail
            </h2>
            <p className="text-xs text-slate-500 dark:text-slate-400">
              Real-time audit log of all gateway decisions
            </p>
          </div>
        </div>
        <span className="text-xs font-semibold px-2.5 py-1 rounded-full bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300">
          {incidents.length} Records
        </span>
      </div>

      <div className="flex-1 overflow-auto space-y-2.5 max-h-[380px] pr-1">
        {incidents.length === 0 ? (
          <p className="text-xs text-slate-500 py-10 text-center">
            No incidents recorded yet. Trigger a request to start logging.
          </p>
        ) : (
          incidents.map((inc) => {
            const isSelected = selectedIncident?.id === inc.id
            const isCache = inc.resolution_mode === 'CACHE_HIT'
            const isGemini = inc.resolution_mode === 'GEMINI_SYNTHESIS'
            const isVeto = inc.resolution_mode === 'SAFETY_VETO' || inc.guardrail?.hallucination_veto
            const isBypass = inc.resolution_mode === 'BYPASS_FAILED'

            return (
              <div
                key={inc.id}
                onClick={() => onSelectIncident(inc)}
                className={`p-3.5 rounded-xl border cursor-pointer transition-all ${
                  isSelected
                    ? 'border-brand-orange bg-orange-50/60 dark:bg-orange-950/20 ring-1 ring-brand-orange/40'
                    : 'border-slate-200 dark:border-slate-800 bg-white dark:bg-canvas-subtleDark/40 hover:border-slate-300 dark:hover:border-slate-700'
                }`}
              >
                <div className="flex items-center justify-between text-xs">
                  <div className="flex items-center gap-2">
                    {isCache && <Zap className="w-4 h-4 text-amber-500" />}
                    {isGemini && <CheckCircle2 className="w-4 h-4 text-emerald-500" />}
                    {isVeto && <AlertOctagon className="w-4 h-4 text-rose-500" />}
                    {isBypass && <ShieldAlert className="w-4 h-4 text-rose-500" />}
                    <span className="font-bold text-slate-900 dark:text-white uppercase font-mono">{inc.service_id}</span>
                    <span className="text-slate-400 font-mono text-[11px]">#{inc.id.slice(0, 8)}</span>
                  </div>
                  <span className="text-slate-400 text-xs">
                    {new Date(inc.timestamp).toLocaleTimeString()}
                  </span>
                </div>

                <div className="flex items-center justify-between mt-2.5 text-xs">
                  <span
                    className={`px-2.5 py-0.5 rounded-full text-[11px] font-semibold ${
                      isCache
                        ? 'bg-amber-100 dark:bg-amber-950/40 text-amber-800 dark:text-amber-300 border border-amber-200 dark:border-amber-800/40'
                        : isGemini
                        ? 'bg-emerald-100 dark:bg-emerald-950/40 text-emerald-800 dark:text-emerald-300 border border-emerald-200 dark:border-emerald-800/40'
                        : 'bg-rose-100 dark:bg-rose-950/40 text-rose-800 dark:text-rose-300 border border-rose-200 dark:border-rose-800/40'
                    }`}
                  >
                    {inc.resolution_mode}
                  </span>

                  <div className="flex items-center gap-2.5 text-slate-500 dark:text-slate-400 font-medium">
                    <span className="font-mono">{(inc.total_latency_ms || 0).toFixed(1)}ms</span>
                    <span className={inc.final_status === 200 ? 'text-emerald-600 dark:text-emerald-400 font-bold font-mono' : 'text-rose-600 dark:text-rose-400 font-bold font-mono'}>
                      HTTP {inc.final_status}
                    </span>
                    <ChevronRight className="w-3.5 h-3.5 text-slate-400" />
                  </div>
                </div>
              </div>
            )
          })
        )}
      </div>
    </div>
  )
}
