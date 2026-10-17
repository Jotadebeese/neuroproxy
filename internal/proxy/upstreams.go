package proxy

import (
	"context"
	"fmt"
	"strings"
	"time"
)

var defaultHTTPForwarder = NewHTTPForwarder()

// GetServiceContracts returns the default target microservice contracts protected by NeuroProxy.
func GetServiceContracts() map[string]ServiceContract {
	return map[string]ServiceContract{
		"checkout": {
			ID:          "checkout",
			Name:        "Order Checkout & Payment Service (v2 + AML Layer)",
			Endpoint:    "/internal/upstream/checkout",
			Description: "2-Layer Validator: Layer 1 checks flat schema (order_id, amount_cents, currency, customer_email). Layer 2 enforces AML compliance ('compliance_verified': true) when amount_cents >= 100000 ($1,000+).",
			RequiredFields: map[string]string{
				"order_id":            "string (non-empty order identifier, e.g. 'ORD-9021')",
				"amount_cents":        "integer (total payment in cents, > 0, e.g. 14900)",
				"currency":            "string (3-letter uppercase ISO currency code, e.g. 'USD', 'THB', 'SGD')",
				"customer_email":      "string (valid email address containing '@')",
				"compliance_verified": "boolean (optional for < $1k; REQUIRED true by Layer 2 Domain Engine when amount_cents >= 100000)",
			},
			SampleValidJSON: `{"order_id":"ORD-9021","amount_cents":14900,"currency":"USD","customer_email":"alex@enterprise.io"}`,
		},
		"inventory": {
			ID:          "inventory",
			Name:        "Warehouse ERP Fulfillment Service",
			Endpoint:    "/internal/upstream/inventory",
			Description: "Strict warehouse stock reservation service expecting sku_code (string), quantity (integer > 0), warehouse_zone (string), and priority_tier (enum: STANDARD, EXPRESS, CRITICAL).",
			RequiredFields: map[string]string{
				"sku_code":       "string (product SKU identifier, e.g. 'SKU-GPU-H100')",
				"quantity":       "integer (units to reserve, > 0)",
				"warehouse_zone": "string (target fulfillment hub, e.g. 'APAC-SG-01')",
				"priority_tier":  "string (strictly one of: 'STANDARD', 'EXPRESS', 'CRITICAL')",
			},
			SampleValidJSON: `{"sku_code":"SKU-GPU-H100","quantity":4,"warehouse_zone":"APAC-SG-01","priority_tier":"EXPRESS"}`,
		},
		"telemetry": {
			ID:          "telemetry",
			Name:        "Industrial IoT Grid Telemetry Ingest",
			Endpoint:    "/internal/upstream/telemetry",
			Description: "Strict industrial telemetry collector expecting device_id (string), temperature_c (number), voltage_v (number), and healthy (boolean).",
			RequiredFields: map[string]string{
				"device_id":     "string (unique hardware node ID, e.g. 'NODE-BKK-404')",
				"temperature_c": "number (temperature in Celsius, float/int, e.g. 42.5)",
				"voltage_v":     "number (bus voltage in Volts, float/int, e.g. 230.2)",
				"healthy":       "boolean (true or false — strict boolean, not string)",
			},
			SampleValidJSON: `{"device_id":"NODE-BKK-404","temperature_c":42.5,"voltage_v":230.2,"healthy":true}`,
		},
	}
}

// ValidateAndExecuteUpstream simulates a strict production microservice or forwards to a live HTTP endpoint.
// Supports real HTTP endpoints, built-in 2-layer services, as well as user-defined custom contracts.
func ValidateAndExecuteUpstream(contract ServiceContract, payload map[string]interface{}) (int, map[string]interface{}, string) {
	if strings.HasPrefix(contract.Endpoint, "http://") || strings.HasPrefix(contract.Endpoint, "https://") {
		res := defaultHTTPForwarder.ForwardRequest(context.Background(), contract.Endpoint, payload, nil)
		return res.StatusCode, res.Body, res.ErrorMsg
	}
	switch contract.ID {
	case "checkout":
		return validateCheckout(payload)
	case "inventory":
		return validateInventory(payload)
	case "telemetry":
		return validateTelemetry(payload)
	default:
		return validateDynamicContract(contract, payload)
	}
}

