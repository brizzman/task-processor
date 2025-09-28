package config

type Log struct {
	Level string `envconfig:"LOG_LEVEL"`
}
