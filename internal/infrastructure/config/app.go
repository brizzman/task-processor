package config

type App struct {
	Name    string `envconfig:"APP_NAME"`
	Version string `envconfig:"APP_VERSION"`
	Env     string `envconfig:"APP_ENV"`
}