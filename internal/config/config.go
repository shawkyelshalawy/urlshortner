package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          int
	Env           string
	RedisAddr     string
	BaseURL       string
	MongoURI      string
	MongoDatabase string
	RateLimit     struct {
		Enabled  bool
		Requests int
		Window   time.Duration
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

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	
	cfg := &Config{
		Port:          8080,
		Env:           env,
		RedisAddr:     "localhost:6379",
		BaseURL:       "http://localhost:8080",
		MongoURI:      "mongodb://localhost:27017",
		MongoDatabase: "urlshortener",
		RateLimit: struct {
			Enabled  bool
			Requests int
			Window   time.Duration
		}{
			Enabled:  true,
			Requests: 10,
			Window:   time.Minute,
		},
	}

	
	if cfg.Env != "development" {
		_ = godotenv.Load() 
	}

	
	if portStr := os.Getenv("PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.Port = port
		}
	}
	if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" {
		cfg.RedisAddr = redisAddr
	}
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		cfg.BaseURL = baseURL
	}
	if mongoURI := os.Getenv("MONGO_URI"); mongoURI != "" {
		cfg.MongoURI = mongoURI
	}
	if mongoDatabase := os.Getenv("MONGO_DATABASE"); mongoDatabase != "" {
		cfg.MongoDatabase = mongoDatabase
	}

	return cfg
}
