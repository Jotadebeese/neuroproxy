package contracts

import (
	"time"
)

// ServiceContract defines the schema and endpoint for an upstream microservice.
type ServiceContract struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Endpoint       string            `json:"endpoint"`
	Description    string            `json:"description"`
	RequiredFields map[string]string `json:"required_fields"`
}

// FieldMapping describes how a drifted source field maps to the expected target schema field.
type FieldMapping struct {
	SourcePath string `json:"source_path"`
	TargetPath string `json:"target_path"`
	Transform  string `json:"transform"`
}

// GuardrailReport contains value conservation verification results.
type GuardrailReport struct {
	Passed            bool                   `json:"passed"`
	HallucinationVeto bool                   `json:"hallucination_veto"`
	SourceValues      map[string]interface{} `json:"source_values"`
	ExtractedValues   map[string]interface{} `json:"extracted_values"`
	ConservedCount    int                    `json:"conserved_count"`
	ViolationCount    int                    `json:"violation_count"`
	Reason            string                 `json:"reason"`
}

// TraceStep records an atomic milestone in the self-healing pipeline.
type TraceStep struct {
	Timestamp  time.Time `json:"timestamp"`
	Phase      string    `json:"phase"`
	Detail     string    `json:"detail"`
	DurationMs float64   `json:"duration_ms"`
}

// LearnedRule is stored in the thread-safe RuleCache keyed by structural schema fingerprint.
type LearnedRule struct {
	Fingerprint   string         `json:"fingerprint"`
	ServiceID     string         `json:"service_id"`
	Description   string         `json:"description"`
	Mappings      []FieldMapping `json:"mappings"`
	GitHubPRTitle string         `json:"github_pr_title"`
	GitHubPRDiff  string         `json:"github_pr_diff"`
	HitCount      int64          `json:"hit_count"`
	CreatedAt     time.Time      `json:"created_at"`
	LastUsedAt    time.Time      `json:"last_used_at"`
}

// IncidentRecord captures a complete self-healing event for the Mission Control inspector.
type IncidentRecord struct {
	ID                string                 `json:"id"`
	Timestamp         time.Time              `json:"timestamp"`
	ServiceID         string                 `json:"service_id"`
	ServiceName       string                 `json:"service_name"`
	SchemaFingerprint string                 `json:"schema_fingerprint"`
	ResolutionMode    string                 `json:"resolution_mode"` // "GEMINI_SYNTHESIS" | "CACHE_HIT" | "BYPASS_FAILED" | "VALID_PASS" | "SAFETY_VETO"
	OriginalStatus    int                    `json:"original_status"`
	FinalStatus       int                    `json:"final_status"`
	AgentTurns        int                    `json:"agent_turns"`
	UpstreamError     string                 `json:"upstream_error"`
	OriginalPayload   map[string]interface{} `json:"original_payload"`
	HealedPayload     map[string]interface{} `json:"healed_payload,omitempty"`
	UpstreamResponse  map[string]interface{} `json:"upstream_response"`
	Mappings          []FieldMapping         `json:"mappings,omitempty"`
	Guardrail         GuardrailReport        `json:"guardrail"`
	RootCauseAnalysis string                 `json:"root_cause_analysis,omitempty"`
	GitHubPRTitle     string                 `json:"github_pr_title,omitempty"`
	GitHubPRDiff      string                 `json:"github_pr_diff,omitempty"`
	TotalLatencyMs    float64                `json:"total_latency_ms"`
	HealingLatencyMs  float64                `json:"healing_latency_ms"`
	Trace             []TraceStep            `json:"trace"`
	ModelUsed         string                 `json:"model_used"`
}

// GatewayMetrics exposes live real-time telemetry for the dashboard.
type GatewayMetrics struct {
	TotalRequests              int64   `json:"total_requests"`
	SelfHealedRequests         int64   `json:"self_healed_requests"`
	ValidPassRequests          int64   `json:"valid_pass_requests"`
	BypassedFailures           int64   `json:"bypassed_failures"`
	CacheHits                  int64   `json:"cache_hits"`
	CacheMisses                int64   `json:"cache_misses"`
	SafetyVetoes               int64   `json:"safety_vetoes"`
	AvgGeminiLatencyMs         float64 `json:"avg_gemini_latency_ms"`
	AvgCacheLatencyMs          float64 `json:"avg_cache_latency_ms"`
	EstimatedDowntimeSavedMin  float64 `json:"estimated_downtime_saved_min"`
}

// ChaosPreset defines a reproducible payload drift scenario for 1-click evaluation.
type ChaosPreset struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	ServiceID     string                 `json:"service_id"`
	Description   string                 `json:"description"`
	DriftScenario string                 `json:"drift_scenario"`
	Payload       map[string]interface{} `json:"payload"`
}

