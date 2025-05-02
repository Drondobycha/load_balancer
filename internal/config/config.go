package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Port                 string          `yaml:"port"`
	Strategy             string          `yaml:"strategy"`
	Backends             []string        `yaml:"backends"`
	HealsthCheckInterval time.Duration   `yaml:"health_check_interval"`
	RateLimit            RateLimitConfig `yaml:"rate_limit"`
}

type RateLimitConfig struct {
	DefaultRate     int `yaml:"default_rate"`
	DefaultCapacity int `yaml:"default_capacity"`
}

func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config path is empty")
	}

	return MustLoadByPath(path)
}

func MustLoadByPath(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}
	return &cfg
}

func fetchConfigPath() string {
	var res string
	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
