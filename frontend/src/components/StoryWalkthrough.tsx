import React, { useState } from 'react'
import { motion } from 'framer-motion'
import {
  AlertTriangle,
  Sparkles,
  ShieldCheck,
  Zap,
  GitPullRequest,
  CheckCircle2,
  Cpu,
  Layers,
  Clock,
  Play,
  ChevronRight,
  ArrowRight,
  Code2,
  Terminal,
} from 'lucide-react'
import { ChaosPreset, IncidentRecord } from '../types'
import { MaskedTextReveal } from './MaskedTextReveal'

interface StoryWalkthroughProps {
  presets: ChaosPreset[]
  lastIncident: IncidentRecord | null
  onTriggerPreset: (serviceId: string, payload: Record<string, any>, selfHeal: boolean) => Promise<void>
  onSwitchToConsole: () => void
  isProcessing: boolean
}

export const StoryWalkthrough: React.FC<StoryWalkthroughProps> = ({
  presets,
  lastIncident,
  onTriggerPreset,
  onSwitchToConsole,
  isProcessing,
}) => {
  const [activeStep, setActiveStep] = useState<number>(1)

  const handleRunPreset = async (presetIndex: number) => {
    const preset = presets[presetIndex] || presets[0]
    if (preset) {
      await onTriggerPreset(preset.service_id, preset.payload, true)
    }
  }

  return (
    <div className="w-full space-y-16 py-6 relative z-10">
      {/* Cuberto-Style Hero Section */}
      <section className="text-center max-w-5xl mx-auto px-4 pt-8 pb-4">
        {/* Animated Pill Badge */}
        <motion.div
          initial={{ opacity: 0, scale: 0.9, y: -10 }}
          animate={{ opacity: 1, scale: 1, y: 0 }}
          transition={{ duration: 0.6, ease: [0.16, 1, 0.3, 1] }}
          className="inline-flex items-center gap-2.5 px-4 py-2 rounded-full text-xs font-semibold bg-brand-orange/10 dark:bg-brand-orange/15 text-brand-orange border border-brand-orange/25 mb-8 shadow-sm"
        >
          <span className="flex h-2 w-2 relative">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-brand-orange opacity-75"></span>
            <span className="relative inline-flex rounded-full h-2 w-2 bg-brand-orange"></span>
          </span>
          <span className="tracking-wide">Autonomous SRE Gateway • Google Cloud AI Builder Cup 2026</span>
        </motion.div>

        {/* Cinematic Masked Display Title */}
        <div className="flex justify-center mb-6">
          <MaskedTextReveal
            text="How NeuroProxy Stops Broken APIs From Costing $5,600 / Minute"
            highlightWords={['$5,600', '/', 'Minute']}
            className="text-4xl sm:text-5xl lg:text-6xl font-extrabold tracking-tight text-slate-900 dark:text-white leading-[1.1] justify-center text-center max-w-4xl"
            delay={0.15}
          />
        </div>

        {/* Smooth Subtitle Entrance */}
        <motion.p
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.7, delay: 0.45, ease: [0.16, 1, 0.3, 1] }}
          className="mt-6 text-lg sm:text-xl text-slate-600 dark:text-slate-300 font-normal leading-relaxed max-w-2xl mx-auto"
        >
          When an upstream service mutates its schema in production, downtime costs $300k/hour. 
          NeuroProxy intercepts breaking drifts, applies zero-hallucination math guardrails, and heals API traffic in <strong className="text-slate-900 dark:text-white font-semibold">0.12 milliseconds</strong>.
        </motion.p>

        {/* Action CTAs with Smooth Transitions */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.7, delay: 0.6, ease: [0.16, 1, 0.3, 1] }}
          className="mt-10 flex flex-wrap items-center justify-center gap-4"
        >
          <motion.button
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
            onClick={() => handleRunPreset(0)}
            disabled={isProcessing}
            className="px-8 py-4 rounded-2xl font-bold text-sm bg-gradient-to-r from-brand-orange via-amber-500 to-orange-500 text-slate-950 shadow-lg shadow-brand-orange/20 hover:shadow-brand-orange/35 hover:brightness-105 transition-all flex items-center gap-2.5 cursor-pointer disabled:opacity-50"
          >
            <Play className="w-4 h-4 fill-slate-950" />
            <span>{isProcessing ? 'Agent Reasoning...' : 'Simulate Drift Incident Live'}</span>
          </motion.button>

          <motion.button
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
            onClick={onSwitchToConsole}
            className="px-8 py-4 rounded-2xl font-bold text-sm bg-white dark:bg-canvas-cardDark border border-slate-200 dark:border-slate-800 text-slate-800 dark:text-slate-200 hover:bg-slate-50 dark:hover:bg-slate-800/80 shadow-card-light dark:shadow-card-dark transition-all flex items-center gap-2 cursor-pointer"
          >
            <span>Open Mission Control</span>
            <ChevronRight className="w-4 h-4 text-slate-400" />
          </motion.button>
        </motion.div>
      </section>

      {/* Sticky Floating Progress Navigation */}
      <div className="sticky top-20 z-30 flex items-center justify-center px-4">
        <motion.div
          initial={{ opacity: 0, y: -10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.8 }}
          className="p-1.5 rounded-full bg-white/95 dark:bg-canvas-cardDark/95 backdrop-blur-xl border border-slate-200/90 dark:border-slate-800 shadow-lg flex items-center gap-1 sm:gap-1.5 max-w-full overflow-x-auto"
        >
          {[
            { num: 1, label: 'Act 1: Breaking Drift' },
            { num: 2, label: 'Act 2: Gemini Synthesis' },
            { num: 3, label: 'Act 3: Math Guardrail' },
            { num: 4, label: 'Act 4: Live Delivery' },
            { num: 5, label: 'Act 5: 0.12ms Hot-Path' },
          ].map((step) => (
            <button
              key={step.num}
              onClick={() => {
                setActiveStep(step.num)
                const el = document.getElementById(`story-act-${step.num}`)
                if (el) el.scrollIntoView({ behavior: 'smooth', block: 'center' })
              }}
              className={`px-3.5 py-1.5 rounded-full text-xs font-semibold whitespace-nowrap transition-all cursor-pointer ${
                activeStep === step.num
                  ? 'bg-slate-900 text-white dark:bg-brand-orange dark:text-slate-950 shadow-md scale-105'
                  : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'
              }`}
            >
              {step.label}
            </button>
          ))}
        </motion.div>
      </div>

      {/* The 5 Acts - Stacking Smooth Cards */}
      <div className="max-w-5xl mx-auto space-y-16 px-4">
        {/* ACT 1 */}
        <motion.section
          id="story-act-1"
          onMouseEnter={() => setActiveStep(1)}
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: '-60px' }}
          transition={{ duration: 0.8, ease: [0.16, 1, 0.3, 1] }}
          className="scroll-mt-32 p-8 sm:p-10 rounded-3xl modern-card modern-card-hover border border-slate-200 dark:border-slate-800/90 relative overflow-hidden group"
        >
          <div className="flex items-center gap-2.5 text-xs font-bold tracking-wider uppercase text-rose-500 mb-3">
            <span className="w-2.5 h-2.5 rounded-full bg-rose-500 animate-ping" />
            <span>Act 1 • 00:00:00 — The Silent Upstream Mutation</span>
          </div>

          <h2 className="text-2xl sm:text-3xl lg:text-4xl font-bold text-slate-900 dark:text-white tracking-tight">
            A Third-Party Partner Modifies Their Webhook Schema
          </h2>

          <p className="mt-4 text-base text-slate-600 dark:text-slate-300 leading-relaxed max-w-3xl">
            Without prior RFC notice, an external vendor migrates to internationalized Spanish keys. 
            <code className="text-xs px-2 py-0.5 rounded-md bg-slate-100 dark:bg-slate-800 text-rose-500 font-mono mx-1">order_id</code> 
            becomes 
            <code className="text-xs px-2 py-0.5 rounded-md bg-slate-100 dark:bg-slate-800 text-rose-500 font-mono mx-1">numero_orden</code>, 
            and integer cents mutate into floating-point dollars.
          </p>

          <div className="mt-8 grid grid-cols-1 md:grid-cols-2 gap-5">
            <div className="p-6 rounded-2xl bg-rose-50/70 dark:bg-rose-950/20 border border-rose-200 dark:border-rose-900/40">
              <div className="flex items-center justify-between text-xs font-semibold text-rose-700 dark:text-rose-400 mb-3">
                <span>Incoming Mutated Payload</span>
                <span className="text-[10px] uppercase font-mono px-2 py-0.5 rounded-full bg-rose-200 dark:bg-rose-900/60 text-rose-800 dark:text-rose-300">
                  Drift Detected
                </span>
              </div>
              <pre className="text-xs font-mono text-rose-800 dark:text-rose-300/90 overflow-x-auto leading-relaxed">
{`{
  "numero_orden": "ORD-ES-789",
  "monto_usd": 149.50,
  "moneda": "USD",
  "cliente": {
    "correo_electronico": "sofia@empresa.es"
  }
}`}
              </pre>
            </div>

            <div className="p-6 rounded-2xl bg-slate-50 dark:bg-slate-900/60 border border-slate-200 dark:border-slate-800 flex flex-col justify-between">
              <div>
                <div className="flex items-center justify-between text-xs font-semibold text-slate-700 dark:text-slate-300 mb-3">
                  <span>Traditional Gateway Fate</span>
                  <span className="text-[10px] uppercase font-mono px-2 py-0.5 rounded-full bg-rose-500 text-white font-bold">
                    400 Bad Request Cascade
                  </span>
                </div>
                <p className="text-xs text-slate-600 dark:text-slate-400 leading-relaxed">
                  In standard gateways (Kong, Envoy, Apigee), strict JSON schema validation drops the request. 
                  Payment pipelines fail, triggering high-severity pager alerts and draining engineering hours.
                </p>
              </div>

              <div className="mt-6 pt-4 border-t border-slate-200 dark:border-slate-800 flex items-center justify-between">
                <span className="text-xs text-slate-500">Industry Mean-Time-To-Repair:</span>
                <span className="text-xs font-bold text-rose-600 dark:text-rose-400">4.2 Hours (Engineering Hotfix)</span>
              </div>
            </div>
          </div>
        </motion.section>

        {/* ACT 2 */}
        <motion.section
          id="story-act-2"
          onMouseEnter={() => setActiveStep(2)}
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: '-60px' }}
          transition={{ duration: 0.8, ease: [0.16, 1, 0.3, 1] }}
          className="scroll-mt-32 p-8 sm:p-10 rounded-3xl modern-card modern-card-hover border border-slate-200 dark:border-slate-800/90 relative overflow-hidden group"
        >
          <div className="flex items-center gap-2.5 text-xs font-bold tracking-wider uppercase text-brand-orange mb-3">
            <Cpu className="w-4 h-4 text-brand-orange" />
            <span>Act 2 • 00:00:01 — Gemini 3.8 Flash Autonomous Diagnosis</span>
          </div>

          <h2 className="text-2xl sm:text-3xl lg:text-4xl font-bold text-slate-900 dark:text-white tracking-tight">
            Zero-Shot Schema Synthesis with Gemini 3.8 Flash
          </h2>

          <p className="mt-4 text-base text-slate-600 dark:text-slate-300 leading-relaxed max-w-3xl">
            NeuroProxy intercepts the rejection in-flight. Gemini 3.8 Flash inspects the AST differences between the incoming JSON and the registered target schema contract, computing intent-preserving field alignments.
          </p>

          <div className="mt-8 p-6 rounded-2xl bg-amber-50/50 dark:bg-amber-950/20 border border-amber-200 dark:border-amber-900/40">
            <h4 className="text-xs font-bold uppercase tracking-wider text-amber-800 dark:text-amber-400 mb-4 flex items-center gap-2">
              <Sparkles className="w-4 h-4" />
              <span>Synthesized AST Mapping Rules Generated by Gemini 3.8 Flash</span>
            </h4>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3.5 text-xs font-mono">
              <div className="p-4 rounded-xl bg-white dark:bg-canvas-cardDark border border-slate-200 dark:border-slate-800">
                <span className="text-slate-400 text-[10px] block mb-1">Key Translation</span>
                <div className="text-rose-500">numero_orden</div>
                <div className="text-slate-400 my-1">↓ identity</div>
                <div className="text-emerald-500 font-bold">order_id</div>
              </div>
              <div className="p-4 rounded-xl bg-white dark:bg-canvas-cardDark border border-slate-200 dark:border-slate-800">
                <span className="text-slate-400 text-[10px] block mb-1">Currency & Unit Conversion</span>
                <div className="text-rose-500">monto_usd ($149.50)</div>
                <div className="text-slate-400 my-1">↓ multiply_100</div>
                <div className="text-emerald-500 font-bold">amount_cents (14950)</div>
              </div>
              <div className="p-4 rounded-xl bg-white dark:bg-canvas-cardDark border border-slate-200 dark:border-slate-800">
                <span className="text-slate-400 text-[10px] block mb-1">Nested Field Flattening</span>
                <div className="text-rose-500">cliente.correo_electronico</div>
                <div className="text-slate-400 my-1">↓ identity</div>
                <div className="text-emerald-500 font-bold">customer_email</div>
              </div>
            </div>
          </div>
        </motion.section>

        {/* ACT 3 */}
        <motion.section
          id="story-act-3"
          onMouseEnter={() => setActiveStep(3)}
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: '-60px' }}
          transition={{ duration: 0.8, ease: [0.16, 1, 0.3, 1] }}
          className="scroll-mt-32 p-8 sm:p-10 rounded-3xl modern-card modern-card-hover border border-slate-200 dark:border-slate-800/90 relative overflow-hidden group"
        >
          <div className="flex items-center gap-2.5 text-xs font-bold tracking-wider uppercase text-emerald-600 dark:text-emerald-400 mb-3">
            <ShieldCheck className="w-4 h-4 text-emerald-500" />
            <span>Act 3 • 00:00:02 — The Deterministic Value Conservation Guardrail</span>
          </div>

          <h2 className="text-2xl sm:text-3xl lg:text-4xl font-bold text-slate-900 dark:text-white tracking-tight">
            Zero-Hallucination Mathematical Verification
          </h2>

          <p className="mt-4 text-base text-slate-600 dark:text-slate-300 leading-relaxed max-w-3xl">
            In enterprise financial pipelines, LLM hallucinations are unacceptable. If an AI miscalculates $149.50 as $14.95 or fabricates discounts, catastrophic financial losses occur. 
            NeuroProxy solves this with a **zero-trust scalar provenance guardrail** executed strictly in compiled Go.
          </p>

          <div className="mt-8 grid grid-cols-1 sm:grid-cols-2 gap-5">
            <div className="p-6 rounded-2xl bg-emerald-50/70 dark:bg-emerald-950/20 border border-emerald-200 dark:border-emerald-900/40">
              <div className="flex items-center gap-2 text-xs font-bold text-emerald-700 dark:text-emerald-400 mb-3">
                <CheckCircle2 className="w-4 h-4" />
                <span>Deterministic Math Invariant Checks</span>
              </div>
              <ul className="text-xs text-slate-600 dark:text-slate-300 space-y-2.5 leading-relaxed">
                <li className="flex items-center gap-2">
                  <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 shrink-0" />
                  <span><strong>Cent Equivalence:</strong> |source * 100 - target| &lt; 0.001</span>
                </li>
                <li className="flex items-center gap-2">
                  <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 shrink-0" />
                  <span><strong>Provenance:</strong> Every field must mathematically trace to source</span>
                </li>
                <li className="flex items-center gap-2">
                  <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 shrink-0" />
                  <span><strong>Currency Integrity:</strong> ISO-4217 validation</span>
                </li>
              </ul>
            </div>

            <div className="p-6 rounded-2xl bg-rose-50/70 dark:bg-rose-950/20 border border-rose-200 dark:border-rose-900/40 flex flex-col justify-between">
              <div>
                <div className="flex items-center gap-2 text-xs font-bold text-rose-700 dark:text-rose-400 mb-3">
                  <AlertTriangle className="w-4 h-4" />
                  <span>Immediate Safety Veto (422 VETO)</span>
                </div>
                <p className="text-xs text-slate-600 dark:text-slate-300 leading-relaxed">
                  If Gemini hallucinates a numerical shift or generates an unauthorized discount, 
                  the Go guardrail intercepts and terminates execution with an immediate <code>422 SAFETY_VETO</code>. No corrupted data ever reaches the database.
                </p>
              </div>
              <span className="mt-4 text-[11px] font-mono text-rose-600 dark:text-rose-400 font-semibold">
                Guaranteed: 100.0% Financial Value Conservation
              </span>
            </div>
          </div>
        </motion.section>

        {/* ACT 4 */}
        <motion.section
          id="story-act-4"
          onMouseEnter={() => setActiveStep(4)}
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: '-60px' }}
          transition={{ duration: 0.8, ease: [0.16, 1, 0.3, 1] }}
          className="scroll-mt-32 p-8 sm:p-10 rounded-3xl modern-card modern-card-hover border border-slate-200 dark:border-slate-800/90 relative overflow-hidden group"
        >
          <div className="flex items-center gap-2.5 text-xs font-bold tracking-wider uppercase text-blue-600 dark:text-blue-400 mb-3">
            <Layers className="w-4 h-4 text-blue-500" />
            <span>Act 4 • 00:00:02 — Live Reverse-Proxy Delivery</span>
          </div>

          <h2 className="text-2xl sm:text-3xl lg:text-4xl font-bold text-slate-900 dark:text-white tracking-tight">
            Transparent In-Flight Payload Adaptation
          </h2>

          <p className="mt-4 text-base text-slate-600 dark:text-slate-300 leading-relaxed max-w-3xl">
            With mathematical safety verified, NeuroProxy dispatches the adapted payload to the destination microservice.
            The client experiences zero downtime, and the microservice receives a contract-compliant payload.
          </p>

          <div className="mt-8 p-6 rounded-2xl bg-blue-50/70 dark:bg-blue-950/20 border border-blue-200 dark:border-blue-900/40 flex flex-wrap items-center justify-between gap-4">
            <div>
              <span className="text-xs font-bold text-blue-800 dark:text-blue-300 block mb-1">
                Upstream Microservice Response
              </span>
              <p className="text-xs text-slate-600 dark:text-slate-300">
                The checkout service receives the perfectly formatted contract and acknowledges with <strong>HTTP 200 OK</strong>.
              </p>
            </div>
            <div className="flex items-center gap-3 font-mono text-xs">
              <span className="px-4 py-1.5 rounded-xl bg-emerald-500 text-white font-bold shadow-sm">
                200 OK
              </span>
              <span className="text-slate-500 dark:text-slate-400">
                Zero Client Modification
              </span>
            </div>
          </div>
        </motion.section>

        {/* ACT 5 */}
        <motion.section
          id="story-act-5"
          onMouseEnter={() => setActiveStep(5)}
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: '-60px' }}
          transition={{ duration: 0.8, ease: [0.16, 1, 0.3, 1] }}
          className="scroll-mt-32 p-8 sm:p-10 rounded-3xl modern-card modern-card-hover border border-slate-200 dark:border-slate-800/90 relative overflow-hidden group"
        >
          <div className="flex items-center gap-2.5 text-xs font-bold tracking-wider uppercase text-indigo-600 dark:text-indigo-400 mb-3">
            <Zap className="w-4 h-4 text-indigo-500" />
            <span>Act 5 • 00:00:03 onwards — Hot-Path Cache &amp; Permanent Codification</span>
          </div>

          <h2 className="text-2xl sm:text-3xl lg:text-4xl font-bold text-slate-900 dark:text-white tracking-tight">
            Sub-Millisecond Hot-Path Healing &amp; Automated PR Synthesis
          </h2>

          <p className="mt-4 text-base text-slate-600 dark:text-slate-300 leading-relaxed max-w-3xl">
            Gemini is not invoked on every request. The learned AST rule is committed to local in-memory LRU cache and Google Cloud Memorystore Redis.
            Subsequent drifted requests heal in <strong>0.12 milliseconds</strong>!
          </p>

          <div className="mt-8 grid grid-cols-1 sm:grid-cols-2 gap-5">
            <div className="p-6 rounded-2xl bg-slate-50 dark:bg-slate-900/60 border border-slate-200 dark:border-slate-800">
              <div className="flex items-center gap-2 text-xs font-bold text-indigo-700 dark:text-indigo-400 mb-2">
                <Clock className="w-4 h-4" />
                <span>100x Latency Reduction</span>
              </div>
              <p className="text-xs text-slate-600 dark:text-slate-400 leading-relaxed">
                Initial Gemini reasoning takes ~800ms. Subsequent requests are resolved via the two-tier cache in <strong>0.12ms</strong> with zero LLM API costs.
              </p>
            </div>

            <div className="p-6 rounded-2xl bg-slate-50 dark:bg-slate-900/60 border border-slate-200 dark:border-slate-800">
              <div className="flex items-center gap-2 text-xs font-bold text-purple-700 dark:text-purple-400 mb-2">
                <GitPullRequest className="w-4 h-4" />
                <span>Permanent GitHub PR Codification</span>
              </div>
              <p className="text-xs text-slate-600 dark:text-slate-400 leading-relaxed">
                While traffic flows uninterrupted, NeuroProxy synthesizes an atomic Go adapter PR patch. SREs can review and merge the permanent fix at their convenience during normal business hours.
              </p>
            </div>
          </div>

          <div className="mt-10 pt-8 border-t border-slate-200 dark:border-slate-800 flex flex-wrap items-center justify-between gap-5">
            <div>
              <h4 className="text-base font-bold text-slate-900 dark:text-white">
                Experience the Full Operational Cockpit
              </h4>
              <p className="text-xs text-slate-500 mt-1">
                Inject custom traffic, test live upstream endpoints, and benchmark 10x concurrent bursts.
              </p>
            </div>
            <motion.button
              whileHover={{ scale: 1.02 }}
              whileTap={{ scale: 0.98 }}
              onClick={onSwitchToConsole}
              className="px-6 py-3 rounded-2xl font-bold text-xs bg-slate-900 text-white dark:bg-white dark:text-slate-950 hover:opacity-90 transition-all flex items-center gap-2 shadow-md cursor-pointer"
            >
              <span>Switch to Mission Control Console</span>
              <ArrowRight className="w-4 h-4" />
            </motion.button>
          </div>
        </motion.section>
      </div>
    </div>
  )
}
