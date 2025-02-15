package config

import (
	"time"
)

type Config struct {
	Port          int
	Env           string
	RedisAddr     string
	BaseURL       string
	MongoURI      string
	MongoDatabase string
	RateLimit     struct {
		Enabled   bool
		Requests  int
		Window    time.Duration
	}
	Limiter struct {
		Enabled bool
		RPS     float64
		Burst   int
	}
	CORS struct {
		TrustedOrigins []string
	}
}

func Load() *Config {
	cfg := &Config{
		Port:          8080,
		Env:          "development",
		RedisAddr:    "localhost:6379",
		BaseURL:      "http://localhost:8080",
		MongoURI:     "mongodb://localhost:27017",
		MongoDatabase: "urlshortener",
		RateLimit: struct {
			Enabled   bool
			Requests  int
			Window    time.Duration
		}{
			Enabled:  true,
			Requests: 10,
			Window:   time.Minute,
		},
	}
	return cfg
}