package cache

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

	"github.com/Jotadebeese/neuroproxy/internal/contracts"
)

// RuleStorage defines the interface for persisting learned schema adaptation rules.
// Supports both in-memory LRU and Google Cloud Memorystore / Redis backends.
type RuleStorage interface {
	Get(fingerprint string) (*contracts.LearnedRule, bool)
	Put(rule *contracts.LearnedRule)
	List() []*contracts.LearnedRule
	Len() int
	Reset()
}

// lruNode represents an element in the doubly linked list.
type lruNode struct {
	key       string
	rule      *contracts.LearnedRule
	expiresAt time.Time
	prev      *lruNode
	next      *lruNode
}

// LRURuleCache is a thread-safe in-memory cache with LRU eviction and TTL expiration.
type LRURuleCache struct {
	mu       sync.RWMutex
	capacity int
	ttl      time.Duration
	items    map[string]*lruNode
	head     *lruNode
	tail     *lruNode
}

// NewLRURuleCache creates an LRU cache with maximum capacity and entry TTL.
func NewLRURuleCache(capacity int, ttl time.Duration) *LRURuleCache {
	if capacity <= 0 {
		capacity = 5000
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &LRURuleCache{
		capacity: capacity,
		ttl:      ttl,
		items:    make(map[string]*lruNode),
	}
}

func (c *LRURuleCache) Get(fingerprint string) (*contracts.LearnedRule, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, exists := c.items[fingerprint]
	if !exists {
		return nil, false
	}

	if time.Now().After(node.expiresAt) {
		c.removeNode(node)
		delete(c.items, fingerprint)
		return nil, false
	}

	c.moveToHead(node)
	node.rule.HitCount++
	node.rule.LastUsedAt = time.Now()
	node.expiresAt = time.Now().Add(c.ttl)

	return node.rule, true
}

func (c *LRURuleCache) Put(rule *contracts.LearnedRule) {
	if rule == nil || rule.Fingerprint == "" {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if node, exists := c.items[rule.Fingerprint]; exists {
		node.rule = rule
		node.expiresAt = time.Now().Add(c.ttl)
		c.moveToHead(node)
		return
	}

	for len(c.items) >= c.capacity && c.tail != nil {
		oldKey := c.tail.key
		c.removeNode(c.tail)
		delete(c.items, oldKey)
	}

	newNode := &lruNode{
		key:       rule.Fingerprint,
		rule:      rule,
		expiresAt: time.Now().Add(c.ttl),
	}
	c.items[rule.Fingerprint] = newNode
	c.addToHead(newNode)
}

func (c *LRURuleCache) List() []*contracts.LearnedRule {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	res := make([]*contracts.LearnedRule, 0, len(c.items))
	curr := c.head
	for curr != nil {
		if now.Before(curr.expiresAt) {
			res = append(res, curr.rule)
		}
		curr = curr.next
	}
	return res
}

func (c *LRURuleCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

func (c *LRURuleCache) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*lruNode)
	c.head = nil
	c.tail = nil
}

func (c *LRURuleCache) addToHead(node *lruNode) {
	node.next = c.head
	node.prev = nil
	if c.head != nil {
		c.head.prev = node
	}
	c.head = node
	if c.tail == nil {
		c.tail = node
	}
}

func (c *LRURuleCache) removeNode(node *lruNode) {
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		c.head = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	} else {
		c.tail = node.prev
	}
}

func (c *LRURuleCache) moveToHead(node *lruNode) {
	if c.head == node {
		return
	}
	c.removeNode(node)
	c.addToHead(node)
}

// ComputeSchemaFingerprint generates an order-independent MD5/SHA256 fingerprint
// representing the exact structural shape of the incoming payload.
func ComputeSchemaFingerprint(payload map[string]interface{}) string {
	paths := extractPayloadPaths(payload, "")
	sort.Strings(paths)
	joined := strings.Join(paths, "|")
	h := sha256.Sum256([]byte(joined))
	return hex.EncodeToString(h[:8])
}

// ApplyRule transforms a drifted incoming payload according to a cached LearnedRule.
func ApplyRule(raw map[string]interface{}, rule *contracts.LearnedRule) map[string]interface{} {
	if rule == nil {
		return raw
	}
	healed := make(map[string]interface{})
	for _, m := range rule.Mappings {
		rawVal, found := extractPath(raw, m.SourcePath)
		if !found {
			continue
		}
		healed[m.TargetPath] = ApplyTransform(rawVal, m.Transform)
	}
	return healed
}

func extractPayloadPaths(val interface{}, prefix string) []string {
	var paths []string
	switch m := val.(type) {
	case map[string]interface{}:
		for k, v := range m {
			currPath := k
			if prefix != "" {
				currPath = prefix + "." + k
			}
			paths = append(paths, currPath)
			paths = append(paths, extractPayloadPaths(v, currPath)...)
		}
	}
	return paths
}

func extractPath(raw map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	var curr interface{} = raw
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

// ApplyTransform performs type coercions and arithmetic transforms on values.
func ApplyTransform(val interface{}, transform string) interface{} {
	t := strings.ToLower(strings.TrimSpace(transform))
	switch {
	case strings.Contains(t, "dollars_to_cents") || strings.Contains(t, "multiply_100"):
		if f, ok := ToFloat(val); ok {
			return int(math.Round(f * 100))
		}
	case strings.Contains(t, "to_int") || strings.Contains(t, "integer"):
		if f, ok := ToFloat(val); ok {
			return int(math.Round(f))
		}
	case strings.Contains(t, "to_float") || strings.Contains(t, "number"):
		if f, ok := ToFloat(val); ok {
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

// ToFloat attempts to parse interface{} values into float64 safely.
func ToFloat(val interface{}) (float64, bool) {
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
