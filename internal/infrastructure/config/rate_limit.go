package config


type RateLimit struct {
	RPS    int           `envconfig:"RATE_LIMIT_RPS"`
}
