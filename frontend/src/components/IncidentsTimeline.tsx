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
    <div className="rounded-xl border border-white/10 bg-canvas-cardDark/70 backdrop-blur-md p-5 flex flex-col h-full">
      <div className="flex items-center justify-between pb-3 border-b border-white/10 mb-3">
        <div className="flex items-center gap-2">
          <Activity className="w-4 h-4 text-emerald-400" />
          <h2 className="text-sm font-semibold text-white tracking-tight">
            Live Incident Audit Trail
          </h2>
        </div>
        <span className="text-xs font-mono text-slate-400">
          {incidents.length} Records
        </span>
      </div>

      <div className="flex-1 overflow-auto space-y-2 max-h-[380px] pr-1">
        {incidents.length === 0 ? (
          <p className="text-xs text-slate-500 font-mono py-8 text-center">
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
                className={`p-3 rounded-lg border cursor-pointer transition-all ${
                  isSelected
                    ? 'border-brand-orange bg-brand-orange/10 shadow-glow-orange'
                    : 'border-white/5 bg-canvas-subtleDark/50 hover:border-white/20 hover:bg-canvas-subtleDark'
                }`}
              >
                <div className="flex items-center justify-between text-xs font-mono">
                  <div className="flex items-center gap-2">
                    {isCache && <Zap className="w-3.5 h-3.5 text-amber-400" />}
                    {isGemini && <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" />}
                    {isVeto && <AlertOctagon className="w-3.5 h-3.5 text-rose-400" />}
                    {isBypass && <ShieldAlert className="w-3.5 h-3.5 text-rose-400" />}
                    <span className="font-bold text-white uppercase">{inc.service_id}</span>
                    <span className="text-slate-400">#{inc.id.slice(0, 8)}</span>
                  </div>
                  <span className="text-slate-400 text-[11px]">
                    {new Date(inc.timestamp).toLocaleTimeString()}
                  </span>
                </div>

                <div className="flex items-center justify-between mt-2 text-[11px] font-mono">
                  <span
                    className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                      isCache
                        ? 'bg-amber-500/20 text-amber-300'
                        : isGemini
                        ? 'bg-emerald-500/20 text-emerald-300'
                        : isVeto
                        ? 'bg-rose-500/20 text-rose-300'
                        : 'bg-rose-500/20 text-rose-300'
                    }`}
                  >
                    {inc.resolution_mode}
                  </span>

                  <div className="flex items-center gap-2 text-slate-400">
                    <span>{(inc.total_latency_ms || 0).toFixed(1)}ms</span>
                    <span className={inc.final_status === 200 ? 'text-emerald-400 font-bold' : 'text-rose-400 font-bold'}>
                      HTTP {inc.final_status}
                    </span>
                    <ChevronRight className="w-3 h-3 text-slate-600" />
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
