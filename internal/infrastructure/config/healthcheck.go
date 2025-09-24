package config

import "time"

type HealthCheck struct {
	Timeout time.Duration `yaml:"timeout"`
}
