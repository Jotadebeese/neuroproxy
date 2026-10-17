import React, { useState, useEffect } from 'react'
import { ChaosPreset, ServiceContract } from '../types'
import { Play, Sparkles, AlertTriangle, ShieldAlert, Code2, Globe, ArrowRight, Check } from 'lucide-react'

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

  // Initialize with first preset
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
    <div className="rounded-xl border border-white/10 bg-canvas-cardDark/70 backdrop-blur-md p-5 flex flex-col h-full">
      <div className="flex items-center justify-between pb-3 border-b border-white/10 mb-4">
        <div className="flex items-center gap-2">
          <Code2 className="w-4 h-4 text-brand-orange" />
          <h2 className="text-sm font-semibold text-white tracking-tight">
            Traffic Injection & Chaos Simulator
          </h2>
        </div>
        <div className="flex items-center gap-2">
          <label className="flex items-center gap-1.5 cursor-pointer text-xs font-mono">
            <span className={selfHeal ? 'text-emerald-400 font-semibold' : 'text-slate-400'}>
              {selfHeal ? 'Autonomous Healer (ON)' : 'Bypass / Native Reject (OFF)'}
            </span>
            <input
              type="checkbox"
              checked={selfHeal}
              onChange={(e) => setSelfHeal(e.target.checked)}
              className="sr-only peer"
            />
            <div className="w-9 h-5 bg-slate-800 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-slate-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-emerald-600 relative"></div>
          </label>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="flex flex-col flex-1 gap-3.5">
        {/* Preset Selector */}
        <div>
          <label className="block text-xs font-mono text-slate-400 mb-1.5">
            Scenario Preset / Schema Drift Pattern:
          </label>
          <select
            value={selectedPresetId}
            onChange={(e) => handlePresetSelect(e.target.value)}
            className="w-full bg-canvas-subtleDark border border-white/10 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-brand-orange font-mono"
          >
            {presets.map((p) => (
              <option key={p.id} value={p.id}>
                [{p.service_id.toUpperCase()}] {p.name}
              </option>
            ))}
          </select>
        </div>

        {/* Target Service ID & Live Upstream URL */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <div>
            <label className="block text-xs font-mono text-slate-400 mb-1">
              Target Contract ID:
            </label>
            <input
              type="text"
              value={serviceId}
              onChange={(e) => setServiceId(e.target.value)}
              placeholder="checkout, inventory, telemetry, or custom"
              className="w-full bg-canvas-subtleDark border border-white/10 rounded-lg px-3 py-1.5 text-xs text-white focus:outline-none focus:border-brand-orange font-mono"
            />
          </div>

          <div>
            <label className="block text-xs font-mono text-slate-400 mb-1 flex items-center justify-between">
              <span className="flex items-center gap-1">
                <Globe className="w-3 h-3 text-cyan-400" />
                Live External Upstream URL (Optional):
              </span>
              <span className="text-[10px] text-slate-500">BYO API</span>
            </label>
            <input
              type="text"
              value={targetUrl}
              onChange={(e) => setTargetUrl(e.target.value)}
              placeholder="e.g. https://httpbin.org/post or mock"
              className="w-full bg-canvas-subtleDark border border-white/10 rounded-lg px-3 py-1.5 text-xs text-white focus:outline-none focus:border-brand-orange font-mono placeholder:text-slate-600"
            />
          </div>
        </div>

        {/* Payload JSON Editor */}
        <div className="flex-1 flex flex-col min-h-[220px]">
          <div className="flex items-center justify-between mb-1">
            <span className="text-xs font-mono text-slate-400">
              Incoming Drifted Payload (HTTP Body):
            </span>
            <button
              type="button"
              onClick={handleFormat}
              className="text-[11px] font-mono text-brand-orange hover:underline cursor-pointer"
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
              rows={10}
              className="w-full h-full bg-[#0a0a0c] border border-white/10 rounded-lg p-3 text-xs font-mono text-emerald-400 focus:outline-none focus:border-brand-orange resize-none selection:bg-brand-orange/30"
              spellCheck={false}
            />
          </div>
          {jsonError && (
            <p className="mt-1 text-xs text-rose-400 font-mono flex items-center gap-1">
              <AlertTriangle className="w-3.5 h-3.5" />
              {jsonError}
            </p>
          )}
        </div>

        {/* Action Button */}
        <div className="pt-2">
          <button
            type="submit"
            disabled={isLoading}
            className={`w-full py-2.5 px-4 rounded-xl font-mono text-xs font-bold uppercase tracking-wider flex items-center justify-center gap-2 transition-all shadow-lg active:scale-[0.99] cursor-pointer disabled:opacity-50 ${
              selfHeal
                ? 'bg-gradient-to-r from-brand-orange to-amber-500 text-slate-950 hover:brightness-110 shadow-glow-orange'
                : 'bg-rose-600/80 hover:bg-rose-600 text-white shadow-rose-900/30'
            }`}
          >
            {isLoading ? (
              <>
                <div className="w-3.5 h-3.5 border-2 border-slate-950 border-t-transparent rounded-full animate-spin" />
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
