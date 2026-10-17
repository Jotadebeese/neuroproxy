import React, { useState } from 'react'
import { IncidentRecord } from '../types'
import { CheckCircle2, AlertOctagon, GitPullRequest, ArrowRight, Copy, Check, Terminal } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'

interface LiveDiffViewerProps {
  incident: IncidentRecord | null
}

export const LiveDiffViewer: React.FC<LiveDiffViewerProps> = ({ incident }) => {
  const [activeTab, setActiveTab] = useState<'payload' | 'diff' | 'mappings'>('payload')
  const [copied, setCopied] = useState(false)

  if (!incident) {
    return (
      <div className="rounded-2xl modern-card p-10 flex flex-col items-center justify-center text-center h-full min-h-[360px]">
        <div className="p-3 rounded-2xl bg-slate-100 dark:bg-canvas-subtleDark border border-slate-200 dark:border-slate-800 text-slate-400 mb-3">
          <Terminal className="w-8 h-8" />
        </div>
        <h3 className="text-sm font-bold text-slate-800 dark:text-slate-200">
          Awaiting Traffic Ingestion
        </h3>
        <p className="text-xs text-slate-500 dark:text-slate-400 max-w-sm mt-1.5 leading-relaxed">
          Select a scenario preset on the left and click "Inject Traffic" to trigger live schema drift synthesis and view the diff.
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
    <div className="rounded-2xl modern-card p-6 flex flex-col h-full">
      {/* Top Status Banner */}
      <div className="pb-4 border-b border-slate-200 dark:border-slate-800 mb-4">
        <div className="flex items-center justify-between flex-wrap gap-2">
          <div className="flex items-center gap-2">
            {isHealed && (
              <span className="flex items-center gap-1.5 px-3 py-1 rounded-lg text-xs font-semibold bg-emerald-50 dark:bg-emerald-950/40 text-emerald-700 dark:text-emerald-400 border border-emerald-200 dark:border-emerald-800/40">
                <CheckCircle2 className="w-3.5 h-3.5" />
                {incident.resolution_mode === 'CACHE_HIT' ? 'Hot-Path Cache Recovery (0.12ms)' : 'Autonomously Self-Healed (200 OK)'}
              </span>
            )}
            {isVeto && (
              <span className="flex items-center gap-1.5 px-3 py-1 rounded-lg text-xs font-semibold bg-rose-50 dark:bg-rose-950/40 text-rose-700 dark:text-rose-400 border border-rose-200 dark:border-rose-800/40">
                <AlertOctagon className="w-3.5 h-3.5" />
                Safety Veto: Value Conservation Guardrail Enforced
              </span>
            )}
            {isBypass && (
              <span className="flex items-center gap-1.5 px-3 py-1 rounded-lg text-xs font-semibold bg-rose-50 dark:bg-rose-950/40 text-rose-700 dark:text-rose-400 border border-rose-200 dark:border-rose-800/40">
                <AlertOctagon className="w-3.5 h-3.5" />
                Upstream 400 Bad Request (Bypass Mode)
              </span>
            )}
          </div>

          <div className="flex items-center gap-2 text-xs text-slate-500 dark:text-slate-400 font-medium">
            <span>Latency: <strong className="text-slate-900 dark:text-white font-mono">{(incident.total_latency_ms || 0).toFixed(1)}ms</strong></span>
            {incident.agent_turns > 0 && (
              <span>• Turns: <strong className="text-brand-orange font-mono">{incident.agent_turns}</strong></span>
            )}
          </div>
        </div>

        {incident.root_cause_analysis && (
          <p className="mt-2.5 text-xs text-slate-700 dark:text-slate-300 bg-slate-50 dark:bg-slate-900/60 p-3 rounded-xl border border-slate-200 dark:border-slate-800 leading-relaxed">
            <strong className="text-brand-orange font-semibold">RCA:</strong> {incident.root_cause_analysis}
          </p>
        )}
      </div>

      {/* Tabs */}
      <div className="flex items-center justify-between border-b border-slate-200 dark:border-slate-800 pb-2.5 mb-4">
        <div className="flex items-center gap-1.5">
          <button
            onClick={() => setActiveTab('payload')}
            className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition-colors cursor-pointer ${
              activeTab === 'payload'
                ? 'bg-slate-900 text-white dark:bg-brand-orange dark:text-slate-950 shadow-sm'
                : 'text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-white'
            }`}
          >
            Payload Diff (Before / After)
          </button>
          <button
            onClick={() => setActiveTab('diff')}
            className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition-colors flex items-center gap-1 cursor-pointer ${
              activeTab === 'diff'
                ? 'bg-purple-600 text-white shadow-sm'
                : 'text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-white'
            }`}
          >
            <GitPullRequest className="w-3 h-3" />
            GitHub PR Patch (.diff)
          </button>
          <button
            onClick={() => setActiveTab('mappings')}
            className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition-colors cursor-pointer ${
              activeTab === 'mappings'
                ? 'bg-blue-600 text-white shadow-sm'
                : 'text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-white'
            }`}
          >
            Mappings ({incident.mappings?.length || 0})
          </button>
        </div>

        {activeTab === 'diff' && incident.github_pr_diff && (
          <button
            onClick={copyDiff}
            className="flex items-center gap-1 text-xs font-medium text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-white cursor-pointer"
          >
            {copied ? <Check className="w-3.5 h-3.5 text-emerald-500" /> : <Copy className="w-3.5 h-3.5" />}
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
              className="grid grid-cols-1 md:grid-cols-2 gap-3.5 h-full"
            >
              {/* Drifted Input */}
              <div className="flex flex-col h-full bg-slate-900 rounded-xl border border-rose-500/30 p-3.5">
                <span className="text-xs font-bold text-rose-400 mb-2 flex items-center justify-between">
                  <span>Drifted Incoming Webhook</span>
                  <span className="text-[10px] text-slate-400 font-mono">Source</span>
                </span>
                <pre className="text-xs font-mono text-rose-300 overflow-auto flex-1 leading-relaxed">
                  {JSON.stringify(incident.original_payload, null, 2)}
                </pre>
              </div>

              {/* Healed Target */}
              <div className="flex flex-col h-full bg-slate-900 rounded-xl border border-emerald-500/30 p-3.5">
                <span className="text-xs font-bold text-emerald-400 mb-2 flex items-center justify-between">
                  <span>Synthesized Target Payload</span>
                  <span className="text-[10px] text-emerald-400 font-mono">Strict Contract Compliant</span>
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
              className="bg-slate-900 rounded-xl border border-purple-500/30 p-4 h-full flex flex-col"
            >
              <div className="text-xs font-bold text-purple-300 mb-2">
                {incident.github_pr_title || 'PR Patch'}
              </div>
              <pre className="text-xs font-mono text-slate-300 overflow-auto flex-1 leading-relaxed whitespace-pre">
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
              className="space-y-2.5"
            >
              {(!incident.mappings || incident.mappings.length === 0) ? (
                <p className="text-xs text-slate-500 py-8 text-center">
                  No explicit mappings recorded for this request.
                </p>
              ) : (
                incident.mappings.map((m, idx) => (
                  <div
                    key={idx}
                    className="p-3 rounded-xl bg-slate-50 dark:bg-canvas-subtleDark border border-slate-200 dark:border-slate-800 flex items-center justify-between text-xs"
                  >
                    <div className="flex items-center gap-2 font-mono">
                      <span className="text-rose-500 font-semibold">{m.source_path}</span>
                      <ArrowRight className="w-3.5 h-3.5 text-slate-400" />
                      <span className="text-emerald-500 font-semibold">{m.target_path}</span>
                    </div>
                    <span className="px-2 py-0.5 rounded text-[11px] font-semibold bg-brand-orange/10 text-brand-orange border border-brand-orange/20 font-mono">
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
