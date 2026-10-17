package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Jotadebeese/neuroproxy/internal/contracts"
	"github.com/Jotadebeese/neuroproxy/internal/forwarder"
)

var defaultForwarder = forwarder.NewHTTPForwarder()

// ValidateAndExecuteUpstream validates the payload against the contract and delivers it
// either to a real live external HTTP microservice or to the built-in offline simulation.
func ValidateAndExecuteUpstream(ctx context.Context, contract contracts.ServiceContract, payload map[string]interface{}) (int, map[string]interface{}, string) {
	// If the contract endpoint is an external HTTP/HTTPS URL, proxy to the real network!
	endpoint := strings.TrimSpace(contract.Endpoint)
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			return http.StatusBadRequest, nil, "Failed to serialize payload: " + err.Error()
		}

		headers := make(http.Header)
		headers.Set("User-Agent", "NeuroProxy-Gateway/2.0 (Google-Cloud-Run)")
		headers.Set("X-NeuroProxy-Origin", "Autonomous-Healer")

		fwdRes, err := defaultForwarder.Dispatch(ctx, endpoint, headers, bodyBytes)
		if err != nil {
			return http.StatusBadGateway, nil, fmt.Sprintf("External upstream network error to %s: %v", endpoint, err)
		}

		var parsedResp map[string]interface{}
		_ = json.Unmarshal(fwdRes.Body, &parsedResp)
		if parsedResp == nil {
			parsedResp = map[string]interface{}{
				"raw_body":    string(fwdRes.Body),
				"status_code": fwdRes.StatusCode,
			}
		}

		if fwdRes.StatusCode >= 400 {
			errStr := fmt.Sprintf("Upstream HTTP %d from %s: %s", fwdRes.StatusCode, endpoint, string(fwdRes.Body))
			return fwdRes.StatusCode, parsedResp, errStr
		}

		return fwdRes.StatusCode, parsedResp, ""
	}

	// Built-in offline simulated targets
	switch contract.ID {
	case "checkout":
		return validateCheckout(payload)
	case "inventory":
		return validateInventory(payload)
	case "telemetry":
		return validateTelemetry(payload)
	default:
		return validateCustomContract(contract, payload)
	}
}

func validateCheckout(payload map[string]interface{}) (int, map[string]interface{}, string) {
	var missing []string
	if val, ok := payload["order_id"]; !ok || fmt.Sprintf("%v", val) == "" {
		missing = append(missing, "order_id (string)")
	}

	centsVal, hasCents := payload["amount_cents"]
	if !hasCents {
		missing = append(missing, "amount_cents (integer cents, e.g. 14950)")
	} else {
		switch v := centsVal.(type) {
		case int:
			if v <= 0 {
				missing = append(missing, "amount_cents (must be positive integer)")
			}
		case float64:
			if v != float64(int(v)) || v <= 0 {
				missing = append(missing, "amount_cents (floating point disallowed; must be integer cents)")
			}
		default:
			missing = append(missing, "amount_cents (invalid type)")
		}
	}

	if val, ok := payload["currency"]; !ok || len(fmt.Sprintf("%v", val)) != 3 {
		missing = append(missing, "currency (3-letter ISO code, e.g. 'USD')")
	}

	if val, ok := payload["customer_email"]; !ok || !strings.Contains(fmt.Sprintf("%v", val), "@") {
		missing = append(missing, "customer_email (valid email string)")
	}

	if len(missing) > 0 {
		errMsg := fmt.Sprintf("Layer1_ContractValidationFailed: Missing or invalid required fields: %s", strings.Join(missing, ", "))
		return http.StatusBadRequest, nil, errMsg
	}

	// Layer 2 Domain Compliance: High-value transactions ($1000+) require KYC/AML verification
	amountCents, _ := toInt(centsVal)
	if amountCents >= 100000 {
		verified, ok := payload["compliance_verified"].(bool)
		if !ok || !verified {
			return http.StatusConflict, nil, "Layer2_DomainComplianceConflict: High-value transaction requires verified customer KYC/AML status (compliance_verified: true)"
		}
	}

	return http.StatusOK, map[string]interface{}{
		"status":         "PROCESSED",
		"order_id":       payload["order_id"],
		"amount_cents":   amountCents,
		"currency":       payload["currency"],
		"customer_email": payload["customer_email"],
		"auth_code":      "AUTH_NP_OK",
	}, ""
}