func validateCheckout(p map[string]interface{}) (int, map[string]interface{}, string) {
	var errs []string

	orderID, ok := p["order_id"].(string)
	if !ok || strings.TrimSpace(orderID) == "" {
		errs = append(errs, "missing or invalid required field 'order_id' (expected non-empty string)")
	}

	amountVal, hasAmount := p["amount_cents"]
	amountInt := 0
	if !hasAmount {
		errs = append(errs, "missing required field 'amount_cents' (expected integer cents > 0)")
	} else {
		switch v := amountVal.(type) {
		case float64:
			if v <= 0 || v != float64(int(v)) {
				errs = append(errs, fmt.Sprintf("field 'amount_cents' must be a positive integer, got %v", v))
			} else {
				amountInt = int(v)
			}
		case int:
			if v <= 0 {
				errs = append(errs, "field 'amount_cents' must be > 0")
			}
			amountInt = v
		default:
			errs = append(errs, fmt.Sprintf("type mismatch on 'amount_cents': expected integer, got %T (%v)", amountVal, amountVal))
		}
	}

	currency, ok := p["currency"].(string)
	if !ok || len(strings.TrimSpace(currency)) != 3 || currency != strings.ToUpper(currency) {
		errs = append(errs, fmt.Sprintf("invalid 'currency': expected 3-letter uppercase ISO string (e.g. 'USD'), got %v", p["currency"]))
	}

	email, ok := p["customer_email"].(string)
	if !ok || !strings.Contains(email, "@") {
		errs = append(errs, "missing or invalid 'customer_email' (expected valid email string)")
	}

	// LAYER 1: Structural JSON Schema Check
	if len(errs) > 0 {
		errMsg := "Layer1_SchemaValidationError [CheckoutService v2.4]: " + strings.Join(errs, "; ")
		return 422, map[string]interface{}{
			"status":     "REJECTED_LAYER_1_SCHEMA",
			"service":    "checkout-service-v2",
			"error_code": "ERR_STRICT_CONTRACT_VIOLATION",
			"details":    errs,
			"trace":      "at CheckoutController.unmarshalStrictV2 (checkout_handler.go:114)",
		}, errMsg
	}

	// LAYER 2: Domain & AML Compliance Rule Engine (only executes after Layer 1 passes!)
	if amountInt >= 100000 {
		compVal, hasComp := p["compliance_verified"]
		compBool, isBool := compVal.(bool)
		if !hasComp || !isBool || !compBool {
			errMsg := fmt.Sprintf("Layer2_DomainComplianceConflict [AML Guard]: High-value institutional checkout (%d cents >= 100000) requires boolean field 'compliance_verified': true mapped from source KYC/AML approval status.", amountInt)
			return 409, map[string]interface{}{
				"status":     "REJECTED_LAYER_2_AML_COMPLIANCE",
				"service":    "checkout-aml-engine",
				"error_code": "ERR_HIGH_VALUE_COMPLIANCE_MISSING",
				"details":    []string{errMsg},
				"trace":      "at ComplianceEngine.verifyInstitutionalOrder (aml_policy.go:57)",
			}, errMsg
		}
	}

	return 200, map[string]interface{}{
		"status":              "CAPTURED",
		"service":             "checkout-service-v2",
		"transaction_id":      fmt.Sprintf("TXN-%d", time.Now().UnixMilli()%100000),
		"order_id":            orderID,
		"charged_cents":       amountInt,
		"currency":            currency,
		"receipt_sent":        email,
		"compliance_verified": p["compliance_verified"] == true,
	}, ""
}

