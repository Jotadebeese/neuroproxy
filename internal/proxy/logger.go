package proxy

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// CloudStructuredLog represents Google Cloud Run / Cloud Logging structured log payload
// matching: https://cloud.google.com/logging/docs/structured-logging
type CloudStructuredLog struct {
	Severity       string         `json:"severity"`
	Message        string         `json:"message"`
	Timestamp      string         `json:"timestamp"`
	Component      string         `json:"component"`
	Trace          string         `json:"logging.googleapis.com/trace,omitempty"`
	SpanID         string         `json:"logging.googleapis.com/spanId,omitempty"`
	TraceSampled   bool           `json:"logging.googleapis.com/trace_sampled,omitempty"`
	ServiceID      string         `json:"service_id,omitempty"`
	ResolutionMode string         `json:"resolution_mode,omitempty"`
	DurationMs     int64          `json:"duration_ms,omitempty"`
	AgentTurns     int            `json:"agent_turns,omitempty"`
	StatusCode     int            `json:"status_code,omitempty"`
	Fingerprint    string         `json:"schema_fingerprint,omitempty"`
	Extra          map[string]any `json:"extra,omitempty"`
}

// LogIncident emits a JSON structured line to stdout for Google Cloud Logging ingestion.
func LogIncident(rec *IncidentRecord, traceHeader string) {
	if rec == nil {
		return
	}

	severity := "INFO"
	switch rec.ResolutionMode {
	case "SAFETY_VETO":
		severity = "CRITICAL"
	case "BYPASS_FAILED":
		severity = "ERROR"
	case "GEMINI_SYNTHESIS":
		severity = "WARNING" // Drift occurred, required agentic healing
	case "CACHE_HIT", "VALID_PASS":
		severity = "INFO"
	}

	traceID, spanID, sampled := parseCloudTraceContext(traceHeader)

	msg := fmt.Sprintf("[%s] Service: %s | Mode: %s | Latency: %.2fms | Status: %d",
		severity, rec.ServiceID, rec.ResolutionMode, rec.TotalLatencyMs, rec.FinalStatus)

	if rec.RootCauseAnalysis != "" {
		msg += " | " + rec.RootCauseAnalysis
	}

	entry := CloudStructuredLog{
		Severity:       severity,
		Message:        msg,
		Timestamp:      time.Now().UTC().Format(time.RFC3339Nano),
		Component:      "neuroproxy-engine",
		Trace:          traceID,
		SpanID:         spanID,
		TraceSampled:   sampled,
		ServiceID:      rec.ServiceID,
		ResolutionMode: rec.ResolutionMode,
		DurationMs:     int64(rec.TotalLatencyMs),
		AgentTurns:     rec.AgentTurns,
		StatusCode:     rec.FinalStatus,
		Fingerprint:    rec.SchemaFingerprint,
	}

	if rec.GitHubPRTitle != "" {
		if entry.Extra == nil {
			entry.Extra = make(map[string]any)
		}
		entry.Extra["github_pr_title"] = rec.GitHubPRTitle
	}

	bytes, err := json.Marshal(entry)
	if err == nil {
		fmt.Println(string(bytes))
	}
}

// parseCloudTraceContext parses the standard Google Cloud `X-Cloud-Trace-Context` header:
// TRACE_ID/SPAN_ID;o=TRACE_TRUE
func parseCloudTraceContext(header string) (traceID string, spanID string, sampled bool) {
	if header == "" {
		return "", "", false
	}

	parts := strings.Split(header, ";")
	idParts := strings.Split(parts[0], "/")
	rawTrace := idParts[0]

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		projectID = os.Getenv("GCP_PROJECT")
	}
	if projectID == "" {
		projectID = "neuroproxy-prod"
	}

	traceID = fmt.Sprintf("projects/%s/traces/%s", projectID, rawTrace)

	if len(idParts) > 1 {
		spanID = idParts[1]
	}

	if len(parts) > 1 && strings.Contains(parts[1], "o=1") {
		sampled = true
	}

	return traceID, spanID, sampled
}