func validateInventory(payload map[string]interface{}) (int, map[string]interface{}, string) {
	var missing []string
	if val, ok := payload["sku_code"]; !ok || !strings.HasPrefix(fmt.Sprintf("%v", val), "SKU-") {
		missing = append(missing, "sku_code (must start with 'SKU-')")
	}

	qtyVal, hasQty := payload["quantity"]
	if !hasQty {
		missing = append(missing, "quantity (strict integer)")
	} else {
		switch v := qtyVal.(type) {
		case int:
			if v <= 0 {
				missing = append(missing, "quantity (must be > 0)")
			}
		case float64:
			if v != float64(int(v)) || v <= 0 {
				missing = append(missing, "quantity (must be whole integer, received float)")
			}
		default:
			missing = append(missing, "quantity (invalid type)")
		}
	}

	if val, ok := payload["priority_tier"]; !ok {
		missing = append(missing, "priority_tier (STANDARD | EXPRESS | CRITICAL)")
	} else {
		s := fmt.Sprintf("%v", val)
		if s != "STANDARD" && s != "EXPRESS" && s != "CRITICAL" {
			missing = append(missing, "priority_tier (must be enum: STANDARD, EXPRESS, CRITICAL)")
		}
	}

	if len(missing) > 0 {
		return http.StatusBadRequest, nil, "Layer1_InventoryValidationFailed: " + strings.Join(missing, ", ")
	}

	return http.StatusOK, map[string]interface{}{
		"status":      "ALLOCATED",
		"sku_code":    payload["sku_code"],
		"quantity":    payload["quantity"],
		"warehouse":   "WH-CENTRAL-01",
		"priority":    payload["priority_tier"],
		"reservation": "RES-SUCCESS",
	}, ""
}

func validateTelemetry(payload map[string]interface{}) (int, map[string]interface{}, string) {
	var missing []string
	if val, ok := payload["device_id"]; !ok || !strings.HasPrefix(fmt.Sprintf("%v", val), "DEV-") {
		missing = append(missing, "device_id (must start with 'DEV-')")
	}

	if _, ok := payload["temperature_c"].(float64); !ok {
		missing = append(missing, "temperature_c (float64)")
	}

	if _, ok := payload["healthy"].(bool); !ok {
		missing = append(missing, "healthy (boolean)")
	}

	if len(missing) > 0 {
		return http.StatusBadRequest, nil, "Layer1_TelemetryValidationFailed: " + strings.Join(missing, ", ")
	}

	return http.StatusOK, map[string]interface{}{
		"status":        "INGESTED",
		"device_id":     payload["device_id"],
		"temperature_c": payload["temperature_c"],
		"healthy":       payload["healthy"],
	}, ""
}

func validateCustomContract(contract contracts.ServiceContract, payload map[string]interface{}) (int, map[string]interface{}, string) {
	var missing []string
	for field, rule := range contract.RequiredFields {
		val, exists := payload[field]
		if !exists || val == nil || fmt.Sprintf("%v", val) == "" {
			missing = append(missing, fmt.Sprintf("%s (%s)", field, rule))
		}
	}

	if len(missing) > 0 {
		return http.StatusBadRequest, nil, fmt.Sprintf("Layer1_CustomContractFailed: Missing %s", strings.Join(missing, ", "))
	}

	return http.StatusOK, map[string]interface{}{
		"status":      "ACCEPTED",
		"contract_id": contract.ID,
		"service":     contract.Name,
		"payload":     payload,
	}, ""
}

func toInt(val interface{}) (int, bool) {
	switch v := val.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	}
	return 0, false
}
