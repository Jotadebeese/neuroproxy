package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Jotadebeese/neuroproxy/internal/config"
	"github.com/Jotadebeese/neuroproxy/internal/contracts"
	"github.com/Jotadebeese/neuroproxy/internal/engine"
	"github.com/Jotadebeese/neuroproxy/internal/telemetry"
)

// Gateway provides the HTTP server router, middleware, and API endpoints.
type Gateway struct {
	cfg    *config.Config
	engine *engine.Engine
	mux    *http.ServeMux
}

// NewGateway creates and initializes all HTTP handlers and routers.
func NewGateway(cfg *config.Config, eng *engine.Engine) *Gateway {
	gw := &Gateway{
		cfg:    cfg,
		engine: eng,
		mux:    http.NewServeMux(),
	}
	gw.routes()
	return gw
}

// Handler returns the HTTP handler wrapped with CORS middleware.
func (gw *Gateway) Handler() http.Handler {
	return corsMiddleware(gw.mux)
}

func (gw *Gateway) routes() {
	// 1. Health checks
	gw.mux.HandleFunc("/healthz", gw.handleHealth)
	gw.mux.HandleFunc("/api/health", gw.handleHealth)

	// 2. Bootstrap data
	gw.mux.HandleFunc("/api/bootstrap", gw.handleBootstrap)

	// 3. Contracts
	gw.mux.HandleFunc("/api/contracts", gw.handleContracts)

	// 4. Export Go Middleware
	gw.mux.HandleFunc("/api/export/middleware.go", gw.handleExportMiddleware)

	// 5. Metrics & Incidents & Rules
	gw.mux.HandleFunc("/api/metrics", gw.handleMetrics)
	gw.mux.HandleFunc("/api/incidents", gw.handleIncidents)
	gw.mux.HandleFunc("/api/rules", gw.handleRules)

	// 6. Reset
	gw.mux.HandleFunc("/api/reset", gw.handleReset)

	// 7. Proxy Endpoint: POST /api/proxy/{service}
	gw.mux.HandleFunc("/api/proxy/", gw.handleProxy)

	// 8. Static Web UI
	gw.setupStaticUI()
}

func (gw *Gateway) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"service": "neuroproxy-gateway",
		"version": "2.0.0",
		"metrics": gw.engine.GetMetrics(),
	})
}

func (gw *Gateway) handleBootstrap(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"contracts": gw.engine.GetContracts(),
		"presets":   contracts.GetChaosPresets(),
		"metrics":   gw.engine.GetMetrics(),
		"incidents": gw.engine.GetIncidents(),
		"rules":     gw.engine.GetRules(),
	})
}

func (gw *Gateway) handleContracts(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, gw.engine.GetContracts())
		return
	}
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "POST required"})
		return
	}
	var contract contracts.ServiceContract
	if err := json.NewDecoder(r.Body).Decode(&contract); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid contract JSON: " + err.Error()})
		return
	}
	if strings.TrimSpace(contract.ID) == "" || len(contract.RequiredFields) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Contract ID and required_fields are required"})
		return
	}
	gw.engine.UpsertContract(contract)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "contract_updated",
		"contracts": gw.engine.GetContracts(),
	})
}

func (gw *Gateway) handleExportMiddleware(w http.ResponseWriter, r *http.Request) {
	code := gw.engine.ExportGoMiddleware()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="neuroproxy_middleware.go"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(code))
}

func (gw *Gateway) handleMetrics(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, gw.engine.GetMetrics())
}

func (gw *Gateway) handleIncidents(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, gw.engine.GetIncidents())
}

func (gw *Gateway) handleRules(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, gw.engine.GetRules())
}

func (gw *Gateway) handleReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "POST required"})
		return
	}
	gw.engine.Reset()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "reset_complete",
		"metrics": gw.engine.GetMetrics(),
	})
}

func (gw *Gateway) handleProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "POST required"})
		return
	}

	serviceID := strings.TrimPrefix(r.URL.Path, "/api/proxy/")
	serviceID = strings.Trim(serviceID, "/")
	if serviceID == "" {
		serviceID = "custom"
	}

	var rawMap map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&rawMap); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body: " + err.Error()})
		return
	}

	payload := rawMap
	selfHeal := true
	if p, ok := rawMap["payload"].(map[string]interface{}); ok {
		payload = p
		if sh, ok := rawMap["self_heal"].(bool); ok {
			selfHeal = sh
		}
	}
	if r.URL.Query().Get("bypass") == "true" || r.Header.Get("X-NeuroProxy-Mode") == "bypass" {
		selfHeal = false
	}

	// Live external HTTP/HTTPS upstream targets
	targetURL := r.Header.Get("X-Target-URL")
	if targetURL == "" {
		targetURL = r.URL.Query().Get("target")
	}
	if targetURL == "" {
		if t, ok := rawMap["target_url"].(string); ok {
			targetURL = t
		}
	}
	if targetURL != "" {
		gw.engine.UpsertContract(contracts.ServiceContract{
			ID:          serviceID,
			Name:        "Live Upstream (" + serviceID + ")",
			Endpoint:    targetURL,
			Description: "Live external HTTP microservice at " + targetURL,
			RequiredFields: map[string]string{
				"status": "expected 200 OK from " + targetURL,
			},
		})
	}

	rec, err := gw.engine.ProcessRequest(r.Context(), serviceID, payload, selfHeal)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Google Cloud Structured Logging emission
	telemetry.LogIncident(rec, r.Header.Get("X-Cloud-Trace-Context"))

	switch rec.ResolutionMode {
	case "GEMINI_SYNTHESIS":
		w.Header().Set("X-Healed-By", fmt.Sprintf("NeuroProxy-Gemini-Turn-%d", rec.AgentTurns))
		w.Header().Set("X-Schema-Fingerprint", rec.SchemaFingerprint)
		w.Header().Set("X-Value-Conservation", "100%-Verified")
	case "CACHE_HIT":
		w.Header().Set("X-Healed-By", "NeuroProxy-Cache")
		w.Header().Set("X-Schema-Fingerprint", rec.SchemaFingerprint)
		w.Header().Set("X-Value-Conservation", "100%-Verified")
	case "SAFETY_VETO":
		w.Header().Set("X-Healed-By", "Blocked-Safety-Veto")
		w.Header().Set("X-Schema-Fingerprint", rec.SchemaFingerprint)
	case "VALID_PASS":
		w.Header().Set("X-Healed-By", "None-Healthy")
	default:
		w.Header().Set("X-Healed-By", "Bypassed-Failed")
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"incident": rec,
		"metrics":  gw.engine.GetMetrics(),
		"rules":    gw.engine.GetRules(),
	})
}

func (gw *Gateway) setupStaticUI() {
	staticDir := gw.cfg.StaticDir
	fileServer := http.FileServer(http.Dir(staticDir))
	gw.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		cleanPath := filepath.Clean(r.URL.Path)
		fullPath := filepath.Join(staticDir, cleanPath)
		if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}
		indexFile := filepath.Join(staticDir, "index.html")
		if _, err := os.Stat(indexFile); err == nil {
			http.ServeFile(w, r, indexFile)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{
			"service": "NeuroProxy Autonomous Self-Healing API Gateway",
			"status":  "API Online",
		})
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-NeuroProxy-Mode, X-Target-URL, X-Cloud-Trace-Context")
		w.Header().Set("Access-Control-Expose-Headers", "X-Healed-By, X-Schema-Fingerprint, X-Value-Conservation")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
