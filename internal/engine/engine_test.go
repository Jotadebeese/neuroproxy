package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/Jotadebeese/neuroproxy/internal/ai"
	"github.com/Jotadebeese/neuroproxy/internal/cache"
	"github.com/Jotadebeese/neuroproxy/internal/contracts"
	"github.com/Jotadebeese/neuroproxy/internal/guardrail"
)

func TestNeuroProxyEndToEndAndCacheConcurrency(t *testing.T) {
	cb := ai.NewCircuitBreaker(100.0, 100, 10, 1*time.Second)
	healer := ai.NewGeminiHealer(cb)
	storage := cache.NewLRURuleCache(1000, 24*time.Hour)
	eng := NewEngine(storage, healer)

	ctx := context.Background()
	presets := contracts.GetChaosPresets()
	if len(presets) == 0 {
		t.Fatal("no presets configured")
	}
	preset := presets[0]

	// 1. First run: Schema Drift triggers Synthesis & Cache Storage
	rec1, err := eng.ProcessRequest(ctx, preset.ServiceID, preset.Payload, true)
	if err != nil {
		t.Fatalf("first request failed: %v", err)
	}
	if rec1.FinalStatus != http.StatusOK {
		t.Fatalf("expected 200 OK after healing, got %d", rec1.FinalStatus)
	}
	if rec1.ResolutionMode != "GEMINI_SYNTHESIS" {
		t.Fatalf("expected GEMINI_SYNTHESIS for first drift encounter, got %s", rec1.ResolutionMode)
	}

	// 2. Second run: Must HIT CACHE in <1ms
	rec2, err := eng.ProcessRequest(ctx, preset.ServiceID, preset.Payload, true)
	if err != nil {
		t.Fatalf("second request failed: %v", err)
	}
	if rec2.ResolutionMode != "CACHE_HIT" {
		t.Fatalf("expected CACHE_HIT on second request, got %s", rec2.ResolutionMode)
	}
	if rec2.HealingLatencyMs > 5.0 {
		t.Errorf("expected cache latency < 5ms, got %.2fms", rec2.HealingLatencyMs)
	}

	// 3. Concurrency Stress Test: 50 concurrent requests hitting compiled cache
	var wg sync.WaitGroup
	errCh := make(chan error, 50)
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			r, err := eng.ProcessRequest(ctx, preset.ServiceID, preset.Payload, true)
			if err != nil {
				errCh <- fmt.Errorf("worker %d failed: %w", workerID, err)
				return
			}
			if r.FinalStatus != http.StatusOK {
				errCh <- fmt.Errorf("worker %d got status %d", workerID, r.FinalStatus)
			}
		}(i)
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Fatal(err)
	}

	metrics := eng.GetMetrics()
	if metrics.CacheHits < 51 {
		t.Errorf("expected at least 51 cache hits, got %d", metrics.CacheHits)
	}
}

func TestRealHTTPUpstreamForwarding(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)

		centsVal, hasCents := body["amount_cents"]
		if !hasCents || centsVal == nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "Field 'amount_cents' is required as integer cents",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "CHARGE_SUCCEEDED",
			"received": body,
		})
	}))
	defer ts.Close()

	cb := ai.NewCircuitBreaker(100.0, 100, 10, 1*time.Second)
	healer := ai.NewGeminiHealer(cb)
	storage := cache.NewLRURuleCache(1000, 24*time.Hour)
	eng := NewEngine(storage, healer)

	eng.UpsertContract(contracts.ServiceContract{
		ID:          "payment-service-live",
		Name:        "External Live Billing Microservice",
		Endpoint:    ts.URL,
		Description: "External live billing server",
		RequiredFields: map[string]string{
			"amount_cents": "int: amount in cents",
		},
	})

	ctx := context.Background()
	driftedPayload := map[string]interface{}{
		"payment": map[string]interface{}{
			"amount_dollars": 149.50,
		},
		"currency": "USD",
	}

	rec, err := eng.ProcessRequest(ctx, "payment-service-live", driftedPayload, true)
	if err != nil {
		t.Fatalf("expected real HTTP reverse proxy to succeed, got error: %v", err)
	}
	if rec.FinalStatus != http.StatusOK {
		t.Fatalf("expected HTTP 200 OK from live upstream, got %d (mode: %s)", rec.FinalStatus, rec.ResolutionMode)
	}
	if rec.HealedPayload["amount_cents"] != 14950 {
		t.Fatalf("expected amount_cents=14950, got %v", rec.HealedPayload["amount_cents"])
	}
	if callCount < 2 {
		t.Fatalf("expected at least 2 calls to live server (1 failed probe + 1 healed execution), got %d", callCount)
	}
}

