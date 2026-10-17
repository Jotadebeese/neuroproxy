package proxy

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RuleCache is a thread-safe in-memory store mapping structural schema fingerprints to learned transformation rules.
type RuleCache struct {
	mu    sync.RWMutex
	rules map[string]*LearnedRule
}

func NewRuleCache() *RuleCache {
	return &RuleCache{
		rules: make(map[string]*LearnedRule),
	}
}

// ComputeSchemaFingerprint generates a deterministic SHA-256 hash of a payload's structural shape
// (serviceID + sorted nested key paths and value types) so identical schema drift hits the cache in <1ms.
func ComputeSchemaFingerprint(serviceID string, payload map[string]interface{}) string {
	var paths []string
	collectSchemaPaths("", payload, &paths)
	sort.Strings(paths)
	raw := serviceID + "::" + strings.Join(paths, "|")
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:8]) // 16-char hex fingerprint
}

func collectSchemaPaths(prefix string, obj map[string]interface{}, out *[]string) {
	for k, v := range obj {
		fullKey := k
		if prefix != "" {
			fullKey = prefix + "." + k
		}
		switch sub := v.(type) {
		case map[string]interface{}:
			collectSchemaPaths(fullKey, sub, out)
		default:
			*out = append(*out, fmt.Sprintf("%s:%T", fullKey, v))
		}
	}
}

// Get retrieves a learned rule by fingerprint and increments its hit counter safely.
func (c *RuleCache) Get(fingerprint string) (*LearnedRule, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	rule, ok := c.rules[fingerprint]
	if !ok {
		return nil, false
	}
	rule.HitCount++
	rule.LastUsedAt = time.Now()
	// Return a copy to avoid data races
	copyRule := *rule
	return &copyRule, true
}

// Set stores a learned transformation rule in memory.
func (c *RuleCache) Set(rule *LearnedRule) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rules[rule.Fingerprint] = rule
}

// List returns all active rules sorted by most recently created.
func (c *RuleCache) List() []LearnedRule {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]LearnedRule, 0, len(c.rules))
	for _, r := range c.rules {
		out = append(out, *r)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

// Count returns the number of cached rules.
func (c *RuleCache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.rules)
}

// Clear purges all cached rules.
func (c *RuleCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rules = make(map[string]*LearnedRule)
}

// ApplyLearnedRule executes the deterministic field mappings on a new incoming payload in <1ms without calling Gemini.
func ApplyLearnedRule(payload map[string]interface{}, rule *LearnedRule) map[string]interface{} {
	healed := make(map[string]interface{})
	for _, m := range rule.Mappings {
		rawVal, found := extractNestedValue(payload, m.SourcePath)
		if !found {
			continue
		}
		healed[m.TargetPath] = applyTransform(rawVal, m.Transform)
	}
	return healed
}

func extractNestedValue(obj map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	var curr interface{} = obj
	for _, part := range parts {
		m, ok := curr.(map[string]interface{})
		if !ok {
			return nil, false
		}
		val, exists := m[part]
		if !exists {
			return nil, false
		}
		curr = val
	}
	return curr, true
}

func applyTransform(val interface{}, transform string) interface{} {
	t := strings.ToLower(strings.TrimSpace(transform))
	switch {
	case strings.Contains(t, "dollars_to_cents") || strings.Contains(t, "multiply_100"):
		if f, ok := toFloat(val); ok {
			return int(math.Round(f * 100))
		}
	case strings.Contains(t, "to_int") || strings.Contains(t, "integer"):
		if f, ok := toFloat(val); ok {
			return int(math.Round(f))
		}
	case strings.Contains(t, "to_float") || strings.Contains(t, "number"):
		if f, ok := toFloat(val); ok {
			return f
		}
	case strings.Contains(t, "uppercase") || strings.Contains(t, "upper"):
		return strings.ToUpper(fmt.Sprintf("%v", val))
	case strings.Contains(t, "enum_priority"):
		s := strings.ToUpper(fmt.Sprintf("%v", val))
		if strings.Contains(s, "CRITICAL") || strings.Contains(s, "EMERGENCY") {
			return "CRITICAL"
		}
		if strings.Contains(s, "URGENT") || strings.Contains(s, "RUSH") || strings.Contains(s, "EXPRESS") || strings.Contains(s, "FAST") {
			return "EXPRESS"
		}
		return "STANDARD"
	case strings.Contains(t, "to_bool") || strings.Contains(t, "boolean"):
		switch v := val.(type) {
		case bool:
			return v
		case string:
			u := strings.ToUpper(strings.TrimSpace(v))
			return u == "TRUE" || u == "1" || u == "ONLINE" || u == "OK" || u == "HEALTHY" || u == "ACTIVE" || u == "APPROVED" || u == "VERIFIED" || u == "PASS" || u == "YES"
		case float64:
			return v > 0
		}
	}
	return val
}

func toFloat(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case string:
		cleaned := strings.TrimSpace(strings.ReplaceAll(v, "$", ""))
		f, err := strconv.ParseFloat(cleaned, 64)
		return f, err == nil
	default:
		return 0, false
	}
}
