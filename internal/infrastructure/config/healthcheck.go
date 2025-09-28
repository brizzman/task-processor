package config

import "time"

type HealthCheck struct {
	Timeout time.Duration `envconfig:"HEALTHCHECK_TIMEOUT"`
}
