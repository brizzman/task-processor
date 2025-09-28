package config

type WorkerPool struct {
	MaxWorkers int `envconfig:"WORKER_POOL_MAX_WORKERS"`
}