// GetChaosPresets returns the 6 realistic real-world drift scenarios for live demonstration.
func GetChaosPresets() []ChaosPreset {
	return []ChaosPreset{
		{
			ID:            "preset-stripe-spanish",
			Name:          "Stripe Partner Webhook Drift (Spanish Keys & Dollar Floats)",
			ServiceID:     "checkout",
			Description:   "LATAM client sends Spanish localized keys (numero_orden, monto_usd) with floating-point dollar price ($149.50) instead of integer cents.",
			DriftScenario: "Translates keys to canonical order_id, customer_email, currency, and converts monto_usd (149.50) to amount_cents (14950) with 100% Value Conservation.",
			Payload: map[string]interface{}{
				"numero_orden": "ORD-LATAM-94821",
				"pago": map[string]interface{}{
					"monto_usd": 149.50,
					"moneda":    "USD",
				},
				"cliente": map[string]interface{}{
					"correo_electronico": "sofia.garcia@empresa.mx",
					"nombre":             "Sofia Garcia",
				},
			},
		},
		{
			ID:            "preset-cascading-kyc",
			Name:          "2-Layer AML Cascade (Drift Healed -> 409 KYC Required)",
			ServiceID:     "checkout",
			Description:   "High-value institutional payment ($250,000) with drifted schema that triggers a Layer 2 domain compliance veto on Turn 1.",
			DriftScenario: "Demonstrates true multi-turn autonomous agent self-correction. Resolves schema drift on Turn 1, intercepts upstream 409, extracts KYC verification, and succeeds on Turn 2.",
			Payload: map[string]interface{}{
				"id": "INST-TX-88301",
				"payment": map[string]interface{}{
					"amount_dollars": 250000.00,
					"iso_currency":   "USD",
				},
				"buyer": map[string]interface{}{
					"contact": map[string]interface{}{
						"email": "treasury@acme-corp.com",
					},
				},
				"kyc_aml_estado": map[string]interface{}{
					"verified": true,
					"provider": "Trulioo-Enterprise",
					"score":    99.4,
				},
			},
		},
		{
			ID:            "preset-hallucination-veto",
			Name:          "Safety Guardrail: Value Conservation Veto",
			ServiceID:     "checkout",
			Description:   "Adversarial request completely missing the required purchase price. Detects potential hallucination and blocks malicious data fabrication.",
			DriftScenario: "Triggers deterministic mathematical guardrail. Halts pipeline and returns HTTP 422 SAFETY_VETO, protecting downstream ledgers from corrupted data.",
			Payload: map[string]interface{}{
				"order_number": "ORD-CORRUPTED-NULL",
				"currency":     "USD",
				"user": map[string]interface{}{
					"email": "attacker@darkweb.xyz",
				},
				"note": "Payment amount field deliberately omitted to test hallucination resistance",
			},
		},
		{
			ID:            "preset-legacy-erp",
			Name:          "Legacy ERP Migration (Units to Integer Coercion)",
			ServiceID:     "inventory",
			Description:   "Supply chain ERP sends warehouse reserve request using SAP naming (item_sku, units as string '42.0') instead of strict integer quantity.",
			DriftScenario: "Extracts item_sku -> sku_code, coerces float string '42.0' to integer 42, and maps express delivery to priority_tier.",
			Payload: map[string]interface{}{
				"item_sku": "SKU-TITAN-X900",
				"units":    "42.0",
				"warehouse": map[string]interface{}{
					"node_id":  "WH-EAST-04",
					"dispatch": "EXPRESS_OVERNIGHT",
				},
			},
		},
		{
			ID:            "preset-iot-telemetry",
			Name:          "IoT Firmware Drift (Boolean/Float Inversion)",
			ServiceID:     "telemetry",
			Description:   "Factory edge sensor firmware v3.2 emits status as boolean flag instead of strict numeric 1.0/0.0.",
			DriftScenario: "Maps sensor payload and coerces booleans into strict domain float ratings with zero data loss.",
			Payload: map[string]interface{}{
				"device_serial": "EDGE-THERMO-118",
				"online_status": "ONLINE",
				"metrics": map[string]interface{}{
					"core_temp_celsius": 73.4,
					"status_rating":     1.0,
				},
			},
		},
		{
			ID:            "preset-valid-control",
			Name:          "Healthy Control (100% Contract Compliant)",
			ServiceID:     "checkout",
			Description:   "Standard perfectly formatted checkout payload. Bypasses AI synthesis and executes in 0ms with 200 OK.",
			DriftScenario: "Control benchmark demonstrating that valid traffic experiences zero proxy overhead.",
			Payload: map[string]interface{}{
				"order_id":            "ORD-CLEAN-001",
				"amount_cents":        4999,
				"currency":            "USD",
				"customer_email":      "alex@example.com",
				"compliance_verified": true,
			},
		},
	}
}
