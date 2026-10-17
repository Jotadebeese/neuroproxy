package ai

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrRateLimitExceeded = errors.New("CIRCUIT_BREAKER_RATE_LIMITED: Token bucket depleted to protect LLM budget")
	ErrCircuitOpen       = errors.New("CIRCUIT_BREAKER_OPEN: Gateway tripped due to consecutive upstream failures")
)

type CircuitState string

const (
	StateClosed   CircuitState = "CLOSED"
	StateHalfOpen CircuitState = "HALF_OPEN"
	StateOpen     CircuitState = "OPEN"
)

// CircuitBreaker guards the Gemini LLM synthesizer against Denial-of-Wallet attacks
// and runaway cost exhaustion using Token Bucket rate limiting and an error circuit breaker.
type CircuitBreaker struct {
	mu           sync.Mutex
	rateLimit    float64       // max tokens per second
	burstLimit   float64       // token capacity
	tokens       float64       // current available tokens
	lastRefill   time.Time
	state        CircuitState
	failureCount int
	threshold    int           // consecutive failures to trip
	cooldown     time.Duration // recovery wait before HALF_OPEN
	lastTrip     time.Time
}

// NewCircuitBreaker creates a circuit breaker tailored for Cloud Run and Gemini 2.5 Flash.
func NewCircuitBreaker(rateLimit float64, burstLimit int, failureThreshold int, cooldown time.Duration) *CircuitBreaker {
	if rateLimit <= 0 {
		rateLimit = 0.5 // 30 req/min
	}
	if burstLimit <= 0 {
		burstLimit = 10
	}
	if failureThreshold <= 0 {
		failureThreshold = 5
	}
	if cooldown <= 0 {
		cooldown = 15 * time.Second
	}

	return &CircuitBreaker{
		rateLimit:  rateLimit,
		burstLimit: float64(burstLimit),
		tokens:     float64(burstLimit),
		lastRefill: time.Now(),
		state:      StateClosed,
		threshold:  failureThreshold,
		cooldown:   cooldown,
	}
}

// Allow checks if the next request is permitted through the circuit breaker.
func (cb *CircuitBreaker) Allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	// Check if cooldown elapsed in StateOpen -> transition to StateHalfOpen
	if cb.state == StateOpen {
		if now.Sub(cb.lastTrip) >= cb.cooldown {
			cb.state = StateHalfOpen
			cb.failureCount = 0
			// Grant immediate probe token for half-open recovery test
			cb.tokens = 1.0
			cb.lastRefill = now
		} else {
			return ErrCircuitOpen
		}
	}

	// Refill token bucket
	elapsed := now.Sub(cb.lastRefill).Seconds()
	cb.tokens += elapsed * cb.rateLimit
	if cb.tokens > cb.burstLimit {
		cb.tokens = cb.burstLimit
	}
	cb.lastRefill = now

	// Consume token
	if cb.tokens < 1.0 {
		return ErrRateLimitExceeded
	}

	cb.tokens -= 1.0
	return nil
}

// RecordSuccess marks a successful execution, resetting circuit state to CLOSED.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount = 0
	if cb.state == StateHalfOpen {
		cb.state = StateClosed
	}
}

// RecordFailure notes an upstream/LLM error, tripping the circuit if threshold is reached.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	if cb.state == StateHalfOpen || cb.failureCount >= cb.threshold {
		cb.state = StateOpen
		cb.lastTrip = time.Now()
	}
}

// State returns the current status of the circuit breaker.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
