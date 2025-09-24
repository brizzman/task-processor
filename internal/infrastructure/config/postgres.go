package config

import "time"

type PG struct {
	URL             string        `yaml:"url"`
	MigrationsDir   string 	  	  `yaml:"migrationsDir"`
	MaxPoolSize     int           `yaml:"max_pool_size"`
	MinPoolSize     int           `yaml:"min_pool_size"`
	MaxConnLifetime time.Duration `yaml:"max_conn_lifetime"`
	MaxConnIdleTime time.Duration `yaml:"max_conn_idle_time"`
	ConnectTimeout  time.Duration `yaml:"connect_timeout"`
}
