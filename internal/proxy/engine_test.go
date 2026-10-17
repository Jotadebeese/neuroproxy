package proxy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNeuroProxyEndToEndAndCacheConcurrency(t *testing.T) {
	eng := NewEngine()
	ctx := context.Background()

	// 1. Standard 1-Turn Stripe Webhook Drift
	driftedPayload := map[string]interface{}{
		"id": "ORD-2026-991",
		"payment": map[string]interface{}{
			"amount_dollars": "149.50",
			"iso_currency":   "usd",
		},
		"buyer": map[string]interface{}{
			"contact": map[string]interface{}{
				"email": "sre@googlecloud.com",
			},
		},
	}

	// Direct Bypass (Self-Healing OFF) must fail with 422
	bypassRec, err := eng.ProcessRequest(ctx, "checkout", driftedPayload, false)
	if err != nil {
		t.Fatalf("unexpected error in bypass mode: %v", err)
	}
	if bypassRec.FinalStatus != 422 || bypassRec.ResolutionMode != "BYPASS_FAILED" {
		t.Fatalf("expected 422 BYPASS_FAILED, got %d (%s)", bypassRec.FinalStatus, bypassRec.ResolutionMode)
	}

	// First Self-Healing Request must intercept 422, heal via Gemini, pass Guardrail, cache rule, and return 200 OK
	healRec, err := eng.ProcessRequest(ctx, "checkout", driftedPayload, true)
	if err != nil {
		t.Fatalf("unexpected error in heal mode: %v", err)
	}
	if healRec.FinalStatus != 200 || healRec.ResolutionMode != "GEMINI_SYNTHESIS" {
		t.Fatalf("expected 200 GEMINI_SYNTHESIS, got %d (%s)", healRec.FinalStatus, healRec.ResolutionMode)
	}
	if !healRec.Guardrail.Passed {
		t.Fatalf("expected Guardrail.Passed=true, got false: %v", healRec.Guardrail.Summary)
	}

	// Concurrent repeat requests with same schema shape must hit in-memory RuleCache with zero data races
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			reqPayload := map[string]interface{}{
				"id": "ORD-CONCURRENT",
				"payment": map[string]interface{}{
					"amount_dollars": "89.99",
					"iso_currency":   "sgd",
				},
				"buyer": map[string]interface{}{
					"contact": map[string]interface{}{
						"email": "user@apac.io",
					},
				},
			}
			rec, err := eng.ProcessRequest(ctx, "checkout", reqPayload, true)
			if err != nil {
				t.Errorf("concurrent request %d failed: %v", idx, err)
				return
			}
			if rec.FinalStatus != 200 || rec.ResolutionMode != "CACHE_HIT" {
				t.Errorf("expected 200 CACHE_HIT, got %d (%s)", rec.FinalStatus, rec.ResolutionMode)
			}
		}(i)
	}
	wg.Wait()

	// 2. Test 2-Layer Cascading Failure (Turn 1 fixes Schema 422 -> Turn 2 fixes Layer 2 AML 409 -> 200 OK)
	cascadingPayload := map[string]interface{}{
		"numero_orden": "LATAM-INST-7740",
		"pago": map[string]interface{}{
			"monto_usd": "2450.00",
			"moneda":    "usd",
		},
		"cliente": map[string]interface{}{
			"correo_electronico": "tesoreria@banco-bogota.co",
		},
		"kyc_aml_estado": "APPROVED",
	}
	cascadeRec, err := eng.ProcessRequest(ctx, "checkout", cascadingPayload, true)
	if err != nil {
		t.Fatalf("unexpected error in cascading 2-layer test: %v", err)
	}
	if cascadeRec.FinalStatus != 200 || cascadeRec.AgentTurns != 2 {
		t.Fatalf("expected 200 OK after 2 AgentTurns, got status=%d turns=%d", cascadeRec.FinalStatus, cascadeRec.AgentTurns)
	}

	// 3. Test Zero-Hallucination Guardrail Safety Veto on Missing Price Trap
	corruptPayload := map[string]interface{}{
		"id": "ORD-CORRUPT-009",
		"buyer": map[string]interface{}{
			"email": "risk-audit@stripe.com",
		},
		"currency": "USD",
	}
	vetoRec, err := eng.ProcessRequest(ctx, "checkout", corruptPayload, true)
	if err != nil {
		t.Fatalf("unexpected error in safety veto test: %v", err)
	}
	if vetoRec.ResolutionMode != "SAFETY_VETO" || vetoRec.Guardrail.Passed {
		t.Fatalf("expected SAFETY_VETO and Guardrail.Passed=false, got mode=%s passed=%v", vetoRec.ResolutionMode, vetoRec.Guardrail.Passed)
	}

	// 4. Test Go Middleware Code Export
	code := eng.ExportGoMiddleware()
	if !strings.Contains(code, "NeuroProxyAdapterMiddleware") {
		t.Fatalf("expected exported Go middleware to contain NeuroProxyAdapterMiddleware")
	}
}

