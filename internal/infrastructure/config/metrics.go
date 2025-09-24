package config

import "time"

type Metrics struct {
	CollectionInterval time.Duration `yaml:"metrics_collection_interval"`
	HTTPPort           string        `yaml:"metrics_http_port"`
}