func validateInventory(p map[string]interface{}) (int, map[string]interface{}, string) {
	var errs []string

	sku, ok := p["sku_code"].(string)
	if !ok || strings.TrimSpace(sku) == "" {
		errs = append(errs, "missing required field 'sku_code' (expected string)")
	}

	qtyVal, hasQty := p["quantity"]
	qtyInt := 0
	if !hasQty {
		errs = append(errs, "missing required field 'quantity' (expected integer > 0)")
	} else {
		switch v := qtyVal.(type) {
		case float64:
			if v <= 0 || v != float64(int(v)) {
				errs = append(errs, fmt.Sprintf("field 'quantity' must be a positive integer, got %v", v))
			} else {
				qtyInt = int(v)
			}
		case int:
			qtyInt = v
		default:
			errs = append(errs, fmt.Sprintf("type mismatch on 'quantity': expected integer, got %T (%v)", qtyVal, qtyVal))
		}
	}

	zone, ok := p["warehouse_zone"].(string)
	if !ok || strings.TrimSpace(zone) == "" {
		errs = append(errs, "missing required field 'warehouse_zone' (expected string)")
	}

	tier, ok := p["priority_tier"].(string)
	validTiers := map[string]bool{"STANDARD": true, "EXPRESS": true, "CRITICAL": true}
	if !ok || !validTiers[tier] {
		errs = append(errs, fmt.Sprintf("invalid enum 'priority_tier': expected one of [STANDARD, EXPRESS, CRITICAL], got '%v'", p["priority_tier"]))
	}

	if len(errs) > 0 {
		errMsg := "ERPContractException [InventoryService]: " + strings.Join(errs, "; ")
		return 422, map[string]interface{}{
			"status":     "ERP_VALIDATION_FAILED",
			"service":    "warehouse-erp-v3",
			"error_code": "ERR_INVALID_FULFILLMENT_PAYLOAD",
			"details":    errs,
			"trace":      "at WarehouseFulfillment.parseReservation (erp_ingest.go:88)",
		}, errMsg
	}

	return 200, map[string]interface{}{
		"status":         "RESERVED",
		"service":        "warehouse-erp-v3",
		"allocation_id":  fmt.Sprintf("ALLOC-%d", time.Now().UnixMilli()%100000),
		"sku_code":       sku,
		"reserved_units": qtyInt,
		"warehouse_zone": zone,
		"priority_tier":  tier,
	}, ""
}

func validateTelemetry(p map[string]interface{}) (int, map[string]interface{}, string) {
	var errs []string

	devID, ok := p["device_id"].(string)
	if !ok || strings.TrimSpace(devID) == "" {
		errs = append(errs, "missing required field 'device_id' (expected string)")
	}

	tempVal, hasTemp := p["temperature_c"]
	tempFloat := 0.0
	if !hasTemp {
		errs = append(errs, "missing required field 'temperature_c' (expected numeric Celsius)")
	} else if vf, ok := tempVal.(float64); ok {
		tempFloat = vf
	} else if vi, ok := tempVal.(int); ok {
		tempFloat = float64(vi)
	} else {
		errs = append(errs, fmt.Sprintf("type mismatch on 'temperature_c': expected number, got %T (%v)", tempVal, tempVal))
	}

	voltVal, hasVolt := p["voltage_v"]
	voltFloat := 0.0
	if !hasVolt {
		errs = append(errs, "missing required field 'voltage_v' (expected numeric Volts)")
	} else if vf, ok := voltVal.(float64); ok {
		voltFloat = vf
	} else if vi, ok := voltVal.(int); ok {
		voltFloat = float64(vi)
	} else {
		errs = append(errs, fmt.Sprintf("type mismatch on 'voltage_v': expected number, got %T (%v)", voltVal, voltVal))
	}

	healthyVal, hasHealthy := p["healthy"]
	healthyBool := false
	if !hasHealthy {
		errs = append(errs, "missing required field 'healthy' (expected strict boolean)")
	} else if vb, ok := healthyVal.(bool); ok {
		healthyBool = vb
	} else {
		errs = append(errs, fmt.Sprintf("type mismatch on 'healthy': expected strict boolean, got %T (%v)", healthyVal, healthyVal))
	}

	if len(errs) > 0 {
		errMsg := "TelemetryIngestPanic [IoTCollector]: " + strings.Join(errs, "; ")
		return 400, map[string]interface{}{
			"status":     "INGEST_DROPPED",
			"service":    "iot-grid-collector",
			"error_code": "ERR_FIRMWARE_SCHEMA_MISMATCH",
			"details":    errs,
			"trace":      "at GridCollector.decodeSensorFrame (telemetry_decoder.go:62)",
		}, errMsg
	}

	return 200, map[string]interface{}{
		"status":        "INGESTED",
		"service":       "iot-grid-collector",
		"frame_id":      fmt.Sprintf("FRM-%d", time.Now().UnixMilli()%100000),
		"device_id":     devID,
		"temperature_c": tempFloat,
		"voltage_v":     voltFloat,
		"healthy":       healthyBool,
	}, ""
}