// TestRealHTTPUpstreamForwarding tests real HTTP reverse-proxying over network to a live server.
func TestRealHTTPUpstreamForwarding(t *testing.T) {
	// Spin up a real HTTP upstream server that validates strict order fields
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)

		orderID, hasOrder := body["order_id"].(string)
		amountCents, hasAmount := body["amount_cents"]

		if !hasOrder || orderID == "" || !hasAmount || amountCents == nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "Strict schema rejection: required fields 'order_id' and 'amount_cents' missing",
				"status":  "REJECTED_422",
				"attempt": requestCount,
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "CAPTURED_200",
			"order_id": orderID,
			"amount":   amountCents,
			"attempt":  requestCount,
		})
	}))
	defer server.Close()

	eng := NewEngine()
	ctx := context.Background()

	// Register the real HTTP server URL into NeuroProxy contracts
	eng.UpsertContract(ServiceContract{
		ID:          "real-payment-api",
		Name:        "Live Microservice on HTTP",
		Endpoint:    server.URL,
		Description: "Real downstream HTTP server testing network proxying",
		RequiredFields: map[string]string{
			"order_id":     "string (order id)",
			"amount_cents": "integer (cents)",
		},
	})

	driftedPayload := map[string]interface{}{
		"id": "ORD-REAL-NET-99",
		"payment": map[string]interface{}{
			"amount_dollars": "149.50",
			"iso_currency":   "usd",
		},
		"buyer": map[string]interface{}{
			"contact": map[string]interface{}{
				"email": "test@googlecloud.com",
			},
		},
	}

	// 1. Direct bypass against real HTTP server must return 422
	bypassRec, err := eng.ProcessRequest(ctx, "real-payment-api", driftedPayload, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bypassRec.FinalStatus != 422 {
		t.Fatalf("expected 422 from real HTTP server, got %d", bypassRec.FinalStatus)
	}

	// 2. Self-healing mode against real HTTP server must intercept 422, heal, re-dispatch, and succeed with 200 OK
	healRec, err := eng.ProcessRequest(ctx, "real-payment-api", driftedPayload, true)
	if err != nil {
		t.Fatalf("unexpected error in real HTTP heal: %v", err)
	}
	if healRec.FinalStatus != 200 {
		t.Fatalf("expected 200 OK from real HTTP server after heal, got %d", healRec.FinalStatus)
	}
	if healRec.ResolutionMode != "GEMINI_SYNTHESIS" {
		t.Fatalf("expected GEMINI_SYNTHESIS, got %s", healRec.ResolutionMode)
	}

	// 3. Repeat request must hit cache and return 200 OK in <1ms without re-synthesizing
	repeatRec, err := eng.ProcessRequest(ctx, "real-payment-api", driftedPayload, true)
	if err != nil {
		t.Fatalf("unexpected error in real HTTP repeat: %v", err)
	}
	if repeatRec.FinalStatus != 200 || repeatRec.ResolutionMode != "CACHE_HIT" {
		t.Fatalf("expected 200 CACHE_HIT on repeat real HTTP request, got status=%d mode=%s", repeatRec.FinalStatus, repeatRec.ResolutionMode)
	}
}

// TestRuleCacheLRUEviction verifies capacity bounds and LRU eviction.
func TestRuleCacheLRUEviction(t *testing.T) {
	cache := NewRuleCacheWithConfig(2, 1*time.Hour) // Max capacity 2

	rule1 := &LearnedRule{Fingerprint: "fp1", Description: "Rule 1", LastUsedAt: time.Now().Add(-10 * time.Minute)}
	rule2 := &LearnedRule{Fingerprint: "fp2", Description: "Rule 2", LastUsedAt: time.Now().Add(-5 * time.Minute)}
	rule3 := &LearnedRule{Fingerprint: "fp3", Description: "Rule 3", LastUsedAt: time.Now()}

	cache.Set(rule1)
	cache.Set(rule2)
	if cache.Count() != 2 {
		t.Fatalf("expected 2 rules, got %d", cache.Count())
	}

	// Inserting 3rd rule must evict fp1 (oldest LastUsedAt)
	cache.Set(rule3)
	if cache.Count() != 2 {
		t.Fatalf("expected 2 rules after LRU eviction, got %d", cache.Count())
	}

	if _, hit := cache.Get("fp1"); hit {
		t.Fatalf("expected fp1 to be evicted by LRU policy")
	}
	if _, hit := cache.Get("fp2"); !hit {
		t.Fatalf("expected fp2 to still be cached")
	}
	if _, hit := cache.Get("fp3"); !hit {
		t.Fatalf("expected fp3 to still be cached")
	}
}

// TestCircuitBreakerProtection verifies rate limits and fail-fast tripping.
func TestCircuitBreakerProtection(t *testing.T) {
	cb := NewCircuitBreaker(3, 50*time.Millisecond, 10.0, 5)

	// Can consume up to burst
	for i := 0; i < 5; i++ {
		if !cb.Allow() {
			t.Fatalf("expected token %d to be allowed", i)
		}
	}

	// 3 consecutive failures should trip breaker to OPEN
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.GetState() != StateOpen {
		t.Fatalf("expected StateOpen, got %s", cb.GetState())
	}

	// While open, Allow() must reject immediately
	if cb.Allow() {
		t.Fatalf("expected Allow() to be false when circuit is OPEN")
	}

	// Wait for reset timeout
	time.Sleep(60 * time.Millisecond)

	// Breaker transitions to HalfOpen
	if !cb.Allow() {
		t.Fatalf("expected Allow() to be true after reset timeout in HalfOpen")
	}
	cb.RecordSuccess()
	if cb.GetState() != StateClosed {
		t.Fatalf("expected StateClosed after success, got %s", cb.GetState())
	}
}
