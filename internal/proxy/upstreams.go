package proxy

import (
	"fmt"
	"strings"
	"time"
)

// GetServiceContracts returns the strict target microservice contracts protected by NeuroProxy.
func GetServiceContracts() map[string]ServiceContract {
	return map[string]ServiceContract{
		"checkout": {
			ID:          "checkout",
			Name:        "Order Checkout & Payment Service (v2)",
			Endpoint:    "/internal/upstream/checkout",
			Description: "Strict payment processor expecting flat v2 fields: order_id (string), amount_cents (integer > 0), currency (3-char uppercase ISO string), customer_email (valid email string).",
			RequiredFields: map[string]string{
				"order_id":       "string (non-empty order identifier, e.g. 'ORD-9021')",
				"amount_cents":   "integer (total payment in cents, > 0, e.g. 14900)",
				"currency":       "string (3-letter uppercase ISO currency code, e.g. 'USD', 'THB', 'SGD')",
				"customer_email": "string (valid email address containing '@')",
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

// ValidateAndExecuteUpstream simulates a strict production microservice validating an incoming payload.
// Returns (HTTPStatusCode, ResponseMap, ErrorMessage).
func ValidateAndExecuteUpstream(serviceID string, payload map[string]interface{}) (int, map[string]interface{}, string) {
	switch serviceID {
	case "checkout":
		return validateCheckout(payload)
	case "inventory":
		return validateInventory(payload)
	case "telemetry":
		return validateTelemetry(payload)
	default:
		return 404, map[string]interface{}{
			"error": fmt.Sprintf("Unknown upstream service '%s'", serviceID),
		}, fmt.Sprintf("Upstream service '%s' not found", serviceID)
	}
}

func validateCheckout(p map[string]interface{}) (int, map[string]interface{}, string) {
	var errs []string

	orderID, ok := p["order_id"].(string)
	if !ok || strings.TrimSpace(orderID) == "" {
		errs = append(errs, "missing or invalid required field 'order_id' (expected non-empty string)")
	}

	// JSON numbers unmarshal into float64 in Go map[string]interface{}
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

	if len(errs) > 0 {
		errMsg := "SchemaValidationError [CheckoutService v2.4]: " + strings.Join(errs, "; ")
		return 422, map[string]interface{}{
			"status":     "REJECTED_SCHEMA_DRIFT",
			"service":    "checkout-service-v2",
			"error_code": "ERR_STRICT_CONTRACT_VIOLATION",
			"details":    errs,
			"trace":      "at CheckoutController.validateStrictV2 (checkout_handler.go:114)",
		}, errMsg
	}

	return 200, map[string]interface{}{
		"status":         "CAPTURED",
		"service":        "checkout-service-v2",
		"transaction_id": fmt.Sprintf("TXN-%d", time.Now().UnixMilli()%100000),
		"order_id":       orderID,
		"charged_cents":  amountInt,
		"currency":       currency,
		"receipt_sent":   email,
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
	} else {
		errs = append(errs, fmt.Sprintf("type mismatch on 'temperature_c': expected number, got %T (%v)", tempVal, tempVal))
	}

	voltVal, hasVolt := p["voltage_v"]
	voltFloat := 0.0
	if !hasVolt {
		errs = append(errs, "missing required field 'voltage_v' (expected numeric Volts)")
	} else if vf, ok := voltVal.(float64); ok {
		voltFloat = vf
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

// GetChaosPresets returns curated real-world breaking webhook scenarios for the simulator UI.
func GetChaosPresets() []ChaosPreset {
	return []ChaosPreset{
		{
			ID:          "stripe-webhook-drift",
			Title:       "Scenario 1: Stripe / Shopify Webhook Drift",
			Subtitle:    "Nested buyer object & dollar float instead of flat integer cents",
			ServiceID:   "checkout",
			Badge:       "Webhook Drift",
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
			ID:          "erp-fulfillment-drift",
			Title:       "Scenario 2: Legacy ERP Type & Enum Mismatch",
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
			Title:       "Scenario 3: IoT Edge Firmware OTA Schema Drift",
			Subtitle:    "Fahrenheit telemetry + string status ('ONLINE') after OTA update",
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
			Title:       "Scenario 4: Healthy Control Payload (Zero Drift)",
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
