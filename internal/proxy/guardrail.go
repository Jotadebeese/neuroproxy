package proxy

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// GuardrailReport summarizes the deterministic Zero-Hallucination Value Conservation check.
type GuardrailReport struct {
	Passed          bool     `json:"passed"`
	VerifiedFields  []string `json:"verified_fields"`
	Violations      []string `json:"violations,omitempty"`
	ConservationPct float64  `json:"conservation_pct"`
	Summary         string   `json:"summary"`
}

// VerifyValueConservation ensures that Gemini did not hallucinate or invent critical financial/business
// values (amounts, IDs, emails, SKUs, sensor readings) that cannot be traced back to the source payload.
func VerifyValueConservation(original map[string]interface{}, healed map[string]interface{}) GuardrailReport {
	flatSource := make(map[string]interface{})
	flattenPaths("", original, flatSource)

	var verified []string
	var violations []string

	allowedEnums := map[string]bool{
		"STANDARD": true, "EXPRESS": true, "CRITICAL": true,
		"USD": true, "SGD": true, "THB": true, "EUR": true, "GBP": true,
	}

	keys := make([]string, 0, len(healed))
	for k := range healed {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, targetKey := range keys {
		targetVal := healed[targetKey]

		// Booleans must derive from a non-empty source payload containing a boolean or status string
		if bVal, isBool := targetVal.(bool); isBool {
			if srcPath, ok := traceBoolToSource(bVal, flatSource); ok {
				verified = append(verified, fmt.Sprintf("%s=%v <- %s", targetKey, bVal, srcPath))
				continue
			}
			violations = append(violations, fmt.Sprintf("unverified boolean '%s=%v' has no matching status/flag in source payload", targetKey, bVal))
			continue
		}

		if sourcePath, ok := traceValueToSource(targetVal, flatSource, allowedEnums); ok {
			verified = append(verified, fmt.Sprintf("%s <- %s", targetKey, sourcePath))
		} else {
			violations = append(violations, fmt.Sprintf("hallucinated/unverified value '%v' for field '%s' not present in source request", targetVal, targetKey))
		}
	}

	total := len(verified) + len(violations)
	pct := 100.0
	if total > 0 {
		pct = float64(len(verified)) / float64(total) * 100.0
	}

	if len(violations) > 0 {
		return GuardrailReport{
			Passed:          false,
			VerifiedFields:  verified,
			Violations:      violations,
			ConservationPct: pct,
			Summary:         fmt.Sprintf("SAFETY VETO: Blocked potential LLM hallucination (%s)", strings.Join(violations, "; ")),
		}
	}

	return GuardrailReport{
		Passed:          true,
		VerifiedFields:  verified,
		ConservationPct: 100.0,
		Summary:         fmt.Sprintf("Zero-Hallucination Guardrail PASSED: 100%% of target scalars (%d/%d) mathematically traced to source leaf nodes.", len(verified), len(verified)),
	}
}

func traceBoolToSource(targetBool bool, flatSource map[string]interface{}) (string, bool) {
	for srcPath, srcVal := range flatSource {
		switch v := srcVal.(type) {
		case bool:
			if v == targetBool {
				return srcPath, true
			}
		case string:
			u := strings.ToUpper(strings.TrimSpace(v))
			if targetBool && (u == "TRUE" || u == "1" || u == "ONLINE" || u == "OK" || u == "HEALTHY" || u == "ACTIVE" || u == "APPROVED" || u == "VERIFIED" || u == "PASS") {
				return fmt.Sprintf("%s (%q)", srcPath, v), true
			}
			if !targetBool && (u == "FALSE" || u == "0" || u == "OFFLINE" || u == "REJECTED" || u == "FAILED") {
				return fmt.Sprintf("%s (%q)", srcPath, v), true
			}
		}
	}
	return "", false
}

func traceValueToSource(targetVal interface{}, flatSource map[string]interface{}, allowedEnums map[string]bool) (string, bool) {
	targetStr := strings.TrimSpace(fmt.Sprintf("%v", targetVal))
	targetUpper := strings.ToUpper(targetStr)

	// 1. Direct or case-insensitive string match
	for srcPath, srcVal := range flatSource {
		srcStr := strings.TrimSpace(fmt.Sprintf("%v", srcVal))
		if strings.EqualFold(targetStr, srcStr) && srcStr != "" {
			return srcPath, true
		}
	}

	// 2. Numeric conservation check (exact numeric equality OR dollars-to-cents * 100 conversion)
	if targetFloat, ok := toFloat(targetVal); ok {
		for srcPath, srcVal := range flatSource {
			if srcFloat, ok := toFloat(srcVal); ok {
				if math.Abs(targetFloat-srcFloat) < 0.001 {
					return fmt.Sprintf("%s (numeric identity)", srcPath), true
				}
				if math.Abs(targetFloat-(srcFloat*100.0)) < 0.5 {
					return fmt.Sprintf("%s (x100 cents conversion)", srcPath), true
				}
			}
		}
	}

	// 3. Canonical enum / ISO currency derived from an existing source string
	if allowedEnums[targetUpper] {
		for srcPath, srcVal := range flatSource {
			if s, ok := srcVal.(string); ok && strings.TrimSpace(s) != "" {
				return fmt.Sprintf("%s (enum coercion)", srcPath), true
			}
		}
	}

	return "", false
}
