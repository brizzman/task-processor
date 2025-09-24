package config

import "time"

type RateLimit struct {
	RPS    int           `yaml:"rps"`
	Burst  int           `yaml:"burst"`
	Period time.Duration `yaml:"period"`
}
