package proxy

import "time"

// ServiceContract defines the strict schema contract expected by a downstream microservice.
type ServiceContract struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Endpoint       string            `json:"endpoint"`
	Description    string            `json:"description"`
	RequiredFields map[string]string `json:"required_fields"`
	SampleValidJSON string           `json:"sample_valid_json"`
}

// TraceStep represents a single stage in the NeuroProxy interception & healing pipeline.
type TraceStep struct {
	Step        int    `json:"step"`
	Phase       string `json:"phase"`
	Status      string `json:"status"` // "ok", "error", "warn", "healed", "cache_hit"
	Detail      string `json:"detail"`
	DurationMs  float64 `json:"duration_ms"`
}

// FieldMapping describes how a drifted source field maps to the expected target schema field.
type FieldMapping struct {
	SourcePath string `json:"source_path"`
	TargetPath string `json:"target_path"`
	Transform  string `json:"transform"`
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
	ResolutionMode    string                 `json:"resolution_mode"` // "GEMINI_SYNTHESIS" | "CACHE_HIT" | "BYPASS_FAILED" | "VALID_PASS"
	OriginalStatus    int                    `json:"original_status"`
	FinalStatus       int                    `json:"final_status"`
	UpstreamError     string                 `json:"upstream_error"`
	OriginalPayload   map[string]interface{} `json:"original_payload"`
	HealedPayload     map[string]interface{} `json:"healed_payload,omitempty"`
	UpstreamResponse  map[string]interface{} `json:"upstream_response"`
	Mappings          []FieldMapping         `json:"mappings,omitempty"`
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
	TotalRequests        int64   `json:"total_requests"`
	InterceptedErrors    int64   `json:"intercepted_errors"`
	GeminiHealedCount    int64   `json:"gemini_healed_count"`
	CacheHitHealedCount  int64   `json:"cache_hit_healed_count"`
	DirectBypassFailures int64   `json:"direct_bypass_failures"`
	ActiveCachedRules    int     `json:"active_cached_rules"`
	AvgGeminiLatencyMs   float64 `json:"avg_gemini_latency_ms"`
	AvgCacheLatencyMs    float64 `json:"avg_cache_latency_ms"`
	UptimeRatePercent    float64 `json:"uptime_rate_percent"`
	GeminiLiveConfigured bool    `json:"gemini_live_configured"`
}

// ChaosPreset provides 1-click realistic schema drift scenarios in the UI.
type ChaosPreset struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Subtitle    string                 `json:"subtitle"`
	ServiceID   string                 `json:"service_id"`
	Badge       string                 `json:"badge"`
	Description string                 `json:"description"`
	Payload     map[string]interface{} `json:"payload"`
}
