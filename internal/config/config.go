package config

import "os"

type Config struct {
	Port      string
	JWTSecret string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-me"
	}

	return Config{
		Port:      port,
		JWTSecret: secret,
	}
}

func (c Config) Address() string {
	return ":" + c.Port
}
