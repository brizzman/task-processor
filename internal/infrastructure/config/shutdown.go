package config

import "time"

type Shutdown struct {
	HTTPTimeout    time.Duration `envconfig:"SHUTDOWN_HTTP_TIMEOUT"`
	HardPeriod     time.Duration `envconfig:"SHUTDOWN_HARD_PERIOD"`
	ReadinessDrain time.Duration `envconfig:"SHUTDOWN_READINESS_DRAIN"`
}
