package config

import "time"

type PG struct {
	URL            string        `envconfig:"POSTGRES_URL"`
	MaxPoolSize    int32         `envconfig:"POSTGRES_MAX_POOL_SIZE"`
	MinPoolSize    int32         `envconfig:"POSTGRES_MIN_POOL_SIZE"`
	MaxConnLife    time.Duration `envconfig:"POSTGRES_MAX_CONN_LIFETIME"`
	MaxConnIdle    time.Duration `envconfig:"POSTGRES_MAX_CONN_IDLE_TIME"`
	ConnectTimeout time.Duration `envconfig:"POSTGRES_CONNECT_TIMEOUT"`
}
