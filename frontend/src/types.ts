export interface FieldMapping {
  source_path: string
  target_path: string
  transform: string
}

export interface GuardrailReport {
  passed: boolean
  hallucination_veto: boolean
  source_values: Record<string, any>
  extracted_values: Record<string, any>
  conserved_count: number
  violation_count: number
  reason: string
}

export interface IncidentRecord {
  id: string
  timestamp: string
  service_id: string
  service_name?: string
  schema_fingerprint: string
  resolution_mode: 'GEMINI_SYNTHESIS' | 'CACHE_HIT' | 'VALID_PASS' | 'BYPASS_FAILED' | 'SAFETY_VETO' | string
  original_status: number
  final_status: number
  agent_turns: number
  upstream_error?: string
  original_payload: Record<string, any>
  healed_payload?: Record<string, any>
  upstream_response?: any
  mappings?: FieldMapping[]
  guardrail?: GuardrailReport
  root_cause_analysis?: string
  github_pr_title?: string
  github_pr_diff?: string
  total_latency_ms: number
  healing_latency_ms?: number
  model_used?: string
}

export interface Metrics {
  total_requests: number
  self_healed_requests: number
  valid_pass_requests: number
  bypassed_failures: number
  cache_hits: number
  cache_misses: number
  safety_vetoes: number
  avg_gemini_latency_ms: number
  avg_cache_latency_ms: number
  estimated_downtime_saved_min: number
}

export interface ServiceContract {
  id: string
  name: string
  endpoint: string
  description: string
  required_fields: Record<string, string>
}

export interface ChaosPreset {
  id: string
  name: string
  service_id: string
  description: string
  drift_scenario: string
  payload: Record<string, any>
}

export interface LearnedRule {
  fingerprint: string
  service_id: string
  description: string
  mappings: FieldMapping[]
  github_pr_title?: string
  github_pr_diff?: string
  hit_count: number
  created_at: string
  last_used_at: string
}

export interface BootstrapResponse {
  contracts: ServiceContract[]
  presets: ChaosPreset[]
  metrics: Metrics
  incidents: IncidentRecord[]
  rules: LearnedRule[]
}

export interface ProxyResponse {
  incident: IncidentRecord
  metrics: Metrics
  rules: LearnedRule[]
}