func validateDynamicContract(contract ServiceContract, p map[string]interface{}) (int, map[string]interface{}, string) {
	var errs []string
	for field, rule := range contract.RequiredFields {
		val, exists := p[field]
		if !exists || val == nil || fmt.Sprintf("%v", val) == "" {
			errs = append(errs, fmt.Sprintf("missing required field '%s' (%s)", field, rule))
			continue
		}
		ruleLower := strings.ToLower(rule)
		if strings.Contains(ruleLower, "int") || strings.Contains(ruleLower, "number") {
			if _, ok := toFloat(val); !ok {
				errs = append(errs, fmt.Sprintf("field '%s' expected numeric value, got %T (%v)", field, val, val))
			}
		} else if strings.Contains(ruleLower, "bool") {
			if _, ok := val.(bool); !ok {
				errs = append(errs, fmt.Sprintf("field '%s' expected boolean, got %T (%v)", field, val, val))
			}
		}
	}

	if len(errs) > 0 {
		errMsg := fmt.Sprintf("CustomContractError [%s]: %s", contract.Name, strings.Join(errs, "; "))
		return 422, map[string]interface{}{
			"status":     "CUSTOM_CONTRACT_REJECTED",
			"service":    contract.ID,
			"error_code": "ERR_DYNAMIC_SCHEMA_MISMATCH",
			"details":    errs,
		}, errMsg
	}

	return 200, map[string]interface{}{
		"status":   "ACCEPTED",
		"service":  contract.ID,
		"received": p,
	}, ""
}

