package config

import "os"

type Config struct {
	Port        string
	JWTSecret   string
	DatabaseURL string
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

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/taskboard?sslmode=disable"
	}

	return Config{
		Port:        port,
		JWTSecret:   secret,
		DatabaseURL: dbURL,
	}
}

func (c Config) Address() string {
	return ":" + c.Port
}
