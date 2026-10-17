import { BootstrapResponse, ProxyResponse, ServiceContract, Metrics, IncidentRecord, LearnedRule } from './types'

const API_BASE = ''

export async function fetchBootstrap(): Promise<BootstrapResponse> {
  const res = await fetch(`${API_BASE}/api/bootstrap`)
  if (!res.ok) {
    throw new Error(`Failed to bootstrap: ${res.statusText}`)
  }
  return res.json()
}

export async function sendProxyRequest(
  serviceId: string,
  payload: Record<string, any>,
  selfHeal: boolean = true,
  targetUrl?: string
): Promise<ProxyResponse> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }
  if (!selfHeal) {
    headers['X-NeuroProxy-Mode'] = 'bypass'
  }
  if (targetUrl && targetUrl.trim().length > 0) {
    headers['X-Target-URL'] = targetUrl.trim()
  }

  const res = await fetch(`${API_BASE}/api/proxy/${encodeURIComponent(serviceId)}`, {
    method: 'POST',
    headers,
    body: JSON.stringify({
      payload,
      self_heal: selfHeal,
      target_url: targetUrl || undefined,
    }),
  })

  if (!res.ok && res.status !== 422 && res.status !== 400) {
    const errorBody = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(errorBody.error || `Request failed with HTTP ${res.status}`)
  }

  return res.json()
}

export async function resetGatewayState(): Promise<{ status: string; metrics: Metrics }> {
  const res = await fetch(`${API_BASE}/api/reset`, { method: 'POST' })
  if (!res.ok) {
    throw new Error(`Reset failed: ${res.statusText}`)
  }
  return res.json()
}

export async function upsertServiceContract(contract: ServiceContract): Promise<{ status: string; contracts: ServiceContract[] }> {
  const res = await fetch(`${API_BASE}/api/contracts`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(contract),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error || `Upsert contract failed: ${res.statusText}`)
  }
  return res.json()
}

export async function fetchMetrics(): Promise<Metrics> {
  const res = await fetch(`${API_BASE}/api/metrics`)
  if (!res.ok) throw new Error('Failed to fetch metrics')
  return res.json()
}

export async function fetchIncidents(): Promise<IncidentRecord[]> {
  const res = await fetch(`${API_BASE}/api/incidents`)
  if (!res.ok) throw new Error('Failed to fetch incidents')
  return res.json()
}

export async function fetchRules(): Promise<LearnedRule[]> {
  const res = await fetch(`${API_BASE}/api/rules`)
  if (!res.ok) throw new Error('Failed to fetch rules')
  return res.json()
}
