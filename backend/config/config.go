package config

import (
	"log"
	"os"
)

type Config struct {
	DatabaseURL      string
	JWTSecret        string
	RateLimitEnabled bool
	RedisURL         string
	TrustedProxies   string
}

func Load() Config {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	rateLimitEnabled := true
	if val := os.Getenv("RATE_LIMIT_ENABLED"); val == "false" {
		rateLimitEnabled = false
	}

	trustedProxies := os.Getenv("TRUSTED_PROXIES")
	if trustedProxies == "" {
		trustedProxies = "127.0.0.1/32,::1/128,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16"
	}

	return Config{
		DatabaseURL:      dbURL,
		JWTSecret:        jwtSecret,
		RateLimitEnabled: rateLimitEnabled,
		RedisURL:         os.Getenv("REDIS_URL"),
		TrustedProxies:   trustedProxies,
	}
}
