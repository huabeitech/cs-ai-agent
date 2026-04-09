package config

var current *Config

func SetCurrent(cfg *Config) {
	current = cfg
}

func Current() Config {
	if current == nil {
		panic("config not initialized")
	}
	return *current
}
