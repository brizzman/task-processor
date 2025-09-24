package config

import "time"

type CircuitBreaker struct {
	Enabled 			bool 		  `yaml:"enabled"`
	MaxRequests         uint32        `yaml:"max_requests"`
	Interval            time.Duration `yaml:"interval"`
	Timeout             time.Duration `yaml:"timeout"`
	ConsecutiveFailures uint32        `yaml:"consecutive_failures"`
}