func TestRuleCacheLRUEviction(t *testing.T) {
	c := cache.NewLRURuleCache(2, 1*time.Hour)

	r1 := &contracts.LearnedRule{Fingerprint: "fp1", ServiceID: "s1"}
	r2 := &contracts.LearnedRule{Fingerprint: "fp2", ServiceID: "s2"}
	r3 := &contracts.LearnedRule{Fingerprint: "fp3", ServiceID: "s3"}

	c.Put(r1)
	c.Put(r2)

	if c.Len() != 2 {
		t.Fatalf("expected 2 items, got %d", c.Len())
	}

	// Access r1 so r2 becomes the LRU
	_, ok := c.Get("fp1")
	if !ok {
		t.Fatal("expected fp1 to be present")
	}

	// Add r3 -> r2 should be evicted!
	c.Put(r3)

	if _, ok := c.Get("fp2"); ok {
		t.Errorf("expected fp2 to be evicted by LRU policy")
	}
	if _, ok := c.Get("fp1"); !ok {
		t.Errorf("expected fp1 to still be retained")
	}
	if _, ok := c.Get("fp3"); !ok {
		t.Errorf("expected fp3 to be present")
	}
}

func TestCircuitBreakerProtection(t *testing.T) {
	cb := ai.NewCircuitBreaker(10.0, 3, 2, 50*time.Millisecond)

	// Consume tokens
	if err := cb.Allow(); err != nil {
		t.Fatalf("token 1 failed: %v", err)
	}
	if err := cb.Allow(); err != nil {
		t.Fatalf("token 2 failed: %v", err)
	}
	if err := cb.Allow(); err != nil {
		t.Fatalf("token 3 failed: %v", err)
	}

	// 4th token within the same second should be rate-limited
	if err := cb.Allow(); err != ai.ErrRateLimitExceeded {
		t.Fatalf("expected ErrRateLimitExceeded, got %v", err)
	}

	// Test circuit trip
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.State() != ai.StateOpen {
		t.Fatalf("expected StateOpen after 2 failures, got %s", cb.State())
	}

	// Wait for cooldown
	time.Sleep(60 * time.Millisecond)
	if err := cb.Allow(); err != nil {
		t.Fatalf("expected half-open probe request to be allowed, got %v", err)
	}
	if cb.State() != ai.StateHalfOpen {
		t.Fatalf("expected StateHalfOpen, got %s", cb.State())
	}

	cb.RecordSuccess()
	if cb.State() != ai.StateClosed {
		t.Fatalf("expected StateClosed after success, got %s", cb.State())
	}
}

func TestValueConservationGuardrail(t *testing.T) {
	contract := contracts.ServiceContract{
		ID:   "checkout",
		Name: "Checkout",
		RequiredFields: map[string]string{
			"amount_cents": "int",
		},
	}

	// Case 1: Valid conversion ($149.50 -> 14950 cents)
	rawValid := map[string]interface{}{"pago": map[string]interface{}{"monto_usd": 149.50}}
	healedValid := map[string]interface{}{"amount_cents": 14950}
	repValid := guardrail.VerifyValueConservation(rawValid, healedValid, contract)
	if !repValid.Passed || repValid.HallucinationVeto {
		t.Fatalf("expected valid conservation to pass, got: %s", repValid.Reason)
	}

	// Case 2: Hallucination (Missing raw price, fabricated 5000 cents)
	rawCorrupted := map[string]interface{}{"order_number": "123"}
	healedFabricated := map[string]interface{}{"amount_cents": 5000}
	repHallucination := guardrail.VerifyValueConservation(rawCorrupted, healedFabricated, contract)
	if repHallucination.Passed || !repHallucination.HallucinationVeto {
		t.Fatal("expected hallucination to be vetoed")
	}
}
