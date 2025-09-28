package config

import "time"

type Redis struct {
	Address      string        `envconfig:"REDIS_ADDRESS"`
	Password     string        `envconfig:"REDIS_PASSWORD"`
	DB           int           `envconfig:"REDIS_DB"`
	PoolSize     int           `envconfig:"REDIS_POOL_SIZE"`
	DialTimeout  time.Duration `envconfig:"REDIS_DIAL_TIMEOUT"`
	ReadTimeout  time.Duration `envconfig:"REDIS_READ_TIMEOUT"`
	WriteTimeout time.Duration `envconfig:"REDIS_WRITE_TIMEOUT"`
}
