package config

import "time"

type CircuitBreaker struct {
	Enabled             bool          `envconfig:"CIRCUIT_BREAKER_ENABLED"`
	MaxRequests         uint32        `envconfig:"CIRCUIT_BREAKER_MAX_REQUESTS"`
	Interval            time.Duration `envconfig:"CIRCUIT_BREAKER_INTERVAL"`
	Timeout             time.Duration `envconfig:"CIRCUIT_BREAKER_TIMEOUT"`
	ConsecutiveFailures uint32        `envconfig:"CIRCUIT_BREAKER_CONSECUTIVE_FAILURES"`
}
