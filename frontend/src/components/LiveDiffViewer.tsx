import React, { useState } from 'react'
import { IncidentRecord } from '../types'
import { CheckCircle2, AlertOctagon, GitPullRequest, ArrowRight, ShieldCheck, Copy, Check, Terminal } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'

interface LiveDiffViewerProps {
  incident: IncidentRecord | null
}

export const LiveDiffViewer: React.FC<LiveDiffViewerProps> = ({ incident }) => {
  const [activeTab, setActiveTab] = useState<'payload' | 'diff' | 'mappings'>('payload')
  const [copied, setCopied] = useState(false)

  if (!incident) {
    return (
      <div className="rounded-xl border border-white/10 bg-canvas-cardDark/70 backdrop-blur-md p-8 flex flex-col items-center justify-center text-center h-full min-h-[350px]">
        <Terminal className="w-10 h-10 text-slate-600 mb-3 animate-pulse" />
        <h3 className="text-sm font-semibold text-slate-300">Awaiting Traffic Ingestion</h3>
        <p className="text-xs text-slate-500 max-w-sm mt-1">
          Select a scenario preset on the left and click "Inject Traffic" to trigger live schema drift synthesis.
        </p>
      </div>
    )
  }

  const isHealed = incident.resolution_mode === 'GEMINI_SYNTHESIS' || incident.resolution_mode === 'CACHE_HIT'
  const isVeto = incident.resolution_mode === 'SAFETY_VETO' || incident.guardrail?.hallucination_veto
  const isBypass = incident.resolution_mode === 'BYPASS_FAILED'

  const copyDiff = () => {
    if (incident.github_pr_diff) {
      navigator.clipboard.writeText(incident.github_pr_diff)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  return (
    <div className="rounded-xl border border-white/10 bg-canvas-cardDark/70 backdrop-blur-md p-5 flex flex-col h-full">
      {/* Top Banner */}
      <div className="pb-3 border-b border-white/10 mb-3">
        <div className="flex items-center justify-between flex-wrap gap-2">
          <div className="flex items-center gap-2">
            {isHealed && (
              <span className="flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-mono font-semibold bg-emerald-500/15 text-emerald-400 border border-emerald-500/30">
                <CheckCircle2 className="w-3.5 h-3.5" />
                {incident.resolution_mode === 'CACHE_HIT' ? 'HOT-PATH CACHE RECOVERY (<1ms)' : 'AUTONOMOUSLY SELF-HEALED (200 OK)'}
              </span>
            )}
            {isVeto && (
              <span className="flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-mono font-semibold bg-rose-500/15 text-rose-400 border border-rose-500/30">
                <AlertOctagon className="w-3.5 h-3.5" />
                SAFETY VETO: VALUE CONSERVATION GUARDRAIL ENFORCED
              </span>
            )}
            {isBypass && (
              <span className="flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-mono font-semibold bg-rose-500/15 text-rose-400 border border-rose-500/30">
                <AlertOctagon className="w-3.5 h-3.5" />
                UPSTREAM 400 REJECTION (BYPASS MODE)
              </span>
            )}
          </div>

          <div className="flex items-center gap-1 text-xs font-mono text-slate-400">
            <span>Latency: <strong className="text-white">{(incident.total_latency_ms || 0).toFixed(1)}ms</strong></span>
            {incident.agent_turns > 0 && (
              <span> • Agent Turns: <strong className="text-brand-orange">{incident.agent_turns}</strong></span>
            )}
          </div>
        </div>

        {incident.root_cause_analysis && (
          <p className="mt-2 text-xs text-slate-300 font-mono bg-black/30 p-2.5 rounded-lg border border-white/5 leading-relaxed">
            <strong className="text-brand-orange">RCA:</strong> {incident.root_cause_analysis}
          </p>
        )}
      </div>

      {/* Tabs */}
      <div className="flex items-center justify-between border-b border-white/5 pb-2 mb-3">
        <div className="flex items-center gap-2">
          <button
            onClick={() => setActiveTab('payload')}
            className={`px-3 py-1 rounded-lg text-xs font-mono transition-colors ${
              activeTab === 'payload'
                ? 'bg-brand-orange text-slate-950 font-bold'
                : 'text-slate-400 hover:text-white hover:bg-white/5'
            }`}
          >
            Payload Diff (Before / After)
          </button>
          <button
            onClick={() => setActiveTab('diff')}
            className={`px-3 py-1 rounded-lg text-xs font-mono transition-colors flex items-center gap-1 ${
              activeTab === 'diff'
                ? 'bg-purple-600 text-white font-bold'
                : 'text-slate-400 hover:text-white hover:bg-white/5'
            }`}
          >
            <GitPullRequest className="w-3 h-3" />
            GitHub PR Patch (.diff)
          </button>
          <button
            onClick={() => setActiveTab('mappings')}
            className={`px-3 py-1 rounded-lg text-xs font-mono transition-colors ${
              activeTab === 'mappings'
                ? 'bg-blue-600 text-white font-bold'
                : 'text-slate-400 hover:text-white hover:bg-white/5'
            }`}
          >
            Mappings ({incident.mappings?.length || 0})
          </button>
        </div>

        {activeTab === 'diff' && incident.github_pr_diff && (
          <button
            onClick={copyDiff}
            className="flex items-center gap-1 text-[11px] font-mono text-slate-400 hover:text-white cursor-pointer"
          >
            {copied ? <Check className="w-3 h-3 text-emerald-400" /> : <Copy className="w-3 h-3" />}
            {copied ? 'Copied' : 'Copy Diff'}
          </button>
        )}
      </div>

      {/* Tab Contents */}
      <div className="flex-1 overflow-auto">
        <AnimatePresence mode="wait">
          {activeTab === 'payload' && (
            <motion.div
              key="payload"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="grid grid-cols-1 md:grid-cols-2 gap-3 h-full"
            >
              {/* Drifted Input */}
              <div className="flex flex-col h-full bg-[#0a0a0c] rounded-lg border border-rose-500/20 p-3">
                <span className="text-[11px] font-mono font-semibold text-rose-400 mb-1.5 flex items-center justify-between">
                  <span>Drifted Incoming Webhook</span>
                  <span className="text-[10px] text-slate-500">Source</span>
                </span>
                <pre className="text-xs font-mono text-rose-300 overflow-auto flex-1 leading-relaxed">
                  {JSON.stringify(incident.original_payload, null, 2)}
                </pre>
              </div>

              {/* Healed Target */}
              <div className="flex flex-col h-full bg-[#0a0a0c] rounded-lg border border-emerald-500/20 p-3">
                <span className="text-[11px] font-mono font-semibold text-emerald-400 mb-1.5 flex items-center justify-between">
                  <span>Synthesized Healed Payload</span>
                  <span className="text-[10px] text-emerald-500">Strict Contract Compliant</span>
                </span>
                <pre className="text-xs font-mono text-emerald-300 overflow-auto flex-1 leading-relaxed">
                  {JSON.stringify(incident.healed_payload, null, 2)}
                </pre>
              </div>
            </motion.div>
          )}

          {activeTab === 'diff' && (
            <motion.div
              key="diff"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="bg-[#0a0a0c] rounded-lg border border-purple-500/20 p-3 h-full flex flex-col"
            >
              <div className="text-[11px] font-mono text-purple-300 mb-2 font-semibold">
                {incident.github_pr_title || 'PR Patch'}
              </div>
              <pre className="text-xs font-mono text-slate-300 overflow-auto flex-1 leading-relaxed whitespace-pre font-mono">
                {incident.github_pr_diff || 'No diff available for this mode.'}
              </pre>
            </motion.div>
          )}

          {activeTab === 'mappings' && (
            <motion.div
              key="mappings"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="space-y-2"
            >
              {(!incident.mappings || incident.mappings.length === 0) ? (
                <p className="text-xs text-slate-500 font-mono p-4 text-center">
                  No explicit mappings recorded for this request.
                </p>
              ) : (
                incident.mappings.map((m, idx) => (
                  <div
                    key={idx}
                    className="p-2.5 rounded-lg bg-black/40 border border-white/10 flex items-center justify-between text-xs font-mono"
                  >
                    <div className="flex items-center gap-2">
                      <span className="text-rose-400 font-semibold">{m.source_path}</span>
                      <ArrowRight className="w-3.5 h-3.5 text-slate-500" />
                      <span className="text-emerald-400 font-semibold">{m.target_path}</span>
                    </div>
                    <span className="px-2 py-0.5 rounded text-[10px] bg-brand-orange/15 text-brand-orange border border-brand-orange/30">
                      {m.transform}
                    </span>
                  </div>
                ))
              )}
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </div>
  )
}
