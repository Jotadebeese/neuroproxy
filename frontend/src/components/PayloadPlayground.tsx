import React, { useState, useEffect } from 'react'
import { ChaosPreset, ServiceContract } from '../types'
import { Sparkles, ShieldAlert, Code2, Globe, AlertTriangle } from 'lucide-react'

interface PayloadPlaygroundProps {
  presets: ChaosPreset[]
  contracts: ServiceContract[]
  onSend: (serviceId: string, payload: Record<string, any>, selfHeal: boolean, targetUrl?: string) => Promise<void>
  isLoading: boolean
}

export const PayloadPlayground: React.FC<PayloadPlaygroundProps> = ({
  presets,
  contracts,
  onSend,
  isLoading,
}) => {
  const [selectedPresetId, setSelectedPresetId] = useState<string>('')
  const [serviceId, setServiceId] = useState<string>('checkout')
  const [payloadText, setPayloadText] = useState<string>('{}')
  const [selfHeal, setSelfHeal] = useState<boolean>(true)
  const [targetUrl, setTargetUrl] = useState<string>('')
  const [jsonError, setJsonError] = useState<string | null>(null)

  useEffect(() => {
    if (presets.length > 0 && !selectedPresetId) {
      const p = presets[0]
      setSelectedPresetId(p.id)
      setServiceId(p.service_id)
      setPayloadText(JSON.stringify(p.payload, null, 2))
    }
  }, [presets])

  const handlePresetSelect = (presetId: string) => {
    setSelectedPresetId(presetId)
    const preset = presets.find((p) => p.id === presetId)
    if (preset) {
      setServiceId(preset.service_id)
      setPayloadText(JSON.stringify(preset.payload, null, 2))
      setJsonError(null)
    }
  }

  const handleFormat = () => {
    try {
      const parsed = JSON.parse(payloadText)
      setPayloadText(JSON.stringify(parsed, null, 2))
      setJsonError(null)
    } catch (err: any) {
      setJsonError(err.message)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      const parsed = JSON.parse(payloadText)
      setJsonError(null)
      await onSend(serviceId, parsed, selfHeal, targetUrl.trim() || undefined)
    } catch (err: any) {
      setJsonError('Invalid JSON: ' + err.message)
    }
  }

  return (
    <div className="rounded-2xl modern-card p-6 flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between pb-4 border-b border-slate-200 dark:border-slate-800 mb-5">
        <div className="flex items-center gap-2.5">
          <div className="p-2 rounded-xl bg-orange-50 dark:bg-orange-950/40 border border-orange-200 dark:border-orange-800/40 text-brand-orange">
            <Code2 className="w-4 h-4" />
          </div>
          <div>
            <h2 className="text-base font-bold text-slate-900 dark:text-white tracking-tight">
              Traffic Injection &amp; Chaos
            </h2>
            <p className="text-xs text-slate-500 dark:text-slate-400">
              Simulate real-time schema mutations and drift
            </p>
          </div>
        </div>

        {/* Self-Healing Switch */}
        <div className="flex items-center gap-2">
          <label className="flex items-center gap-2 cursor-pointer text-xs font-medium">
            <span className={selfHeal ? 'text-emerald-600 dark:text-emerald-400 font-semibold' : 'text-slate-500 dark:text-slate-400'}>
              {selfHeal ? 'Healer Active' : 'Native Reject'}
            </span>
            <input
              type="checkbox"
              checked={selfHeal}
              onChange={(e) => setSelfHeal(e.target.checked)}
              className="sr-only peer"
            />
            <div className="w-10 h-5 bg-slate-300 dark:bg-slate-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-emerald-500 relative"></div>
          </label>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="flex flex-col flex-1 gap-4">
        {/* Scenario Preset Dropdown */}
        <div>
          <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1.5">
            Scenario Preset / Mutation Pattern:
          </label>
          <select
            value={selectedPresetId}
            onChange={(e) => handlePresetSelect(e.target.value)}
            className="w-full bg-slate-50 dark:bg-canvas-subtleDark border border-slate-200 dark:border-slate-800 rounded-xl px-3.5 py-2.5 text-xs text-slate-800 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-brand-orange/40 font-medium"
          >
            {presets.map((p) => (
              <option key={p.id} value={p.id}>
                [{p.service_id.toUpperCase()}] {p.name}
              </option>
            ))}
          </select>
        </div>

        {/* Target Contract ID & BYO API Upstream URL */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3.5">
          <div>
            <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1.5">
              Target Contract ID:
            </label>
            <input
              type="text"
              value={serviceId}
              onChange={(e) => setServiceId(e.target.value)}
              placeholder="e.g. checkout, inventory"
              className="w-full bg-slate-50 dark:bg-canvas-subtleDark border border-slate-200 dark:border-slate-800 rounded-xl px-3.5 py-2 text-xs text-slate-800 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-brand-orange/40 font-mono"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1.5 flex items-center justify-between">
              <span className="flex items-center gap-1.5">
                <Globe className="w-3.5 h-3.5 text-cyan-500" />
                <span>External Upstream URL:</span>
              </span>
              <span className="text-[10px] text-slate-400 font-normal">Optional</span>
            </label>
            <input
              type="text"
              value={targetUrl}
              onChange={(e) => setTargetUrl(e.target.value)}
              placeholder="https://httpbin.org/post"
              className="w-full bg-slate-50 dark:bg-canvas-subtleDark border border-slate-200 dark:border-slate-800 rounded-xl px-3.5 py-2 text-xs text-slate-800 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-brand-orange/40 font-mono placeholder:text-slate-400"
            />
          </div>
        </div>

        {/* Payload JSON Editor */}
        <div className="flex-1 flex flex-col min-h-[200px]">
          <div className="flex items-center justify-between mb-1.5">
            <span className="text-xs font-semibold text-slate-700 dark:text-slate-300">
              Incoming HTTP Request Body:
            </span>
            <button
              type="button"
              onClick={handleFormat}
              className="text-xs font-medium text-brand-orange hover:underline cursor-pointer"
            >
              Format JSON
            </button>
          </div>
          <div className="relative flex-1">
            <textarea
              value={payloadText}
              onChange={(e) => {
                setPayloadText(e.target.value)
                setJsonError(null)
              }}
              rows={8}
              className="w-full h-full bg-slate-900 text-emerald-400 border border-slate-800 rounded-xl p-3.5 text-xs font-mono focus:outline-none focus:ring-2 focus:ring-brand-orange/40 resize-none selection:bg-brand-orange/30 leading-relaxed"
              spellCheck={false}
            />
          </div>
          {jsonError && (
            <p className="mt-1.5 text-xs text-rose-500 font-medium flex items-center gap-1">
              <AlertTriangle className="w-3.5 h-3.5" />
              {jsonError}
            </p>
          )}
        </div>

        {/* Submit Action */}
        <div className="pt-2">
          <button
            type="submit"
            disabled={isLoading}
            className={`w-full py-3 px-4 rounded-xl font-bold text-xs uppercase tracking-wider flex items-center justify-center gap-2 transition-all shadow-md active:scale-[0.99] cursor-pointer disabled:opacity-50 ${
              selfHeal
                ? 'bg-gradient-to-r from-brand-orange to-amber-500 text-slate-950 hover:brightness-105 shadow-glow-orange'
                : 'bg-rose-600 hover:bg-rose-700 text-white'
            }`}
          >
            {isLoading ? (
              <>
                <div className="w-4 h-4 border-2 border-slate-950 border-t-transparent rounded-full animate-spin" />
                <span>Simulating Autonomous Agent Turn...</span>
              </>
            ) : selfHeal ? (
              <>
                <Sparkles className="w-4 h-4 text-slate-950" />
                <span>Inject Traffic &amp; Execute Autonomous Healing</span>
              </>
            ) : (
              <>
                <ShieldAlert className="w-4 h-4 text-white" />
                <span>Send Without Self-Healing (Expect Upstream 400 Failure)</span>
              </>
            )}
          </button>
        </div>
      </form>
    </div>
  )
}
