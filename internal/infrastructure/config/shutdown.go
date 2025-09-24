package config

import "time"

type Shutdown struct {
	HTTPTimeout    time.Duration `yaml:"http_timeout"`
	HardPeriod     time.Duration `yaml:"hard_period"`
	ReadinessDrain time.Duration `yaml:"readiness_drain"`
}
