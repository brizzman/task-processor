package config

import (
	"fmt"
	"sync"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	App            App
	HTTP           HTTP
	Log            Log
	PG             PG
	Shutdown       Shutdown
	WorkerPool     WorkerPool
	RateLimit      RateLimit
	Redis          Redis
	HealthCheck    HealthCheck
	CircuitBreaker CircuitBreaker
}

var (
	instance *Config
	once     sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{}
		if err := envconfig.Process("", instance); err != nil {
			panic(fmt.Sprintf("Failed to load env config: %s", err))
		}
	})
	return instance
}