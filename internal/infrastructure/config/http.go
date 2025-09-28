package config

import "time"

type HTTP struct {
	Port         string        `envconfig:"HTTP_PORT"`
	ReadTimeout  time.Duration `envconfig:"HTTP_READ_TIMEOUT"`
	WriteTimeout time.Duration `envconfig:"HTTP_WRITE_TIMEOUT"`
	IdleTimeout  time.Duration `envconfig:"HTTP_IDLE_TIMEOUT"`
}
