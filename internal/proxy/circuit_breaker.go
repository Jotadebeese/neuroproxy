package proxy

import (
	"sync"
	"time"
)

// CircuitState represents the state of the LLM circuit breaker.
type CircuitState string

const (
	StateClosed   CircuitState = "CLOSED"
	StateHalfOpen CircuitState = "HALF_OPEN"
	StateOpen     CircuitState = "OPEN"
)

// CircuitBreaker protects the gateway from Denial-of-Wallet attacks and LLM cascading failures.
type CircuitBreaker struct {
	mu           sync.Mutex
	state        CircuitState
	failures     int
	threshold    int
	resetTimeout time.Duration
	lastFailure  time.Time

	// Token-bucket rate limiter
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

// NewCircuitBreaker creates a circuit breaker with a token bucket limiter.
func NewCircuitBreaker(threshold int, resetTimeout time.Duration, ratePerSec float64, burst int) *CircuitBreaker {
	now := time.Now()
	return &CircuitBreaker{
		state:        StateClosed,
		threshold:    threshold,
		resetTimeout: resetTimeout,
		tokens:       float64(burst),
		maxTokens:    float64(burst),
		refillRate:   ratePerSec,
		lastRefill:   now,
	}
}

// Allow checks if a new Gemini API call is permitted under current rate & circuit limits.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	// 1. Check Circuit Breaker state
	if cb.state == StateOpen {
		if now.Sub(cb.lastFailure) > cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.tokens = 1.0
			cb.lastRefill = now
			return true // Allow the single probe request in HalfOpen!
		} else {
			return false // Circuit open: fail fast to protect budget
		}
	}

	// 2. Refill token bucket
	elapsed := now.Sub(cb.lastRefill).Seconds()
	cb.tokens = cb.tokens + (elapsed * cb.refillRate)
	if cb.tokens > cb.maxTokens {
		cb.tokens = cb.maxTokens
	}
	cb.lastRefill = now

	// 3. Consume token
	if cb.tokens >= 1.0 {
		cb.tokens -= 1.0
		return true
	}

	return false // Rate limit exceeded
}

// RecordSuccess records a successful Gemini API call.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	cb.state = StateClosed
}

// RecordFailure records a failed Gemini API call and potentially trips the circuit breaker.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailure = time.Now()
	if cb.failures >= cb.threshold {
		cb.state = StateOpen
	}
}

// GetState returns the current circuit breaker state.
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
