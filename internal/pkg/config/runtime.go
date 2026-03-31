package config

var current *Config

func SetCurrent(cfg *Config) {
	current = cfg
}

func Current() *Config {
	return current
}
