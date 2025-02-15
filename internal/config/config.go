package config



type Config struct {
	ServerPort    string
	RedisAddr     string
	BaseURL       string
}

func Load() *Config {
	return &Config{
		ServerPort:    "8080",
		RedisAddr:     "localhost:6379",
		BaseURL:       "http://localhost:8080/",
	}
}

