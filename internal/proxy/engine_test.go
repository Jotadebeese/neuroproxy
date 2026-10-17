package proxy

import (
	"context"
	"strings"
	"sync"
	"testing"
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