// GetChaosPresets returns curated real-world breaking webhook scenarios for the simulator UI.
func GetChaosPresets() []ChaosPreset {
	return []ChaosPreset{
		{
			ID:          "stripe-webhook-drift",
			Title:       "Scenario 1: Stripe / Shopify Webhook Drift",
			Subtitle:    "Nested buyer object & dollar string instead of flat integer cents",
			ServiceID:   "checkout",
			Badge:       "1-Turn Heal",
			Description: "Partner webhook upgraded from v1 to v2 overnight: 'order_id' renamed to 'id', 'amount_cents' sent as nested dollar string 'payment.amount_dollars': '149.50', and 'customer_email' nested inside 'buyer.contact.email'.",
			Payload: map[string]interface{}{
				"id": "ORD-2026-8841",
				"payment": map[string]interface{}{
					"amount_dollars": "149.50",
					"iso_currency":   "usd",
				},
				"buyer": map[string]interface{}{
					"contact": map[string]interface{}{
						"email": "cto@fintech-apac.sg",
					},
				},
				"webhook_version": "2026-10-breaking",
			},
		},
		{
			ID:          "cascading-2layer-multilingual",
			Title:       "Scenario 2: 2-Layer Cascading Agentic Loop (Multilingual + AML)",
			Subtitle:    "Spanish LATAM payload ($2,450) -> Turn 1 fixes Schema (422) -> Turn 2 fixes AML Rule (409)",
			ServiceID:   "checkout",
			Badge:       "2-Turn Agent Loop",
			Description: "Demonstrates multi-turn agentic self-correction! Turn 1 maps Spanish keys ('numero_orden', 'monto_usd': '2450.00') to English schema. Upstream then triggers Layer 2 AML Compliance (409 Conflict: orders >= $1,000 require 'compliance_verified': true). Turn 2 autonomously maps 'kyc_aml_estado': 'APPROVED' -> 'compliance_verified': true!",
			Payload: map[string]interface{}{
				"numero_orden": "LATAM-INST-7740",
				"pago": map[string]interface{}{
					"monto_usd": "2450.00",
					"moneda":    "usd",
				},
				"cliente": map[string]interface{}{
					"correo_electronico": "tesoreria@banco-bogota.co",
				},
				"kyc_aml_estado": "APPROVED",
			},
		},
		{
			ID:          "hallucination-safety-trap",
			Title:       "Scenario 3: Hallucination Safety Trap (Missing Price Veto)",
			Subtitle:    "Corrupted webhook missing payment amount -> Guardrail blocks LLM from inventing a price",
			ServiceID:   "checkout",
			Badge:       "🛡️ Safety Veto",
			Description: "Proves production safety: this corrupted webhook has NO price/amount field at all. If an AI gateway invents a default price (e.g. 14950 cents), it would overcharge the customer. NeuroProxy's Zero-Hallucination Value Conservation Guardrail detects the unverified number and triggers a SAFETY_VETO.",
			Payload: map[string]interface{}{
				"id": "ORD-CORRUPT-009",
				"buyer": map[string]interface{}{
					"email": "risk-audit@stripe.com",
				},
				"currency": "USD",
				"note":     "Corrupted payload: price field was dropped by client bug",
			},
		},
		{
			ID:          "erp-fulfillment-drift",
			Title:       "Scenario 4: Legacy ERP Type & Enum Mismatch",
			Subtitle:    "String quantity & non-standard urgency label ('urgent_rush')",
			ServiceID:   "inventory",
			Badge:       "Enum / Type Drift",
			Description: "Regional SAP/ERP connector dispatches 'item_sku' instead of 'sku_code', passes 'qty' as string '12', 'hub_location' instead of 'warehouse_zone', and 'urgency': 'urgent_rush' instead of enum [STANDARD, EXPRESS, CRITICAL].",
			Payload: map[string]interface{}{
				"item_sku":     "SKU-TPU-V5P-POD",
				"qty":          "12",
				"hub_location": "APAC-SIN-EAST",
				"urgency":      "urgent_rush",
			},
		},
		{
			ID:          "iot-firmware-mutation",
			Title:       "Scenario 5: IoT Edge Firmware OTA Schema Drift",
			Subtitle:    "String telemetry metrics + 'ONLINE' status string after OTA update",
			ServiceID:   "telemetry",
			Badge:       "IoT Firmware Drift",
			Description: "An OTA firmware update on industrial substation sensors changed 'device_id' to 'node_serial', nested metrics under 'readings.temp_celsius' and 'readings.bus_voltage', and changed 'healthy' boolean to string 'status': 'ONLINE'.",
			Payload: map[string]interface{}{
				"node_serial": "SUB-BKK-NORTH-09",
				"readings": map[string]interface{}{
					"temp_celsius": "46.8",
					"bus_voltage":  "231.4",
				},
				"status": "ONLINE",
			},
		},
		{
			ID:          "valid-checkout-control",
			Title:       "Scenario 6: Healthy Control Payload (Zero Drift)",
			Subtitle:    "Complies 100% with Checkout v2 contract (passes straight through)",
			ServiceID:   "checkout",
			Badge:       "Healthy Control",
			Description: "Validates that healthy traffic passes through NeuroProxy with zero modification and sub-millisecond overhead.",
			Payload: map[string]interface{}{
				"order_id":       "ORD-HEALTHY-100",
				"amount_cents":   29900,
				"currency":       "SGD",
				"customer_email": "ops@googlecloud.com",
			},
		},
	}
}
