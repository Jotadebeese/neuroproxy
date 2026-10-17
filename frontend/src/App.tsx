import React, { useState, useEffect } from 'react'
import { Header } from './components/Header'
import { MetricsBar } from './components/MetricsBar'
import { FlowVisualizer } from './components/FlowVisualizer'
import { PayloadPlayground } from './components/PayloadPlayground'
import { LiveDiffViewer } from './components/LiveDiffViewer'
import { IncidentsTimeline } from './components/IncidentsTimeline'
import { RulesTable } from './components/RulesTable'
import { ContractEditorModal } from './components/ContractEditorModal'
import {
  fetchBootstrap,
  sendProxyRequest,
  resetGatewayState,
  upsertServiceContract,
} from './api'
import {
  BootstrapResponse,
  ChaosPreset,
  ServiceContract,
  Metrics,
  IncidentRecord,
  LearnedRule,
} from './types'

export const App: React.FC = () => {
  const [contracts, setContracts] = useState<ServiceContract[]>([])
  const [presets, setPresets] = useState<ChaosPreset[]>([])
  const [metrics, setMetrics] = useState<Metrics>({
    total_requests: 0,
    self_healed_requests: 0,
    valid_pass_requests: 0,
    bypassed_failures: 0,
    cache_hits: 0,
    cache_misses: 0,
    safety_vetoes: 0,
    avg_gemini_latency_ms: 0,
    avg_cache_latency_ms: 0,
    estimated_downtime_saved_min: 0,
  })
  const [incidents, setIncidents] = useState<IncidentRecord[]>([])
  const [rules, setRules] = useState<LearnedRule[]>([])
  const [lastIncident, setLastIncident] = useState<IncidentRecord | null>(null)
  const [selectedIncident, setSelectedIncident] = useState<IncidentRecord | null>(null)
  const [isLoading, setIsLoading] = useState<boolean>(false)
  const [isBenchmarking, setIsBenchmarking] = useState<boolean>(false)
  const [isContractModalOpen, setIsContractModalOpen] = useState<boolean>(false)

  const loadData = async () => {
    try {
      const data = await fetchBootstrap()
      setContracts(data.contracts || [])
      setPresets(data.presets || [])
      setMetrics(data.metrics)
      setIncidents(data.incidents || [])
      setRules(data.rules || [])
      if (data.incidents && data.incidents.length > 0) {
        setLastIncident(data.incidents[0])
        setSelectedIncident(data.incidents[0])
      }
    } catch (err) {
      console.error('Failed to bootstrap:', err)
    }
  }

  useEffect(() => {
    loadData()
  }, [])

  const handleSendRequest = async (
    serviceId: string,
    payload: Record<string, any>,
    selfHeal: boolean,
    targetUrl?: string
  ) => {
    setIsLoading(true)
    try {
      const res = await sendProxyRequest(serviceId, payload, selfHeal, targetUrl)
      setLastIncident(res.incident)
      setSelectedIncident(res.incident)
      setMetrics(res.metrics)
      setRules(res.rules || [])
      setIncidents((prev) => [res.incident, ...prev.filter((i) => i.id !== res.incident.id)])
    } catch (err: any) {
      alert(`Request execution failed: ${err.message}`)
    } finally {
      setIsLoading(false)
    }
  }

  const handleBurstBenchmark = async () => {
    setIsBenchmarking(true)
    try {
      // Pick the first preset (checkout with Spanish drift)
      const preset = presets[0] || {
        service_id: 'checkout',
        payload: {
          numero_orden: 'ORD-BURST-100',
          monto_usd: 129.99,
          moneda: 'USD',
          cliente: { correo_electronico: 'burst@test.io' },
        },
      }

      // Execute 1st request to guarantee cache compilation if empty
      await sendProxyRequest(preset.service_id, preset.payload, true)

      // Burst 10 requests concurrently
      const promises = Array.from({ length: 10 }).map((_, i) =>
        sendProxyRequest(preset.service_id, {
          ...preset.payload,
          numero_orden: `ORD-BURST-${i + 1}`,
        }, true)
      )

      const results = await Promise.all(promises)
      if (results.length > 0) {
        const latest = results[results.length - 1]
        setLastIncident(latest.incident)
        setSelectedIncident(latest.incident)
        setMetrics(latest.metrics)
        setRules(latest.rules || [])
        // Refresh full incident log
        await loadData()
      }
    } catch (err: any) {
      console.error('Benchmark failed:', err)
    } finally {
      setIsBenchmarking(false)
    }
  }

  const handleReset = async () => {
    try {
      const res = await resetGatewayState()
      setMetrics(res.metrics)
      setIncidents([])
      setRules([])
      setLastIncident(null)
      setSelectedIncident(null)
    } catch (err: any) {
      alert('Failed to reset: ' + err.message)
    }
  }

  const handleSaveContract = async (contract: ServiceContract) => {
    const res = await upsertServiceContract(contract)
    setContracts(res.contracts)
  }

  return (
    <div className="min-h-screen bg-canvas-dark text-slate-100 flex flex-col selection:bg-brand-orange/30">
      {/* Top Header */}
      <Header
        onReset={handleReset}
        onBurstBenchmark={handleBurstBenchmark}
        onOpenContractModal={() => setIsContractModalOpen(true)}
        isBenchmarking={isBenchmarking}
      />

      {/* Main Content Dashboard */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6 space-y-6">
        {/* Telemetry Metrics Bar */}
        <MetricsBar metrics={metrics} />

        {/* Dynamic Pipeline Flow Graph */}
        <FlowVisualizer
          lastIncident={lastIncident}
          isProcessing={isLoading || isBenchmarking}
        />

        {/* Traffic Injection & Live Diff Viewer (2-Column Grid) */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
          <div className="lg:col-span-5 flex flex-col">
            <PayloadPlayground
              presets={presets}
              contracts={contracts}
              onSend={handleSendRequest}
              isLoading={isLoading || isBenchmarking}
            />
          </div>
          <div className="lg:col-span-7 flex flex-col">
            <LiveDiffViewer incident={selectedIncident || lastIncident} />
          </div>
        </div>

        {/* Incidents Timeline & Active Rule Cache (2-Column Grid) */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <IncidentsTimeline
            incidents={incidents}
            selectedIncident={selectedIncident}
            onSelectIncident={(inc) => setSelectedIncident(inc)}
          />
          <RulesTable rules={rules} />
        </div>
      </main>

      {/* Contract Editor Modal */}
      <ContractEditorModal
        contracts={contracts}
        isOpen={isContractModalOpen}
        onClose={() => setIsContractModalOpen(false)}
        onSaveContract={handleSaveContract}
      />

      {/* Footer */}
      <footer className="border-t border-white/5 py-4 px-6 text-center text-xs font-mono text-slate-500 bg-canvas-cardDark/50">
        Google Cloud AI Builder Cup 2026 • Autonomous Self-Healing API Gateway • Built with Go, Cloud Run &amp; Gemini 2.5 Flash
      </footer>
    </div>
  )
}
