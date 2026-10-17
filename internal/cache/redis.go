package cache

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/Jotadebeese/neuroproxy/internal/contracts"
)

// RedisRuleStorage connects to Google Cloud Memorystore (Redis) using raw TCP RESP protocol,
// ensuring zero external dependencies while enabling multi-instance cache sharing.
type RedisRuleStorage struct {
	addr     string
	fallback *LRURuleCache
	prefix   string
}

// NewRuleStorage returns a Redis-backed storage if redisURL is provided and reachable,
// otherwise returns a high-performance in-memory LRURuleCache.
func NewRuleStorage(redisURL string) RuleStorage {
	localLRU := NewLRURuleCache(5000, 24*time.Hour)
	if strings.TrimSpace(redisURL) == "" {
		return localLRU
	}

	addr := strings.TrimPrefix(redisURL, "redis://")
	if !strings.Contains(addr, ":") {
		addr += ":6379"
	}

	// Test connection with a short timeout
	conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
	if err != nil {
		fmt.Printf("⚠️ Redis at %s unreachable (%v), falling back to in-memory LRU cache\n", addr, err)
		return localLRU
	}
	_ = conn.Close()

	fmt.Printf("✅ Google Cloud Memorystore (Redis) connected at %s\n", addr)
	return &RedisRuleStorage{
		addr:     addr,
		fallback: localLRU,
		prefix:   "neuroproxy:rule:",
	}
}

func (r *RedisRuleStorage) Get(fingerprint string) (*contracts.LearnedRule, bool) {
	// Check fallback first for ultra-fast local memory lookup
	if rule, ok := r.fallback.Get(fingerprint); ok {
		return rule, true
	}

	// Fetch from Redis
	conn, err := net.DialTimeout("tcp", r.addr, 500*time.Millisecond)
	if err != nil {
		return nil, false
	}
	defer conn.Close()

	key := r.prefix + fingerprint
	cmd := fmt.Sprintf("*2\r\n$3\r\nGET\r\n$%d\r\n%s\r\n", len(key), key)
	if _, err := conn.Write([]byte(cmd)); err != nil {
		return nil, false
	}

	// Parse simple RESP response
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil || n <= 0 {
		return nil, false
	}
	respStr := string(buf[:n])
	if strings.HasPrefix(respStr, "$-1") {
		return nil, false // Nil
	}

	// Extract JSON payload after the length line
	parts := strings.SplitN(respStr, "\r\n", 3)
	if len(parts) >= 2 && len(parts[1]) > 0 && !strings.HasPrefix(parts[1], "$") {
		var rule contracts.LearnedRule
		if err := json.Unmarshal([]byte(parts[1]), &rule); err == nil {
			r.fallback.Put(&rule) // warm local L1 cache
			return &rule, true
		}
	}

	return nil, false
}

func (r *RedisRuleStorage) Put(rule *contracts.LearnedRule) {
	if rule == nil {
		return
	}
	// Always warm local L1 cache
	r.fallback.Put(rule)

	conn, err := net.DialTimeout("tcp", r.addr, 500*time.Millisecond)
	if err != nil {
		return
	}
	defer conn.Close()

	data, err := json.Marshal(rule)
	if err != nil {
		return
	}

	key := r.prefix + rule.Fingerprint
	val := string(data)
	// SET key val EX 86400 (24 hours TTL)
	cmd := fmt.Sprintf("*5\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n$2\r\nEX\r\n$5\r\n86400\r\n",
		len(key), key, len(val), val)
	_, _ = conn.Write([]byte(cmd))
}

func (r *RedisRuleStorage) List() []*contracts.LearnedRule {
	return r.fallback.List()
}

func (r *RedisRuleStorage) Len() int {
	return r.fallback.Len()
}

func (r *RedisRuleStorage) Reset() {
	r.fallback.Reset()
}
