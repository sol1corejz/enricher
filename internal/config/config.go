package config

import (
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	Env  string
	Port string
}

func MustLoad() *Config {
	var cfg Config
	if err := godotenv.Load(); err == nil {
		env := os.Getenv("ENV")
		port := os.Getenv("PORT")

		cfg.Env = env
		cfg.Port = port
	}

	return &cfg
}
