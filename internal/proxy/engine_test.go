package proxy

import (
	"context"
	"sync"
	"testing"
)

func TestNeuroProxyEndToEndAndCacheConcurrency(t *testing.T) {
	eng := NewEngine()
	ctx := context.Background()

	// Drifted Stripe/Shopify checkout webhook payload
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

	// 1. Direct Bypass (Self-Healing OFF) must fail with 422
	bypassRec, err := eng.ProcessRequest(ctx, "checkout", driftedPayload, false)
	if err != nil {
		t.Fatalf("unexpected error in bypass mode: %v", err)
	}
	if bypassRec.FinalStatus != 422 || bypassRec.ResolutionMode != "BYPASS_FAILED" {
		t.Fatalf("expected 422 BYPASS_FAILED, got %d (%s)", bypassRec.FinalStatus, bypassRec.ResolutionMode)
	}

	// 2. First Self-Healing Request must intercept 422, heal via Gemini, cache rule, and return 200 OK
	healRec, err := eng.ProcessRequest(ctx, "checkout", driftedPayload, true)
	if err != nil {
		t.Fatalf("unexpected error in heal mode: %v", err)
	}
	if healRec.FinalStatus != 200 || healRec.ResolutionMode != "GEMINI_SYNTHESIS" {
		t.Fatalf("expected 200 GEMINI_SYNTHESIS, got %d (%s)", healRec.FinalStatus, healRec.ResolutionMode)
	}
	if healRec.HealedPayload["amount_cents"] != 14950 {
		t.Fatalf("expected amount_cents=14950, got %v", healRec.HealedPayload["amount_cents"])
	}

	// 3. Concurrent repeat requests with same schema shape must hit in-memory RuleCache with zero data races
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

	metrics := eng.GetMetrics()
	if metrics.GeminiHealedCount != 1 {
		t.Fatalf("expected GeminiHealedCount=1, got %d", metrics.GeminiHealedCount)
	}
	if metrics.CacheHitHealedCount != 20 {
		t.Fatalf("expected CacheHitHealedCount=20, got %d", metrics.CacheHitHealedCount)
	}
}
