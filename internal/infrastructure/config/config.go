package config

import (
	"sync"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App            App            `yaml:"app"`
	HTTP           HTTP           `yaml:"http"`
	Log            Log            `yaml:"log"`
	PG             PG             `yaml:"postgres"`
	Shutdown       Shutdown       `yaml:"shutdown"`
	WorkerPool     WorkerPool     `yaml:"worker_pool"`
	RateLimit      RateLimit      `yaml:"rate_limit"`
	Redis          Redis          `yaml:"redis"`
	HealthCheck    HealthCheck    `yaml:"healthcheck"`
	CircuitBreaker CircuitBreaker `yaml:"circuit_breaker"`
}

var (
	instance *Config
 	once sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{}
		if err := cleanenv.ReadConfig("config-local.yaml", instance); err != nil {
			panic(err)
		}
	})
	return instance
}