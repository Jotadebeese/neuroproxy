package guardrail

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/Jotadebeese/neuroproxy/internal/contracts"
)

// VerifyValueConservation verifies that numeric and critical business values in the healed
// payload match or mathematically derive from values in the raw input.
// Prevents LLM hallucinations (e.g. fabricating missing amounts).
func VerifyValueConservation(raw map[string]interface{}, healed map[string]interface{}, contract contracts.ServiceContract) contracts.GuardrailReport {
	report := contracts.GuardrailReport{
		Passed:          true,
		SourceValues:    make(map[string]interface{}),
		ExtractedValues: make(map[string]interface{}),
	}

	flatRaw := flattenPayload(raw, "")
	flatHealed := flattenPayload(healed, "")

	// Check amount / price conservation
	targetCentsKey := findKeyContaining(flatHealed, "cents")
	if targetCentsKey != "" {
		healedCentsVal := flatHealed[targetCentsKey]
		healedCents, isInt := toInt(healedCentsVal)
		report.ExtractedValues[targetCentsKey] = healedCentsVal

		sourceAmtKey, sourceAmtVal := findNumericSource(flatRaw, []string{"dollar", "usd", "amount", "monto", "precio", "price", "total"})
		if sourceAmtKey != "" {
			report.SourceValues[sourceAmtKey] = sourceAmtVal
			sourceFloat, _ := toFloat(sourceAmtVal)

			if isInt {
				var expectedCents int
				if strings.Contains(strings.ToLower(sourceAmtKey), "dollar") ||
					strings.Contains(strings.ToLower(sourceAmtKey), "usd") ||
					strings.Contains(strings.ToLower(sourceAmtKey), "monto") ||
					strings.Contains(fmt.Sprintf("%v", sourceAmtVal), ".") {
					expectedCents = int(math.Round(sourceFloat * 100))
				} else {
					expectedCents = int(math.Round(sourceFloat))
				}

				diff := math.Abs(float64(healedCents - expectedCents))
				if diff > 1.0 {
					report.Passed = false
					report.HallucinationVeto = true
					report.ViolationCount++
					report.Reason = fmt.Sprintf("VALUE_INTEGRITY_VIOLATION: Target %s (%d) does not conserve source %s (%v -> expected %d cents)",
						targetCentsKey, healedCents, sourceAmtKey, sourceAmtVal, expectedCents)
					return report
				}
				report.ConservedCount++
			}
		} else {
			// No source amount existed in raw input, yet healed payload has a non-zero price -> Hallucination!
			if healedCents > 0 {
				report.Passed = false
				report.HallucinationVeto = true
				report.ViolationCount++
				report.Reason = fmt.Sprintf("HALLUCINATION_VETO: Required field %s (%d) was fabricated without any source numeric path in incoming payload",
					targetCentsKey, healedCents)
				return report
			}
		}
	}

	// Check quantity conservation
	targetQtyKey := findKeyContaining(flatHealed, "quantity", "qty")
	if targetQtyKey != "" {
		healedQtyVal := flatHealed[targetQtyKey]
		healedQty, _ := toInt(healedQtyVal)
		report.ExtractedValues[targetQtyKey] = healedQtyVal

		sourceQtyKey, sourceQtyVal := findNumericSource(flatRaw, []string{"unit", "qty", "quantity", "count", "amount"})
		if sourceQtyKey != "" {
			report.SourceValues[sourceQtyKey] = sourceQtyVal
			sourceQty, _ := toInt(sourceQtyVal)
			if healedQty != sourceQty {
				report.Passed = false
				report.HallucinationVeto = true
				report.ViolationCount++
				report.Reason = fmt.Sprintf("QUANTITY_CONSERVATION_VIOLATION: Target %s (%d) != source %s (%d)",
					targetQtyKey, healedQty, sourceQtyKey, sourceQty)
				return report
			}
			report.ConservedCount++
		}
	}

	return report
}

func flattenPayload(val interface{}, prefix string) map[string]interface{} {
	flat := make(map[string]interface{})
	switch m := val.(type) {
	case map[string]interface{}:
		for k, v := range m {
			currPath := k
			if prefix != "" {
				currPath = prefix + "." + k
			}
			flat[currPath] = v
			childFlat := flattenPayload(v, currPath)
			for ck, cv := range childFlat {
				flat[ck] = cv
			}
		}
	}
	return flat
}

func findKeyContaining(flat map[string]interface{}, substrings ...string) string {
	for k := range flat {
		kLower := strings.ToLower(k)
		for _, sub := range substrings {
			if strings.Contains(kLower, sub) {
				return k
			}
		}
	}
	return ""
}

func findNumericSource(flat map[string]interface{}, keywords []string) (string, interface{}) {
	for _, kw := range keywords {
		for k, v := range flat {
			if strings.Contains(strings.ToLower(k), kw) {
				if _, ok := toFloat(v); ok {
					return k, v
				}
			}
		}
	}
	return "", nil
}

func toFloat(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case string:
		clean := strings.Trim(strings.ReplaceAll(v, "$", ""), " ")
		if f, err := strconv.ParseFloat(clean, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

func toInt(val interface{}) (int, bool) {
	switch v := val.(type) {
	case int:
		return v, true
	case float64:
		return int(math.Round(v)), true
	case string:
		clean := strings.Trim(strings.ReplaceAll(v, "$", ""), " ")
		if i, err := strconv.Atoi(clean); err == nil {
			return i, true
		}
		if f, err := strconv.ParseFloat(clean, 64); err == nil {
			return int(math.Round(f)), true
		}
	}
	return 0, false
}
